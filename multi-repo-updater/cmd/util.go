package cmd

import (
	"github.com/google/go-github/v32/github"

	"github.com/sirupsen/logrus"
)

var (
	// CircleCIConfigPath is the default filepath at which we expect the Circle CI config file
	CircleCIConfigPath = ".circleci/config.yml"
)

// ConvertReposContexts first fetches all repositories for the given org
// Next, filters them down to only those repositories that actually have .circleci/config.yml files
// Finally, processes each of those repositories according to the logic defined in yaml.go, essentially adding the required Gruntwork Admin context to any Workflows -> Jobs -> Context arrays that don't already have it
func ConvertReposContexts(GithubClient *github.Client, GithubOrg string, allowedRepos []*AllowedRepo, stats *RunStats) {

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

	processReposWithCircleCIConfigs(GithubClient, reposToIterate, stats)

	// Print out the final report of what was done during this run
}
