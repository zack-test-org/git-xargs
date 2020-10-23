package cmd

import (
	"os"
	"os/exec"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

var (
	AllOrgRepos                          []*github.Repository
	OrgReposWithCircleCIConfig           []*github.Repository
	OrgReposWithNoCircleCIConfig         []*github.Repository
	OrgReposWithCorrectContextAlreadySet []*github.Repository
	CircleCIConfigPath                   = ".circleci/config.yml"
)

func dependencyInstalled(dep string) bool {
	_, err := exec.LookPath(dep)
	if err != nil {
		return false
	}
	return true
}

// A given third party binary that must be installed on the operator's system in order for them to use this tool
type Dependency struct {
	Name string
	URL  string
}

// Accepts a slice of dependencies, and FREAKS OUT if any of them are missing
func MustHaveDependenciesInstalled(deps []Dependency) {

	for _, d := range deps {

		if !dependencyInstalled(d.Name) {
			log.WithFields(logrus.Fields{
				"Dependency":         d.Name,
				"Install / info URL": d.URL,
			}).Debug("Missing dependency. Please install it before using this tool")
			os.Exit(1)
		}
	}
}

// First, fetches all repositories for the given org
// Next, filters them down to only those repositories that actually have .circleci/config.yml files
// Finally, processes each of those repositories according to the logic defined in yaml.go, essentially adding the required Gruntwork Admin context to any Workflows -> Jobs -> Context arrays that don't already have it
func ConvertReposContexts() {
	repos, err := getReposByOrg(GithubOrg)
	if err != nil {
		log.WithFields(logrus.Fields{
			"Error":        err,
			"Organization": GithubOrg,
		}).Debug("Failure looking up repos for organization")
		return
	}

	for _, repo := range repos {
		log.WithFields(logrus.Fields{
			"Organization": GithubOrg,
			"Repository":   repo.GetName(),
		}).Debug("Found repository")

		AllOrgRepos = append(AllOrgRepos, repo)
	}

	// Set aside the repos that do have a .circleci/config.yml file
	OrgReposWithCircleCIConfig = processReposWithCircleCIConfigs(AllOrgRepos)
}
