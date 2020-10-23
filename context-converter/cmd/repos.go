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
		return nil, errors.New("No repositories found!")
	}

	log.WithFields(logrus.Fields{
		"Repo count": repoCount,
	}).Debug(fmt.Sprintf("Fetched repos from Github organization: %s", GithubOrg))

	return allRepos, nil
}

func processReposWithCircleCIConfigs(repos []*github.Repository) []*github.Repository {
	opt := &github.RepositoryContentGetOptions{}

	var reposWithCircleCIConfigs []*github.Repository

	for _, repo := range repos {
		repositoryFile, _, _, err := GithubClient.Repositories.GetContents(context.Background(), GithubOrg, repo.GetName(), CircleCIConfigPath, opt)

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

		fmt.Println("*****************************************")
		fmt.Printf("PRE UPDATING YAML DOCUMENT %s\n", string(fileContents))
		fmt.Println("*****************************************")

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

			if updatedYAML == nil {

				log.WithFields(logrus.Fields{
					"Repo Name": repo.Name,
				}).Debug("YAML was NOT updated for repo")
				continue
			}

			fmt.Println("*****************************************")
			fmt.Printf("POST UPDATING YAML DOCUMENT: %s\n", string(updatedYAML))
			fmt.Println("*****************************************")
		}
	}
	return reposWithCircleCIConfigs
}
