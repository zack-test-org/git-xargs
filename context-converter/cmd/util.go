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
