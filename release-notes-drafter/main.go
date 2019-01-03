package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

var (
	// GithubWebhookSecretkey is the secret key used to validate the incoming webhook requests so that we can make sure
	// it is coming from github.
	GithubWebhookSecretKey = os.Getenv("GITHUB_WEBHOOK_SECRET")
	// GithubApiKey is the personal access token to use to read pull request info and update release notes.
	GithubApiKey = os.Getenv("GITHUB_API_KEY")
	// IsLocal determines whether or not to run in local mode (basic web server as opposed to lambda function)
	IsLocal = os.Getenv("IS_LOCAL")
)

func main() {
	if IsLocal != "" {
		serveLocal()
	} else {
		serveLambda()
	}
}

// serveLambda will serve the handler as a lambda function.
func serveLambda() {
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
