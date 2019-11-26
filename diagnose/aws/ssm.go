package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/gruntwork-io/gruntwork-cli/collections"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/prototypes/diagnose/options"
	"strings"
	"time"
)

func RunShellCommandViaSsm(command string, instanceIds []string, opts *options.Options) error {
	session, err := NewAuthenticatedSession()
	if err != nil {
		return err
	}

	client := ssm.New(session)

	runningCommand, err := runCommand(command, instanceIds, client)
	if err != nil {
		return err
	}

	outputs, err := waitForCommandToComplete(runningCommand, client, opts)
	if err != nil {
		return err
	}

	return checkOutputs(outputs, command, opts)
}

func runCommand(command string, instanceIds []string, client *ssm.SSM) (*ssm.Command, error) {
	parameters := map[string][]*string{
		"executionTimeout": aws.StringSlice([]string{"5"}),
		"commands":         aws.StringSlice([]string{command}),
	}

	input := ssm.SendCommandInput{
		Comment:         aws.String("diagnose utility executing local commands to test a web service"),
		DocumentName:    aws.String("AWS-RunShellScript"),
		DocumentVersion: aws.String("1"),
		InstanceIds:     aws.StringSlice(instanceIds),
		Parameters:      parameters,
	}

	output, err := client.SendCommand(&input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return output.Command, nil
}

func checkOutputs(outputs []*ssm.GetCommandInvocationOutput, command string, opts *options.Options) error {
	errorsFound := false

	opts.Logger.Infof("Results of running command '%s' on localhost on the following instances:", command)

	for _, output := range outputs {
		opts.Logger.Infof("Instance: '%s'. Status: '%s'. StatusDetails: '%s'. ExitCode: '%d'. StdOut: '%s'. StdErr: '%s'.", aws.StringValue(output.InstanceId), aws.StringValue(output.Status), aws.StringValue(output.StatusDetails), aws.Int64Value(output.ResponseCode), aws.StringValue(output.StandardOutputContent), aws.StringValue(output.StandardErrorContent))
		if aws.StringValue(output.Status) != ssm.CommandInvocationStatusSuccess {
			errorsFound = true
		}
	}

	if errorsFound {
		return errors.WithStackTrace(ErrorsFoundRunningCommand(command))
	}

	return nil
}

func waitForCommandToComplete(runningCommand *ssm.Command, client *ssm.SSM, opts *options.Options) ([]*ssm.GetCommandInvocationOutput, error) {
	maxRetries := 5
	timeBetweenRetries := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		outputs, err := getCommandStatusForAllInstances(runningCommand, client)

		fmt.Printf("Got outputs: %v\n", outputs)

		if err == nil && allCommandsCompleted(runningCommand, outputs) {
			return outputs, nil
		}

		if err != nil {
			if strings.Contains(err.Error(), "InvocationDoesNotExist") {
				opts.Logger.Infof("Invocation '%s' does not yet exist. This is probably due to eventual consistency in AWS.", aws.StringValue(runningCommand.CommandId))
			} else {
				return nil, errors.WithStackTrace(err)
			}
		}

		opts.Logger.Infof("The command is still running. Will sleep for %s and check again.", timeBetweenRetries)
		time.Sleep(timeBetweenRetries)
	}

	return nil, errors.WithStackTrace(CommandTimedOut{RunningCommand: runningCommand, Retries: maxRetries})
}

func getCommandStatusForAllInstances(runningCommand *ssm.Command, client *ssm.SSM) ([]*ssm.GetCommandInvocationOutput, error) {
	var outputs []*ssm.GetCommandInvocationOutput

	for _, instanceId := range runningCommand.InstanceIds {
		input := ssm.GetCommandInvocationInput{
			CommandId:  runningCommand.CommandId,
			InstanceId: instanceId,
		}

		output, err := client.GetCommandInvocation(&input)
		if err != nil {
			return nil, errors.WithStackTrace(err)
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

var commandCompletedStatuses = []string{
	ssm.CommandInvocationStatusCancelled,
	ssm.CommandInvocationStatusFailed,
	ssm.CommandInvocationStatusSuccess,
	ssm.CommandInvocationStatusTimedOut,
}

func allCommandsCompleted(runningCommand *ssm.Command, outputs []*ssm.GetCommandInvocationOutput) bool {
	if len(outputs) != len(runningCommand.InstanceIds) {
		return false
	}

	for _, output := range outputs {
		if !collections.ListContainsElement(commandCompletedStatuses,aws.StringValue(output.Status)) {
			return false
		}
	}

	return true
}

type CommandTimedOut struct {
	RunningCommand *ssm.Command
	Retries        int
}

func (err CommandTimedOut) Error() string {
	return fmt.Sprintf("Command %s still not completed after %d retries.", aws.StringValue(err.RunningCommand.CommandId), err.Retries)
}

type ErrorsFoundRunningCommand string

func (err ErrorsFoundRunningCommand) Error() string {
	return fmt.Sprintf("There were errors running command '%s' on localhost on the EC2 instances. See log output above for details.", string(err))
}