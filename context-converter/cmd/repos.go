package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v32/github"
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

func getMasterBranchGitRef(repo *github.Repository) *github.Reference {

	ref, _, err := GithubClient.Git.GetRef(context.Background(), GithubOrg, repo.GetName(), "heads/master")

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error retrieving head commit SHA")
	}

	return ref
}

func createProjectBranchIfNotExists(repo *github.Repository) {

	existingRef, getResponse, getErr := GithubClient.Git.GetRef(context.Background(), GithubOrg, repo.GetName(), RefsTargetBranch)

	if getErr != nil {
		log.WithFields(logrus.Fields{
			"Error":     getErr,
			"Branch":    TargetBranch,
			"Repo name": repo.GetName(),
		}).Debug("Error checking if project branch already exists")
	}

	if getResponse.StatusCode == 404 {
		log.WithFields(logrus.Fields{
			"Branch":    TargetBranch,
			"Repo name": repo.GetName(),
		}).Debug("Target branch was not found for repo - will attempt to create it")
	} else if existingRef != nil {
		log.WithFields(logrus.Fields{
			"Branch":    TargetBranch,
			"Repo name": repo.GetName(),
		}).Debug(fmt.Sprintf("Project branch already exists for repo - will not attempt to create it again"))
		return
	}

	masterGitRef := getMasterBranchGitRef(repo)

	// Update the ref's name with our new desired branch name, which will be POSTed via the Github API
	// to create a new branch by that name. Note, however, that the ref object still comes from master, so that its Ref.object.SHA will still point to master
	// This tells the Github API that we want to create a new branch with our provided name, with the HEAD of master's SHA as the starting point. In other words, branch off the HEAD of master.
	masterGitRef.Ref = &RefsTargetBranch

	_, _, err := GithubClient.Git.CreateRef(context.Background(), GithubOrg, repo.GetName(), masterGitRef)

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error creating new branch from master")
		return
	}

	log.WithFields(logrus.Fields{
		"Repo name": repo.GetName(),
	}).Debug(fmt.Sprintf("Created new branch %s off of master for repo %s", TargetBranch, repo.GetName()))
}

// Update the file via the Github API, on a special branch specific to this tool, which can then be PR'd against master
func updateFileOnBranch(repo *github.Repository, path string, sha *string, b []byte) {

	opt := &github.RepositoryContentFileOptions{
		Branch:  github.String(TargetBranch),
		Content: b,
		SHA:     sha,
		Message: github.String("Context converter programmatically repairing CircleCI config!"),
	}

	_, _, err := GithubClient.Repositories.UpdateFile(context.Background(), GithubOrg, repo.GetName(), path, opt)

	if err != nil {
		log.WithFields(logrus.Fields{
			"Err":    err,
			"Path":   path,
			"Branch": TargetBranch,
		}).Debug("Error updating file on branch")
	}
}

func openPullRequest(repo *github.Repository) {

	body := "This pull request was programmatically opened by the context-converter program. It should be adding the 'Gruntwork Admin' context to any Workflows -> Jobs nodes and should also be leaving the rest of the .circleci/config.yml file alone. \n\n This PR was opened so that all our repositories' .circleci/config.yml files can be converted to use the same CircleCI context, which will make rotating secrets much easier in the future."

	newPR := &github.NewPullRequest{
		Title:               github.String("Fix CircleCI Contexts"),
		Head:                github.String(TargetBranch),
		Base:                github.String("master"),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := GithubClient.PullRequests.Create(context.Background(), GithubOrg, repo.GetName(), newPR)

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
			"Head":  TargetBranch,
			"Base":  "master",
			"Body":  body,
		}).Debug("Error opening Pull request")
	} else {
		log.WithFields(logrus.Fields{
			"Pull request URL": *pr.URL,
		}).Debug("Successfully opened pull request")
	}
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
			updatedYAMLBytes := UpdateYamlDocument([]byte(fileContents))

			if updatedYAMLBytes == nil {

				log.WithFields(logrus.Fields{
					"Repo Name": repo.Name,
				}).Debug("YAML was NOT updated for repo")
				continue
			}

			fmt.Println("*****************************************")
			fmt.Printf("POST UPDATING YAML DOCUMENT: %s\n", string(updatedYAMLBytes))
			fmt.Println("*****************************************")

			log.Debug("Attempting to update file on branch")

			createProjectBranchIfNotExists(repo)
			updateFileOnBranch(repo, CircleCIConfigPath, repositoryFile.SHA, updatedYAMLBytes)
			openPullRequest(repo)

		}
	}

	return reposWithCircleCIConfigs
}
