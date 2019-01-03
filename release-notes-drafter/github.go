package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// getGithubClient returns an authenticated github client
func getGithubClient(ctx context.Context) *github.Client {
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubApiKey},
	)
	oauth2Client := oauth2.NewClient(ctx, tokenSource)
	return github.NewClient(oauth2Client)
}

func makeRequest(url string) (string, error) {
	ctx := context.Background()
	client := getGithubClient(ctx)
	request, err := client.NewRequest("GET", url, "")
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	buf := bytes.NewBufferString("")
	_, err = client.Do(ctx, request, buf)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	return buf.String(), nil
}

// updateReleaseDescription will update the given release with the provided description
func updateReleaseDescription(
	logger *logrus.Entry,
	repo *github.Repository,
	release *github.RepositoryRelease,
	description string,
) error {
	logger.Infof("Updating release %s", release.GetURL())
	defer logger.Infof("Finished updating release %s", release.GetURL())

	release.Body = github.String(description)

	ctx := context.Background()
	client := getGithubClient(ctx)
	_, _, err := client.Repositories.EditRelease(ctx, repo.GetOwner().GetLogin(), repo.GetName(), release.GetID(), release)
	return LogIfError(logger, fmt.Sprintf("Error while updating release %s", release.GetURL()), err)
}

// createReleaseDraftWithClient will create a new empty release for the provided repo in the draft state. The tag
// defaults to a patch release.
func createReleaseDraftWithClient(
	logger *logrus.Entry,
	ctx context.Context,
	client *github.Client,
	repo *github.Repository,
	lastRelease *github.RepositoryRelease,
) (*github.RepositoryRelease, error) {
	logger.Infof("Creating new release in draft state for repo %s", repo.GetFullName())
	defer logger.Infof("Finished creating new release for repo %s", repo.GetFullName())

	tagName, err := bumpPatchVersion(lastRelease)
	err = LogIfError(
		logger,
		fmt.Sprintf("Error while parsing release version (%s) to semantic version", lastRelease.GetTagName()),
		err,
	)
	if err != nil {
		return nil, err
	}

	newRelease := github.RepositoryRelease{TagName: github.String(tagName), Draft: github.Bool(true)}
	release, _, err := client.Repositories.CreateRelease(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &newRelease)
	err = LogIfError(logger, fmt.Sprintf("Error creating new draft release for repository %s", repo.GetFullName()), err)
	return release, err
}

// getOrCreateReleaseDraft will return the latest release if it is in draft state. Otherwise, it will create a new
// release in draft state.
func getOrCreateReleaseDraft(logger *logrus.Entry, repo *github.Repository) (*github.RepositoryRelease, error) {
	logger.Infof("Retrieving release note draft for repository %s", repo.GetFullName())
	defer logger.Infof("Finished retrieving release note draft for repository %s", repo.GetFullName())

	ctx := context.Background()
	client := getGithubClient(ctx)

	releases, _, err := client.Repositories.ListReleases(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &github.ListOptions{})
	err = LogIfError(logger, fmt.Sprintf("Error retrieving draft release for repository %s", repo.GetFullName()), err)
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		logger.Infof("Found no releases for repository %s. Creating.", repo.GetFullName())
		return createReleaseDraftWithClient(logger, ctx, client, repo, nil)
	}
	if !releases[0].GetDraft() {
		logger.Infof("Latest release for repository %s is not in draft state. Creating.", repo.GetFullName())
		return createReleaseDraftWithClient(logger, ctx, client, repo, releases[0])
	}
	logger.Infof("Latest release for repository %s is in draft state.", repo.GetFullName())
	return releases[0], nil
}

func bumpPatchVersion(lastRelease *github.RepositoryRelease) (string, error) {
	if lastRelease == nil {
		return "v0.0.1", nil
	}
	lastTagName := lastRelease.GetTagName()
	if strings.HasPrefix(lastTagName, "v") {
		lastTagName = lastTagName[1:]
	}
	lastVersion, err := semver.Make(lastTagName)
	if err != nil {
		return "", err
	}
	lastVersion.Patch += 1
	return fmt.Sprintf("v%s", lastVersion), nil
}
