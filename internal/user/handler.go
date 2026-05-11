package user

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	DB *sql.DB
}

// DTOs
type AddLibraryRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Status  string `json:"status" binding:"required"` // reading, completed, plan_to_read
}

type UpdateProgressRequest struct {
	MangaID        string `json:"manga_id" binding:"required"`
	CurrentChapter int    `json:"chapter" binding:"required"`
}

// Thêm manga vào thư viện cá nhân
func (h *UserHandler) AddToLibrary(c *gin.Context) {
	// Lấy user_id từ JWTMiddleware
	userID := c.GetString("user_id")

	var req AddLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Thêm vào user_progress
	_, err := h.DB.Exec(`
		INSERT OR REPLACE INTO user_progress (user_id, manga_id, current_chapter, status, updated_at) 
		VALUES (?, ?, COALESCE((SELECT current_chapter FROM user_progress WHERE user_id = ? AND manga_id = ?), 0), ?, CURRENT_TIMESTAMP)`,
		userID, req.MangaID, userID, req.MangaID, req.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to library"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully added to library"})
}

// Lấy danh sách truyện đang theo dõi của user
func (h *UserHandler) GetLibrary(c *gin.Context) {
	userID := c.GetString("user_id")
	statusFilter := c.Query("status") // lọc ?status=reading

	query := `
		SELECT up.manga_id, m.title, up.current_chapter, m.total_chapters, up.status, up.updated_at 
		FROM user_progress up 
		JOIN manga m ON up.manga_id = m.id 
		WHERE up.user_id = ?`

	var rows *sql.Rows
	var err error

	if statusFilter != "" {
		query += " AND up.status = ?"
		rows, err = h.DB.Query(query, userID, statusFilter)
	} else {
		rows, err = h.DB.Query(query, userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	type LibraryItem struct {
		MangaID        string    `json:"manga_id"`
		Title          string    `json:"title"`
		CurrentChapter int       `json:"current_chapter"`
		TotalChapters  int       `json:"total_chapters"`
		Status         string    `json:"status"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	var library []LibraryItem
	for rows.Next() {
		var item LibraryItem
		if err := rows.Scan(&item.MangaID, &item.Title, &item.CurrentChapter, &item.TotalChapters, &item.Status, &item.UpdatedAt); err != nil {
			continue
		}
		library = append(library, item)
	}

	if library == nil {
		library = []LibraryItem{}
	}

	c.JSON(http.StatusOK, gin.H{"library": library, "count": len(library)})
}

// Cập nhật chương đang đọc
func (h *UserHandler) UpdateProgress(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	// Cập nhật chapter và thời gian
	result, err := h.DB.Exec(`
		UPDATE user_progress 
		SET current_chapter = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = ? AND manga_id = ?`,
		req.CurrentChapter, userID, req.MangaID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found in your library. Add it first."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Progress updated successfully",
		"manga_id": req.MangaID,
		"chapter": req.CurrentChapter,
	})

	// TODO: Nơi đây sẽ kích hoạt TCP Broadcast trong Giai đoạn 2
}