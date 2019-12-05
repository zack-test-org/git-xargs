package cmd

import (
	"os"

	"github.com/gruntwork-io/gruntwork-cli/files"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/gruntwork-io/prototypes/tgscaffold/config"
	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	logLevelFlag = cli.StringFlag{
		Name:  "loglevel",
		Value: logrus.InfoLevel.String(),
	}
	configPathFlag = cli.StringFlag{
		Name:  "config",
		Value: "~/.config/terragrunt/scaffold.toml",
	}

	GlobalFlags = []cli.Flag{
		logLevelFlag,
		configPathFlag,
	}
)

func withInitialization(action func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := initCli(c); err != nil {
			return err
		}
		return action(c)
	}
}

// initCli initializes the CLI app before any command is actually executed. This function will handle all the
// setup code, such as setting up the logger with the appropriate log level.
func initCli(c *cli.Context) error {
	// Set logging level
	logLevel := c.String(logLevelFlag.Name)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return errors.WithStackTrace(err)
	}
	logging.SetGlobalLogLevel(level)

	// If logging level is for debugging (debug or trace), enable stacktrace debugging
	if level == logrus.DebugLevel || level == logrus.TraceLevel {
		os.Setenv("GRUNTWORK_DEBUG", "true")
	}

	// Ensure config exists
	configPath := c.String(configPathFlag.Name)
	configPathExpanded, err := homedir.Expand(configPath)
	if err != nil {
		return errors.WithStackTrace(err)
	}
	if !files.FileExists(configPathExpanded) {
		err := config.InitializeConfig(configPathExpanded)
		if err != nil {
			return errors.WithStackTrace(err)
		}
	}
	return nil
}
