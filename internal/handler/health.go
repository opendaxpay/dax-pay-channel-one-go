package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daxpay.open/dax-pay-channel-one-go/internal/version"
)

// Health：对标 Spring Actuator /actuator/health
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "UP",
		"version":   version.Version,
		"gitCommit": version.GitCommit,
		"buildTime": version.BuildTime,
	})
}
