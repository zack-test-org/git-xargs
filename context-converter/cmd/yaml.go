package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var YQ_BINARY = "yq"

// Accept an arbitrary number of string arguments to pass to the yq binary
// Run yq with the supplied arguments and return its output as a byte slice
func runYqCommand(args ...string) []byte {
	cmd := exec.Command(YQ_BINARY, args...)
	stdout, err := cmd.Output()

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error":   err,
			"args...": args,
		}).Debug(fmt.Sprintf("Error running command against %s", YQ_BINARY))
	}

	return stdout
}

// Count the number of workflows blocks defined in the config file, as we can only programmatically operate
// on workflows blocks that already exist
func ensureConfigFileHasWorkflowsBlock(filename string) bool {

	workflowsCount := getIntFromCommand("r", filename, "--length", "workflows")

	if workflowsCount < 1 {
		log.WithFields(logrus.Fields{
			"Error": "This config file does not already use workflows. Cannot programmatically build it",
		}).Debug("Config file missing workflows block")

		return false
	}

	return true
}

// Ensure the config file's Workflows block is using at least syntax version 2.0, which
// contains support for contexts
func ensureWorkflowSyntaxVersion(filename string) bool {

	syntaxVersion := getFloatFromCommand("r", filename, "workflows.version")

	if syntaxVersion < 2.0 {
		log.WithFields(logrus.Fields{
			"Parsed version": syntaxVersion,
		}).Debug("Workflows syntax version too low to support contexts")

		return false
	}

	return true
}

func getFloatFromCommand(args ...string) float64 {

	cmdOutput := runYqCommand(args...)

	cmdOutputString := string(cmdOutput)

	strippedOutput := strings.ReplaceAll(cmdOutputString, "\\n", "")
	cleanedOutput := strings.TrimSpace(strippedOutput)

	parsedFloat, err := strconv.ParseFloat(cleanedOutput, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
			"Args":  args,
		}).Debug("Error parsing float from cmd output")
		return 0
	}
	return parsedFloat
}

func getIntFromCommand(args ...string) int64 {

	cmdOutput := runYqCommand(args...)

	cmdOutputString := string(cmdOutput)

	strippedOutput := strings.ReplaceAll(cmdOutputString, "\\n", "")
	cleanedOutput := strings.TrimSpace(strippedOutput)

	// yq will return nothing to STDOUT if the count is empty
	if cleanedOutput == "" {
		cleanedOutput = "0"
	}

	parsedInt, err := strconv.ParseInt(cleanedOutput, 10, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
			"Args":  args,
		}).Debug("Error parsing int from cmd output")
		return 0
	}
	return parsedInt
}

// Count the number of nested Workflows -> Jobs -> Context fields in the YAML document
func configFileHasContexts(filename string) bool {

	contextsCount := getIntFromCommand("r", filename, "--length", "--collect", "workflows.*.jobs.*.*.context")

	if contextsCount < 1 {
		return false
	}
	return true
}

func appendContextNodes(filename string) {

	cmdOutput := runYqCommand("w", "-i", filename, "workflows.*.jobs.*.context[+]", "Gruntwork Admin")

	log.WithFields(logrus.Fields{
		"cmdOutput": cmdOutput,
	}).Debug("appendContextNodes ran command to add and populate context nodes")
}

func writeYamlToTempFile(b []byte) *os.File {

	tmpFile, err := ioutil.TempFile("", "circle-ci-context")
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Fatal("Error creating temporary YAML file")
	}

	if _, writeErr := tmpFile.Write(b); writeErr != nil {
		log.WithFields(logrus.Fields{
			"Error": writeErr,
		}).Debug("Error writing YAML to temporary file")
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		log.WithFields(logrus.Fields{
			"Error": closeErr,
		}).Debug("Error closing temporary file after writing YAML")
	}

	return tmpFile
}

// Use yq to make the required updates to the supplied YAML file
func UpdateYamlDocument(b []byte) []byte {

	tmpFile := writeYamlToTempFile(b)
	tmpFileName := tmpFile.Name()

	// Clean up the temp file when we're done with it
	defer os.Remove(tmpFileName)

	// Only operate on files with `Workflows` blocks already defined. Currently, we cannot programmatically build out the workflows block
	if !ensureConfigFileHasWorkflowsBlock(tmpFileName) {
		return nil
	}

	if !ensureWorkflowSyntaxVersion(tmpFileName) {
		return nil
	}

	// If none of the config file's Workflow -> Jobs nodes have context fields, append them
	// Note this function will both append the context arrays and add the correct "Gruntwork Admin" member
	if !configFileHasContexts(tmpFileName) {
		appendContextNodes(tmpFileName)
	}

	updatedYamlBytes, readErr := ioutil.ReadFile(tmpFileName)

	if readErr != nil {
		log.WithFields(logrus.Fields{
			"Error": readErr,
		}).Debug("Could not read updated YAML file into bytes")
	}

	return updatedYamlBytes
}
