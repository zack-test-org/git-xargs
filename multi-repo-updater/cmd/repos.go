package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

func getFileDefinedRepos(GithubClient *github.Client, allowedRepos []*AllowedRepo) ([]*github.Repository, error) {
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
		}

		if resp.StatusCode == 200 {
			log.WithFields(logrus.Fields{
				"Organization": allowedRepo.Organization,
				"Name":         allowedRepo.Name,
			}).Debug("Successfully fetched repo")
		}

		allRepos = append(allRepos, repo)
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

func getMasterBranchGitRef(GithubClient *github.Client, GithubOrg string, repo *github.Repository) (*github.Reference, error) {

	ref, _, err := GithubClient.Git.GetRef(context.Background(), GithubOrg, repo.GetName(), "heads/master")

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
		}).Debug("Error retrieving head commit SHA")
		return nil, err
	}

	return ref, nil
}

func createProjectBranchIfNotExists(DryRun bool, GithubClient *github.Client, GithubOrg string, repo *github.Repository, stats *RunStats) error {

	if DryRun {
		log.WithFields(logrus.Fields{
			"Repo": repo.GetName(),
		}).Debug("DryRun is set to true, so skipping creation of new branch!")
		// Keep track of the repos that were not affected by any changes because dry-run was set to true
		stats.TrackSingle(DryRunSet, repo)

		return nil
	}

	existingRef, getResponse, getErr := GithubClient.Git.GetRef(context.Background(), GithubOrg, repo.GetName(), RefsTargetBranch)

	if getErr != nil && getResponse.StatusCode == 404 {
		log.WithFields(logrus.Fields{
			"Branch":    TargetBranch,
			"Repo name": repo.GetName(),
		}).Debug("Target branch was not found for repo - will attempt to create it")

		stats.TrackSingle(TargetBranchNotFound, repo)

	} else if existingRef != nil {
		log.WithFields(logrus.Fields{
			"Branch":    TargetBranch,
			"Repo name": repo.GetName(),
		}).Debug(fmt.Sprintf("Project branch already exists for repo - will not attempt to create it again"))

		stats.TrackSingle(TargetBranchAlreadyExists, repo)

		return nil
	} else if getErr != nil {
		log.WithFields(logrus.Fields{
			"Error":     getErr,
			"Branch":    TargetBranch,
			"Repo name": repo.GetName(),
		}).Debug("Error checking if project branch already exists")

		stats.TrackSingle(TargetBranchLookupErr, repo)
	}

	masterGitRef, err := getMasterBranchGitRef(GithubClient, GithubOrg, repo)

	if err != nil {
		log.Debug("Error retrieving git ref for master branch - can't create branch")
		return err
	}

	// Update the ref's name with our new desired branch name, which will be POSTed via the Github API
	// to create a new branch by that name. Note, however, that the ref object still comes from master, so that its Ref.object.SHA will still point to master
	// This tells the Github API that we want to create a new branch with our provided name, with the HEAD of master's SHA as the starting point. In other words, branch off the HEAD of master.
	masterGitRef.Ref = &RefsTargetBranch

	_, _, createRefErr := GithubClient.Git.CreateRef(context.Background(), GithubOrg, repo.GetName(), masterGitRef)

	if createRefErr != nil {
		log.WithFields(logrus.Fields{
			"Error": createRefErr,
		}).Debug("Error creating new branch from master")
		return createRefErr
	}

	log.WithFields(logrus.Fields{
		"Repo name": repo.GetName(),
	}).Debug(fmt.Sprintf("Created new branch %s off of master for repo %s", TargetBranch, repo.GetName()))

	stats.TrackSingle(TargetBranchSuccessfullyCreated, repo)

	return nil
}

// Update the file via the Github API, on a special branch specific to this tool, which can then be PR'd against master
func updateFileOnBranch(DryRun bool, GithubClient *github.Client, GithubOrg string, repo *github.Repository, path string, sha *string, fileContents []byte, stats *RunStats) {

	if DryRun {
		log.WithFields(logrus.Fields{
			"Repo": repo.GetName(),
		}).Debug("DryRun is set to true, so skipping file update!")
		return
	}

	opt := &github.RepositoryContentFileOptions{
		Branch:  github.String(TargetBranch),
		Content: fileContents,
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

		stats.TrackSingle(PullRequestOpenErr, repo)
	}
}

