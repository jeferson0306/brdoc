package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func limitedRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(RateLimit())
	router.GET("/validate", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}

func call(router *gin.Engine, path, address string) int {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.RemoteAddr = address + ":54321"
	router.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestBurstIsAllowedThenThrottled(t *testing.T) {
	router := limitedRouter(t)

	// The burst is spent first; what follows is refilled at the configured rate,
	// which over a tight loop means refused.
	allowed := 0
	for i := 0; i < defaultBurst+40; i++ {
		if call(router, "/validate", "203.0.113.10") == http.StatusOK {
			allowed++
		}
	}

	if allowed == 0 {
		t.Fatal("no request got through; the limiter is refusing everything")
	}
	if allowed > defaultBurst+5 {
		t.Fatalf("the limiter let %d requests through, expected around the burst of %d", allowed, defaultBurst)
	}
}

func TestClientsAreLimitedIndependently(t *testing.T) {
	router := limitedRouter(t)

	for i := 0; i < defaultBurst+40; i++ {
		call(router, "/validate", "203.0.113.10")
	}

	// A second address must be unaffected by the first exhausting its bucket.
	if code := call(router, "/validate", "203.0.113.99"); code != http.StatusOK {
		t.Fatalf("a different address was throttled by its neighbour: got %d", code)
	}
}

// A throttled monitor reports an outage that is not happening.
func TestHealthIsNeverThrottled(t *testing.T) {
	router := limitedRouter(t)

	for i := 0; i < defaultBurst+200; i++ {
		if code := call(router, "/health", "203.0.113.10"); code != http.StatusOK {
			t.Fatalf("health was throttled after %d calls", i)
		}
	}
}

func TestThrottledResponseTellsTheCallerWhen(t *testing.T) {
	router := limitedRouter(t)

	var recorder *httptest.ResponseRecorder
	for i := 0; i < defaultBurst+40; i++ {
		recorder = httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/validate", nil)
		request.RemoteAddr = "203.0.113.55:1234"
		router.ServeHTTP(recorder, request)
		if recorder.Code == http.StatusTooManyRequests {
			break
		}
	}

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatal("never reached the limit")
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("a 429 without Retry-After leaves the caller guessing")
	}
}
