package main

import (
	"DataValidatorAPI/docs"
	"DataValidatorAPI/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func main() {
	docs.SwaggerInfo.BasePath = "/"
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/validate", func(c *gin.Context) {
		handlers.ValidateHandler(c.Writer, c.Request)
	})
	router.Run(":8080")
}
