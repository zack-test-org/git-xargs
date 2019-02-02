package http_server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"

	"github.com/gruntwork-io/prototypes/gw-support/csrf"
	"github.com/gruntwork-io/prototypes/gw-support/logging"
)

var (
	LocalCache     *cache.Cache
	ShutdownServer chan os.Signal
	ServerPort     int
)

func StartServer(port int) error {
	logger := logging.GetProjectLogger()

	// Initialize the cache
	LocalCache = cache.New(4*time.Hour, 15*time.Minute)
	// Record port so we can use it later
	ServerPort = port

	router := gin.Default()

	// require authorization to retrieve credentials and to shutdown the server
	token, err := csrf.GetOrCreateCsrfToken("")
	if err != nil {
		return err
	}
	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		fmt.Sprintf("%s", csrf.Username): token, // user:houston password:csrf token
	}))
	authorized.GET("shutdown", shutdownController)
	authorized.GET("status", statusController)
	authorized.GET("credentials", credentialsController)

	// These endpoints (Oauth2 flow) can be retrieved without authorization
	router.GET("login", initiateOauthFlowController)
	router.GET("callback", oauthCallbackController)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: router,
	}

	go func() {
		// service connections
		if err := server.ListenAndServe(); err != nil {
			logger.Infof("listen: %s", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	ShutdownServer = make(chan os.Signal)
	signal.Notify(ShutdownServer, os.Interrupt)
	<-ShutdownServer
	logger.Infof("Shutdown gw-support server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}
