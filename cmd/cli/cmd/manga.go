package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var mangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Manga management commands",
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for manga",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := loadToken()
		if token == "" {
			fmt.Println("❌ Vui lòng đăng nhập trước: mangahub auth login")
			return
		}

		query := url.QueryEscape(args[0])
		req, _ := http.NewRequest("GET", "http://172.20.10.3:8080/manga?q="+query, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("❌ Lỗi kết nối đến server.")
			return
		}
		defer resp.Body.Close()

		var result struct {
			Results []struct {
				ID            string `json:"id"`
				Title         string `json:"title"`
				Author        string `json:"author"`
				Status        string `json:"status"`
				TotalChapters int    `json:"total_chapters"`
			} `json:"results"`
			Count int `json:"count"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		fmt.Printf("Searching for \"%s\"...\n", args[0])
		fmt.Printf("Found %d results:\n\n", result.Count)

		if result.Count == 0 {
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tAUTHOR\tSTATUS\tCHAPTERS")
		fmt.Fprintln(w, "--\t-----\t------\t------\t--------")
		for _, m := range result.Results {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", m.ID, m.Title, m.Author, m.Status, m.TotalChapters)
		}
		w.Flush()
		fmt.Println("\nUse 'mangahub library add --manga-id <id>' to add to your library")
	},
}

func init() {
	rootCmd.AddCommand(mangaCmd)
	mangaCmd.AddCommand(searchCmd)
}