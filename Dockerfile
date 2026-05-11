# ==========================================
# Bước 1: Build Stage (Môi trường biên dịch)
# ==========================================
FROM golang:1.22-alpine AS builder

# Cài đặt gcc và musl-dev (Bắt buộc để compile SQLite qua CGO)
RUN apk add --no-cache gcc musl-dev

# Thiết lập thư mục làm việc trong container
WORKDIR /app

# Copy file quản lý thư viện và tải dependencies trước (tận dụng cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source code vào container
COPY . .

# Biên dịch ứng dụng API Server
# CGO_ENABLED=1 là bắt buộc cho SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/mangahub-server ./cmd/api-server/main.go

# ==========================================
# Bước 2: Final Stage (Môi trường chạy thật)
# ==========================================
FROM alpine:latest

WORKDIR /app

# Khởi tạo thư mục data để chứa file SQLite
RUN mkdir -p data

# Chỉ copy file thực thi (binary) từ bước 1 sang, giúp Image cực kỳ nhẹ
COPY --from=builder /app/mangahub-server .
COPY --from=builder /app/.env ./.env 

# Khai báo 5 cổng mạng của hệ thống MangaHub
EXPOSE 8080 
EXPOSE 9090 
EXPOSE 9091/udp 
EXPOSE 9092 
EXPOSE 9093

# Lệnh khởi chạy server
CMD ["./mangahub-server"]