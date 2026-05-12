package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	mangaID    string
	statusFlag string
)

var libraryCmd = &cobra.Command{Use: "library", Short: "Manage your manga library"}

var addLibraryCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a manga to your library",
	Run: func(cmd *cobra.Command, args []string) {
		token := loadToken()
		if token == "" {
			fmt.Println("❌ Vui lòng đăng nhập trước!")
			return
		}

		reqBody, _ := json.Marshal(map[string]string{
			"manga_id": mangaID,
			"status":   statusFlag,
		})
		
		req, _ := http.NewRequest("POST", "http://172.20.10.3:8080/users/library", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("❌ Lỗi kết nối đến server.")
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("✓ Đã thêm truyện %s vào thư viện với trạng thái: %s\n", mangaID, statusFlag)
		} else {
			fmt.Println("❌ Thêm thất bại. Truyện có thể đã có trong thư viện.")
		}
	},
}

var listLibraryCmd = &cobra.Command{
	Use:   "list",
	Short: "View your personal library and reading progress",
	Run: func(cmd *cobra.Command, args []string) {
		token := loadToken()
		if token == "" {
			fmt.Println("❌ Vui lòng đăng nhập trước!")
			return
		}

		fmt.Println("📚 Đang tải thư viện của bạn...")

		req, _ := http.NewRequest("GET", "http://172.20.10.3:8080/users/library", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("❌ Lỗi kết nối đến server.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("❌ Không thể lấy dữ liệu thư viện.")
			return
		}

		var result struct {
			Library []struct {
				MangaID        string `json:"manga_id"`
				Title          string `json:"title"`
				Status         string `json:"status"`
				CurrentChapter int    `json:"current_chapter"`
			} `json:"library"`
		}

		json.NewDecoder(resp.Body).Decode(&result)

		if len(result.Library) == 0 {
			fmt.Println("📭 Thư viện của bạn đang trống. Hãy dùng 'mangahub library add' để thêm truyện nhé!")
			return
		}

		fmt.Printf("\n--- Thư viện của bạn (%d bộ) ---\n\n", len(result.Library))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tREADING")
		fmt.Fprintln(w, "--\t-----\t------\t-------")
		for _, item := range result.Library {
			title := item.Title
			if title == "" {
				title = "(Đang cập nhật tên)"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\tCh. %d\n", 
				item.MangaID, title, item.Status, item.CurrentChapter)
		}
		w.Flush()
		fmt.Println("")
	},
}

func init() {
	rootCmd.AddCommand(libraryCmd)
	libraryCmd.AddCommand(addLibraryCmd, listLibraryCmd)

	addLibraryCmd.Flags().StringVarP(&mangaID, "manga-id", "m", "", "ID of the manga")
	addLibraryCmd.Flags().StringVarP(&statusFlag, "status", "s", "reading", "Reading status (reading, completed, dropped)")
	addLibraryCmd.MarkFlagRequired("manga-id")
}