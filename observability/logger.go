// Package observability wires request logging that is safe to keep.
//
// The data this service receives — CPF, RG, card numbers — is exactly the data
// that must never reach a log file, so logging is built around omitting the
// value rather than around printing the request.
package observability

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey is where the per-request id is stored for handlers to read.
const RequestIDKey = "request_id"

// Logger returns the process logger. JSON in production so a log platform can
// index it; text locally, where a human is reading.
func Logger() *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
		level = slog.LevelDebug
	}

	if gin.Mode() == gin.ReleaseMode {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// RequestLogger records one line per request, deliberately without the query
// string.
//
// gin.Logger() formats its line as `path + "?" + RawQuery`, which for this
// service means every CPF, RG and card number it receives is written to stdout
// in the clear. Only the route, the outcome and the timing are recorded here;
// the value being validated never is.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-Id", requestID)

		c.Next()

		// Which parameter was asked for is useful; its value is not.
		parameters := make([]string, 0, len(c.Request.URL.Query()))
		for key := range c.Request.URL.Query() {
			parameters = append(parameters, key)
		}

		logger.Info("request",
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("route", c.FullPath()),
			slog.Any("parameters", parameters),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
		)
	}
}
