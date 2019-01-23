package http_server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gruntwork-io/prototypes/gw-support/logging"
)

var shutdownServer chan os.Signal

func StartServer(port int) error {
	logger := logging.GetProjectLogger()

	router := gin.Default()

	router.GET("status", statusController)
	router.GET("shutdown", shutdownController)

	// TODO: implement similar system to houston-cli with csrf

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
	shutdownServer = make(chan os.Signal)
	signal.Notify(shutdownServer, os.Interrupt)
	<-shutdownServer
	logger.Infof("Shutdown gw-support server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}
