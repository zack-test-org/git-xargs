package check

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/prototypes/diagnose/options"
)

func ElbHealthChecksPassing(targets []*elbv2.TargetHealthDescription, opts *options.Options) error {
	errorsFound := false

	opts.Logger.Info("ELB health check statuses for your EC2 Instances:")

	for _, target := range targets {
		opts.Logger.Infof("Instance: '%s'. Status: '%s'. Reason: '%s'. Description: '%s'.", aws.StringValue(target.Target.Id), aws.StringValue(target.TargetHealth.State), aws.StringValue(target.TargetHealth.Reason), aws.StringValue(target.TargetHealth.Description))

		if aws.StringValue(target.TargetHealth.State) != elbv2.TargetHealthStateEnumHealthy {
			errorsFound = true
		}
	}

	if errorsFound {
		return errors.WithStackTrace(FailingHealthChecksFound(""))
	}

	return nil
}

type FailingHealthChecksFound string

func (err FailingHealthChecksFound) Error() string {
	return fmt.Sprintf("There were failing health checksin the ELB. See log output above for details.")
}