package checks

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/gruntwork-io/prototypes/diagnose/aws"
	"github.com/gruntwork-io/prototypes/diagnose/options"
	"github.com/gruntwork-io/prototypes/diagnose/output"
)

func TestInstancesViaLocalhost(targets []*elbv2.TargetHealthDescription, opts *options.Options) error {
	port := aws.GetPortForTargets(targets)
	instanceIds := aws.GetInstanceIdsForTargets(targets)

	command := fmt.Sprintf("curl --silent --location --fail --show-error localhost:%s", port)
	opts.Logger.Infof("Using SSM to run command on all ELB targets to check local connectivity: %s", command)

	err := aws.RunShellCommandViaSsm(command, instanceIds, opts)
	if err != nil {
		output.ShowDiagnosis(fmt.Sprintf("Testing the instances via localhost failed. This most likely means your web service is not running or not listening on the port (%d) you expect.", port), opts)
		// TODO: we could run commands via SSM to check if anything is running or listening on that port!
		return err
	}

	return nil
}