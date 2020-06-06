package main

import (
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gruntwork-io/gruntwork-cli/entrypoint"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/gruntwork-io/usage-patterns/scripts/refarch-init/form"
	http_client "github.com/gruntwork-io/usage-patterns/scripts/refarch-init/http-client"
	http_server "github.com/gruntwork-io/usage-patterns/scripts/refarch-init/http-server"
)

const DefaultPort = 46548

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

	app.Name = "customer-users"
	app.Author = "Gruntwork <www.gruntwork.io>"
	app.Before = initCli
	app.Flags = []cli.Flag{
		logLevelFlag,
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "lookup",
			Usage:  "Lookup information on all authorized users for a subscription.",
			Flags:  []cli.Flag{portFlag},
			Action: lookupCmd,
		},
		cli.Command{
			Name:  "start",
			Usage: "Start the background HTTP server process to manage credentials.",
			Description: `To be able to access the Google API, we need a webserver that can initiate and complete the Oauth flow to retrieve credentials to access the API. To cache the credentials in memory, the server is run in the background so that it can be called upon when ever the command needs credentials.

You can use the 'customer-users stop' command to shut down the HTTP server.

Note: Any command that needs google credentials will automatically start the HTTP server.`,
			Flags:  []cli.Flag{portFlag, foregroundFlag},
			Action: startServer,
		},
		cli.Command{
			Name:        "stop",
			Usage:       "Stop the background HTTP server process used to manage credentials.",
			Description: `Stop the background HTTP server process used to manage credentials. See 'customer-users start' for more details on the HTTP server.`,
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

// lookupCmd is the main routine of this CLI. This will:
// - start the google credentials server if it is not running
// - retrieve google oauth credentials for API access
// - use the google oauth credentials to setup an API client for google sheets
// - access the Gruntwork Customers google spreadsheet to lookup information about Gruntwork Customers
func lookupCmd(cliContext *cli.Context) error {
	logger := GetProjectLogger()

	port := cliContext.Int(portFlag.Name)

	if !http_client.ServerRunning(port) {
		logger.Infof("google credentials server is not running. Starting server in background.")
		err := callStartServer(port)
		if err != nil {
			logger.Errorf("Error starting server: %s", err)
			return err
		}
		time.Sleep(10 * time.Second)
	}

	token, err := http_client.EnsureCredentials(port)
	if err != nil {
		return err
	}
	oauthConfig, err := form.PrepareOauthConfig(port)
	if err != nil {
		return err
	}
	client, err := form.NewSheetsClient(oauthConfig, token)
	if err != nil {
		return err
	}
	return lookupUsers(client)
}

// startServer is the command entrypoint to start. The start command starts the HTTP server that is used to manage
// credentials to access the Gsuites API.
func startServer(cliContext *cli.Context) error {
	logger := GetProjectLogger()

	port := cliContext.Int(portFlag.Name)
	foreground := cliContext.Bool(foregroundFlag.Name)

	if http_client.ServerRunning(port) {
		logger.Warnf("google credentials server is already running on port %d", port)
		return nil
	}

	if foreground {
		return http_server.StartServer(port)
	} else {
		logger.Infof("starting google credentials http server on port %d", port)
		return rerunForeground()
	}
}

func stopServer(cliContext *cli.Context) error {
	port := cliContext.Int(portFlag.Name)
	return http_client.StopServer(port)
}

func callStartServer(port int) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.WithStackTrace(err)
	}

	args := []string{os.Args[0], "start", "--port", strconv.Itoa(port), "--foreground"}
	return runCommandInBackground(args, cwd)
}

// rerunForeground runs 'customer-users start --foreground' as a background process, detached from the current process.
// This works like houston-cli.
func rerunForeground() error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.WithStackTrace(err)
	}

	args := append(os.Args, "--foreground")
	return runCommandInBackground(args, cwd)
}

func runCommandInBackground(args []string, cmdDir string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cmdDir
	err := cmd.Start()
	if err != nil {
		return errors.WithStackTrace(err)
	}
	return errors.WithStackTrace(cmd.Process.Release())
}
