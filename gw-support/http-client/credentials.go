package http_client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/houston-cli/browser"

	"github.com/gruntwork-io/prototypes/gw-support/logging"
)

const (
	MaxRetries          = 40
	SleepBetweenRetries = 5 * time.Second
)

// EnsureCredentials will initiate the oauth flow if the Google credentials do not exist on the local server. See the
// root README for more information on the local server design. This will error out if it encounters a fatal error.
func EnsureCredentials(port int) (map[string]string, error) {
	logger := logging.GetProjectLogger()

	token, found, err := GetCredentials(port)
	if err != nil {
		logger.Errorf("Error getting credentials: %s", err)
		return nil, err
	}
	if !found {
		logger.Infof("No credentials found. Initiating flow.")
		loginUrl := GetPath(port, "/login")
		err := browser.OpenBrowser(loginUrl)
		if err != nil {
			return nil, errors.WithStackTrace(err)
		}
		return waitUntilCredentialsFound(port)
	}
	return token, nil
}

func GetCredentials(port int) (map[string]string, bool, error) {
	logger := logging.GetProjectLogger()

	response, err := makeRequest("GET", port, "/credentials")
	if err != nil {
		logger.Errorf("Error getting credentials: %s", err)
		return nil, false, err
	}
	data, err := getJsonBody(response)
	if err != nil {
		logger.Errorf("Error reading json data from message body (status: %d): %s", response.StatusCode, err)
		return nil, false, err
	}
	if response.StatusCode != 200 {
		logger.Warnf("Credentials not found: %s", data["error"])
		return nil, false, nil
	}
	logger.Debugf("Found and received credentials from server")
	return data, true, nil
}

func getJsonBody(response *http.Response) (map[string]string, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	data := map[string]string{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return data, nil
}

func waitUntilCredentialsFound(port int) (map[string]string, error) {
	logger := logging.GetProjectLogger()
	logger.Infof("Waiting for oauth flow to complete and retrieve credentials.")

	for i := 0; i <= MaxRetries; i++ {
		logger.Info("Attempting to get credentials")
		token, found, err := GetCredentials(port)
		if err != nil {
			logger.Errorf("Fatal error waiting for credentials: %s", err)
			return nil, err
		}
		if !found {
			logger.Infof("Did not find credentials. Sleeping for %s before retrying.", SleepBetweenRetries)
			time.Sleep(SleepBetweenRetries)
		} else {
			logger.Info("Found credentials")
			return token, nil
		}
	}
	logger.Errorf("Timedout waiting for credentials")
	return nil, errors.WithStackTrace(MaxRetriesError{"waiting for credentials"})
}
