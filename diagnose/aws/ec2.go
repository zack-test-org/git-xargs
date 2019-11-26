package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/gruntwork-io/gruntwork-cli/collections"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"regexp"
)

func FindSecurityGroupIdsForInstances(instanceIds []string) ([]*ec2.SecurityGroup, error) {
	session, err := NewAuthenticatedSession()
	if err != nil {
		return nil, err
	}

	client := ec2.New(session)

	instances, err := describeAllInstances(instanceIds, client, nil)
	if err != nil {
		return nil, err
	}

	var securityGroupIds []string
	for _, instance := range instances {
		for _, securityGroup := range instance.SecurityGroups {
			securityGroupId := aws.StringValue(securityGroup.GroupId)
			if !collections.ListContainsElement(securityGroupIds, securityGroupId) {
				securityGroupIds = append(securityGroupIds, securityGroupId)
			}
		}
	}

	return describeAllSecurityGroups(securityGroupIds, client, nil)
}

func FindEnisForElb(elb *elbv2.LoadBalancer) ([]*ec2.NetworkInterface, error) {
	session, err := NewAuthenticatedSession()
	if err != nil {
		return nil, err
	}

	client := ec2.New(session)

	fullElbName, err := getElbFullName(elb)
	if err != nil {
		return nil, err
	}

	// See https://aws.amazon.com/premiumsupport/knowledge-center/elb-find-load-balancer-IP/ for context
	elbNameForEniSearch := fmt.Sprintf("ELB %s", fullElbName)
	input := ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("description"), Values: aws.StringSlice([]string{fmt.Sprintf(elbNameForEniSearch)})},
		},
	}

	output, err := client.DescribeNetworkInterfaces(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return output.NetworkInterfaces, nil
}

func describeAllInstances(instanceIds []string, client *ec2.EC2, nextToken *string) ([]*ec2.Instance, error) {
	input := ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instanceIds),
		NextToken:   nextToken,
	}

	output, err := client.DescribeInstances(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	var instances []*ec2.Instance
	for _, reservation := range output.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	if aws.StringValue(output.NextToken) == "" {
		return instances, nil
	}

	rest, err := describeAllInstances(instanceIds, client, output.NextToken)
	if err != nil {
		return nil, err
	}

	return append(instances, rest...), nil
}

func describeAllSecurityGroups(securityGroupIds []string, client *ec2.EC2, nextToken *string) ([]*ec2.SecurityGroup, error) {
	input := ec2.DescribeSecurityGroupsInput{
		GroupIds:  aws.StringSlice(securityGroupIds),
		NextToken: nextToken,
	}

	output, err := client.DescribeSecurityGroups(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if aws.StringValue(output.NextToken) == "" {
		return output.SecurityGroups, nil
	}

	rest, err := describeAllSecurityGroups(securityGroupIds, client, output.NextToken)
	if err != nil {
		return nil, err
	}

	return append(output.SecurityGroups, rest...), nil
}

// Parse an ELB name from an ELB ARN.
// arn:aws:elasticloadbalancing:<REGION>:<ACCOUNT_ID>:loadbalancer/<NAME>
var elbNameFromElbArnRegex = regexp.MustCompile("arn:aws:elasticloadbalancing:.+?:.+?:loadbalancer/(.+)")

// Get the full ELB name for the given ELB. Note that the "full" name includes the ELB type and a unique ID, and not
// just the human-friendly name you pass in. E.g., For an ALB named foo, the full name may be app/foo/abcdef12345.
func getElbFullName(elb *elbv2.LoadBalancer) (string, error) {
	arn := aws.StringValue(elb.LoadBalancerArn)
	matches := elbNameFromElbArnRegex.FindStringSubmatch(arn)
	if len(matches) != 2 {
		return "", errors.WithStackTrace(InvalidElbArn(arn))
	}
	return matches[1], nil
}

type InvalidElbArn string

func (err InvalidElbArn) Error() string {
	return fmt.Sprintf("Unable to parse invalid ELB ARN: %s", string(err))
}
