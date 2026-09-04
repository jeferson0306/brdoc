package observability

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const sensitiveCPF = "529.982.247-25"

func logOneRequest(t *testing.T, target string) string {
	t.Helper()

	var captured bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&captured, nil))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger(logger))
	router.GET("/validate", func(c *gin.Context) { c.Status(http.StatusOK) })

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, target, nil))
	return captured.String()
}

// The whole point of the middleware: gin's default logger writes
// `path + "?" + RawQuery`, which for this service is a CPF in the clear.
func TestValueNeverReachesTheLog(t *testing.T) {
	line := logOneRequest(t, "/validate?cpf="+sensitiveCPF)

	for _, forbidden := range []string{sensitiveCPF, "52998224725", "529.982"} {
		if strings.Contains(line, forbidden) {
			t.Fatalf("the log line contains the value being validated: %s", line)
		}
	}
}

func TestParameterNameIsKept(t *testing.T) {
	// Which parameter was asked for is operationally useful and identifies nobody.
	line := logOneRequest(t, "/validate?cpf="+sensitiveCPF)
	if !strings.Contains(line, `"cpf"`) {
		t.Fatalf("expected the parameter name to be logged, got: %s", line)
	}
}

func TestRouteIsLoggedNotTheRawPath(t *testing.T) {
	line := logOneRequest(t, "/validate?cpf="+sensitiveCPF)
	if !strings.Contains(line, `"route":"/validate"`) {
		t.Fatalf("expected the matched route, got: %s", line)
	}
}

func TestRequestIDIsIssuedAndReturned(t *testing.T) {
	var captured bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&captured, nil))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger(logger))
	router.GET("/validate", func(c *gin.Context) { c.Status(http.StatusOK) })

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/validate?cpf=1", nil))

	header := recorder.Header().Get("X-Request-Id")
	if header == "" {
		t.Fatal("expected an X-Request-Id header so a caller can quote it in a report")
	}
	if !strings.Contains(captured.String(), header) {
		t.Fatal("the id returned to the caller does not match the one in the log")
	}
}
