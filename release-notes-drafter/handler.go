package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-github/github"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/sirupsen/logrus"
)

// Response represents a generic response returned by this app. This is minimal because it won't be used for further
// processing. Rely on the application logs for debugging purposes.
type Response struct {
	Status string `json:"status"`
}

// ErrorResponse is a generic error response to return when we encounter an error
var ErrorResponse = Response{Status: "error"}

// proxyRequestAsRequest will take an APIGatewayProxyRequest and convert it to a http.Request object that the go-github
// client understands.
func proxyRequestAsRequest(request events.APIGatewayProxyRequest) http.Request {
	headers := http.Header{}
	for key, value := range request.Headers {
		headers[key] = []string{value}
	}
	_, hasKey := headers["Content-Type"]
	if !hasKey {
		// Assume application/json if not provided
		headers["Content-Type"] = []string{"application/json"}
	}
	return http.Request{
		Method: request.HTTPMethod,
		Header: headers,
		Body:   ioutil.NopCloser(bytes.NewReader([]byte(request.Body))),
	}
}

// processPullRequestMerged will process the provided github pull request event as a merge event. This will
// - Get the latest release note draft if it exists. Otherwise, create a new one with an empty body.
// - Extract the release note draft body. If it is empty, insert the template.
// - Append the information from the merged pull request. This includes: (1) modules affected based on files changed;
//   (2) placeholder description which consists of the PR title; (3) PR link
// - Update the release note draft with the new description
// Note: This assumes the passed in event is a pull request merged event
func processPullRequestMerged(logger *logrus.Entry, event *github.PullRequestEvent) error {
	logger.Infof("Processing pull request %s/%d merge", event.GetRepo().GetFullName(), event.GetNumber())

	// Assert that this is a pull request merge event
	if event.GetAction() != "closed" || !event.GetPullRequest().GetMerged() {
		return errors.WithStackTrace(IncorrectHandlerError{event.GetAction()})
	}

	if ProjectLockTableName != "" {
		// Synchronize on the repository to ensure that only one instance of the function is exectued per repo.
		logger.Info("Lock table is configured, so synchronizing event processing on repository.")
		lockString := event.GetRepo().GetFullName()
		defer ReleaseLock(lockString)
		err := BlockingAcquireLock(lockString)
		if err != nil {
			return err
		}
	}

	logger.Infof("Getting or creating release draft for repo %s", event.GetRepo().GetFullName())
	draftRelease, err := getOrCreateReleaseDraft(logger, event.GetRepo())
	if err != nil {
		return err
	}
	logger.Infof("Done getting or creating release draft for repo %s", event.GetRepo().GetFullName())

	draftBody := draftRelease.GetBody()
	if strings.TrimSpace(draftBody) == "" {
		draftBody = ReleaseNoteTemplate
		logger.Info("Draft body was empty. Using template.")
	}

	logger.Infof("Appending release info")
	pullRequest := event.GetPullRequest()
	modulesAffected, err := getModulesAffected(pullRequest)
	if err != nil {
		return err
	}
	draftBody, err = addModulesAffected(draftBody, modulesAffected)
	if err != nil {
		return err
	}
	description := getDescription(pullRequest)
	draftBody, err = addDescription(draftBody, description)
	if err != nil {
		return err
	}
	link := getLink(pullRequest)
	draftBody, err = addRelatedLink(draftBody, link)
	if err != nil {
		return err
	}
	logger.Infof("Done appending release info")

	err = updateReleaseDescription(logger, event.GetRepo(), draftRelease, draftBody)
	if err == nil {
		logger.Infof("Successfully processed pull request %s/%d merge", event.GetRepo().GetFullName(), event.GetNumber())
	}
	return err
}

// processEvent will process the provided event. This entails looking up the event type and discarding anything that
// doesn't have a processor (currently we will only process pull request merge events).
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

// Handler is the main request handler. This will take an http request object and goes through the following sequence:
// - Validate that the received request is a proper github webhook event.
// - Parse and process the webhook, extracting the pull request details.
// - As part of processing the webhook event:
//    * Get or create a release note draft.
//    * Parse the release note draft body into a ReleaseNote object for easier manipulation.
//    * Add to the sections based on pull request merge info
//    * Render the updated release notes and update the github release object
func Handler(request *http.Request) (Response, error) {
	logger := GetProjectLogger()
	logger.Info("Received new event. Beginning Processing.")

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
