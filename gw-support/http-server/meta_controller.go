package http_server

import (
	"os"
	"syscall"

	"github.com/gin-gonic/gin"
)

func shutdownController(ginCtx *gin.Context) {
	shutdownServer <- os.Signal(syscall.SIGINT)
	ginCtx.JSON(200, gin.H{"status": "shutting down"})
}

func statusController(ginCtx *gin.Context) {
	ginCtx.JSON(200, gin.H{"status": "running"})
}
