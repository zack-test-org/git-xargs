package cmd

import (
	"github.com/google/go-github/v32/github"

	"github.com/sirupsen/logrus"
)

// There are two ways to select repos to operate on via this tool: 1. the --github-org flag, which accepts the Github
// organization name to look up all the repos for and to page through them programmatically and
// 2. the user-defined flatfile of repos in the format of 'gruntwork-io/cloud-nuke' with one repos defined one per line
// This function acts as a switch, depending upon whether or not the user provided an explicit list of repos to operate
// However, even though there are two methods for users to select repos, we still only want a single uniform interface
// for dealing with a repo throughout this tool, and that is the *github.Repository type provided by the go-github
// library. Therefore, this function serves the purpose of creating that uniform interface, by looking up flatfile-provided
// repos via go-github, so that we're only ever dealing with pointers to github.Repositories going forward
func OperateOnRepos(GithubClient *github.Client, GithubOrg string, allowedRepos []*AllowedRepo, scripts ScriptCollection, stats *RunStats) {

	var reposToIterate []*github.Repository
	// Prefer repos passed in via file over the user-supplied command line flag for GithubOrg
	if len(allowedRepos) > 0 {

		log.Debug("Allowed repos were provided via file, preferring them over -o --github-org flag's value")

		// Per the comment above, this helper method turns all the flatfile defined repos into pointers to the
		// github.Repository type provided by go-github
		repos, err := getFileDefinedRepos(GithubClient, allowedRepos, stats)
		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":         err,
				"Allowed Repos": allowedRepos,
			}).Debug("error looking up filename provided repos")
		}

		reposToIterate = repos
	} else {

		// In this code path, the user did not provide a flatfile, so we're just looking up all the Github
		// repos via their Organization name via the Github API
		repos, err := getReposByOrg(GithubClient, GithubOrg, stats)
		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":        err,
				"Organization": GithubOrg,
			}).Debug("Failure looking up repos for organization")
			return
		}

		reposToIterate = repos
	}

	// Track the repos selected for processing
	stats.TrackMultiple(ReposSelected, reposToIterate)

	for _, repo := range reposToIterate {
		log.WithFields(logrus.Fields{
			"Repository": repo.GetName(),
		}).Debug("Repo will have all targeted scripts run against it")
	}

	// Now that we've gathered up the repos we're going to operate on, do the actual processing by running the
	// user-defined scripts against each repo and handling the resulting git operations that follow
	processRepos(DryRun, GithubClient, reposToIterate, scripts, stats)
}
