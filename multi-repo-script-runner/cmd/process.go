package cmd

import (
	"sync"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// Loop through every repo we've selected and use a WaitGroup so that the processing can happen in parallel
func processRepos(dryRun bool, GithubClient *github.Client, repos []*github.Repository, scriptsCollection ScriptCollection, stats *RunStats) {
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func(dryRun bool, GithubClient *github.Client, repo *github.Repository, scriptsCollection ScriptCollection, stats *RunStats) {
			defer wg.Done()
			processRepo(dryRun, GithubClient, repo, scriptsCollection, stats)
		}(dryRun, GithubClient, repo, scriptsCollection, stats)
	}
	wg.Wait()
}

// 1. Attempt to clone it to the local filesystem. To avoid conflicts, this generates a new directory for each repo FOR EACH run, so heavy use of this tool may inflate your /tmp/ directory size
// 2. Look up the HEAD ref of the repo, and create a new branch from that ref, specific to this tool so that we can
// safely make our changes in the branch
// 3. Loop through all the supplied and validated scripts, executing them against the locally cloned repo in sequence
// 4. Look up any worktree changes (deleted files, modified files, new and untracked files) and ADD THEM ALL to the stage
// 5. Commit these changes with the optionally configurable git commit message, or fall back to the default if it was not provided by the user
// 6. Push the branch containing the new commit to the remote origin
// 7. Via the Github API, open a pull request of the newly pushed branch against the main branch of the repo
// 8. Track all successfully opened pull requests via the stats tracker so that we can print them out as part of our final
// run report that is displayed in table format to the operator following each run
func processRepo(dryRun bool, GithubClient *github.Client, repo *github.Repository, scriptsCollection ScriptCollection, stats *RunStats) {
	log.WithFields(logrus.Fields{
		"Repo": repo.GetName(),
	}).Debug("Attempting to clone repository using GITHUB_OAUTH_TOKEN")

	// Create a new temporary directory in the default temp directory of the system, but append
	// multi-repo-script-runner-<repo-name> to it so that it's easier to find when you're looking for it
	repositoryDir, localRepository, cloneErr := cloneLocalRepository(repo, stats)

	if cloneErr != nil {
		return
	}

	// Get HEAD ref from the repo
	ref, headRefErr := getLocalRepoHeadRef(localRepository, repo, stats)
	if headRefErr != nil {
		return
	}

	// Get the worktree for the given local repository so we can examine any changes made by script operations
	worktree, worktreeErr := getLocalWorkTree(repositoryDir, localRepository, repo)

	if worktreeErr != nil {
		// If we couldn't get the worktree for the local repo, skip on to the next one
		return
	}

	// Create a branch in the locally cloned copy of the repo to hold all the changes that may result from script execution
	branchName, branchErr := checkoutLocalBranch(ref, worktree, repo, localRepository, stats)
	if branchErr != nil {
		// If checking out a local branch failed, skip to the next repo
		return
	}

	// At this point, the repo has been successfully cloned, a fresh branch has been checked out, and it is ready to have the target scripts run against it
	runAllTargetedScripts(repositoryDir, scriptsCollection, repo, worktree, stats)

	// All scripts have now been run against the local clone of the repository in the tmp directory

	// Commit any untracked files, modified or deleted files that resulted from script execution
	commitLocalChanges(worktree, repo, localRepository, stats)

	// Push the local branch containing all of our changes from executing the target scripts
	pushLocalBranch(dryRun, repo, localRepository, stats)

	// Open a pull request on Github, of the recently pushed branch against master
	openPullRequest(dryRun, GithubClient, repo, branchName.String(), stats)
}
