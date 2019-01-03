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
	return nil
}

/*
	// We assume the passed in event is a pull request merged event
	logger.Infof("Processing pull request %s/%d merge", event.GetRepo().GetFullName(), event.GetNumber())
	defer logger.Infof("Processed pull request %s/%d merge", event.GetRepo().GetFullName(), event.GetNumber())

	draftRelease, err := getOrCreateReleaseDraft(logger, event.GetRepo())
	if err != nil {
		return err
	}

	releaseNote, err := parseReleaseNoteBody(draftRelease.GetBody())
	if err != nil {
		return err
	}

	pullRequest := event.GetPullRequest()
	modulesAffected, err := getModulesAffected(pullRequest)
	if err != nil {
		return err
	}
	releaseNote = appendModulesAffected(releaseNote, modulesAffected)
	description := getDescription(pullRequest)
	releaseNote = appendDescription(releaseNote, description)
	link := getLink(pullRequest)
	releaseNote = appendRelatedLinks(releaseNote, link)

	return updateReleaseDescription(logger, event.GetRepo(), draftRelease, RenderReleaseNote(releaseNote))
}*/

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

func Handler(request events.APIGatewayProxyRequest) (Response, error) {
	logger := GetProjectLogger()
	logger.Infof("Request data: %v", request.Body)

	httpRequest := proxyRequestAsRequest(request)

	payload, err := github.ValidatePayload(&httpRequest, []byte(GithubWebhookSecretKey))
	if err != nil {
		logger.Errorf("Error validating event payload: %s", err)
		return ErrorResponse, errors.WithStackTrace(err)
	}
	event, err := github.ParseWebHook(github.WebHookType(&httpRequest), payload)
	if err != nil {
		logger.Errorf("Error parsing webhook event payload: %s", err)
		return ErrorResponse, errors.WithStackTrace(err)
	}
	err = processEvent(logger, github.WebHookType(&httpRequest), event)
	if err != nil {
		logger.Errorf("Error while processing webhook event: %s", err)
		return ErrorResponse, errors.WithStackTrace(err)
	}

	return Response{"ok"}, nil
}
