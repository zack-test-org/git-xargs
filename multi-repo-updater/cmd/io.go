package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Handles input, output operations, tempfiles, etc

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

func dumpTempFileContents(tmpFile *os.File) {

	fileBytes, readErr := ioutil.ReadFile(tmpFile.Name())

	if readErr != nil {
		log.WithFields(logrus.Fields{
			"Error": readErr,
		}).Debug("Error reading temp file contents for debugging purposes")
	}

	fmt.Printf("%s\n", fileBytes)
}

func processAllowedRepos(filepath string) ([]*AllowedRepo, error) {
	file, err := os.Open(filepath)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error":    err,
			"Filepath": filepath,
		}).Debug("Could not open")
	}
	defer file.Close()

	var allowedRepos []*AllowedRepo

	// The regex for all common special characters to remove from the repo lines in the allowed repos file
	charRegex := regexp.MustCompile(`['",!]`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		trimmedLine := strings.TrimSpace(scanner.Text())
		cleanedLine := charRegex.ReplaceAllString(trimmedLine, "")
		orgAndRepoSlice := strings.Split(cleanedLine, "/")
		// TODO add validation here and make this less naive
		repo := &AllowedRepo{
			Organization: orgAndRepoSlice[0],
			Name:         orgAndRepoSlice[1],
		}
		allowedRepos = append(allowedRepos, repo)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return allowedRepos, nil
}
