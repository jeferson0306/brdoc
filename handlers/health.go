package handlers

import (
	"net/http"

	"DataValidatorAPI/utils"

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
	cache := "unreachable"
	if utils.CacheHealthy() {
		cache = "ok"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"cache":  cache,
	})
}
