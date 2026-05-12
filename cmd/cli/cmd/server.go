package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/anhkhoa13-dev/mangahub/internal/auth"
	grpcServer "github.com/anhkhoa13-dev/mangahub/internal/grpc"
	"github.com/anhkhoa13-dev/mangahub/internal/manga"
	"github.com/anhkhoa13-dev/mangahub/internal/tcp"
	"github.com/anhkhoa13-dev/mangahub/internal/udp"
	"github.com/anhkhoa13-dev/mangahub/internal/user"
	"github.com/anhkhoa13-dev/mangahub/internal/websocket"
	"github.com/anhkhoa13-dev/mangahub/pkg/database"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Administer the MangaHub backend servers",
}

var statusServerCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of all backend services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Sử dụng lệnh 'mangahub server start' để chạy hệ thống.")
	},
}

var startServerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all MangaHub backend servers directly from CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Đang khởi động hệ thống MangaHub Servers...")

		// 1. Load file .env (nếu có)
		godotenv.Load(".env")

		// 2. Tự động tạo thư mục data ở nơi bạn gõ lệnh để lưu database
		os.MkdirAll("data", os.ModePerm)
		dbPath := filepath.Join("data", "mangahub.db")

		db, err := database.InitDB(dbPath)
		if err != nil {
			log.Fatalf("❌ Failed to init database: %v", err)
		}
		defer db.Close()

		// 3. KHỞI CHẠY CÁC SERVER NGẦM
		tcpServer := tcp.NewServer(":9090")
		go tcpServer.Start()

		udpServer := udp.NewServer(":9091")
		go udpServer.Start()

		chatHub := websocket.NewHub()
		go chatHub.Run()
		
		go func() {
			wsRouter := gin.Default()
			wsRouter.GET("/chat", auth.JWTMiddleware(), chatHub.HandleConnection)
			if err := wsRouter.Run(":9093"); err != nil {
				log.Fatalf("❌ WS server error: %v", err)
			}
		}()

		go grpcServer.Start(":9092", db, tcpServer.Broadcast)

		// 4. KHỞI CHẠY HTTP SERVE
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()

		authHandler := &auth.AuthHandler{DB: db}
		mangaHandler := &manga.MangaHandler{DB: db}
		userHandler := &user.UserHandler{DB: db, Broadcast: tcpServer.Broadcast}

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
				var notifyData struct {
					MangaID string `json:"manga_id"`
					Message string `json:"message"`
				}

				if err := c.ShouldBindJSON(&notifyData); err != nil {
					c.JSON(400, gin.H{"error": "Invalid JSON format"})
					return
				}
				udpServer.Broadcast(notifyData.MangaID, notifyData.Message)

				c.JSON(200, gin.H{
					"status": "Notification sent via UDP!",
					"data":   notifyData,
				})
			})
		}

		log.Println("✅ All systems go! HTTP API Server is listening on :8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Kiểm tra trạng thái thực tế của các dịch vụ MangaHub",
	Run: func(cmd *cobra.Command, args []string) {
		// Danh sách các dịch vụ cần kiểm tra
		services := []struct {
			Name string
			Port string
			Type string
		}{
			{"HTTP API", "8080", "tcp"},
			{"TCP Sync", "9090", "tcp"},
			{"UDP Notify", "9091", "udp"},
			{"gRPC Svc", "9092", "tcp"},
			{"WS Chat", "9093", "tcp"},
		}

		fmt.Println("🏥 Đang kiểm tra sức khỏe hệ thống...")
		fmt.Println("---------------------------------------")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "SERVICE\tPORT\tPROTOCOL\tSTATUS")
		fmt.Fprintln(w, "-------\t----\t--------\t------")

		for _, s := range services {
			status := "🟢 ONLINE"
			
			// Thử kết nối tới port trong vòng 1 giây
			address := net.JoinHostPort("172.20.10.3", s.Port) // Dùng IP của bạn
			conn, err := net.DialTimeout(s.Type, address, 1*time.Second)
			
			if err != nil {
				status = "🔴 OFFLINE"
			} else {
				conn.Close()
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, s.Port, s.Type, status)
		}
		w.Flush()
		fmt.Println("")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(statusServerCmd, startServerCmd)
}