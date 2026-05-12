package grpc

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/anhkhoa13-dev/mangahub/internal/tcp"
	pb "github.com/anhkhoa13-dev/mangahub/proto"
)

type MangaGrpcServer struct {
	pb.UnimplementedMangaServiceServer
	DB        *sql.DB
	Broadcast chan tcp.ProgressUpdate
}

// Lấy chi tiết một bộ truyện qua gRPC
func (s *MangaGrpcServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.MangaResponse, error) {
	var res pb.MangaResponse
	var genresJSON string

	err := s.DB.QueryRow("SELECT id, title, author, genres, status, total_chapters, description FROM manga WHERE id = ?", req.Id).
		Scan(&res.Id, &res.Title, &res.Author, &genresJSON, &res.Status, &res.TotalChapters, &res.Description)

	if err != nil {
		return nil, err
	}
	
	res.Genres = genresJSON 

	return &res, nil
}

// Tìm kiếm truyện
func (s *MangaGrpcServer) SearchManga(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	searchTerm := "%" + req.Query + "%"
	rows, err := s.DB.Query("SELECT id, title, author, status, total_chapters FROM manga WHERE title LIKE ? OR author LIKE ? LIMIT 10", searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*pb.MangaResponse
	for rows.Next() {
		var m pb.MangaResponse
		if err := rows.Scan(&m.Id, &m.Title, &m.Author, &m.Status, &m.TotalChapters); err == nil {
			results = append(results, &m)
		}
	}

	return &pb.SearchResponse{Results: results}, nil
}

// Cập nhật chương đang đọc và phát sóng TCP
func (s *MangaGrpcServer) UpdateProgress(ctx context.Context, req *pb.ProgressRequest) (*pb.ProgressResponse, error) {
	result, err := s.DB.Exec(`
		UPDATE user_progress 
		SET current_chapter = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = ? AND manga_id = ?`,
		req.Chapter, req.UserId, req.MangaId)

	if err != nil {
		return &pb.ProgressResponse{Success: false, Message: "Database error"}, nil
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return &pb.ProgressResponse{Success: false, Message: "Manga not found in user library"}, nil
	}

	// Phát sóng tiến trình lên TCP Sync Server
	if s.Broadcast != nil {
		s.Broadcast <- tcp.ProgressUpdate{
			UserID:    req.UserId,
			MangaID:   req.MangaId,
			Chapter:   int(req.Chapter),
			Timestamp: time.Now().Unix(),
		}
	}

	return &pb.ProgressResponse{Success: true, Message: "Progress updated successfully via gRPC"}, nil
}

// Kích hoạt gRPC Server
func Start(port string, db *sql.DB, broadcast chan tcp.ProgressUpdate) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("[gRPC] Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	
	// Đăng ký Service vào gRPC Server
	pb.RegisterMangaServiceServer(grpcServer, &MangaGrpcServer{
		DB:        db,
		Broadcast: broadcast,
	})

	log.Printf("[gRPC] Internal Service is listening on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[gRPC] Failed to serve: %v", err)
	}
}