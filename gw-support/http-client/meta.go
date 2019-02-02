package http_client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gruntwork-io/gruntwork-cli/errors"

	"github.com/gruntwork-io/prototypes/gw-support/csrf"
	"github.com/gruntwork-io/prototypes/gw-support/logging"
)

func makeRequest(method string, port int, path string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, GetPath(port, path), nil)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	// update request with csrf before making request
	username := csrf.Username
	password, err := csrf.GetOrCreateCsrfToken("")
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return resp, nil
}

func ServerRunning(port int) bool {
	logger := logging.GetProjectLogger()

	result, err := makeRequest("GET", port, "/status")
	if err != nil {
		logger.Debugf("Error checking server status: %s", err)
		logger.Debugf("This may be intentional if the server has not been started yet.")
		return false
	}
	if result.StatusCode != 200 {
		logger.Warnf("Error reaching status endpoint. Response status code: %d", result.StatusCode)
		return false
	}
	return true
}

func StopServer(port int) error {
	logger := logging.GetProjectLogger()

	_, err := makeRequest("GET", port, "/shutdown")
	if err != nil {
		logger.Errorf("Error shutting down server: %s", err)
		if strings.Contains(errors.Unwrap(err).Error(), "connection refused") {
			return errors.WithStackTrace(fmt.Errorf("unable to stop gw-support http server: server not running"))
		}
		return errors.WithStackTrace(fmt.Errorf("unable to stop gw-support http server: %s", err))
	}
	logger.Infof("Successfully shut down server listening on port %d", port)
	return csrf.DeleteCsrfToken("")
}
