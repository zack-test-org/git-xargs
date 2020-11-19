package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// Count the number of workflows blocks defined in the config file, as we can only programmatically operate
// on workflows blocks that already exist
func ensureConfigFileHasWorkflowsBlock(filename string) bool {

	workflowsCount, err := getIntFromCommand("r", filename, "-X", "--length", "workflows")

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

	syntaxKeyCount, err := getIntFromCommand("r", filename, "-X", "--length", "workflows.version")

	if err != nil {
		log.Debug("Unable to verify workflows block declares a syntax version")
		return false
	}

	if syntaxKeyCount < 1 {
		log.Debug("Could not find workflows.version key, so can't programmatically operate on this YAML file")
		return false
	}

	syntaxVersion, versionLookupErr := getFloatFromCommand("r", filename, "-X", "workflows.version")

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

	contextsCount, err := getIntFromCommand("r", filename, "-X", "--length", "--collect", "workflows.*.jobs.*.*.context")

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
	originalVersion, err := getIntFromCommand("r", filename, "-X", "workflows.version")

	if err != nil {
		log.Debug("Unable to lookup workflows syntax version - can't safely operate on file")
		return
	}

	cmdOutput, err := runYqCommand("w", "-i", filename, "workflows.*.jobs[*].*.context[+]", TargetContext)

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

// convertScalarJobNodes handles the edge case that is not currently supported by `yq` version 3's path expressions: https://mikefarah.gitbook.io/yq/usage/path-expressions,
// namely that if a job is a scalar value (like a string, and not a map or array) and exists, for example, under workflows -> nightly -> jobs -> test, it will not be converted into a map or object by the same write query used in appendContextNodes
// In these cases - we need to read out all such jobs that exist as a single string - and programmatically generate their path expressions and then use a separate write command to update them to be a map with a key: contexts (which will be the usual array containing the TargetContext)
func convertScalarJobNodes(filename string) {

	scalarJobOutput, err := runYqCommand("r", filename, "--printMode", "pv", "workflows.*.jobs.[*]")

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error reading scalar job nodes via yq")
	}

	log.WithFields(logrus.Fields{
		"Scalar job output": string(scalarJobOutput),
	}).Debug("Read scalar job output")

	// Regular expression for extracting scalar job nodes from the yq read output
	scalarJobsRe := regexp.MustCompile(`workflows\.\w+\.jobs\.\[\d\]\:\s\w+`)
	// Regular expression for converting the scalar job node name into a valid yq path expression
	jobNameCleanRe := regexp.MustCompile(`\:\s`)

	// Pull all the scalar (e.g; single string job names like "-test") into a slice
	scalarJobNames := scalarJobsRe.FindAllString(string(scalarJobOutput), -1)

	log.WithFields(logrus.Fields{
		"Scalar Job Names": scalarJobNames,
	}).Debug("Read Scalar jobs from file")

	// loop through scalar job names and format the yq path expression to address them directly in a write query
	for _, job := range scalarJobNames {
		jobPath := jobNameCleanRe.ReplaceAllString(job, ".")
		jobPathExpression := fmt.Sprintf("%s.context[+]", jobPath)

		log.WithFields(logrus.Fields{
			"jobPathExpression": jobPathExpression,
		}).Debug("convertScalarNodes using parsed scalar job path expression to append context")

		scalarJobWriteOutput, writeErr := runYqCommand("w", "-i", filename, jobPathExpression, TargetContext)
		if writeErr != nil {
			log.WithFields(logrus.Fields{
				"Error":    writeErr,
				"Filename": filename,
			}).Debug("Error appending contexts to scalar job")
		}

		log.WithFields(logrus.Fields{
			"Command output": scalarJobWriteOutput,
		}).Debug("convertScalarNodes write command output")
	}
}

// Get the count of the all the context nodes under the path Workflows -> Jobs -> Context
func countTotalContexts(filename string) int64 {

	countTotalContexts, err := getIntFromCommand("r", filename, "-X", "--length", "--collect", "workflows.*.jobs.*.*.context")

	if err != nil {
		log.Debug("Unable to count number of contexts defined in file")
		return 0
	}

	log.WithFields(logrus.Fields{
		"TotalContextsInFile": countTotalContexts,
	}).Debug("countTotalContexts found contexts")

	return countTotalContexts
}

