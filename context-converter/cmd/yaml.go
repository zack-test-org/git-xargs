package cmd

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

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

// Count the number of nested Workflows -> Jobs -> Context fields in the YAML document
func configFileHasContexts(filename string) bool {

	contextsCount := getIntFromCommand("r", filename, "--length", "--collect", "workflows.*.jobs.*.*.context")

	if contextsCount < 1 {
		return false
	}
	return true
}

// Append the Workflows -> Jobs -> Context arrays to the YAML document (and add the "Gruntwork Admin" member to these context arrays)
// Note that yq's append behavior is similar to `mkdir -p` in that it will add missing nodes as needed
// to satisy the path expression passed into yq (workflows.*.jobs.*.context[+])
// Therefore, this method can be called once it's determined that none of the YAML document's Workflows -> Jobs nodes have any context arrays
func appendContextNodes(filename string) {

	cmdOutput := runYqCommand("w", "-i", filename, "workflows.*.jobs.*.context[+]", "Gruntwork Admin")

	log.WithFields(logrus.Fields{
		"cmdOutput": cmdOutput,
	}).Debug("appendContextNodes ran command to add and populate context nodes")
}

// Takes in the raw YAML file bytes and creates a temporary file to write them to
// This temporary file is then further processed by the various methods, with updates made in-place via yq's -i flag
// When processing is complete, the final contents of this temporary file are read again and then PUT against the original file via the Github API in order to update it
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
// First, the YAML is written to temporary file, and then the temporary file is updated in place
// When processing is complete, the final temp file contents are read out again and returned as bytes, suitable for making updates via the Github API
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

	// By this point, all processing of the tempfile via yq is complete, so its contents can
	// be read out again
	updatedYamlBytes, readErr := ioutil.ReadFile(tmpFileName)

	if readErr != nil {
		log.WithFields(logrus.Fields{
			"Error": readErr,
		}).Debug("Could not read updated YAML file into bytes")
	}

	return updatedYamlBytes
}
