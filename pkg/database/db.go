package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	username TEXT UNIQUE,
	password_hash TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS manga (
	id TEXT PRIMARY KEY,
	title TEXT,
	author TEXT,
	genres TEXT,
	status TEXT,
	total_chapters INTEGER,
	description TEXT
);

CREATE TABLE IF NOT EXISTS user_progress (
	user_id TEXT,
	manga_id TEXT,
	current_chapter INTEGER,
	status TEXT,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id, manga_id)
);
`

// Tạo kết nối và setup cấu trúc bảng
func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully!")
	return db, nil
}