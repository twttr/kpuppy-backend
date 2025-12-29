package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(1)

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			kinopub_username TEXT UNIQUE NOT NULL,
			avatar TEXT,
			is_banned INTEGER DEFAULT 0,
			created_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(kinopub_username)`,

		`CREATE TABLE IF NOT EXISTS content (
			id TEXT PRIMARY KEY,
			kinopub_item_id INTEGER UNIQUE NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_content_kinopub ON content(kinopub_item_id)`,

		`CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			content_id TEXT NOT NULL REFERENCES content(id),
			user_id TEXT NOT NULL REFERENCES users(id),
			text TEXT NOT NULL,
			spoiler INTEGER DEFAULT 0,
			parent_id TEXT REFERENCES comments(id),
			edited_at INTEGER,
			deleted_at INTEGER,
			created_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_content ON comments(content_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id)`,

		`CREATE TABLE IF NOT EXISTS moderation_log (
			id TEXT PRIMARY KEY,
			admin_user_id TEXT NOT NULL,
			action TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			reason TEXT,
			created_at INTEGER NOT NULL
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}
