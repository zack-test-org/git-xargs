package http_server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"

	"github.com/gruntwork-io/prototypes/gw-support/google"
)

func getOrSetOauthConfig() (*oauth2.Config, error) {
	cachedConf, found := LocalCache.Get("oauthConfig")
	if found {
		conf := cachedConf.(*oauth2.Config)
		return conf, nil
	}
	conf, err := google.PrepareOauthConfig(ServerPort)
	if err != nil {
		return conf, err
	}
	LocalCache.Set("oauthConfig", conf, cache.DefaultExpiration)
	return conf, nil
}

// initiateOauthFlowController will initiate the Oauth2 flow with Google, redirecting the user to the Google login
// screen and consent for authorization.
func initiateOauthFlowController(ginCtx *gin.Context) {
	conf, err := getOrSetOauthConfig()
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	url := google.GetAuthCodeURL(conf)
	ginCtx.Redirect(http.StatusFound, url)
}

// oauthCallbackController will consume the oauth2 callback and cache the token in the server so that it can be used to
// query the google api.
func oauthCallbackController(ginCtx *gin.Context) {
	code := ginCtx.Query("code")

	conf, err := getOrSetOauthConfig()
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	LocalCache.Set("token", tok, cache.DefaultExpiration)

	ginCtx.String(http.StatusOK, "The CLI has successfully receieved the authorization token. You can close this tab now.")
}
