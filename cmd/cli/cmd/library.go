package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var mangaID string
var statusFlag string

var libraryCmd = &cobra.Command{Use: "library", Short: "Manage your manga library"}

var addLibraryCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a manga to your library",
	Run: func(cmd *cobra.Command, args []string) {
		token := loadToken()
		reqBody, _ := json.Marshal(map[string]string{
			"manga_id": mangaID,
			"status":   statusFlag,
		})
		
		req, _ := http.NewRequest("POST", "http://localhost:8080/users/library", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		
		client := &http.Client{}
		resp, _ := client.Do(req)
		
		if resp.StatusCode == http.StatusOK {
			fmt.Println("✓ Đã thêm truyện vào thư viện!")
		} else {
			fmt.Println("❌ Thêm thất bại. Hãy kiểm tra lại Manga ID.")
		}
	},
}

var listLibraryCmd = &cobra.Command{
	Use:   "list",
	Short: "View your library",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📚 Đang tải thư viện của bạn...")
		// TODO: Gọi GET /users/library và dùng tabwriter in ra bảng (giống lệnh manga search)
	},
}

func init() {
	rootCmd.AddCommand(libraryCmd)
	libraryCmd.AddCommand(addLibraryCmd, listLibraryCmd)
	addLibraryCmd.Flags().StringVarP(&mangaID, "manga-id", "m", "", "Manga ID (e.g., mal-21)")
	addLibraryCmd.Flags().StringVarP(&statusFlag, "status", "s", "reading", "Status (reading/completed)")
	addLibraryCmd.MarkFlagRequired("manga-id")
}