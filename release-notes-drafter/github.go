package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/v31/github"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// getGithubClient returns an authenticated github client.
func getGithubClient(ctx context.Context) *github.Client {
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubApiKey},
	)
	oauth2Client := oauth2.NewClient(ctx, tokenSource)
	return github.NewClient(oauth2Client)
}

// makeRequest will make an authenticated GET request to the provided url string, and return the response body as a
// string (if there is no error). Use for retrieving text data that has no corresponding API endpoint (e.g diff url).
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

// updateReleaseDescription will update the given release with the provided description.
func updateReleaseDescription(
	logger *logrus.Entry,
	repo *github.Repository,
	release *github.RepositoryRelease,
	description string,
) error {
	logger.Infof("Updating release %s", release.GetURL())

	release.Body = github.String(description)

	ctx := context.Background()
	client := getGithubClient(ctx)
	_, _, err := client.Repositories.EditRelease(ctx, repo.GetOwner().GetLogin(), repo.GetName(), release.GetID(), release)
	if err != nil {
		logger.Errorf("Error while updating release %s: %s", release.GetURL(), err)
		return errors.WithStackTrace(err)
	}
	logger.Infof("Successfully finished updating release %s", release.GetURL())
	return nil
}

// createReleaseDraftWithClient will create a new empty release for the provided repo in the draft state. The tag
// defaults to a patch release.
func createReleaseDraftWithClient(
	ctx context.Context,
	logger *logrus.Entry,
	client *github.Client,
	repo *github.Repository,
	lastRelease *github.RepositoryRelease,
) (*github.RepositoryRelease, error) {
	logger.Infof("Creating new release in draft state for repo %s", repo.GetFullName())

	tagName, err := bumpPatchVersion(lastRelease)
	if err != nil {
		logger.Errorf(
			"Error while parsing release version (%s) to semantic version: %s", lastRelease.GetTagName(), err,
		)
		return nil, errors.WithStackTrace(err)
	}

	newRelease := github.RepositoryRelease{TagName: github.String(tagName), Draft: github.Bool(true)}
	release, _, err := client.Repositories.CreateRelease(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &newRelease)
	if err != nil {
		logger.Errorf("Error creating new draft release for repository %s: %s", repo.GetFullName(), err)
		return nil, errors.WithStackTrace(err)
	}
	logger.Infof("Finished creating new release for repo %s", repo.GetFullName())
	return release, nil
}

// getOrCreateReleaseDraft will return the latest release if it is in draft state. Otherwise, it will create a new
// release in draft state.
func getOrCreateReleaseDraft(logger *logrus.Entry, repo *github.Repository) (*github.RepositoryRelease, error) {
	logger.Infof("Retrieving release note draft for repository %s", repo.GetFullName())

	ctx := context.Background()
	client := getGithubClient(ctx)

	releases, _, err := client.Repositories.ListReleases(ctx, repo.GetOwner().GetLogin(), repo.GetName(), &github.ListOptions{})
	if err != nil {
		logger.Errorf("Error retrieving draft release for repository %s: %s", repo.GetFullName(), err)
		return nil, errors.WithStackTrace(err)
	}
	if len(releases) == 0 {
		logger.Infof("Found no releases for repository %s. Creating.", repo.GetFullName())
		return createReleaseDraftWithClient(ctx, logger, client, repo, nil)
	}
	if !releases[0].GetDraft() {
		logger.Infof("Latest release for repository %s is not in draft state. Creating.", repo.GetFullName())
		return createReleaseDraftWithClient(ctx, logger, client, repo, releases[0])
	}
	logger.Infof("Latest release for repository %s is in draft state.", repo.GetFullName())
	logger.Infof("Successfully retrieved release note draft for repository %s", repo.GetFullName())
	return releases[0], nil
}

// getPullRequestDiffSummary will retrieve the comparison of the commits in the PR against the base when it was created.
func getPullRequestDiffSummary(logger *logrus.Entry, pullRequest *github.PullRequest) (*github.CommitsComparison, error) {
	logger.Infof("Retrieving diff for pull request %s", pullRequest.GetHTMLURL())

	repo := pullRequest.GetBase().GetRepo()
	prHeadSha := pullRequest.GetHead().GetSHA()
	prBaseSha := pullRequest.GetBase().GetSHA()

	ctx := context.Background()
	client := getGithubClient(ctx)
	logger.Infof("Comparing state from commit %s (PR Base) to commit %s (PR Head)", prBaseSha, prHeadSha)
	comparison, _, err := client.Repositories.CompareCommits(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		prBaseSha,
		prHeadSha,
	)
	if err != nil {
		logger.Errorf(
			"Error retrieving comparison of state from commit %s (PR Base) to commit %s (PR Head): %s",
			prBaseSha,
			prHeadSha,
			err,
		)
		return nil, errors.WithStackTrace(err)
	}
	logger.Infof(
		"Successfully retrieved comparison of state from commit %s (PR Base) to commit %s (PR Head)",
		prBaseSha,
		prHeadSha,
	)
	return comparison, nil
}

// bumpPatchVersion will take the version string from the last release and return the semantic version with the patch
// version bumped.
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
	lastVersion.Patch++
	return fmt.Sprintf("v%s", lastVersion), nil
}
