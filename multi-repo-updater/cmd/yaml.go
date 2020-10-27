package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Count the number of workflows blocks defined in the config file, as we can only programmatically operate
// on workflows blocks that already exist
func ensureConfigFileHasWorkflowsBlock(filename string) bool {

	workflowsCount, err := getIntFromCommand("r", filename, "--length", "workflows")

	if err != nil {
		log.Debug("Unable to verify presence of workflows block for repo")
		return false
	}

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

	syntaxKeyCount, err := getIntFromCommand("r", filename, "--length", "workflows.version")

	if err != nil {
		log.Debug("Unable to verify workflows block declares a syntax version")
		return false
	}

	if syntaxKeyCount < 1 {
		log.Debug("Could not find workflows.version key, so can't programmatically operate on this YAML file")
		return false
	}

	syntaxVersion, versionLookupErr := getFloatFromCommand("r", filename, "workflows.version")

	if versionLookupErr != nil {
		log.Debug("Unable to look up workflows syntax version")
		return false
	}

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

	contextsCount, err := getIntFromCommand("r", filename, "--length", "--collect", "workflows.*.jobs.*.*.context")

	if err != nil {
		log.Debug("Unable to verify file has context nodes")
		return false
	}

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

	// Get the original workflows.version value from the YAML file
	originalVersion, err := getIntFromCommand("r", filename, "workflows.version")

	if err != nil {
		log.Debug("Unable to lookup workflows syntax version - can't safely operate on file")
		return
	}

	cmdOutput, err := runYqCommand("w", "-i", filename, "workflows.*.jobs[*].*.context[+]", "Gruntwork Admin")

	if err != nil {
		log.Debug("Unable to append desired context to Workflows -> jobs nodes")
		return
	}

	// This is profoundly annoying, but for the time being we need to repair the workflows version field that gets nuked by the above write query
	versionRepairOutput, err := runYqCommand("w", "-i", filename, "workflows.version", strconv.FormatInt(originalVersion, 10))

	if err != nil {
		log.Debug("Unable to replace original Workflows -> Version field value")
		return
	}

	log.WithFields(logrus.Fields{
		"cmdOutput":           cmdOutput,
		"versionRepairOutput": versionRepairOutput,
	}).Debug("appendContextNodes ran command to add and populate context nodes")
}

// Get the count of the all the context nodes under the path Workflows -> Jobs -> Context
func countTotalContexts(filename string) int64 {

	countTotalContexts, err := getIntFromCommand("r", filename, "--length", "--collect", "workflows.*.jobs.*.*.context")

	if err != nil {
		log.Debug("Unable to count number of contexts defined in file")
		return 0
	}

	return countTotalContexts
}

// Get the count of all the context arrays under the path Workflows -> Jobs -> Contexts that already contain "Gruntwork Admin" as a member
func countContextsWithMember(filename string) int64 {

	pathExpression := fmt.Sprintf("workflows.*.jobs.*.*.context(.==%s)", TargetContext)

	countContextsCorrectlySet, err := getIntFromCommand("r", filename, "--length", "--collect", pathExpression)

	if err != nil {
		log.Debug("Unable to count number of contexts arrays with desired context member")
		return 0
	}

	return countContextsCorrectlySet
}

// Checks if the config file already has the expected contexts set, by comparing the count of total context arrays
// with the count of context arrays that contain the TargetContext as a member
func correctContextsAlreadyPresent(filename string) bool {
	return countTotalContexts(filename) == countContextsWithMember(filename)
}

// Takes in the raw YAML file bytes and creates a temporary file to write them to
// This temporary file is then further processed by the various methods, with updates made in-place via yq's -i flag
// When processing is complete, the final contents of this temporary file are read again and then PUT against the original file via the Github API in order to update it
func writeYamlToTempFile(yamlBytes []byte) *os.File {

	tmpFile, err := ioutil.TempFile("", "circle-ci-context")
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Fatal("Error creating temporary YAML file")
	}

	if _, writeErr := tmpFile.Write(yamlBytes); writeErr != nil {
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
func UpdateYamlDocument(yamlBytes []byte) []byte {

	tmpFile := writeYamlToTempFile(yamlBytes)
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
	} else {
		// If the config file's Workflows -> Jobs -> Contexts nodes already have the desired context set, return because there's nothing to do
		// This is determined by checking if the count of context nodes is equal to the number of context nodes that contain the "Gruntwork Admin" member
		if correctContextsAlreadyPresent(tmpFileName) {

			log.Debug("All contexts have the correct member - Gruntwork Admin already. Skipping this file!")

			return nil
		}
	}

	// By this point, all processing of the tempfile via yq is complete, so its contents can
	// be read out again
	updatedYamlBytes, readErr := ioutil.ReadFile(tmpFileName)

	// Unfortunately there's an unknown bug in either yq or the underlying go-yaml library that results in our modified files
	// having their *yaml.Node `tag` values written in front of the field itself - e.g.; !!merge <<: *stuff instead of just <<: *stuff, which would be correct
	// Workaround it for the time being by replacing any instances of this - since it's also invalid YAML
	sanitizedUpdatedYamlBytes := strings.ReplaceAll(string(updatedYamlBytes), "!!merge ", "")

	if readErr != nil {
		log.WithFields(logrus.Fields{
			"Error": readErr,
		}).Debug("Could not read updated YAML file into bytes")
	}

	return []byte(sanitizedUpdatedYamlBytes)
}
