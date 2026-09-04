package handlers

import (
	"net/http"

	"github.com/jeferson0306/api-data-validator/internal/cache"

	"github.com/gin-gonic/gin"
)

// HealthHandler reports whether the service is up and whether its cache is
// reachable.
//
// The cache is reported but never gates the status: validation is pure
// computation and stays correct with Redis down, so a degraded cache must not
// take the service out of a load balancer. It exists to make the degradation
// visible — a silently unreachable Redis is how this service spent its first
// deployment paying latency for a cache that never returned a hit.
func HealthHandler(c *gin.Context) {
	// A local named "cache" would shadow the package; more importantly, there
	// are now three states rather than two. "disabled" is not a fault.
	status := "disabled"
	switch {
	case !cache.Enabled():
	case cache.Healthy():
		status = "ok"
	default:
		status = "unreachable"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"cache":  status,
	})
}
