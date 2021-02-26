package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// AllowedReposFile is the path to the file containing the names  of the repos that are safe for this tool to operate on, each on its own line
	AllowedReposFile string
	// Debug will dump the YAML pre and post processing to STDOUT for easier debugging, at the cost of extreme verbosity and terminal spew
	Debug bool
	// DryRun is a boolean flag - when set to true, only proposed YAML operations will be dumped to STDOUT - no branches, file changes or pull requests will be made
	DryRun bool
	// GithubOrg is the name of the organization that this tool will list repositories from
	GithubOrg string
	// TargetScripts represents the scripts to run on the given repo
	TargetScripts []string
	// CommitMessage will be used when committing any file changes to the branch
	CommitMessage string
	// The optional branch name the user can provide. Otherwise, this tool will default to its fallback of "multi-repo-script-runner"
	BranchName string
	// PullRequestTitle will be used when opening the PR - so name it generally after what you are accomplishing with this run
	PullRequestTitle string
	// PullRequestDescription will be used when opening the PR - so provide some context around the changes you will be making with this run
	PullRequestDescription string

	log = logrus.New()
)

func init() {
	// Log messages that are of DEBUG level
	log.SetLevel(logrus.DebugLevel)

	rootCmd.PersistentFlags().StringVarP(&GithubOrg, "github-org", "o", "", "The Github organization whose repos should be operated on")

	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dry-run", "d", false, "When dry-run is set to true, only proposed YAML updates will be output, but not changes in Github will be made (no branches will be created, no files updated, no PRs opened)")

	rootCmd.PersistentFlags().StringVarP(&AllowedReposFile, "allowed-repos-filepath", "a", "", "The path to the file containing repos this tool is allowed to operate on, each repo in format: gruntwork-io/terraform-aws-eks, one repo per line")

	rootCmd.PersistentFlags().StringSliceVarP(&TargetScripts, "scripts", "s", []string{}, "The scripts to run against the selected repos. These scripts must exist in the ./scripts directory and be executable.")

	rootCmd.PersistentFlags().StringVarP(&BranchName, "branch-name", "b", "multi-repo-script-runner", "The name of the branch you want created to hold your changes")

	rootCmd.PersistentFlags().StringVarP(&CommitMessage, "commit-message", "m", "Tis I, the multi-repo-script-runner!", "The commit message to use for any programmatic commits made by this tool")

	rootCmd.PersistentFlags().StringVarP(&PullRequestTitle, "pull-request-title", "t", "Multi Repo Updater Programmatic PR", "The title to add to the pull requests that will be opened by this run")

	rootCmd.PersistentFlags().StringVarP(&PullRequestDescription, "pull-request-description", "e", "This pull request was opened programmatically by the multi repo script runner CLI.", "The description to add to the pull requests that will be opened by this run")

	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the multi-repo-script-runner's version number",
	Long:  "For realzingtons, get the version number for this CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Gruntwork Multi Repo Script Runner v0.0.1")
	},
}

// Function that runs prior to execution of the main command. Useful for performing setup and verification tasks
// such as checking for required user inputs, env vars, dependencies, etc
func persistentPreRun(cmd *cobra.Command, args []string) {
	// Begin startup sanity checks on user provided input

	// Ensure the required dependencies are installed on the operator's system
	requiredDeps := []Dependency{
		{Name: "yq", URL: "https://mikefarah.gitbook.io/yq/"},
	}

	if ok, missingDeps := verifyDependenciesInstalled(requiredDeps); !ok {
		for _, d := range missingDeps {
			log.WithFields(logrus.Fields{
				"Dependency":         d.Name,
				"Install / info URL": d.URL,
			}).Debug("Missing dependency. Please install it before using this tool")
		}
		log.Fatal("All required dependencies must be installed prior to running this tool")
	}

	// If DryRun is enabled, notify user that no file changes will be made
	if DryRun {
		log.Debug("Dry run setting enabled. No actual file changes, branches or PRs will be created in Github")
	}

	// If user didn't provide either means of looking up repos, bail out with a helpful error
	if !ensureValidOptionsPassed(AllowedReposFile, GithubOrg) {
		log.Fatal("You must either provide an AllowedReposFile path or a GithubOrg. See ./multi-repo-script-runner help")

	}
}

var rootCmd = &cobra.Command{
	Use:              "multi-repo-script-runner",
	Short:            "Multi repo script runner CLI",
	Long:             "Multi repo script runner executes arbitrary bash scripts against any repos you select, handling all git operations that result and opening configurable pull requests",
	PersistentPreRun: persistentPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Multi repo script runner running...")

		// Verify the scripts that will be run against the repos and package them into a ScriptCollection
		scriptCollection, verifyErr := VerifyScripts(TargetScripts, "")

		if verifyErr != nil {
			log.WithFields(logrus.Fields{
				"Error": verifyErr,
			}).Fatal("Error verifying scripts passed via --scripts flag. Please fix scripts with issues and re-run")
		}

		// If no valid scripts were returned by the validation function, we have nothing to execute, so must exit with an error
		if len(scriptCollection.Scripts) == 0 {
			log.WithFields(logrus.Fields{
				"User provided scripts": TargetScripts,
			}).Fatal("No valid scripts found to execute. Ensure each script exists in the ./scripts directory, is executable, and was not misspelled when provided via the --scripts flag")
		}

		// Configure the client that will make Github API calls on our behalf, using the user-provided Github personal access token
		GithubClient := ConfigureGithubClient()

		// Configure a stats tracker that can be passed along to keep tallies of which repos fell into which categories, how many were modified, etc
		stats := NewStatsTracker()

		var fileProvidedRepos []*AllowedRepo

		// User provided a flatfile of repos to explicitly operate on, which we'll prefer over --github-org
		if AllowedReposFile != "" {
			// Call the allowed repos parsing function
			allowedRepos, err := processAllowedRepos(AllowedReposFile)
			if err != nil {
				log.WithFields(logrus.Fields{
					"Error":    err,
					"Filepath": AllowedReposFile,
				}).Debug("error processing allowed repos from file")
			}
			// fileProvidedRepos, when set, will be preferred by ConvertReposContexts over the user-passed in github-org flag
			fileProvidedRepos = allowedRepos

			// Update count of number of repos the the tool read in from the provided file
			stats.SetFileProvidedRepos(allowedRepos)

		}
		// Update repos to use the target context, where applicable
		OperateOnRepos(GithubClient, GithubOrg, fileProvidedRepos, scriptCollection, stats)

		// Once all processing is complete, print out the summary of what was done
		stats.PrintReport()
	},
}

// Execute is the main entrypoint to the cmd package. Its sole responsibility is to invoke the rootCmd's Execute method
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Debug(err)
		return
	}
}
