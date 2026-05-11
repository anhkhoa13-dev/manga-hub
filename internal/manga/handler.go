package manga

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/anhkhoa13-dev/mangahub/pkg/models"
)

type MangaHandler struct {
	DB *sql.DB
}

// Tìm kiếm và lấy danh sách truyện
func (h *MangaHandler) SearchManga(c *gin.Context) {
	query := c.Query("q") // /manga?q=
	
	var rows *sql.Rows
	var err error

	if query != "" {
		// Tìm kiếm theo tên hoặc tác giả
		searchTerm := "%" + query + "%"
		rows, err = h.DB.Query("SELECT id, title, author, genres, status, total_chapters, description FROM manga WHERE title LIKE ? OR author LIKE ?", searchTerm, searchTerm)
	} else {
		rows, err = h.DB.Query("SELECT id, title, author, genres, status, total_chapters, description FROM manga LIMIT 50")
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var mangas []models.Manga
	for rows.Next() {
		var m models.Manga
		var genresJSON string 
		
		if err := rows.Scan(&m.ID, &m.Title, &m.Author, &genresJSON, &m.Status, &m.TotalChapters, &m.Description); err != nil {
			continue
		}
	
		m.Genres = []string{genresJSON} 
		mangas = append(mangas, m)
	}

	c.JSON(http.StatusOK, gin.H{
		"results": mangas,
		"count":   len(mangas),
	})
}

func (h *MangaHandler) GetMangaDetails(c *gin.Context) {
	mangaID := c.Param("id")

	var m models.Manga
	var genresJSON string

	err := h.DB.QueryRow("SELECT id, title, author, genres, status, total_chapters, description FROM manga WHERE id = ?", mangaID).
		Scan(&m.ID, &m.Title, &m.Author, &genresJSON, &m.Status, &m.TotalChapters, &m.Description)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	m.Genres = []string{genresJSON}

	c.JSON(http.StatusOK, m)
}