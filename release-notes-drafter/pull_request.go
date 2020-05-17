package main

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v31/github"
)

// getModulesAffected will process the diff of the pull request and look for updates to modules, extracting them as
// strings. This assumes modules refer to folders in the repo under the modules directory in the root.
func getModulesAffected(pullRequest *github.PullRequest) ([]string, error) {
	logger := GetProjectLogger()
	comparison, err := getPullRequestDiffSummary(logger, pullRequest)
	if err != nil {
		return []string{}, err
	}
	return extractModulesAffectedFromDiff(comparison), nil
}

// extractModulesAffectedFromDiff takes the pull request diff summary and determine modules affected.. A module is
// affected if:
// - A new file was added.
// - An existing file was changed.
// - A file was removed.
// TODO: Specially render new files as "[NEW] module"
func extractModulesAffectedFromDiff(pullRequestDiff *github.CommitsComparison) []string {
	modulesAffected := []string{}
	for _, file := range pullRequestDiff.Files {
		var filePath string
		if file.GetStatus() == "deleted" {
			filePath = file.GetPreviousFilename()
		} else {
			filePath = file.GetFilename()
		}
		maybeModuleAffected := getModuleString(filePath)
		if maybeModuleAffected != "" {
			modulesAffected = append(modulesAffected, maybeModuleAffected)
		}
	}
	return modulesAffected
}

// getModuleString will extract the module name given the path to a file that changed, returning empty string if it is
// not a module file.
func getModuleString(path string) string {
	if !strings.HasPrefix(path, "modules") {
		return ""
	}
	items := strings.Split(path, "/")
	if len(items) <= 2 {
		return ""
	}
	return items[1]
}

// getDescription will take a pull request object and construct a place holder description string.
func getDescription(pullRequest *github.PullRequest) string {
	return fmt.Sprintf("TODO: %s", pullRequest.GetTitle())
}

// getContributor will take a pull request object and find the contributor to thank.
func getContributor(pullRequest *github.PullRequest) string {
	return pullRequest.GetUser().GetLogin()
}

// getLink will take a pull request object and return the URL to it.
func getLink(pullRequest *github.PullRequest) string {
	return pullRequest.GetHTMLURL()
}
