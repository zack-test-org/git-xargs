package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// cloneLocalRepository clones a remote Github repo via SSH to a local temporary directory so that scripts can be run
// against the repo locally and any git changes handled thereafter. The local directory has
// multi-repo-script-runner-<repo-name> appended to it to make it easier to find when you are looking for it while debugging
func cloneLocalRepository(repo *github.Repository, stats *RunStats) (string, *git.Repository, error) {
	log.WithFields(logrus.Fields{
		"Repo": repo.GetName(),
	}).Debug("Attempting to clone repository using GITHUB_OAUTH_TOKEN")

	repositoryDir, tmpDirErr := ioutil.TempDir("", fmt.Sprintf("multi-repo-script-runner-%s", repo.GetName()))
	if tmpDirErr != nil {
		log.WithFields(logrus.Fields{
			"Error": tmpDirErr,
			"Repo":  repo.GetName(),
		}).Debug("Failed to create temporary directory to hold repo")
		return repositoryDir, nil, tmpDirErr
	}

	localRepository, err := git.PlainClone(repositoryDir, false, &git.CloneOptions{
		URL:      repo.GetCloneURL(),
		Progress: os.Stdout,
		Auth: &http.BasicAuth{
			Username: repo.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
	})

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
			"Repo":  repo.GetName(),
		}).Debug("Error cloning repository")

		// Track failure to clone for our final run report
		stats.TrackSingle(RepoFailedToClone, repo)

		return repositoryDir, nil, err
	}

	stats.TrackSingle(RepoSuccessfullyCloned, repo)

	return repositoryDir, localRepository, nil
}

// getLocalRepoHeadRef looks up the HEAD reference of the locally cloned git repository, which is required by
// downstream operations such as branching
func getLocalRepoHeadRef(localRepository *git.Repository, repo *github.Repository, stats *RunStats) (*plumbing.Reference, error) {
	ref, headErr := localRepository.Head()
	if headErr != nil {
		log.WithFields(logrus.Fields{
			"Error": headErr,
			"Repo":  repo.GetName(),
		}).Debug("Error getting HEAD ref from local repo")

		stats.TrackSingle(GetHeadRefFailed, repo)

		return nil, headErr
	}
	return ref, nil
}

// runAllTargetedScripts loops through the collection of verified scripts and runs each against the currently targeted
// locally cloned repository, tracking any exceptions that may be thrown during execution
func runAllTargetedScripts(repositoryDir string, scriptsCollection ScriptCollection, repo *github.Repository, worktree *git.Worktree, stats *RunStats) error {
	for _, script := range scriptsCollection.Scripts {
		cmd := exec.Command(script.Path)
		cmd.Dir = repositoryDir

		log.WithFields(logrus.Fields{
			"Repo":      repo.GetName(),
			"Directory": repositoryDir,
			"Script":    script,
		}).Debug("Executing script against local clone of repo...")

		stdoutStdErr, err := cmd.CombinedOutput()

		if err != nil {
			log.WithFields(logrus.Fields{
				"Error": err,
			}).Debug("Error getting output of script execution")
			// Track the script error against the repo
			stats.TrackSingle(ScriptErrorOcurredDuringExecution, repo)
			return err
		}

		log.WithFields(logrus.Fields{
			"CombinedOutput": string(stdoutStdErr),
		}).Debug("Received output of script run")

		status, statusErr := worktree.Status()

		if statusErr != nil {
			log.WithFields(logrus.Fields{
				"Error": statusErr,
				"Repo":  repo.GetName(),
				"Dir":   repositoryDir,
			}).Debug("Error looking up worktree status")

			// Track the status check failure
			stats.TrackSingle(WorktreeStatusCheckFailed, repo)
			return statusErr
		}

		// If our scripts made any file changes, we need to stage, add and commit them
		if !status.IsClean() {
			log.WithFields(logrus.Fields{
				"Repo": repo.GetName(),
			}).Debug("Local repository worktree no longer clean, will stage and add new files and commit changes")

			// Track the fact that worktree changes were made following execution
			stats.TrackSingle(WorktreeStatusDirty, repo)

			for filepath := range status {
				if status.IsUntracked(filepath) {
					fmt.Printf("Found untracked file %s. Adding to stage", filepath)
					_, addErr := worktree.Add(filepath)
					if addErr != nil {
						log.WithFields(logrus.Fields{
							"Error":    addErr,
							"Filepath": filepath,
						}).Debug("Error adding file to git stage")
						// Track the file staging failure
						stats.TrackSingle(WorktreeAddFileFailed, repo)
						return addErr
					}
				}
			}

		} else {
			log.WithFields(logrus.Fields{
				"Repo": repo.GetName(),
			}).Debug("Local repository status is clean - nothing to stage or commit")

			// Track the fact that repo had no file changes post script execution
			stats.TrackSingle(WorktreeStatusClean, repo)
		}
	}

	return nil
}

// getLocalWorkTree looks up the working tree of the locally cloned repository and returns it if possible, or an error
func getLocalWorkTree(repositoryDir string, localRepository *git.Repository, repo *github.Repository) (*git.Worktree, error) {
	worktree, worktreeErr := localRepository.Worktree()

	if worktreeErr != nil {
		log.WithFields(logrus.Fields{
			"Error": worktreeErr,
			"Repo":  repo.GetName(),
			"Dir":   repositoryDir,
		}).Debug("Error looking up local repository's worktree")

		return nil, worktreeErr
	}
	return worktree, nil
}

