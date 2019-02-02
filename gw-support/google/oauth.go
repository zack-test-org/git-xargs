package google

import (
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gruntwork-io/prototypes/gw-support/http-client"
	"github.com/gruntwork-io/prototypes/gw-support/keybase"
)

// asyncGetKeybaseSecret works with awaitGetKeybaseSecret to implement a crude version of async await syntax in golang.
// To use, first create a new channel for each intended async call that will be used to message pass between the async
// function and the main thread.
// Then, spawn goroutines to execute the asyncGetKeybaseSecret function on the encrypted string.
// Finally, await the results by using awaitGetKeybaseSecret and passing in the channels.
func asyncGetKeybaseSecret(out chan<- interface{}, encryptedStr string) {
	secretVal, err := keybase.DecodeSecret(encryptedStr)
	if err != nil {
		out <- err
	} else {
		out <- secretVal
	}
}

func awaitGetKeybaseSecret(in <-chan interface{}) (string, error) {
	result := <-in
	switch result := result.(type) {
	case string:
		return result, nil
	case error:
		return "", result
	default:
		// This should never happen
		return "", UnknownReturnType{Data: result}
	}
}

// PrepareOauthConfig will generate a new oauth2.Config struct that can be used with the google API for initiating and
// completing the Oauth flow.
func PrepareOauthConfig(port int) (*oauth2.Config, error) {
	// Decode the oauth secrets in parallel
	clientIDChan := make(chan interface{}, 1)
	go asyncGetKeybaseSecret(clientIDChan, EncryptedClientID)

	clientSecretChan := make(chan interface{}, 1)
	go asyncGetKeybaseSecret(clientSecretChan, EncryptedClientSecret)

	clientID, err := awaitGetKeybaseSecret(clientIDChan)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	clientSecret, err := awaitGetKeybaseSecret(clientSecretChan)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	// Then pass the oauth secrets to generate the actual struct
	return PrepareOauthConfigFromInfo(port, clientID, clientSecret), nil
}

// PrepareOauthConfigFromInfo will generate a new oauth2.Config struct based on provided oauth secrets that can be used
// with the google API for initiating and completing the Oauth flow.
func PrepareOauthConfigFromInfo(port int, clientID string, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  http_client.GetPath(port, "/callback"),
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events.readonly",
			"https://www.googleapis.com/auth/calendar.readonly",
		},
		Endpoint: google.Endpoint,
	}
}

// GetAuthCodeURL will return the URL that can be used to authorize access to the Google API via Oauth
func GetAuthCodeURL(config *oauth2.Config) string {
	return config.AuthCodeURL("state", oauth2.AccessTypeOnline)
}
