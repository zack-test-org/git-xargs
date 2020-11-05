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
	// TargetContext is the name of the CircleCI context that we want added to the context arrays of the workflow jobs
	TargetContext string
	// TargetBranch is the default name this tool will use when creating new branches with code changes in Github
	TargetBranch = "IAC-1616-programmatically-fix-context"
	// RefsTargetBranch is the name of the branch with the "heads/" prefix, as required by some Github API calls
	RefsTargetBranch = fmt.Sprintf("heads/%s", TargetBranch)
	log              = logrus.New()
)

func init() {
	// Log messages that are of DEBUG level
	log.SetLevel(logrus.DebugLevel)

	rootCmd.PersistentFlags().StringVarP(&GithubOrg, "github-org", "o", "gruntwork-io", "The Github organization whose repos should be operated on")
	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dry-run", "d", false, "When dry-run is set to true, only proposed YAML updates will be output, but not changes in Github will be made (no branches will be created, no files updated, no PRs opened)")

	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "x", false, "When debug is set to true, the YAML file contents for each considered repo will be written to STDOUT both PRE and POST processing for easier debugging")

	rootCmd.PersistentFlags().StringVarP(&AllowedReposFile, "", "a", "allowed-repos-filepath", "The path to the file containing repos this tool is allowed to operate on, each repo in format: gruntwork-io/terraform-aws-eks, one repo per line")

}

// Function that runs prior to execution of the main command. Useful for performing setup and verification tasks
// such as checking for required user inputs, env vars, dependencies, etc
func persistentPreRun(cmd *cobra.Command, args []string) {

	requiredDeps := []Dependency{
		{Name: "yq", URL: "https://mikefarah.gitbook.io/yq/"},
		{Name: "yamllint", URL: "https://yamllint.readthedocs.io/en/stable/quickstart.html#installing-yamllint"},
	}

	// Ensure that operator has all required dependencies installed
	MustHaveDependenciesInstalled(requiredDeps)
}

var rootCmd = &cobra.Command{
	Use:              "multi-repo-updater",
	Short:            "Multi repo updater CLI",
	Long:             "Multi repo updater programmatically looks up Gruntwork repos and adds a configurable context to their CircleCI config workflow jobs",
	PersistentPreRun: persistentPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Context converter running...")

		GithubClient := ConfigureGithubClient()

		var fileProvidedRepos []*AllowedRepo

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
		}
		// Update repos to use the target context, where applicable
		ConvertReposContexts(GithubClient, GithubOrg, fileProvidedRepos)
	},
}

// Execute is the main entrypoint to the cmd package. Its sole responsibility is to invoke the rootCmd's Execute method
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Debug(err)
		return
	}
}
