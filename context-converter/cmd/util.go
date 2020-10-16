package cmd

import (
	"context"
	"errors"
	"fmt"

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

func getReposByOrg(org string) ([]*github.Repository, error) {

	opts := &github.RepositoryListByOrgOptions{}

	repos, _, err := GithubClient.Repositories.ListByOrg(context.Background(), org, opts)

	if len(repos) == 0 {
		return nil, errors.New("No repositories found!")
	}
	return repos, err
}

func filterReposWithCircleCIConfig(repos []*github.Repository) []*github.Repository {
	rgco := &github.RepositoryContentGetOptions{}

	for _, repo := range repos {
		repositoryFile, _, _, err := GithubClient.Repositories.GetContents(context.Background(), GithubOrg, repo.GetName(), CircleCIConfigPath, rgco)

		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":    err,
				"Owner":    repo.GetOwner().GetName(),
				"Repo":     repo.GetName(),
				"Filepath": CircleCIConfigPath,
			}).Debug("Error fetching file content! Repository does not have a CircleCI config file")

			OrgReposWithNoCircleCIConfig = append(OrgReposWithNoCircleCIConfig, repo)

			continue

		}

		fileContents, fileGetContentsErr := repositoryFile.GetContent()

		if fileGetContentsErr != nil {
			log.WithFields(logrus.Fields{
				"Error": fileGetContentsErr,
				"Path":  CircleCIConfigPath,
			}).Debug("Error reading file contents!")
		}

		// If the file contents is an empty string, that means there is no config file at the expected path
		if fileContents == "" {
			log.WithFields(logrus.Fields{
				"Repo": repo.GetName(),
			}).Debug("Repository does not have CircleCI config file")

			OrgReposWithNoCircleCIConfig = append(OrgReposWithNoCircleCIConfig, repo)
		} else {

			fmt.Printf("GOT FILE CONTENTS: %s\n", fileContents)
			OrgReposWithCircleCIConfig = append(OrgReposWithCircleCIConfig, repo)
		}
	}
	return repos
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
	OrgReposWithCircleCIConfig = filterReposWithCircleCIConfig(AllOrgRepos)
}
