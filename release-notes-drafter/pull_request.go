package main

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"github.com/gruntwork-io/gruntwork-cli/collections"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/waigani/diffparser"
)

// getModulesAffected will process the diff of the pull request and look for updates to modules, extracting them as
// strings. This assumes modules refer to folders in the repo under the modules directory in the root.
func getModulesAffected(pullRequest *github.PullRequest) ([]string, error) {
	bodyString, err := makeRequest(pullRequest.GetDiffURL())
	if err != nil {
		return []string{}, err
	}
	return extractModulesAffectedFromDiff(bodyString)
}

// extractModulesAffectedFromDiff takes a git diff as a string, parses it and extracts the modules that were affected as
// strings. A module is affected if:
// - A new file was added.
// - An existing file was changed.
// - A file was removed.
// TODO: Specially render new files as "[NEW] module"
func extractModulesAffectedFromDiff(diffString string) ([]string, error) {
	diff, err := diffparser.Parse(diffString)
	if err != nil {
		return []string{}, errors.WithStackTrace(err)
	}
	modulesAffected := map[string]string{}
	for _, file := range diff.Files {
		fileNameToUse := file.OrigName
		if fileNameToUse == "" {
			fileNameToUse = file.NewName
		}
		maybeModule := getModuleString(fileNameToUse)
		if maybeModule != "" {
			// NOTE: The value doesn't matter, so we use empty string. Ideally we will use bool (and thus true), but
			// collections.Keys only works with map[string]string.
			modulesAffected[maybeModule] = ""
		}
	}
	return collections.Keys(modulesAffected), nil
}

// getModuleString will extract the module name given the path to a file that changed, returning empty string if it is
// not a module file.
func getModuleString(path string) string {
	if !strings.HasPrefix(path, "modules") {
		return ""
	}
	items := strings.Split(path, "/")
	return items[1]
}

// getDescription will take a pull request object and construct a place holder description string.
func getDescription(pullRequest *github.PullRequest) string {
	return fmt.Sprintf("TODO: %s", pullRequest.GetTitle())
}

// getLink will take a pull request object and return the URL to it.
func getLink(pullRequest *github.PullRequest) string {
	return pullRequest.GetHTMLURL()
}
