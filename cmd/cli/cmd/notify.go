package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Lắng nghe thông báo mới từ hệ thống qua UDP",
	Run: func(cmd *cobra.Command, args []string) {

		serverAddr, err := net.ResolveUDPAddr("udp", "172.20.10.3:9091")
		if err != nil {
			fmt.Println("❌ Lỗi cấu hình địa chỉ UDP:", err)
			return
		}

		conn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			fmt.Println("❌ Không thể kết nối tới server UDP:", err)
			return
		}
		defer conn.Close()

		// Gửi tín hiệu SUBSCRIBE để server ghi nhận IP/Port của máy này
		_, err = conn.Write([]byte("SUBSCRIBE"))
		if err != nil {
			fmt.Println("❌ Gửi yêu cầu subscribe thất bại:", err)
			return
		}

		fmt.Println("🔔 Đang lắng nghe thông báo mới (Nhấn Ctrl+C để thoát)...")
		fmt.Println("-----------------------------------------------------------")

		buffer := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("❌ Lỗi nhận dữ liệu:", err)
				continue
			}

			var data struct {
				Type      string `json:"type"`
				MangaID   string `json:"manga_id"`
				Message   string `json:"message"`
				Timestamp int64  `json:"timestamp"`
			}

			err = json.Unmarshal(buffer[:n], &data)
			if err != nil {
				fmt.Printf("ℹ️ %s\n", string(buffer[:n]))
				continue
			}
			t := time.Unix(data.Timestamp, 0).Format("15:04:05")

			fmt.Printf("[%s] 📢 THÔNG BÁO: %s\n", t, data.Message)
			fmt.Printf("        ID Truyện: %s | Loại: %s\n\n", data.MangaID, data.Type)
		}
	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)
}