// checkoutLocalBranch creates a local branch specific to this tool in the locally checked out copy of the repo in the /tmp folder
func checkoutLocalBranch(ref *plumbing.Reference, worktree *git.Worktree, remoteRepository *github.Repository, localRepository *git.Repository, stats *RunStats) (plumbing.ReferenceName, error) {
	// BranchName is a global variable that is set in cmd/root.go. It is override-able by the operator via the --branch-name or -b flag. It defaults to "multi-repo-script-runner"
	branchName := plumbing.NewBranchReferenceName(BranchName)
	log.WithFields(logrus.Fields{
		"Branch Name": branchName,
		"Repo":        remoteRepository.GetName(),
	}).Debug("Created branch")

	// Create a branch specific to the multi repo script runner
	co := &git.CheckoutOptions{
		Hash:   ref.Hash(),
		Branch: branchName,
		Create: true,
	}

	// Attempt to checkout the new tool-specific branch on which all scripts will be executed
	checkoutErr := worktree.Checkout(co)

	if checkoutErr != nil {
		log.WithFields(logrus.Fields{
			"Error": checkoutErr,
			"Repo":  remoteRepository.GetName(),
		}).Debug("Error creating new branch")

		// Track the error checking out the branch
		stats.TrackSingle(BranchCheckoutFailed, remoteRepository)

		return branchName, checkoutErr
	}

	return branchName, nil
}

// commitLocalChanges will create a commit using the supplied or default commit message and will add any untracked, deleted
// or modified files that resulted from script execution
func commitLocalChanges(worktree *git.Worktree, remoteRepository *github.Repository, localRepository *git.Repository, stats *RunStats) error {

	// With all our untracked files staged, we can now create a commit, passing the All
	// option when configuring our commit option so that all modified and deleted files
	// will have their changes committed
	commitOps := &git.CommitOptions{
		All: true,
	}

	_, commitErr := worktree.Commit(CommitMessage, commitOps)

	if commitErr != nil {
		log.WithFields(logrus.Fields{
			"Error": commitErr,
			"Repo":  remoteRepository.GetName(),
		})

		// If we reach this point, we were unable to commit our changes, so we'll
		// continue rather than attempt to push an empty branch and open an empty PR
		stats.TrackSingle(CommitChangesFailed, remoteRepository)
		return commitErr
	}
	return nil
}

// pushLocalBranch pushes the branch in the local clone of the /tmp/ directory repository to the Github remote origin
// so that a pull request can be opened against it via the Github API
func pushLocalBranch(dryRun bool, remoteRepository *github.Repository, localRepository *git.Repository, stats *RunStats) error {
	if dryRun {

		log.WithFields(logrus.Fields{
			"Repo": remoteRepository.GetName(),
		}).Debug("Skipping branch push to remote origin because --dry-run flag is set")

		stats.TrackSingle(PushBranchSkipped, remoteRepository)

		return nil
	}
	// Push the changes to the remote repo
	po := &git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: remoteRepository.GetOwner().GetLogin(),
			Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
		},
	}
	pushErr := localRepository.Push(po)

	if pushErr != nil {
		log.WithFields(logrus.Fields{
			"Error": pushErr,
			"Repo":  remoteRepository.GetName(),
		}).Debug("Error pushing new branch to remote origin")

		// Track the push failure
		stats.TrackSingle(PushBranchFailed, remoteRepository)
		return pushErr
	}

	log.WithFields(logrus.Fields{
		"Repo": remoteRepository.GetName(),
	}).Debug("Successfully pushed local branch to remote origin")

	return nil
}

// Attempt to open a pull request via the Github API, of the supplied branch specific to this tool, against the main
// branch for the remote origin
func openPullRequest(dryRun bool, githubClient *github.Client, repo *github.Repository, branch string, stats *RunStats) error {

	if dryRun {
		log.WithFields(logrus.Fields{
			"Repo": repo.GetName(),
		}).Debug("dryRun is set to true, so skipping opening a pull request!")
		return nil
	}

	// Configure pull request options that the Github client accepts when making calls to open new pull requests
	newPR := &github.NewPullRequest{
		Title:               github.String(PullRequestTitle),
		Head:                github.String(branch),
		Base:                github.String("master"),
		Body:                github.String(PullRequestDescription),
		MaintainerCanModify: github.Bool(true),
	}

	// Make a pull request via the Github API
	pr, _, err := githubClient.PullRequests.Create(context.Background(), *repo.GetOwner().Login, repo.GetName(), newPR)

	if err != nil {
		log.WithFields(logrus.Fields{
			"Error": err,
			"Head":  branch,
			"Base":  "master",
			"Body":  PullRequestDescription,
		}).Debug("Error opening Pull request")

		// Track pull request open failure
		stats.TrackSingle(PullRequestOpenErr, repo)

		return err
	}

	log.WithFields(logrus.Fields{
		"Pull Request URL": pr.GetHTMLURL(),
	}).Debug("Successfully opened pull request")

	// Track successful opening of the pull request, extracting the HTML url to the PR itself for easier review
	stats.TrackPullRequest(repo.GetName(), pr.GetHTMLURL())
	return nil
}
