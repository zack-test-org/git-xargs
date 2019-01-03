package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	GithubWebhookSecretKey = os.Getenv("GITHUB_WEBHOOK_SECRET")
	GithubApiKey           = os.Getenv("GITHUB_API_KEY")
)

func main() {
	lambda.Start(Handler)
}
