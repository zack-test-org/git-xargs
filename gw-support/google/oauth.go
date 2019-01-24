package google

import (
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gruntwork-io/prototypes/gw-support/http-client"
	"github.com/gruntwork-io/prototypes/gw-support/keybase"
)

func PrepareOauthConfig(port int) (*oauth2.Config, error) {
	clientID, err := keybase.DecodeSecret(EncryptedClientID)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	clientSecret, err := keybase.DecodeSecret(EncryptedClientSecret)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  http_client.GetPath(port, "/callback"),
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events.readonly",
			"https://www.googleapis.com/auth/calendar.readonly",
		},
		Endpoint: google.Endpoint,
	}, nil
}

func GetAuthCodeURL(config *oauth2.Config) string {
	return config.AuthCodeURL("state", oauth2.AccessTypeOnline)
}
