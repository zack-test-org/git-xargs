package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// exec contains functions that deal with shelling out to external binaries and processing their output

// YqBinary is the name of the underlying yaml processing CLI tool that this tool wraps
const YqBinary = "yq"

// Accept an arbitrary number of string arguments to pass to the yq binary
// Run yq with the supplied arguments and return its output as a byte slice
func runYqCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(YqBinary, args...)
	stdout, err := cmd.Output()

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error":      err,
			"args...":    args,
			"Cmd Stdout": stdout,
		}).Debug(fmt.Sprintf("Error running command against %s", YqBinary))
		return nil, err
	}

	return stdout, nil
}

// Take in an arbitrary number of string arguments, and pass them along to the yq binary, attempting
// to extract a float (e.g.; 2.0) from the command output
func getFloatFromCommand(args ...string) (float64, error) {

	cmdOutput, err := runYqCommand(args...)

	if err != nil {
		return 0, err
	}

	cmdOutputString := string(cmdOutput)

	strippedOutput := strings.ReplaceAll(cmdOutputString, "\\n", "")
	cleanedOutput := strings.TrimSpace(strippedOutput)

	parsedFloat, err := strconv.ParseFloat(cleanedOutput, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
			"Args":  args,
		}).Debug("Error parsing float from cmd output")

		return 0, err
	}
	return parsedFloat, nil
}

// Take in an arbitrary number of string arguments, and pass them along to the yq binary, attempting
// to extract an int (e.g.; 4) from the command output
func getIntFromCommand(args ...string) (int64, error) {

	cmdOutput, err := runYqCommand(args...)

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
		return 0, nil
	}
	return parsedInt, nil
}
