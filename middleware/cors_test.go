package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.GET("/validate", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	router.OPTIONS("/validate", func(c *gin.Context) { c.String(http.StatusOK, "unreachable") })
	return router
}

func request(t *testing.T, method, origin string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/validate?cpf=529.982.247-25", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	newRouter().ServeHTTP(recorder, req)
	return recorder
}

func TestAllowedOriginIsEchoed(t *testing.T) {
	res := request(t, http.MethodGet, "https://jeferson0306.github.io")
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://jeferson0306.github.io" {
		t.Fatalf("expected the origin to be echoed, got %q", got)
	}
}

func TestUnknownOriginGetsNoHeader(t *testing.T) {
	res := request(t, http.MethodGet, "https://example.com")
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header, got %q", got)
	}
}

func TestVaryIsAlwaysSet(t *testing.T) {
	// Without this a shared cache can serve one site the header computed for another.
	res := request(t, http.MethodGet, "https://example.com")
	if got := res.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("expected Vary: Origin, got %q", got)
	}
}

func TestPreflightIsAnsweredWithoutReachingTheHandler(t *testing.T) {
	res := request(t, http.MethodOptions, "https://jeferson0306.github.io")
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.Code)
	}
	if res.Body.String() != "" {
		t.Fatalf("preflight reached the handler: %q", res.Body.String())
	}
}
