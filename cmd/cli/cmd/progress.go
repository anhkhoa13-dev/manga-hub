package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/spf13/cobra"
)

var chapter int

var progressCmd = &cobra.Command{Use: "progress", Short: "Update and sync reading progress"}

var updateProgressCmd = &cobra.Command{
	Use:   "update",
	Short: "Update your current reading chapter",
	Run: func(cmd *cobra.Command, args []string) {
		token := loadToken()
		reqBody, _ := json.Marshal(map[string]interface{}{
			"manga_id": mangaID,
			"chapter":  chapter,
		})
		req, _ := http.NewRequest("PUT", "http://172.20.10.3:8080/users/progress", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		
		client := &http.Client{}
		resp, _ := client.Do(req)
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("✓ Đã cập nhật truyện %s lên chapter %d\n", mangaID, chapter)
		}
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Start a real-time TCP sync listener",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📡 Đang kết nối tới máy chủ đồng bộ TCP (Port 9090)...")
		conn, err := net.Dial("tcp", "172.20.10.3:9090")
		if err != nil {
			fmt.Println("❌ Không thể kết nối tới TCP Server.")
			return
		}
		defer conn.Close()

		fmt.Println("✓ Đã kết nối! Đang chờ dữ liệu đồng bộ (Nhấn Ctrl+C để thoát)...")
		
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			var update map[string]interface{}
			json.Unmarshal(scanner.Bytes(), &update)
			fmt.Printf("\n[TCP SYNC] User %s vừa cập nhật %s lên chapter %v!\n", 
				update["user_id"], update["manga_id"], update["chapter"])
		}
	},
}

func init() {
	rootCmd.AddCommand(progressCmd)
	progressCmd.AddCommand(updateProgressCmd, syncCmd)

	updateProgressCmd.Flags().StringVarP(&mangaID, "manga-id", "m", "", "Manga ID")
	updateProgressCmd.Flags().IntVarP(&chapter, "chapter", "c", 1, "Current Chapter")
	updateProgressCmd.MarkFlagRequired("manga-id")
}