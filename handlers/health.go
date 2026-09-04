package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler reports that the service is up.
//
// There is nothing else to report. Validation is arithmetic with no
// dependencies: if the process is answering, it is working. A health check that
// enumerates subsystems the service does not have is theatre.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
