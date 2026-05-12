package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{Use: "chat", Short: "MangaHub community chat"}

var joinChatCmd = &cobra.Command{
	Use:   "join",
	Short: "Join the live WebSocket chat room",
	Run: func(cmd *cobra.Command, args []string) {
		token := loadToken()
		if token == "" {
			fmt.Println("❌ Vui lòng đăng nhập trước!")
			return
		}

		header := http.Header{}
		header.Add("Authorization", "Bearer "+token)

		fmt.Println("🔌 Đang kết nối phòng chat WebSocket (Port 9093)...")
		conn, _, err := websocket.DefaultDialer.Dial("ws://172.20.10.3:9093/chat", header)
		if err != nil {
			fmt.Println("❌ Lỗi kết nối WebSocket:", err)
			return
		}
		defer conn.Close()
		fmt.Println("✓ Đã vào phòng chat! Hãy gõ tin nhắn và nhấn Enter. (Nhấn Ctrl+C để thoát)\n---")

		// Goroutine để LẮNG NGHE tin nhắn từ server
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					fmt.Println("\r🔌 Mất kết nối tới server.")
					return
				}

				var msgData struct {
					Username string `json:"username"`
					Message  string `json:"message"`
				}

				err = json.Unmarshal(message, &msgData)
				if err == nil {
					// KIỂM TRA NẾU LÀ TIN NHẮN HỆ THỐNG (JOIN/LEAVE)
					if msgData.Username == "System" {
						// In thông báo căn giữa hoặc in nghiêng cho đẹp
						fmt.Printf("\r\t%s\n> ", msgData.Message)
					} else {
						// In ra format chuẩn: [username]: message
						fmt.Printf("\r[%s]: %s\n> ", msgData.Username, msgData.Message)
					}
				} else {
					fmt.Printf("\r⚠️ Lỗi JSON: %v | Raw: %s\n> ", err, string(message))
				}
			}
		}()

		// Vòng lặp chính để GỬI tin nhắn từ bàn phím
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")
		for scanner.Scan() {
			text := scanner.Text()
			if text != "" {
				err := conn.WriteMessage(websocket.TextMessage, []byte(text))
				if err != nil {
					fmt.Println("❌ Không thể gửi tin nhắn.")
					break
				}
			}
			fmt.Print("> ")
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.AddCommand(joinChatCmd)
}