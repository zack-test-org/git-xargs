package google

import (
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gruntwork-io/prototypes/gw-support/http-client"
	"github.com/gruntwork-io/prototypes/gw-support/keybase"
)

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
		panic(result)
	}
}

func PrepareOauthConfig(port int) (*oauth2.Config, error) {
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
	return PrepareOauthConfigFromInfo(port, clientID, clientSecret), nil
}

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

func GetAuthCodeURL(config *oauth2.Config) string {
	return config.AuthCodeURL("state", oauth2.AccessTypeOnline)
}
