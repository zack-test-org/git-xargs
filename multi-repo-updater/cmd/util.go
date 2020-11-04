package cmd

import (
	"github.com/google/go-github/v32/github"

	"github.com/sirupsen/logrus"
)

var (
	AllOrgRepos                          []*github.Repository
	OrgReposWithCircleCIConfig           []*github.Repository
	OrgReposWithNoCircleCIConfig         []*github.Repository
	OrgReposWithCorrectContextAlreadySet []*github.Repository
	CircleCIConfigPath                   = ".circleci/config.yml"
)

// First, fetches all repositories for the given org
// Next, filters them down to only those repositories that actually have .circleci/config.yml files
// Finally, processes each of those repositories according to the logic defined in yaml.go, essentially adding the required Gruntwork Admin context to any Workflows -> Jobs -> Context arrays that don't already have it
func ConvertReposContexts(GithubClient *github.Client, GithubOrg string) {
	repos, err := getReposByOrg(GithubClient, GithubOrg)
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
	OrgReposWithCircleCIConfig = processReposWithCircleCIConfigs(GithubClient, GithubOrg, AllOrgRepos)
}
