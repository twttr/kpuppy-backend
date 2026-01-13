package sqlite

import (
	"database/sql"
	"fmt"
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

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{1, `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		user_hash TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL,
		avatar TEXT,
		is_banned INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL
	)`},
	{2, `CREATE INDEX IF NOT EXISTS idx_users_hash ON users(user_hash)`},
	{3, `CREATE TABLE IF NOT EXISTS content (
		id TEXT PRIMARY KEY,
		kinopub_item_id INTEGER UNIQUE NOT NULL,
		created_at INTEGER NOT NULL
	)`},
	{4, `CREATE INDEX IF NOT EXISTS idx_content_kinopub ON content(kinopub_item_id)`},
	{5, `CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		content_id TEXT NOT NULL REFERENCES content(id),
		user_id TEXT NOT NULL REFERENCES users(id),
		text TEXT NOT NULL,
		spoiler INTEGER DEFAULT 0,
		parent_id TEXT REFERENCES comments(id),
		edited_at INTEGER,
		deleted_at INTEGER,
		created_at INTEGER NOT NULL
	)`},
	{6, `CREATE INDEX IF NOT EXISTS idx_comments_content ON comments(content_id)`},
	{7, `CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_id)`},
	{8, `CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id)`},
	{9, `CREATE TABLE IF NOT EXISTS moderation_log (
		id TEXT PRIMARY KEY,
		admin_user_id TEXT NOT NULL,
		action TEXT NOT NULL,
		target_type TEXT NOT NULL,
		target_id TEXT NOT NULL,
		reason TEXT,
		created_at INTEGER NOT NULL
	)`},
}

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	var currentVersion int
	err = db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	for _, m := range migrations {
		if m.version <= currentVersion {
			continue
		}

		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %d: %w", m.version, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, m.version); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
	}

	return nil
}
