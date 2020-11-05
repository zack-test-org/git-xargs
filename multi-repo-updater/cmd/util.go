package cmd

import (
	"github.com/google/go-github/v32/github"

	"github.com/sirupsen/logrus"
)

var (
	// AllOrgRepos is the slice containing every Github organization's repository considered by this tool when run (whether or not it was eligible for programmatic upgrade)
	AllOrgRepos []*github.Repository
	// OrgReposWithCircleCIConfig is the slice containing only those considered repos with files at the path .circleci/config.yml
	OrgReposWithCircleCIConfig []*github.Repository
	// OrgReposWithNoCircleCIConfig is the slice containing only those considered repos that DO NOT HAVE files at the path .circleci/config.yml
	OrgReposWithNoCircleCIConfig []*github.Repository
	// OrgReposWithCorrectContextAlreadySet is the slice containing only those considered repos that already had the correct context values according to this tool
	OrgReposWithCorrectContextAlreadySet []*github.Repository
	// CircleCIConfigPath is the default path in the Github repository where the Circle CI config file is expected to be found
	CircleCIConfigPath = ".circleci/config.yml"
)

// ConvertReposContexts first fetches all repositories for the given org
// Next, filters them down to only those repositories that actually have .circleci/config.yml files
// Finally, processes each of those repositories according to the logic defined in yaml.go, essentially adding the required Gruntwork Admin context to any Workflows -> Jobs -> Context arrays that don't already have it
func ConvertReposContexts(GithubClient *github.Client, GithubOrg string, allowedRepos []*AllowedRepo) {

	var reposToIterate []*github.Repository
	// Prefer repos passed in via file over the user-supplied command line flag for GithubOrg
	if len(allowedRepos) > 0 {

		log.Debug("Allowed repos were provided via file, preferring them over -o --github-org flag's value")

		repos, err := getFileDefinedRepos(GithubClient, allowedRepos)

		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":         err,
				"Allowed Repos": allowedRepos,
			}).Debug("error looking up filename provided repos")
		}

		reposToIterate = repos
	} else {

		repos, err := getReposByOrg(GithubClient, GithubOrg)
		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":        err,
				"Organization": GithubOrg,
			}).Debug("Failure looking up repos for organization")
			return
		}

		reposToIterate = repos
	}

	for _, repo := range reposToIterate {
		log.WithFields(logrus.Fields{
			"Repository": repo.GetName(),
		}).Debug("Considering repo for upgrade")
	}

	// Set aside the repos that do have a .circleci/config.yml file
	OrgReposWithCircleCIConfig = processReposWithCircleCIConfigs(GithubClient, reposToIterate)
}
