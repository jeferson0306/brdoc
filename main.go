package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jeferson0306/api-data-validator/docs"
	"github.com/jeferson0306/api-data-validator/handlers"
	"github.com/jeferson0306/api-data-validator/middleware"
	"github.com/jeferson0306/api-data-validator/observability"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const shutdownGrace = 10 * time.Second

func main() {
	logger := observability.Logger()
	slog.SetDefault(logger)

	docs.SwaggerInfo.BasePath = "/"

	// gin.New(), not gin.Default(): the default logger writes the query string,
	// which for this service means writing CPFs and card numbers to stdout.
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(observability.RequestLogger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())

	configureClientAddressing(router, logger)

	router.GET("/health", handlers.HealthHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/validate", func(c *gin.Context) {
		handlers.ValidateHandler(c.Writer, c.Request)
	})
	router.POST("/validate/batch", func(c *gin.Context) {
		handlers.BatchHandler(c.Writer, c.Request)
	})

	server := &http.Server{
		Addr:              ":" + port(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Every deploy sends SIGTERM. Without this, requests in flight are killed
	// mid-response and the caller sees a connection reset rather than an answer.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", slog.String("error", err.Error()))
	}
}

// configureClientAddressing decides whose address the rate limiter counts.
//
// Behind a platform proxy the socket address belongs to the proxy, so without
// configuration every caller shares one bucket and the limit becomes global.
// The obvious fix — trusting X-Forwarded-For — is worse: on a public host that
// header is written by whoever is calling, so a limit keyed on it is a limit
// anyone can step around by changing a string.
//
// So it is explicit. TRUSTED_PLATFORM=cloudflare reads CF-Connecting-IP, which
// Cloudflare overwrites on every request and a caller therefore cannot forge;
// Render fronts services with Cloudflare, which is what makes this the right
// setting there. TRUSTED_PROXIES takes CIDRs for anything else. Unset, nothing
// is trusted: the limit is global, which is safe and honest rather than
// silently forgeable.
func configureClientAddressing(router *gin.Engine, logger *slog.Logger) {
	if platform := strings.ToLower(os.Getenv("TRUSTED_PLATFORM")); platform == "cloudflare" {
		router.TrustedPlatform = gin.PlatformCloudflare
		logger.Info("client address taken from CF-Connecting-IP")
		return
	}

	if proxies := os.Getenv("TRUSTED_PROXIES"); proxies != "" {
		if err := router.SetTrustedProxies(strings.Split(proxies, ",")); err != nil {
			logger.Error("invalid TRUSTED_PROXIES", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("client address taken from X-Forwarded-For", slog.String("trusted", proxies))
		return
	}

	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Error("could not clear trusted proxies", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Warn("no trusted proxy configured; the rate limit applies to all callers together")
}

// port honours the value the platform assigns; Render sets PORT and would
// otherwise route to a port nothing is listening on.
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
