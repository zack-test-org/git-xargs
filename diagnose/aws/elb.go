package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"strconv"
	"strings"
)

func FindTargetsRegisteredInElb(elb *elbv2.LoadBalancer) ([]*elbv2.TargetHealthDescription, error) {
	session, err := NewAuthenticatedSession()
	if err != nil {
		return nil, err
	}

	client := elbv2.New(session)

	targetGroups, err := getAllTargetGroupsForLoadBalancer(elb, client, nil)
	if err != nil {
		return nil, err
	}

	var allTargetHealthDescriptions []*elbv2.TargetHealthDescription
	for _, targetGroup := range targetGroups {
		targetHealthDescriptions, err := getTargetHealthDescriptions(targetGroup, client)
		if err != nil {
			return nil, err
		}
		allTargetHealthDescriptions = append(allTargetHealthDescriptions, targetHealthDescriptions...)
	}

	return allTargetHealthDescriptions, nil
}

func GetPortForTargets(targets []*elbv2.TargetHealthDescription) (int, error) {
	portAsString := aws.StringValue(targets[0].HealthCheckPort)
	port, err := strconv.Atoi(portAsString)
	if err != nil {
		return -1, errors.WithStackTrace(err)
	}
	return port, nil
}

func GetInstanceIdsForTargets(targets []*elbv2.TargetHealthDescription) []string {
	var instanceIds []string

	for _, target := range targets {
		instanceIds = append(instanceIds, aws.StringValue(target.Target.Id))
	}

	return instanceIds
}

func FindElbForResourceRecordSet(recordSet *route53.ResourceRecordSet) (*elbv2.LoadBalancer, error) {
	if recordSet.AliasTarget == nil {
		return nil, errors.WithStackTrace(InvalidRecordSetForElb{RecordSet: recordSet})
	}

	return FindElbForDomainName(cleanupDnsNameForElbSearch(aws.StringValue(recordSet.AliasTarget.DNSName)))
}

func FindElbForDomainName(domainName string) (*elbv2.LoadBalancer, error) {
	session, err := NewAuthenticatedSession()
	if err != nil {
		return nil, err
	}

	client := elbv2.New(session)
	return findElbForDomainName(domainName, client, nil)
}

func getTargetHealthDescriptions(targetGroup *elbv2.TargetGroup, client *elbv2.ELBV2) ([]*elbv2.TargetHealthDescription, error) {
	input := elbv2.DescribeTargetHealthInput{
		TargetGroupArn: targetGroup.TargetGroupArn,
	}
	output, err := client.DescribeTargetHealth(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return output.TargetHealthDescriptions, nil
}

func getAllTargetGroupsForLoadBalancer(elb *elbv2.LoadBalancer, client *elbv2.ELBV2, marker *string) ([]*elbv2.TargetGroup, error) {
	input := elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: elb.LoadBalancerArn,
		Marker:          marker,
	}
	output, err := client.DescribeTargetGroups(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if aws.StringValue(output.NextMarker) == "" {
		return output.TargetGroups, nil
	}

	rest, err := getAllTargetGroupsForLoadBalancer(elb, client, output.NextMarker)
	if err != nil {
		return nil, err
	}

	return append(output.TargetGroups, rest...), nil
}

func findElbForDomainName(domainName string, client *elbv2.ELBV2, marker *string) (*elbv2.LoadBalancer, error) {
	input := elbv2.DescribeLoadBalancersInput{
		Marker: marker,
	}

	output, err := client.DescribeLoadBalancers(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	for _, elb := range output.LoadBalancers {
		if aws.StringValue(elb.DNSName) == domainName {
			return elb, nil
		}
	}

	if aws.StringValue(output.NextMarker) != "" {
		return findElbForDomainName(domainName, client, output.NextMarker)
	}

	return nil, nil
}

// The Route 53 APIs return alias resource record sets with DNSNames of the format:
//
// dualstack.<NAME>.<REGION>.elb.amazonaws.com.
//
// The ELB APIs return DNSNames of the format:
//
// <NAME>.<REGION>.elb.amazonaws.com
//
// (That is, no "dualstack" prefix or dot suffix).
//
// This method cleans up DNSNames of the former variety so that they will match DNSNames of the latter variety.
func cleanupDnsNameForElbSearch(dnsName string) string {
	return strings.TrimSuffix(strings.TrimPrefix(dnsName, "dualstack."), ".")
}

type InvalidRecordSetForElb struct {
	RecordSet *route53.ResourceRecordSet
}

func (err InvalidRecordSetForElb) Error() string {
	return fmt.Sprintf("Invalid Record Set for an ELB, as it does not contain an AliasTarget: %v", err.RecordSet)
}
