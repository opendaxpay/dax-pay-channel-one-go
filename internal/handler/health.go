package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health：对标 Spring Actuator /actuator/health
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}
