package main

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/waigani/diffparser"
)

func getModulesAffected(pullRequest *github.PullRequest) ([]string, error) {
	bodyString, err := makeRequest(pullRequest.GetDiffURL())
	if err != nil {
		return []string{}, err
	}
	return extractModulesAffectedFromDiff(bodyString)
}

// TODO: Specially render new files
func extractModulesAffectedFromDiff(diffString string) ([]string, error) {
	diff, err := diffparser.Parse(diffString)
	if err != nil {
		return []string{}, errors.WithStackTrace(err)
	}
	modulesAffected := []string{}
	for _, file := range diff.Files {
		fileNameToUse := file.OrigName
		if fileNameToUse == "" {
			fileNameToUse = file.NewName
		}
		maybeModule := getModuleString(fileNameToUse)
		if maybeModule != "" {
			modulesAffected = append(modulesAffected, maybeModule)
		}
	}
	return modulesAffected, nil
}

func getModuleString(path string) string {
	if !strings.HasPrefix(path, "modules") {
		return ""
	}
	items := strings.Split(path, "/")
	return items[1]
}

func getDescription(pullRequest *github.PullRequest) string {
	return fmt.Sprintf("TODO: %s", pullRequest.GetTitle())
}

func getLink(pullRequest *github.PullRequest) string {
	return pullRequest.GetHTMLURL()
}
