package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

var (
	GithubWebhookSecretKey = os.Getenv("GITHUB_WEBHOOK_SECRET")
	GithubApiKey           = os.Getenv("GITHUB_API_KEY")
	IsLocal                = os.Getenv("IS_LOCAL")
)

func main() {
	if IsLocal != "" {
		serveLocal()
	} else {
		serveLambda()
	}
}

func serveLambda() {
	lambda.Start(LambdaHandler)
}

func serveLocal() {
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
