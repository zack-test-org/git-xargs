package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	DryRun           bool
	GithubOrg        string
	TargetContext    string
	TargetBranch     = "IAC-1616-programmatically-fix-context"
	RefsTargetBranch = fmt.Sprintf("heads/%s", TargetBranch)
	log              = logrus.New()
)

func init() {
	// Log messages that are of DEBUG level
	log.SetLevel(logrus.DebugLevel)

	rootCmd.PersistentFlags().StringVarP(&GithubOrg, "github-org", "o", "gruntwork-io", "The Github organization whose repos should be operated on")
	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dry-run", "d", true, "When dry-run is set to true, only proposed YAML updates will be output, but not changes in Github will be made (no branches will be created, no files updated, no PRs opened)")

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

		// Update repos to use the target context, where applicable
		ConvertReposContexts(GithubClient, GithubOrg)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Debug(err)
		return
	}
}
