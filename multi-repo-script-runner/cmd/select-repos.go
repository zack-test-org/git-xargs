package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

func getFileDefinedRepos(GithubClient *github.Client, allowedRepos []*AllowedRepo, stats *RunStats) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	for _, allowedRepo := range allowedRepos {

		log.WithFields(logrus.Fields{
			"Organization": allowedRepo.Organization,
			"Name":         allowedRepo.Name,
		}).Debug("Looking up filename provided repo")

		repo, resp, err := GithubClient.Repositories.Get(context.Background(), allowedRepo.Organization, allowedRepo.Name)

		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":                err,
				"Response Status Code": resp.StatusCode,
				"AllowedRepoOwner":     allowedRepo.Organization,
				"AllowedRepoName":      allowedRepo.Name,
			}).Debug("error getting single repo")

			if resp.StatusCode == 404 {
				// This repo does not exist / could not be fetched as named, so we won't include it in the list of repos to process

				// create an empty github repo object to satisfy the stats tracking interface
				missingRepo := &github.Repository{
					Owner: &github.User{Login: github.String(allowedRepo.Organization)},
					Name:  github.String(allowedRepo.Name),
				}
				stats.TrackSingle(RepoNotExists, missingRepo)
				continue
			}
		}

		if resp.StatusCode == 200 {
			log.WithFields(logrus.Fields{
				"Organization": allowedRepo.Organization,
				"Name":         allowedRepo.Name,
			}).Debug("Successfully fetched repo")

			allRepos = append(allRepos, repo)
		}
	}
	return allRepos, nil
}

// Get all the repos for a given Github organization
func getReposByOrg(GithubClient *github.Client, GithubOrg string, stats *RunStats) ([]*github.Repository, error) {
	// Page through all of the organization's repos, collecting them in this slice
	var allRepos []*github.Repository

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := GithubClient.Repositories.ListByOrg(context.Background(), GithubOrg, opt)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	repoCount := len(allRepos)

	if repoCount == 0 {
		return nil, errors.New("no repositories found")
	}

	log.WithFields(logrus.Fields{
		"Repo count": repoCount,
	}).Debug(fmt.Sprintf("Fetched repos from Github organization: %s", GithubOrg))

	stats.TrackMultiple(FetchedViaGithubAPI, allRepos)

	return allRepos, nil
}
