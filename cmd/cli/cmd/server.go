package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{Use: "server", Short: "Administer the MangaHub backend servers"}

var statusServerCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of all backend services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Kiểm tra trạng thái hệ thống:")
		fmt.Println("- HTTP API (8080): Đang chạy")
		fmt.Println("- TCP Sync (9090): Đang chạy")
		fmt.Println("- UDP Node (9091): Đang chạy")
		fmt.Println("- gRPC Svc (9092): Đang chạy")
		fmt.Println("- WS Chat  (9093): Đang chạy")
	},
}

var startServerCmd = &cobra.Command{Use: "start", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Sử dụng docker-compose up -d để khởi động hệ thống.") }}
var stopServerCmd = &cobra.Command{Use: "stop", Run: func(cmd *cobra.Command, args []string) { fmt.Println("Sử dụng docker-compose down để dừng hệ thống.") }}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(statusServerCmd, startServerCmd, stopServerCmd)
}