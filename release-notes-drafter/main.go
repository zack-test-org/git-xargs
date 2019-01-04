package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/github"
	"github.com/gruntwork-io/gruntwork-cli/entrypoint"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	// GithubWebhookSecretkey is the secret key used to validate the incoming webhook requests so that we can make sure
	// it is coming from github. Only used for local or lambda modes.
	GithubWebhookSecretKey = os.Getenv("GITHUB_WEBHOOK_SECRET")
	// GithubApiKey is the personal access token to use to read pull request info and update release notes.
	GithubApiKey = os.Getenv("GITHUB_TOKEN")
	// GithubEventPath is a path on disk to a json file that contains the webhook data
	GithubEventPath = os.Getenv("GITHUB_EVENT_PATH")
	// GithubEventType is the webhook event type. See https://developer.github.com/v3/activity/events/types/.
	GithubEventType = os.Getenv("GITHUB_EVENT_NAME")

	logLevelFlag = cli.StringFlag{
		Name:  "loglevel",
		Value: logrus.InfoLevel.String(),
	}
	jsonPathFlag = cli.StringFlag{
		Name: "event-json",
		// The environment variable is set by Github Action
		Usage: "Path to the webhook event data as a json. This can also be set as the environment variable GITHUB_EVENT_PATH.",
	}
	eventTypeFlag = cli.StringFlag{
		Name: "event-type",
		// The environment variable is set by Github Action
		Usage: "The webhook event type. This can also be set as the environment variable GITHUB_EVENT_NAME.",
	}
)

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

func main() {
	app := entrypoint.NewApp()
	entrypoint.HelpTextLineWidth = 120

	app.Name = "release-notes-drafter"
	app.Author = "Gruntwork <www.gruntwork.io>"
	app.Description = "A tool to draft release notes based on pull request merge events. Can be used as a web server locally, in lambda, or as a cli action."

	app.Before = initCli
	app.Flags = []cli.Flag{
		logLevelFlag,
	}
	app.Commands = []cli.Command{
		{
			Name:        "action",
			Usage:       "CLI action to handle events",
			Description: "Run the handler as a standalone function, handling the input event directly. This should be used in the context of a CI job, or Github Action.",
			Flags: []cli.Flag{
				jsonPathFlag,
				eventTypeFlag,
			},
			Action: runAction,
		},
		{
			Name:        "local",
			Usage:       "Local web server",
			Description: "Start a local webserver that can interpret the github webhook events. This should be hooked to the repository via a github webhook.",
			Action: func(cliContext *cli.Context) error {
				serveLocal()
				return nil
			},
		},
		{
			Name:        "lambda",
			Usage:       "Start an AWS lambda handler",
			Description: "Start an AWS lambda handler that can interpret the github webhook events. This is intended to be run in the context of an AWS lambda function, with an API gateway. This should be hooked to the repository via a github webhook.",
			Action: func(cliContext *cli.Context) error {
				serveLambda()
				return nil
			},
		},
	}

	entrypoint.RunApp(app)
}

// serveLambda will serve the handler as a lambda function.
func serveLambda() {
	logger := GetProjectLogger()
	logger.Infof("Starting as lambda function")
	lambda.Start(lambdaHandler)
}

// lambdaHandler does not get a http.Request object, so convert the provided APIGatewayProxyRequest object into the http
// request object.
func lambdaHandler(request events.APIGatewayProxyRequest) (Response, error) {
	httpRequest := proxyRequestAsRequest(request)
	response, err := Handler(&httpRequest)
	if err != nil {
		GetProjectLogger().Errorf("%s", errors.PrintErrorWithStackTrace(err))
	}
	return response, err
}

// serveLocal will serve the handler as a basic web server.
func serveLocal() {
	logger := GetProjectLogger()
	logger.Infof("Starting as local web server")
	http.HandleFunc("/", httpHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func httpHandler(writer http.ResponseWriter, request *http.Request) {
	response, err := Handler(request)
	if err != nil {
		logger := GetProjectLogger()
		logger.Errorf("%s", errors.PrintErrorWithStackTrace(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Write(data)
}

// runAction will interpret the passed in parameters and run the handler for the provided event.
func runAction(cliContext *cli.Context) error {
	logger := GetProjectLogger()
	logger.Infof("Running handler as a CLI action.")

	var jsonPath string
	if cliContext.String(jsonPathFlag.Name) != "" {
		jsonPath = cliContext.String(jsonPathFlag.Name)
	} else if GithubEventPath != "" {
		jsonPath = GithubEventPath
	} else {
		logger.Errorf("Github event path is required to execute release-notes-drafter as a CLI action.")
		return MissingRequiredParameter{jsonPathFlag.Name}
	}

	var eventType string
	if cliContext.String(eventTypeFlag.Name) != "" {
		eventType = cliContext.String(eventTypeFlag.Name)
	} else if GithubEventType != "" {
		eventType = GithubEventType
	} else {
		logger.Errorf("Github event type is required to execute release-notes-drafter as a CLI action.")
		return MissingRequiredParameter{eventTypeFlag.Name}
	}

	logger.Infof("Reading event json data from %s", jsonPath)
	jsonData, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	logger.Infof("Successfully read event json data from %s", jsonPath)

	logger.Infof("Parsing event json data as %s", eventType)
	event, err := github.ParseWebHook(eventType, jsonData)
	if err != nil {
		return err
	}
	logger.Infof("Successfully parsed event json data as %s", eventType)

	logger.Infof("Processing event data")
	err = processEvent(logger, eventType, event)
	if err == nil {
		logger.Infof("Successfully processed event data")
	}
	return err
}
