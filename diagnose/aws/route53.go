package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/prototypes/diagnose/collections"
	"github.com/gruntwork-io/prototypes/diagnose/options"
	"strings"
)

// Find the hosted zone that configures DNS settings for the URL in the given opts
func FindResourceRecordSetForUrl(opts *options.Options) (*route53.ResourceRecordSet, error) {
	session, err := NewAuthenticatedSession()
	if err != nil {
		return nil, err
	}

	client := route53.New(session)

	opts.Logger.Info("Fetching all Route53 Hosted Zones...")
	allHostedZones, err := getAllHostedZones(client, nil, nil)
	if err != nil {
		return nil, err
	}

	opts.Logger.Infof("Looking for Hosted Zone that matches domain in URL '%s'...", opts.Url)
	hostedZonesForDomain, err := findHostedZonesForDomain(opts.Url, allHostedZones)
	if err != nil {
		return nil, err
	}

	opts.Logger.Infof("Looking for matching record set in hosted zones...")
	return findMatchingRecordSet(client, opts.Url, hostedZonesForDomain)
}

// Fetch all hosted zones in the current AWS account
func getAllHostedZones(client *route53.Route53, dnsName *string, hostedZoneId *string) ([]*route53.HostedZone, error) {
	input := route53.ListHostedZonesByNameInput{
		DNSName:      dnsName,
		HostedZoneId: hostedZoneId,
	}

	output, err := client.ListHostedZonesByName(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if !aws.BoolValue(output.IsTruncated) {
		return output.HostedZones, nil
	}

	rest, err := getAllHostedZones(client, output.NextDNSName, output.NextHostedZoneId)
	if err != nil {
		return nil, err
	}

	return append(output.HostedZones, rest...), nil
}

// Find all the hosted zones from the given list that could configure DNS settings for the domain in the given URL
func findHostedZonesForDomain(url string, hostedZones []*route53.HostedZone) ([]*route53.HostedZone, error) {
	var matchingZones []*route53.HostedZone
	domainParts := formatDomainForComparison(url)

	for _, zone := range hostedZones {
		zoneDomainParts := formatDomainForComparison(aws.StringValue(zone.Name))
		if len(zoneDomainParts) > 0 && collections.IsSubsetOf(domainParts, zoneDomainParts) {
			matchingZones = append(matchingZones, zone)
		}
	}

	return matchingZones, nil
}

// From the given list of hosted zones, find the record set that actually configures the domain name for the given
// URL
func findMatchingRecordSet(client *route53.Route53, url string, hostedZones []*route53.HostedZone) (*route53.ResourceRecordSet, error) {
	domain := formatDomainForSearch(url)

	for _, zone := range hostedZones {
		input := route53.ListResourceRecordSetsInput{
			HostedZoneId:    zone.Id,
			StartRecordName: aws.String(domain),
		}

		output, err := client.ListResourceRecordSets(&input)
		if err != nil {
			return nil, errors.WithStackTrace(err)
		}

		for _, recordSet := range output.ResourceRecordSets {
			if aws.StringValue(recordSet.Name) == domain {
				return recordSet, nil
			}
		}
	}

	return nil, nil
}

// Format the domain name in the given URL for comparison by:
//
// 1. Drop any trailing dots.
// 2. Split the given domain name into parts based on dots within.
// 3. Reverse the order
//
// Example:
//
// formatDomainForComparison("foo.bar.com.") // Returns []string{"com", "bar", "foo"}
func formatDomainForComparison(url string) []string {
	return collections.ReverseSlice(domainutil.SplitDomain(strings.TrimSuffix(url, ".")))
}

// Format the domain name in the given URL for search in Route 53 by extracting the domain and adding a trailing dot.
func formatDomainForSearch(url string) string {
	return fmt.Sprintf("%s.", strings.Join(domainutil.SplitDomain(url), "."))
}
