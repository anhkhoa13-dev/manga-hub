package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

type ChatMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// Quản lý danh sách client và phát sóng tin nhắn
type ChatHub struct {
	Clients    map[*websocket.Conn]string
	Broadcast  chan ChatMessage
	Register   chan struct {
		Conn     *websocket.Conn
		Username string
	}
	Unregister chan *websocket.Conn
	mu         sync.Mutex // Khóa bảo vệ chống Data
}

func NewHub() *ChatHub {
	return &ChatHub{
		Clients:   make(map[*websocket.Conn]string),
		Broadcast: make(chan ChatMessage),
		Register: make(chan struct {
			Conn     *websocket.Conn
			Username string
		}),
		Unregister: make(chan *websocket.Conn),
	}
}

// Khởi chạy Hub
func (h *ChatHub) Run() {
	for {
		select {
		case clientData := <-h.Register:
			h.mu.Lock()
			h.Clients[clientData.Conn] = clientData.Username
			h.mu.Unlock()
			log.Printf("[WebSocket] %s joined chat. Total: %d", clientData.Username, len(h.Clients))

		case conn := <-h.Unregister:
			h.mu.Lock()
			if username, ok := h.Clients[conn]; ok {
				delete(h.Clients, conn)
				conn.Close()
				log.Printf("[WebSocket] %s left chat. Total: %d", username, len(h.Clients))
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.Lock()
			msgBytes, _ := json.Marshal(message)
			for conn := range h.Clients {
				err := conn.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					conn.Close()
					delete(h.Clients, conn)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Kết nối WebSocket từ client và đăng ký vào Hub
func (h *ChatHub) HandleConnection(c *gin.Context) {
	// Lấy thông tin user từ JWT Middleware
	userID := c.GetString("user_id")
	username := c.GetString("username")
	if username == "" {
		username = "Anonymous"
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	// Đăng ký client vào Hub
	h.Register <- struct {
		Conn     *websocket.Conn
		Username string
	}{Conn: conn, Username: username}

	go func() {
		defer func() {
			h.Unregister <- conn
		}()
		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				break 
			}

			chatMsg := ChatMessage{
				UserID:    userID,
				Username:  username,
				Message:   string(msgBytes),
				Timestamp: time.Now().Unix(),
			}
			
			h.Broadcast <- chatMsg
		}
	}()
}