package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

// Get all the repos for a given Github organization
func getReposByOrg(org string) ([]*github.Repository, error) {

	opts := &github.RepositoryListByOrgOptions{}

	repos, _, err := GithubClient.Repositories.ListByOrg(context.Background(), org, opts)
	if len(repos) == 0 {
		return nil, errors.New("No repositories found!")
	}
	return repos, err
}

func processReposWithCircleCIConfigs(repos []*github.Repository) []*github.Repository {
	rgco := &github.RepositoryContentGetOptions{}

	var reposWithCircleCIConfigs []*github.Repository

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

		fmt.Printf("PRE UPDATING YAML DOCUMENT %s\n", string(fileContents))

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

			reposWithCircleCIConfigs = append(reposWithCircleCIConfigs, repo)

			// Process .circleci/config.yml file, updating context nodes as necessary
			updatedYAML := UpdateYamlDocument([]byte(fileContents))

			fmt.Printf("POST UPDATING YAML DOCUMENT: %s\n", string(updatedYAML))
		}
	}
	return reposWithCircleCIConfigs
}
