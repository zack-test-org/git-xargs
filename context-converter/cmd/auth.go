package cmd

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func ConfigureGithubClient() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubOauthToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	GithubClient = client
	log.Debug("Github client instantiated!")

}
