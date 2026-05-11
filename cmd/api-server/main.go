package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/anhkhoa13-dev/mangahub/grpc"
	"github.com/anhkhoa13-dev/mangahub/internal/auth"
	"github.com/anhkhoa13-dev/mangahub/internal/manga"
	"github.com/anhkhoa13-dev/mangahub/internal/tcp"
	"github.com/anhkhoa13-dev/mangahub/internal/udp"
	"github.com/anhkhoa13-dev/mangahub/internal/user"
	"github.com/anhkhoa13-dev/mangahub/internal/websocket"
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

	// Setup TCP Sync Server
	tcpServer := tcp.NewServer(":9090")
	go tcpServer.Start()

	// Setup UDP Notification Server
	udpServer := udp.NewServer(":9091") 
	go udpServer.Start()

	chatHub := websocket.NewHub()
	go chatHub.Run()

	// Setup WebSocket
	go func() {
		wsRouter := gin.Default()
		
		wsRouter.GET("/chat", auth.JWTMiddleware(), chatHub.HandleConnection)
		
		log.Println("[WebSocket] Chat Server is listening on :9093")
		if err := wsRouter.Run(":9093"); err != nil {
			log.Fatalf("Failed to run WebSocket server: %v", err)
		}
	}()

	// Setup gRPC Server
	go grpc.Start(":9092", db, tcpServer.Broadcast)

	// Setup handlers
	authHandler := &auth.AuthHandler{DB: db}
	mangaHandler := &manga.MangaHandler{DB: db}
	userHandler := &user.UserHandler{
		DB:        db,
		Broadcast: tcpServer.Broadcast, 
	}
	
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
		mangaGroup := protectedGroup.Group("/manga")
		{
			mangaGroup.GET("/", mangaHandler.SearchManga)
			mangaGroup.GET("/:id", mangaHandler.GetMangaDetails)
		}

		userGroup := protectedGroup.Group("/users")
		{
			userGroup.POST("/library", userHandler.AddToLibrary)
			userGroup.GET("/library", userHandler.GetLibrary)
			userGroup.PUT("/progress", userHandler.UpdateProgress)
		}
	}

	adminGroup := r.Group("/admin")
	{
		adminGroup.POST("/notify", func(c *gin.Context) {
			var req struct {
				MangaID string `json:"manga_id" binding:"required"`
				Message string `json:"message" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format"})
				return
			}

			udpServer.Broadcast(req.MangaID, req.Message)

			c.JSON(http.StatusOK, gin.H{"status": "Notification broadcasted"})
		})
	}
	
	log.Println("Starting HTTP API Server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}