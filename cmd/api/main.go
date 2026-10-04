package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "libratrack/docs"

	"libratrack/internal/handler"
	"libratrack/internal/middleware"
	"libratrack/internal/repository"
	"net/http"
)

// @title Libratrack API
// @version 1.0
// @description This is a sample server for Libratrack
// @BasePath /api/v1

func main() {
	r := gin.New()

	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())

	repo := repository.NewInMemoryBookRepository()
	bookHandler := handler.NewBookHandler(repo)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/books", bookHandler.CreateBook)
		v1.GET("/books", bookHandler.ListBooks)
		v1.GET("/books/:id", bookHandler.GetBook)
		v1.PUT("/books/:id", bookHandler.UpdateBook)
		v1.DELETE("/books/:id", bookHandler.DeleteBook)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}
