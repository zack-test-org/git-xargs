package http_server

import (
	"net/http"
	"os"
	"syscall"

	"github.com/gin-gonic/gin"
)

func shutdownController(ginCtx *gin.Context) {
	ShutdownServer <- os.Signal(syscall.SIGINT)
	ginCtx.JSON(http.StatusOK, gin.H{"status": "shutting down"})
}

func statusController(ginCtx *gin.Context) {
	ginCtx.JSON(http.StatusOK, gin.H{"status": "running"})
}