func openPullRequest(DryRun bool, GithubClient *github.Client, GithubOrg string, repo *github.Repository, stats *RunStats) {

	if DryRun {
		log.WithFields(logrus.Fields{
			"Repo": repo.GetName(),
		}).Debug("DryRun is set to true, so skipping opening a pull request!")
		return
	}

	body := "This pull request was programmatically opened by the multi-repo-updater program. It should be adding the 'Gruntwork Admin' context to any Workflows -> Jobs nodes and should also be leaving the rest of the .circleci/config.yml file alone. \n\n This PR was opened so that all our repositories' .circleci/config.yml files can be converted to use the same CircleCI context, which will make rotating secrets much easier in the future."

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
			"Pull request URL": *pr.HTMLURL,
		}).Debug("Successfully opened pull request")
	}
}

// Loop through every passed in repository and look up the file contents of the config.yml file via Github API
// Filter down from ALL repos to only those repos containing files at the expected path: .circleci/config.yml
// Then, process only those repos that do contain config files, updating their YAML by first writing the file contents available at the HEAD of the main branch to a tempfile. The tempfile is then further processed via commands shelled out to the `yq` binary to modify the tempfile in place, and the final results of the tempfile are read out again before being PUT via the Github API to a special project branch
func processReposWithCircleCIConfigs(GithubClient *github.Client, repos []*github.Repository, stats *RunStats) {

	opt := &github.RepositoryContentGetOptions{}

	for _, repo := range repos {

		repositoryFile, _, _, err := GithubClient.Repositories.GetContents(context.Background(), *repo.GetOwner().Login, repo.GetName(), CircleCIConfigPath, opt)

		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":    err,
				"Owner":    repo.GetOwner().GetName(),
				"Repo":     repo.GetName(),
				"Filepath": CircleCIConfigPath,
			}).Debug("Error fetching file content! Repository does not have a CircleCI config file")

			// Add repo to the set of those missing Circle CI configs
			stats.TrackSingle(ConfigNotFound, repo)

			continue
		}

		// By this point, we're operating on a repository that contains a .circleci/config.yml file
		fileContents, fileGetContentsErr := repositoryFile.GetContent()

		if Debug {
			fmt.Println("*****************************************")
			fmt.Printf("PRE UPDATING YAML DOCUMENT %s\n", strings.ToUpper(*repo.Name))
			fmt.Println("*****************************************")
			fmt.Printf("%s\n", string(fileContents))
		}

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

			stats.TrackSingle(ConfigNotFound, repo)
		} else {

			stats.TrackSingle(ConfigFound, repo)

			// Process .circleci/config.yml file, updating context nodes as necessary
			updatedYAMLBytes := UpdateYamlDocument([]byte(fileContents), Debug)

			if updatedYAMLBytes == nil {

				log.WithFields(logrus.Fields{
					"Repo Name": *repo.Name,
				}).Debug("YAML was NOT updated for repo")

				// Track this repo as not having its config file updated
				stats.TrackSingle(YamlNotUpdated, repo)

				continue
			} else if !DryRun {
				stats.TrackSingle(YamlUpdated, repo)
			}

			if Debug {
				fmt.Println("*****************************************")
				fmt.Printf("POST UPDATING YAML DOCUMENT %s\n", strings.ToUpper(*repo.Name))
				fmt.Println("*****************************************")
				fmt.Printf("%s\n", string(updatedYAMLBytes))
			}

			createBranchErr := createProjectBranchIfNotExists(DryRun, GithubClient, GithubOrg, repo, stats)

			// If createBranchErr is not equal to nil, then that means we both a). needed to create a branch, because it didn't already exist and b). failed to do so, so we can't proceed with file updates or PR
			if createBranchErr == nil {
				updateFileOnBranch(DryRun, GithubClient, GithubOrg, repo, CircleCIConfigPath, repositoryFile.SHA, updatedYAMLBytes, stats)
				openPullRequest(DryRun, GithubClient, GithubOrg, repo, stats)

			}
		}
	}
}
