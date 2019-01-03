package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/waigani/diffparser"
)

func getModulesAffected(pullRequest *github.PullRequest) ([]string, error) {
	response, err := makeRequest(pullRequest.GetDiffURL())
	if err != nil {
		return []string{}, err
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []string{}, err
	}
	diff, err := diffparser.Parse(string(bodyBytes))
	if err != nil {
		return []string{}, err
	}
	modulesAffected := []string{}
	for _, file := range diff.Files {
		maybeModule := getModuleString(file.OrigName)
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
	items := filepath.SplitList(path)
	return items[1]
}

func getDescription(pullRequest *github.PullRequest) string {
	return fmt.Sprintf("TODO: %s", pullRequest.GetTitle())
}

func getLink(pullRequest *github.PullRequest) string {
	return pullRequest.GetURL()
}
