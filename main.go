package main

import (
	"DataValidatorAPI/docs"
	"DataValidatorAPI/handlers"
	"DataValidatorAPI/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func main() {
	docs.SwaggerInfo.BasePath = "/"
	router := gin.Default()
	// Browsers block a cross-origin fetch unless the service says otherwise.
	router.Use(middleware.CORS())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/validate", func(c *gin.Context) {
		handlers.ValidateHandler(c.Writer, c.Request)
	})
	router.Run(":8080")
}
