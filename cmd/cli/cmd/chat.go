package cmd

import (
	"bufio"
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
		conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:9093/chat", header)
		if err != nil {
			fmt.Println("❌ Lỗi kết nối WebSocket:", err)
			return
		}
		defer conn.Close()
		fmt.Println("✓ Đã vào phòng chat! Hãy gõ tin nhắn và nhấn Enter. (Nhấn Ctrl+C để thoát)\n---")

		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return
				}
				// Xóa dòng nhập hiện tại, in tin nhắn mới ra, rồi in lại dấu nhắc
				fmt.Printf("\r💬 %s\n> ", string(message))
			}
		}()

		// Vòng lặp chính để GỬI tin nhắn từ bàn phím
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")
		for scanner.Scan() {
			text := scanner.Text()
			if text != "" {
				conn.WriteMessage(websocket.TextMessage, []byte(text))
			}
			fmt.Print("> ")
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.AddCommand(joinChatCmd) 
}