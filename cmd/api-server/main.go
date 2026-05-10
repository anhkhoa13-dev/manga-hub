package main

import (
	"log"

	"github.com/anhkhoa13-dev/mangahub/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// Setup database
	db, err := database.InitDB("../../data/mangahub.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("DB Connection active!")
	
	// Setup Gin Router
	r := gin.Default()

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Register endpoint"}) })
		authGroup.POST("/login", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Login endpoint"}) })
	}

	mangaGroup := r.Group("/manga")
	{
		mangaGroup.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Search manga"}) })
		mangaGroup.GET("/:id", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Get manga details"}) })
	}

	userGroup := r.Group("/users")
	{
		userGroup.POST("/library", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Add to library"}) })
		userGroup.GET("/library", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Get library"}) })
		userGroup.PUT("/progress", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Update progress"}) })
	}
	
	log.Println("Starting HTTP API Server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}