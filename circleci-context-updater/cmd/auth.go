package cmd

import (
	"context"
	"os"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"

	"golang.org/x/oauth2"
)

// ConfigureGithubClient creates a Github API client using the user-supplied GITHUB_OAUTH_TOKEN and return the configured Github client
func ConfigureGithubClient() *github.Client {
	// Ensure user provided a GITHUB_OAUTH_TOKEN
	GithubOauthToken := os.Getenv("GITHUB_OAUTH_TOKEN")
	if GithubOauthToken == "" {
		log.WithFields(logrus.Fields{
			"Error": "You must set a Github personal access token with access to Gruntwork repos via the Env var GITHUB_OAUTH_TOKEN",
		}).Debug("Missing GITHUB_OAUTH_TOKEN")
		os.Exit(1)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubOauthToken},
	)

	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	log.Debug("Github client instantiated!")

	return client
}
