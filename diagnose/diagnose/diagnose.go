package diagnose

import (
	"fmt"
	"github.com/gruntwork-io/prototypes/diagnose/aws"
	"github.com/gruntwork-io/prototypes/diagnose/check"
	"github.com/gruntwork-io/prototypes/diagnose/options"
	"github.com/gruntwork-io/prototypes/diagnose/output"
)

func Diagnose(opts *options.Options) error {
	recordSet, err := aws.FindResourceRecordSetForUrl(opts)
	if err != nil {
		return err
	}

	if recordSet == nil {
		opts.Logger.Errorf("Could not find a Route 53 Resource Record Set for URL '%s'.", opts.Url)
		return nil
	}

	opts.Logger.Infof("Found Route53 RecordSet URL '%s': %v", opts.Url, recordSet)

	elb, err := aws.FindElbForResourceRecordSet(recordSet)
	if err != nil {
		return err
	}

	opts.Logger.Infof("Found ELB for RecordSet: %v", elb)

	targets, err := aws.FindTargetsRegisteredInElb(elb)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		opts.Logger.Error("Could not find any targets in ELB.")
		return nil
	}

	opts.Logger.Infof("Found targets for ELB: %v", targets)

	if err := check.CanAccessWebServiceViaLocalhost(targets, opts); err != nil {
		return err
	}

	if err := check.InstanceSecurityGroupsAllowInboundFromELB(elb, targets, opts); err != nil {
		return err
	}

	if err := check.ElbSecurityGroupsAllowOutboundToInstances(elb, targets, opts); err != nil {
		return err
	}

	if err := check.ElbHealthChecksPassing(targets, opts); err != nil {
		return err
	}

	// TODO: instance subnet NACLs allow inbound requests from ELB
	// TODO: instance subnet NACLs allow outbound response from instances
	// TODO: ELB subnet NACLs allow outbound requests to instances
	// TODO: ELB subnet NACLs allow inbound responses from instances
	// TODO: ELB allows inbound from outside world (from your computer's IP)

	output.ShowDiagnosis(fmt.Sprintf("Everything appears to be working. We were not able to find any errors talking to URL %s.", opts.Url), opts)

	return nil
}
