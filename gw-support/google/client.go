package google

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gruntwork-io/gruntwork-cli/errors"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

func NewClient(oauthConfig *oauth2.Config, tokenJson map[string]string) (*http.Client, error) {
	data, err := json.Marshal(tokenJson)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(data, token)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	client := oauthConfig.Client(context.Background(), token)
	return client, nil
}

func NewCalendarClient(oauthConfig *oauth2.Config, tokenJson map[string]string) (*calendar.Service, error) {
	httpClient, err := NewClient(oauthConfig, tokenJson)
	if err != nil {
		return nil, err
	}
	return calendar.New(httpClient)
}
