package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"DataValidatorAPI/docs"
	"DataValidatorAPI/handlers"
	"DataValidatorAPI/middleware"
	"DataValidatorAPI/observability"

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

// port honours the value the platform assigns; Render sets PORT and would
// otherwise route to a port nothing is listening on.
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
