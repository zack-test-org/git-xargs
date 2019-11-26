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

	if err := check.SecurityGroupAllowsInboundFromELB(elb, targets, opts); err != nil {
		return err
	}

	// TODO: ELB SG allows outbound to instance

	if err := check.ElbHealthChecksPassing(targets, opts); err != nil {
		return err
	}

	// TODO: ELB allows inbound from outside world?

	output.ShowDiagnosis(fmt.Sprintf("Everything appears to be working. We were not able to find any errors talking to URL %s.", opts.Url), opts)

	return nil
}