// Get the count of all the context arrays under the path Workflows -> Jobs -> Contexts that already contain "Gruntwork Admin" as a member
func countContextsWithMember(filename string) int64 {

	pathExpression := fmt.Sprintf("workflows.*.jobs.*.*.context(.==%s)", TargetContext)

	log.WithFields(logrus.Fields{
		"pathExpression": pathExpression,
	}).Debug("yq pathExpression used to count number of contexts already correctly set")

	countContextsCorrectlySet, err := getIntFromCommand("r", filename, "-X", "--length", "--collect", pathExpression)

	if err != nil {
		log.Debug("Unable to count number of contexts arrays with desired context member")
		return 0
	}

	log.WithFields(logrus.Fields{
		"ContextsWithTargetMember": countContextsCorrectlySet,
	}).Debug("countContextsWithMember found number of contexts already set correctly")

	return countContextsCorrectlySet
}

// Checks if the config file already has the expected contexts set, by comparing the count of total context arrays
// with the count of context arrays that contain the TargetContext as a member
func correctContextsAlreadyPresent(filename string) bool {
	log.WithFields(logrus.Fields{
		"Filename": filename,
	}).Debug("Checking if correct Contexts already in place...")
	return countTotalContexts(filename) == countContextsWithMember(filename)
}

// UpdateYamlDocument uses yq to make the required updates to the supplied YAML file
// First, the YAML is written to temporary file, and then the temporary file is updated in place
// When processing is complete, the final temp file contents are read out again and returned as bytes, suitable for making updates via the Github API
func UpdateYamlDocument(yamlBytes []byte, debug bool, repo *github.Repository, stats *RunStats) []byte {

	tmpFile := writeYamlToTempFile(yamlBytes)
	tmpFileName := tmpFile.Name()

	// Clean up the temp file when we're done with it
	defer func() {
		removeErr := os.Remove(tmpFileName)
		if removeErr != nil {
			log.WithFields(logrus.Fields{
				"Error": removeErr,
			}).Debug("Error deleting YAML tempfile")
		}
	}()

	// Only operate on files with `Workflows` blocks already defined. Currently, we cannot programmatically build out the workflows block
	if !ensureConfigFileHasWorkflowsBlock(tmpFileName) {
		stats.TrackSingle(WorkflowsMissing, repo)
		return nil
	}

	if !ensureWorkflowSyntaxVersion(tmpFileName) {
		stats.TrackSingle(WorkflowsSyntaxOutdated, repo)
		return nil
	}

	if debug {
		fmt.Println("*** DEBUG - PRIOR TO YQ WRITING TO TEMPFILE IN PLACE ***")
		dumpTempFileContents(tmpFile)
	}

	// If none of the config file's Workflow -> Jobs nodes have context fields, append them
	// Note this function will both append the context arrays and add the correct "Gruntwork Admin" member
	if !configFileHasContexts(tmpFileName) {

		// To all jobs that are of object type, append the expected context
		appendContextNodes(tmpFileName)

		// For all jobs that are of scalar types (single string names in YAML) append the expected context
		convertScalarJobNodes(tmpFileName)

		if debug {
			fmt.Println("*** DEBUG - POST YQ WRITING TO TEMPFILE IN PLACE ***")
			dumpTempFileContents(tmpFile)
		}
	} else {
		// If the config file's Workflows -> Jobs -> Contexts nodes already have the desired context set, return because there's nothing to do
		// This is determined by checking if the count of context nodes is equal to the number of context nodes that contain the "Gruntwork Admin" member
		if correctContextsAlreadyPresent(tmpFileName) {

			log.Debug("All contexts have the correct member - Gruntwork Admin already. Skipping this file!")

			stats.TrackSingle(ContextAlreadySet, repo)
			return nil
		}

		log.WithFields(logrus.Fields{
			"Filename": tmpFileName,
		}).Debug("File was NOT detected as already having all correct contexts set")
	}

	// By this point, all processing of the tempfile via yq is complete, so its contents can
	// be read out again
	updatedYamlBytes, readErr := ioutil.ReadFile(tmpFileName)

	// When yq writes to the tempfile to append the context nodes, it leaves these !!merge tags, which are the node types as determined by the underlying go-yaml v3 package
	// Hopefully when yq v4 beta is released, we can try updating to it and improving some of the yq commands and hopefully drop this unfortunate manual sanitization step, too
	sanitizedUpdatedYamlBytes := strings.ReplaceAll(string(updatedYamlBytes), "!!merge ", "")

	if readErr != nil {
		log.WithFields(logrus.Fields{
			"Error": readErr,
		}).Debug("Could not read updated YAML file into bytes")
	}

	return []byte(sanitizedUpdatedYamlBytes)
}
