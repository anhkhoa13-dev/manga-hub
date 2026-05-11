# MangaHub - Technical Documentation

**Môn học:** Network-centric Programming (IT096IU)
**Thành viên thực hiện:**
1. Đỗ Huỳnh Duy Tiến
2. Nguyễn Viết Anh Khoa

---

## 1. Architecture Overview (Tổng quan kiến trúc)

MangaHub là một hệ thống phân tán được thiết kế theo mô hình client-server, hỗ trợ đa giao thức (Multi-protocol) để tối ưu hóa từng loại tác vụ cụ thể. Hệ thống bao gồm 5 dịch vụ mạng chạy song song và một ứng dụng dòng lệnh (CLI) cho phía Client.

### 1.1. Các thành phần giao thức cốt lõi:
* **HTTP REST API (Port 8080):** Xử lý các tác vụ CRUD tiêu chuẩn như xác thực người dùng (Auth JWT), tìm kiếm truyện (Manga Search), và quản lý thư viện cá nhân.
* **TCP Progress Sync (Port 9090):** Xử lý đồng bộ hóa thời gian thực (Real-time). Khi một user cập nhật tiến trình đọc (chapter) qua HTTP, server sẽ tự động phát sóng (broadcast) trạng thái mới tới tất cả các TCP Client đang lắng nghe.
* **UDP Notification Node (Port 9091):** Hệ thống thông báo một chiều (Fire-and-forget). Được thiết kế để đẩy thông báo nhanh chóng (vd: "Có chapter mới") tới các client đã Subscribe, không yêu cầu thiết lập kết nối liên tục nhằm tiết kiệm tài nguyên.
* **gRPC Internal Service (Port 9092):** Phục vụ giao tiếp nội bộ tốc độ cao với dữ liệu nhị phân (Protobuf). Hỗ trợ các Unary RPC như `GetManga`, `SearchManga`, và `UpdateProgress`.
* **WebSocket Chat Hub (Port 9093):** Cung cấp phòng chat thời gian thực (Bi-directional) cho cộng đồng, sử dụng cơ chế Hub & Spoke.

### 1.2. Cơ sở dữ liệu:
* Sử dụng **SQLite** (`github.com/mattn/go-sqlite3`) cho tính di động cao, được nhúng trực tiếp cùng ứng dụng mà không cần cài đặt database engine bên ngoài.

---

## 2. Setup Instructions (Hướng dẫn cài đặt)

### Yêu cầu hệ thống:
* Go 1.22+
* Docker & Docker Compose (Khuyên dùng)
* Nmap/Ncat (Để test TCP/UDP)
* Thư viện CGO (gcc/MinGW trên Windows nếu build từ source)

### Cách 1: Triển khai bằng Docker Compose (Dễ nhất)
Đây là cách triển khai khuyên dùng vì mọi dependency đã được đóng gói sẵn.

1.  Clone repository và di chuyển vào thư mục gốc dự án.
2.  Mở Terminal và chạy lệnh:
    ```bash
    docker-compose up --build -d
    ```
3.  Kiểm tra logs để đảm bảo 5 cổng đã được mở:
    ```bash
    docker logs mangahub_core
    ```
4.  Dữ liệu database được map ra ngoài thư mục `./data` trên máy tính thật.

### Cách 2: Triển khai từ Source Code
1.  Đảm bảo CGO đã được kích hoạt (cần thiết cho SQLite).
2.  Cài đặt các thư viện:
    ```bash
    go mod download
    ```
3.  Khởi chạy tất cả Server đồng thời:
    ```bash
    go run cmd/api-server/main.go
    ```

### Biên dịch và sử dụng Client (CLI)
Công cụ quản trị hệ thống bằng dòng lệnh được viết bằng Cobra.
1.  Biên dịch CLI (Trên Windows):
    ```bash
    go build -o mangahub.exe cmd/cli/main.go
    ```
2.  Kiểm tra CLI hoạt động:
    ```bash
    ./mangahub.exe help
    ```

---

## 3. API Documentation (Tài liệu API)

### 3.1. HTTP REST API (Base: `http://localhost:8080`)

| Method | Endpoint | Description | Auth Required | Payload / Query |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/auth/register` | Đăng ký tài khoản mới | No | `{"username":"", "password":""}` |
| `POST` | `/auth/login` | Lấy JWT Token | No | `{"username":"", "password":""}` |
| `GET` | `/manga?q={query}` | Tìm kiếm manga | Yes | Query param: `q` |
| `GET` | `/manga/:id` | Xem chi tiết manga | Yes | Path param: `id` |
| `POST` | `/users/library` | Thêm vào thư viện | Yes | `{"manga_id":"", "status":""}` |
| `GET` | `/users/library` | Xem thư viện | Yes | None |
| `PUT` | `/users/progress` | Cập nhật chapter đang đọc | Yes | `{"manga_id":"", "chapter": 0}` |
| `POST` | `/admin/notify` | Bắn thông báo UDP | No (Test) | `{"manga_id":"", "message":""}` |

*Lưu ý: Các API có Auth Required yêu cầu Header: `Authorization: Bearer <token>`.*

### 3.2. TCP Sync Protocol (Port `9090`)
* **Kết nối:** `ncat localhost 9090`
* **Hành vi:** Client duy trì kết nối. Server sẽ tự động gửi chuỗi JSON chứa thông tin user, manga_id và số chapter mới mỗi khi có bất kỳ ai gọi HTTP API `PUT /users/progress`. Dữ liệu được phân tách bằng ký tự `
`.

### 3.3. UDP Notification Protocol (Port `9091`)
* **Đăng ký (Subscribe):** Gửi chuỗi văn bản `SUBSCRIBE` (Giao thức UDP) tới port 9091. Server sẽ phản hồi `REGISTER_SUCCESS`.
* **Nhận thông báo:** Khi Admin gọi API `/admin/notify`, server phát sóng gói tin JSON (chứa `manga_id` và `message`) tới tất cả IP/Port đã đăng ký.
* **Test:** `ncat -u localhost 9091`

### 3.4. WebSocket Chat (Port `9093`)
* **Endpoint:** `ws://localhost:9093/chat`
* **Xác thực:** Gửi Header `Authorization: Bearer <token>` khi thiết lập handshake HTTP.
* **Giao tiếp:** Tin nhắn gửi đi định dạng Text thô. Trả về cấu trúc JSON chứa `user_id`, `username`, `message`, `timestamp`.

### 3.5. gRPC Internal Service (Port `9092`)
Sử dụng file định nghĩa `manga.proto`.
* `GetManga(GetMangaRequest) returns (MangaResponse)`: Lấy chi tiết truyện.
* `SearchManga(SearchRequest) returns (SearchResponse)`: Tìm kiếm.
* `UpdateProgress(ProgressRequest) returns (ProgressResponse)`: Cập nhật tiến trình và **tự động kích hoạt TCP Broadcast**.
