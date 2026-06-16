package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open creates the data directory (if needed), opens the SQLite database at
// the given path, and applies schema migrations + seed data.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1) // modernc sqlite + WAL: keep writes serialized

	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := conn.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	if err := migrate(conn); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := seed(conn); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	return conn, nil
}
