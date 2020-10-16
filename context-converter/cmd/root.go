package cmd

import (
	"os"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	GithubOauthToken string
	TargetContext    string
	GithubClient     *github.Client
	GithubOrg        string
	log              = logrus.New()
)

func init() {
	// Log messages that are of DEBUG level
	log.SetLevel(logrus.DebugLevel)

	rootCmd.PersistentFlags().StringVarP(&GithubOrg, "github-org", "o", "gruntwork-io", "The Github organization whose repos should be operated on")

}

func persistentPreRun(cmd *cobra.Command, args []string) {

	// Ensure user provided a GITHUB_OAUTH_TOKEN
	userProvidedToken := os.Getenv("GITHUB_OAUTH_TOKEN")
	if userProvidedToken == "" {
		panic("You must set a Github personal access token with access to Gruntwork repos via the Env var GITHUB_OAUTH_TOKEN")
	}

	GithubOauthToken = userProvidedToken

}

var rootCmd = &cobra.Command{
	Use:              "context-converter",
	Short:            "Context converter CLI",
	Long:             "Context converter programmatically looks up Gruntwork repos and adds a configurable context to their CircleCI config workflow jobs",
	PersistentPreRun: persistentPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Context converter running...")

		// Configure Github client, with user-provided personal access token
		ConfigureGithubClient()

		// Update repos to use the target context, where applicable
		ConvertReposContexts()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Debug(err)
		return
	}
}
