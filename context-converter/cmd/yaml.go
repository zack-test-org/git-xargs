package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

var YQ_BINARY = "yq"

// yq w config.yml 'workflows.*.jobs.*.*.context[+]' "Gruntwork Admin"

func runYqCommand(args ...string) {
	cmd := exec.Command(YQ_BINARY, args...)
	stdout, err := cmd.Output()

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug(fmt.Sprintf("Error running command against %s", YQ_BINARY))
	}

	fmt.Println(string(stdout))
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

	// Clean up the temp file when we're done with it
	defer os.Remove(tmpFile.Name())

	// Run yq command on the tempfile to update it
	runYqCommand("w", "-i", tmpFile.Name(), "workflows.*.jobs.*.*.context[+]", "Gruntwork Admin")

	b, readErr := ioutil.ReadFile(tmpFile.Name())

	if readErr != nil {
		log.WithFields(logrus.Fields{
			"Error": readErr,
		}).Debug("Error reading temp file after writing updated YAML to it")
	}

	return b
}
