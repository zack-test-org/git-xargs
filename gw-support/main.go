package main

import (
	"os"
	"os/exec"

	"github.com/gruntwork-io/gruntwork-cli/entrypoint"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/gruntwork-io/prototypes/gw-support/http-client"
	"github.com/gruntwork-io/prototypes/gw-support/http-server"
	project_logging "github.com/gruntwork-io/prototypes/gw-support/logging"
)

const DefaultPort = 56789

var (
	logLevelFlag = cli.StringFlag{
		Name:  "loglevel",
		Value: logrus.InfoLevel.String(),
	}
	portFlag = cli.IntFlag{
		Name:  "port",
		Value: DefaultPort,
		Usage: "The TCP port the http server is running on.",
	}
	foregroundFlag = cli.BoolFlag{
		Name:  "foreground",
		Usage: "Whether or not to run the server in the foreground.",
	}
)

func main() {
	// Create a new CLI app. This will return a urfave/cli App with some
	// common initialization.
	app := entrypoint.NewApp()

	app.Name = "gw-support"
	app.Author = "Gruntwork <www.gruntwork.io>"
	app.Flags = []cli.Flag{
		logLevelFlag,
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "start",
			Usage: "Start the background HTTP server process to manage credentials.",
			Description: `To be able to access the Google calendar API, we need a webserver that can initiate and complete the Oauth flow to retrieve credentials to access the API. To cache the credentials in memory, the server is run in the background so that it can be called upon when ever the command needs credentials.
			
You can use the 'gw-support stop' command to shut down the HTTP server.

Note: Any command that needs google credentials will automatically start the HTTP server.`,
			Flags:  []cli.Flag{portFlag, foregroundFlag},
			Action: startServer,
		},
		cli.Command{
			Name:        "stop",
			Usage:       "Stop the background HTTP server process used to manage credentials.",
			Description: `Stop the background HTTP server process used to manage credentials. See 'gw-support start' for more details on the HTTP server.`,
			Flags:       []cli.Flag{portFlag},
			Action:      stopServer,
		},
	}

	// Run your app using the entrypoint package, which will take care of exit codes, stack traces, and panics
	entrypoint.RunApp(app)
}

// initCli initializes the CLI app before any command is actually executed. This function will handle all the setup
// code, such as setting up the logger with the appropriate log level.
func initCli(cliContext *cli.Context) error {
	// Set logging level
	logLevel := cliContext.String(logLevelFlag.Name)
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return errors.WithStackTrace(err)
	}
	logging.SetGlobalLogLevel(level)
	return nil
}

// startServer is the command entrypoint to start. The start command starts the HTTP server that is used to manage
// credentials to access the Gsuites API.
func startServer(cliContext *cli.Context) error {
	logger := project_logging.GetProjectLogger()

	port := cliContext.Int(portFlag.Name)
	foreground := cliContext.Bool(foregroundFlag.Name)
	if http_client.ServerRunning(port) {
		logger.Warnf("gw-support server is already running on port %d", port)
		return nil
	}

	if foreground {
		return http_server.StartServer(port)
	} else {
		logger.Infof("starting gw-support http server on port %d", port)
		return rerunForeground()
	}
}

func stopServer(cliContext *cli.Context) error {
	port := cliContext.Int(portFlag.Name)
	return http_client.StopServer(port)
}

// rerunForeground runs 'gw-support start --foreground' as a background process, detached from the current process.
// This works like houston-cli.
func rerunForeground() error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.WithStackTrace(err)
	}

	args := append(os.Args, "--foreground")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	err = cmd.Start()
	if err != nil {
		return errors.WithStackTrace(err)
	}
	return errors.WithStackTrace(cmd.Process.Release())
}
