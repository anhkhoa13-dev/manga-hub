package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/anhkhoa13-dev/mangahub/pkg/database"
)

// Jikan
type JikanResponse struct {
	Data []JikanManga `json:"data"`
}

type JikanManga struct {
	MalID    int    `json:"mal_id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Chapters int    `json:"chapters"`
	Synopsis string `json:"synopsis"`
	Authors  []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Genres []struct {
		Name string `json:"name"`
	} `json:"genres"`
}

func main() {
	log.Println("Crawling manga data...")

	searchQueries := []string{"action", "romance", "mystery", "comedy"}
	totalInserted := 0
	client := &http.Client{Timeout: 10 * time.Second}

	// Kết nối db
	db, err := database.InitDB("../../data/mangahub.db")
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	// Insert statement
	stmt, err := db.Prepare("INSERT OR REPLACE INTO manga (id, title, author, genres, status, total_chapters, description) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Fetch manga
	for _, query := range searchQueries {
		
		apiURL := fmt.Sprintf("https://api.jikan.moe/v4/manga?q=%s&sfw=true&limit=15&order_by=popularity&sort=desc", url.QueryEscape(query))
		log.Printf("Searching for '%s' -> %s\n", query, apiURL)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			log.Printf("Failed to create request for %s: %v\n", query, err)
			continue
		}
		
		req.Header.Set("User-Agent", "MangaHub-StudentProject/1.0 (IT096IU)")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to fetch API for %s: %v\n", query, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("API Error for '%s'! Status Code: %d. Bỏ qua và đi tiếp...\n", query, resp.StatusCode)
			resp.Body.Close()
			continue
		}

		var result JikanResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Failed to decode JSON for %s: %v\n", query, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Lưu vào db
		count := 0
		for _, m := range result.Data {
			id := "mal-" + strconv.Itoa(m.MalID)
			author := "Unknown"
			if len(m.Authors) > 0 {
				author = m.Authors[0].Name
			}

			genreNames := []string{}
			for _, g := range m.Genres {
				genreNames = append(genreNames, g.Name)
			}
			genresJSON, _ := json.Marshal(genreNames)

			status := "ongoing"
			if m.Status == "Finished" {
				status = "completed"
			}

			_, err := stmt.Exec(id, m.Title, author, string(genresJSON), status, m.Chapters, m.Synopsis)
			if err == nil {
				count++
				totalInserted++
			}
		}
		log.Printf("-> Saved %d manga for '%s'\n", count, query)

		time.Sleep(2 * time.Second)
	}

	log.Printf("Successfully inserted %d manga into the database.\n", totalInserted)
}
	