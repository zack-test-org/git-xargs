package cmd

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// This utility function accepts a path to the flatfile in which the user has defined their explicitly allowed repos
// It expects repos to be defined one per line in the following format: `gruntwork-io/cloud-nuke` with optional commas
// Stray single and double quotes are also handled and stripped out if they are encountered, and spacing is irrelevant
func processAllowedRepos(filepath string) ([]*AllowedRepo, error) {
	var allowedRepos []*AllowedRepo

	file, err := os.Open(filepath)

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error":    err,
			"Filepath": filepath,
		}).Debug("Could not open")

		return allowedRepos, err
	}

	// By wrapping the file.Close in a deferred anonymous function, we are able to avoid a nasty edge-case where
	// an actual closeErr would not be checked or handled properly in the more common `defer file.Close()`
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			log.WithFields(logrus.Fields{
				"Error": closeErr,
			}).Debug("Error closing allowed repos file")
		}
	}()

	// The regex for all common special characters to remove from the repo lines in the allowed repos file
	charRegex := regexp.MustCompile(`['",!]`)

	// Read through the file line by line, extracting the repo organization and name by splitting on the / char
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		trimmedLine := strings.TrimSpace(scanner.Text())
		cleanedLine := charRegex.ReplaceAllString(trimmedLine, "")
		orgAndRepoSlice := strings.Split(cleanedLine, "/")
		// Guard against stray lines, extra dangling single quotes, etc
		if len(orgAndRepoSlice) < 2 {
			continue
		}

		// Validate both the org and name are not empty
		parsedOrg := orgAndRepoSlice[0]
		parsedName := orgAndRepoSlice[1]

		// If both org name and repo name are present, create a new allowed repo and add it to the list
		if parsedOrg != "" && parsedName != "" {
			repo := &AllowedRepo{
				Organization: parsedOrg,
				Name:         parsedName,
			}
			allowedRepos = append(allowedRepos, repo)

		}

	}

	if err := scanner.Err(); err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error parsing line from allowed repos file")
	}

	return allowedRepos, nil
}
