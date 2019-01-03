package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-github/github"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Status string `json:"status"`
}

var ErrorResponse = Response{Status: "error"}

func proxyRequestAsRequest(request events.APIGatewayProxyRequest) http.Request {
	headers := http.Header{}
	for key, value := range request.Headers {
		headers[key] = []string{value}
	}
	return http.Request{
		Method: request.HTTPMethod,
		Header: headers,
		Body:   ioutil.NopCloser(bytes.NewReader([]byte(request.Body))),
	}
}

func processPullRequestMerged(logger *logrus.Entry, event *github.PullRequestEvent) error {
	// We assume the passed in event is a pull request merged event
	logger.Infof("Processing pull request %s/%d merge", event.GetRepo().GetFullName(), event.GetNumber())
	defer logger.Infof("Processed pull request %s/%d merge", event.GetRepo().GetFullName(), event.GetNumber())

	logger.Infof("Getting or creating release draft for repo %s", event.GetRepo().GetFullName())
	draftRelease, err := getOrCreateReleaseDraft(logger, event.GetRepo())
	if err != nil {
		return err
	}
	logger.Infof("Done getting or creating release draft for repo %s", event.GetRepo().GetFullName())

	draftBody := draftRelease.GetBody()
	logger.Infof("Parsing release note body for draft release %s", draftBody)
	releaseNote, err := parseReleaseNoteBody(draftBody)
	if err != nil {
		return err
	}
	logger.Infof("Done parsing release note body for draft release %s", draftBody)

	logger.Infof("Appending release info")
	pullRequest := event.GetPullRequest()
	modulesAffected, err := getModulesAffected(pullRequest)
	if err != nil {
		return err
	}
	for _, module := range modulesAffected {
		releaseNote = appendModulesAffected(releaseNote, module)
	}
	description := getDescription(pullRequest)
	releaseNote = appendDescription(releaseNote, description)
	link := getLink(pullRequest)
	releaseNote = appendRelatedLink(releaseNote, link)
	logger.Infof("Done appending release info")

	return updateReleaseDescription(logger, event.GetRepo(), draftRelease, RenderReleaseNote(releaseNote))
}

func processEvent(logger *logrus.Entry, webhookType string, event interface{}) error {
	switch event := event.(type) {
	case *github.PullRequestEvent:
		action := event.GetAction()
		logger.Infof("Received pull request event %s", action)
		if action != "closed" {
			logger.Infof("Ignoring non pull request merge event %s", action)
		} else if !event.GetPullRequest().GetMerged() {
			logger.Infof("Ignoring pull request close event")
		} else {
			logger.Info("Detected pull request merge event")
			return processPullRequestMerged(logger, event)
		}
	default:
		logger.Infof("Ignoring non pull request merge event %s", webhookType)
	}
	return nil
}

func LambdaHandler(request events.APIGatewayProxyRequest) (Response, error) {
	httpRequest := proxyRequestAsRequest(request)
	return Handler(&httpRequest)

}

func Handler(request *http.Request) (Response, error) {
	logger := GetProjectLogger()

	logger.Info("Validating request")
	payload, err := github.ValidatePayload(request, []byte(GithubWebhookSecretKey))
	if err != nil {
		logger.Errorf("Error validating event payload: %s", err)
		return ErrorResponse, errors.WithStackTrace(err)
	}
	logger.Info("Validated request is webhook from github")

	webhookType := github.WebHookType(request)
	logger.Infof("Parsing webhook type %s payload", webhookType)
	event, err := github.ParseWebHook(webhookType, payload)
	if err != nil {
		logger.Errorf("Error parsing webhook event payload: %s", err)
		return ErrorResponse, errors.WithStackTrace(err)
	}
	logger.Infof("Successfully parsed webhook type %s payload", webhookType)

	logger.Infof("Processing webhook type %s event", webhookType)
	err = processEvent(logger, webhookType, event)
	if err != nil {
		logger.Errorf("Error while processing webhook event: %s", err)
		return ErrorResponse, errors.WithStackTrace(err)
	}
	logger.Infof("Processed webhook type %s event", webhookType)

	return Response{"ok"}, nil
}
