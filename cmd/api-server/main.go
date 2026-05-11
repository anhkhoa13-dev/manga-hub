package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/anhkhoa13-dev/mangahub/internal/auth"
	"github.com/anhkhoa13-dev/mangahub/internal/user"
	"github.com/anhkhoa13-dev/mangahub/pkg/database"
)

func main() {
	// Load .env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: No .env file found or error loading it. Using OS environment variables.")
	}

	// Setup database
	db, err := database.InitDB("../../data/mangahub.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("DB Connection active!")

	// Setup handlers
	authHandler := &auth.AuthHandler{DB: db}
	userHandler := &user.UserHandler{DB: db}
	
	// Setup Gin Router
	r := gin.Default()

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}
	protectedGroup := r.Group("/")
	protectedGroup.Use(auth.JWTMiddleware())
	{
		mangaGroup := r.Group("/manga")
		{
			mangaGroup.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Search manga"}) })
			mangaGroup.GET("/:id", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Get manga details"}) })
		}

		userGroup := protectedGroup.Group("/users")
		{
			userGroup.POST("/library", userHandler.AddToLibrary)
			userGroup.GET("/library", userHandler.GetLibrary)
			userGroup.PUT("/progress", userHandler.UpdateProgress)
		}
	}
	
	log.Println("Starting HTTP API Server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}