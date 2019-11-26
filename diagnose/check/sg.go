package check

import (
	"fmt"
	rawAws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/prototypes/diagnose/aws"
	"github.com/gruntwork-io/prototypes/diagnose/options"
	"github.com/gruntwork-io/prototypes/diagnose/output"
	"net"
)

func SecurityGroupAllowsInboundFromELB(elb *elbv2.LoadBalancer, targets []*elbv2.TargetHealthDescription, opts *options.Options) error {
	enisForElb, err := aws.FindEnisForElb(elb)
	if err != nil {
		return err
	}

	var instanceIds []string
	for _, target := range targets {
		instanceIds = append(instanceIds, rawAws.StringValue(target.Target.Id))
	}

	sgsForInstances, err := aws.FindSecurityGroupIdsForInstances(instanceIds)
	if err != nil {
		return err
	}

	port, err := aws.GetPortForTargets(targets)
	if err != nil {
		return err
	}

	for _, sg := range sgsForInstances {
		allowed, err := sgAllowsInBoundFromELB(port, sg, elb, enisForElb)
		if err != nil {
			return err
		}
		if allowed {
			opts.Logger.Infof("Security Group '%s' on the EC2 Instances allows inbound requests from the ELB!", rawAws.StringValue(sg.GroupId))
			return nil
		}
	}

	output.ShowDiagnosis("None of the Security Groups attached to the EC2 Instances seem to allow inbound requests from the ELB. You should update those Security Groups so the ELB can send requests to the apps running on those Instances!", opts)
	return errors.WithStackTrace(InboundNotAllowed{InstanceSecurityGroups: sgsForInstances, Elb: elb})
}

func sgAllowsInBoundFromELB(port int, sg *ec2.SecurityGroup, elb *elbv2.LoadBalancer, enisForElb []*ec2.NetworkInterface) (bool, error) {

	for _, ipPermission := range sg.IpPermissions {
		// TODO: this needs to check protocol too
		if portInRange(port, ipPermission) {
			if elbSecurityGroupIsWhiteListed(elb, ipPermission) {
				return true, nil
			}

			ipWhitelisted, err := elbIpInRange(enisForElb, ipPermission)
			if err != nil {
				return false, nil
			}
			if ipWhitelisted {
				return true, nil
			}
		}
	}

	return false, nil
}

func portInRange(port int, ipPermission *ec2.IpPermission) bool {
	return rawAws.Int64Value(ipPermission.FromPort) <= int64(port) && rawAws.Int64Value(ipPermission.ToPort) >= int64(port)
}

func elbIpInRange(enisForElb []*ec2.NetworkInterface, ipPermission *ec2.IpPermission) (bool, error) {
	for _, eni := range enisForElb {
		elbPrivateIp := net.ParseIP(rawAws.StringValue(eni.PrivateIpAddress))

		for _, ipRange := range ipPermission.IpRanges {
			allowed, err := ipRangeContainsIp(rawAws.StringValue(ipRange.CidrIp), elbPrivateIp)
			if err != nil {
				return false, err
			}
			if allowed {
				return true, nil
			}
		}

		for _, ipRange := range ipPermission.Ipv6Ranges {
			allowed, err := ipRangeContainsIp(rawAws.StringValue(ipRange.CidrIpv6), elbPrivateIp)
			if err != nil {
				return false, err
			}
			if allowed {
				return true, nil
			}
		}
	}

	return false, nil
}

func ipRangeContainsIp(ipRange string, ip net.IP) (bool, error) {
	_, allowedCidr, err := net.ParseCIDR(ipRange)
	if err != nil {
		return false, errors.WithStackTrace(err)
	}

	if allowedCidr.Contains(ip) {
		return true, nil
	}

	return false, nil
}

func elbSecurityGroupIsWhiteListed(elb *elbv2.LoadBalancer, ipPermission *ec2.IpPermission) bool {
	for _, userIdPair := range ipPermission.UserIdGroupPairs {
		for _, sg := range elb.SecurityGroups {
			if rawAws.StringValue(sg) == rawAws.StringValue(userIdPair.GroupId) {
				return true
			}
		}
	}

	return false
}

type InboundNotAllowed struct {
	InstanceSecurityGroups []*ec2.SecurityGroup
	Elb                    *elbv2.LoadBalancer
}

func (err InboundNotAllowed) Error() string {
	var sgIds []string
	for _, sg := range err.InstanceSecurityGroups {
		sgIds = append(sgIds, rawAws.StringValue(sg.GroupId))
	}
	return fmt.Sprintf("None of the Security Groups attached to the EC2 Instances ('%v') seem to allow inbound requests from the ELB '%s'", sgIds, rawAws.StringValue(err.Elb.LoadBalancerName))
}
