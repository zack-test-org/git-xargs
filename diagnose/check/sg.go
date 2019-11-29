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

func InstanceSecurityGroupsAllowInboundFromELB(elb *elbv2.LoadBalancer, targets []*elbv2.TargetHealthDescription, opts *options.Options) error {
	opts.Logger.Infof("Looking up the ELB's private IP addresses...")
	enisForElb, err := aws.FindEnisForElb(elb)
	if err != nil {
		return err
	}

	var instanceIds []string
	for _, target := range targets {
		instanceIds = append(instanceIds, rawAws.StringValue(target.Target.Id))
	}

	opts.Logger.Infof("Looking up the EC2 Instances...")
	instances, err := aws.DescribeAllInstances(instanceIds)
	if err != nil {
		return err
	}

	opts.Logger.Infof("Looking up the Security Groups attached to the Instances...")
	sgsForInstances, err := aws.FindSecurityGroupIdsForInstances(instances)
	if err != nil {
		return err
	}

	port, err := aws.GetPortForTargets(targets)
	if err != nil {
		return err
	}

	opts.Logger.Infof("Checking the instances allow inbound requests from the ELB...")
	for _, sg := range sgsForInstances {
		allowed, err := instanceSecurityGroupAllowsInBoundFromELB(port, sg, elb, enisForElb)
		if err != nil {
			return err
		}
		if allowed {
			opts.Logger.Infof("Security Group '%s' on the EC2 Instances allows inbound requests from the ELB!", rawAws.StringValue(sg.GroupId))
			return nil
		}
	}

	output.ShowDiagnosis("None of the Security Groups attached to the EC2 Instances seem to allow inbound requests from the ELB. You should update those Security Groups so the ELB can send requests (including health checks) to the apps running on those Instances!", opts)
	return errors.WithStackTrace(InstancesDontAllowInbound{InstanceSecurityGroups: sgsForInstances, Elb: elb})
}

func ElbSecurityGroupsAllowOutboundToInstances(elb *elbv2.LoadBalancer, targets []*elbv2.TargetHealthDescription, opts *options.Options) error {
	var instanceIds []string
	for _, target := range targets {
		instanceIds = append(instanceIds, rawAws.StringValue(target.Target.Id))
	}

	opts.Logger.Infof("Looking up the EC2 Instances...")
	instances, err := aws.DescribeAllInstances(instanceIds)
	if err != nil {
		return err
	}

	opts.Logger.Infof("Looking up the Security Groups attached to the Instances...")
	sgsForInstances, err := aws.FindSecurityGroupIdsForInstances(instances)
	if err != nil {
		return err
	}

	opts.Logger.Infof("Looking up the Security Groups attached to the ELB...")
	sgsForElb, err := aws.DescribeAllSecurityGroups(rawAws.StringValueSlice(elb.SecurityGroups))
	if err != nil {
		return err
	}

	port, err := aws.GetPortForTargets(targets)
	if err != nil {
		return err
	}

	opts.Logger.Infof("Checking the ELB allows outbound requests to the instances...")
	for _, sg := range sgsForElb {
		allowed, err := elbSecurityGroupAllowsOutboundToInstances(port, sg, instances, sgsForInstances)
		if err != nil {
			return err
		}
		if allowed {
			opts.Logger.Infof("Security Group '%s' on the ELB allows outbound requests to the Instances!", rawAws.StringValue(sg.GroupId))
			return nil
		}
	}

	output.ShowDiagnosis("None of the Security Groups attached to the ELB seem to allow outbound requests to the Instances. You should update those Security Groups on the ELB so it can send requests (including health checks) to your Instances!", opts)
	return errors.WithStackTrace(ElbDoesntAllowOutbound{ElbSecurityGroups: sgsForElb, Elb: elb})
}

func elbSecurityGroupAllowsOutboundToInstances(port int, elbSecurityGroup *ec2.SecurityGroup, instances []*ec2.Instance, instanceSecurityGroups []*ec2.SecurityGroup) (bool, error) {
	for _, ipPermission := range elbSecurityGroup.IpPermissionsEgress {
		// TODO: this needs to check protocol too
		if portInRange(port, ipPermission) {
			if instanceSecurityGroupIsWhiteListed(instanceSecurityGroups, ipPermission) {
				return true, nil
			}

			ipWhitelisted, err := instanceIpInRange(instances, ipPermission)
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

func instanceSecurityGroupIsWhiteListed(instanceSecurityGroups []*ec2.SecurityGroup, ipPermission *ec2.IpPermission) bool {
	for _, userIdPair := range ipPermission.UserIdGroupPairs {
		for _, instanceSg := range instanceSecurityGroups {
			if rawAws.StringValue(userIdPair.GroupId) == rawAws.StringValue(instanceSg.GroupId) {
				return true
			}
		}
	}

	return false
}

func instanceIpInRange(instances []*ec2.Instance, ipPermission *ec2.IpPermission) (bool, error) {
	for _, instance := range instances {
		instancePrivateIp := net.ParseIP(rawAws.StringValue(instance.PrivateIpAddress))

		for _, ipRange := range ipPermission.IpRanges {
			allowed, err := ipRangeContainsIp(rawAws.StringValue(ipRange.CidrIp), instancePrivateIp)
			if err != nil {
				return false, err
			}
			if allowed {
				return true, nil
			}
		}

		for _, ipRange := range ipPermission.Ipv6Ranges {
			allowed, err := ipRangeContainsIp(rawAws.StringValue(ipRange.CidrIpv6), instancePrivateIp)
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

func instanceSecurityGroupAllowsInBoundFromELB(port int, instanceSecurityGroup *ec2.SecurityGroup, elb *elbv2.LoadBalancer, enisForElb []*ec2.NetworkInterface) (bool, error) {
	for _, ipPermission := range instanceSecurityGroup.IpPermissions {
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
	// Port must be in range or port values must be nil, which AWS seems to use to mean "any port"
	return (ipPermission.FromPort == nil || rawAws.Int64Value(ipPermission.FromPort) <= int64(port)) && (ipPermission.ToPort == nil || rawAws.Int64Value(ipPermission.ToPort) >= int64(port))
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

type InstancesDontAllowInbound struct {
	InstanceSecurityGroups []*ec2.SecurityGroup
	Elb                    *elbv2.LoadBalancer
}

func (err InstancesDontAllowInbound) Error() string {
	var sgIds []string
	for _, sg := range err.InstanceSecurityGroups {
		sgIds = append(sgIds, rawAws.StringValue(sg.GroupId))
	}
	return fmt.Sprintf("None of the Security Groups attached to the EC2 Instances ('%v') seem to allow inbound requests from the ELB '%s'", sgIds, rawAws.StringValue(err.Elb.LoadBalancerName))
}

type ElbDoesntAllowOutbound struct {
	ElbSecurityGroups []*ec2.SecurityGroup
	Elb                *elbv2.LoadBalancer
}

func (err ElbDoesntAllowOutbound) Error() string {
	var sgIds []string
	for _, sg := range err.ElbSecurityGroups {
		sgIds = append(sgIds, rawAws.StringValue(sg.GroupId))
	}
	return fmt.Sprintf("None of the Security Groups attached to the ELB '%s' ('%v') seem to allow outbound requests to the Instances", rawAws.StringValue(err.Elb.LoadBalancerName), sgIds)
}
