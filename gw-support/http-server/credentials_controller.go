package http_server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func credentialsController(ginCtx *gin.Context) {
	token, found := LocalCache.Get("token")
	if found {
		ginCtx.JSON(http.StatusOK, token)
	} else {
		msg := fmt.Sprintf("Credentials not found")
		ginCtx.JSON(http.StatusNotFound, gin.H{"error": msg})
	}
}
