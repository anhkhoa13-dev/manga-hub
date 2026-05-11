package tcp

import (
	"encoding/json"
	"log"
	"net"
)

type ProgressUpdate struct {
	UserID    string `json:"user_id"`
	MangaID   string `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Timestamp int64  `json:"timestamp"`
}

// Quản lý các kết nối và broadcast thông tin tiến độ
type ProgressSyncServer struct {
	Port        string
	Connections map[string]net.Conn
	Broadcast   chan ProgressUpdate
	Register    chan net.Conn
	Unregister  chan net.Conn
}

// Khởi tạo Server với các Channels
func NewServer(port string) *ProgressSyncServer {
	return &ProgressSyncServer{
		Port:        port,
		Connections: make(map[string]net.Conn),
		Broadcast:   make(chan ProgressUpdate),
		Register:    make(chan net.Conn),
		Unregister:  make(chan net.Conn),
	}
}

// Xử lý Channels
func (s *ProgressSyncServer) Run() {
	for {
		// Chờ tín hiệu từ nhiều Channel cùng một lúc
		select {

		case conn := <-s.Register:
			s.Connections[conn.RemoteAddr().String()] = conn
			log.Printf("[TCP] Client connected: %s. Total: %d", conn.RemoteAddr().String(), len(s.Connections))
		
		case conn := <-s.Unregister:
			if _, ok := s.Connections[conn.RemoteAddr().String()]; ok {
				delete(s.Connections, conn.RemoteAddr().String())
				conn.Close()
				log.Printf("[TCP] Client disconnected: %s. Total: %d", conn.RemoteAddr().String(), len(s.Connections))
			}

		case update := <-s.Broadcast:
			msg, err := json.Marshal(update)
			if err != nil {
				continue
			}
			msg = append(msg, '\n')

			for _, conn := range s.Connections {
				_, err := conn.Write(msg)
				if err != nil {
					log.Printf("[TCP] Failed to send to %s, disconnecting...", conn.RemoteAddr().String())
					s.Unregister <- conn
				}
			}
		}
	}
}

// Mở port và lắng nghe kết nối
func (s *ProgressSyncServer) Start() {
	listener, err := net.Listen("tcp", s.Port)
	if err != nil {
		log.Fatalf("[TCP] Failed to start server: %v", err)
	}
	defer listener.Close()
	log.Printf("[TCP] Sync Server is listening on %s", s.Port)

	// Chạy bộ quản lý state ở một Goroutine riêng
	go s.Run()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[TCP] Accept error: %v", err)
			continue
		}
		// Thêm connection mới vào channel Register
		s.Register <- conn
	}
}