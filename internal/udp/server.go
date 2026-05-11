package udp

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Notification struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// Quản lý danh sách client UDP
type NotificationServer struct {
	Port    string
	conn    *net.UDPConn
	clients map[string]*net.UDPAddr 
	mu      sync.RWMutex            // Khóa bảo vệ chống Data Race
}

func NewServer(port string) *NotificationServer {
	return &NotificationServer{
		Port:    port,
		clients: make(map[string]*net.UDPAddr),
	}
}

// Mở cổng UDP và lắng nghe đăng ký
func (s *NotificationServer) Start() {
	addr, err := net.ResolveUDPAddr("udp", s.Port)
	if err != nil {
		log.Fatalf("[UDP] Resolve error: %v", err)
	}

	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("[UDP] Failed to start server: %v", err)
	}
	defer s.conn.Close()

	log.Printf("[UDP] Notification Server is listening on %s", s.Port)

	buffer := make([]byte, 1024)
	for {
		
		n, clientAddr, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("[UDP] Read error: %v", err)
			continue
		}

		msg := strings.TrimSpace(string(buffer[:n]))
		
		// Lưu IP/Port của client
		if msg == "SUBSCRIBE" {
			s.mu.Lock()
			s.clients[clientAddr.String()] = clientAddr
			s.mu.Unlock()
			
			log.Printf("[UDP] New client registered: %s. Total: %d", clientAddr.String(), len(s.clients))
			
			s.conn.WriteToUDP([]byte("REGISTER_SUCCESS\n"), clientAddr)
		}
	}
}

// Gửi thông báo đến TẤT CẢ client đã đăng ký
func (s *NotificationServer) Broadcast(mangaID, message string) {
	notification := Notification{
		Type:      "NEW_CHAPTER",
		MangaID:   mangaID,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("[UDP] Marshal error: %v", err)
		return
	}
	data = append(data, '\n')

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, clientAddr := range s.clients {
		_, err := s.conn.WriteToUDP(data, clientAddr)
		if err != nil {
			log.Printf("[UDP] Failed to send to %s: %v", clientAddr.String(), err)
		}
	}
	log.Printf("[UDP] Broadcasted notification to %d clients", len(s.clients))
}