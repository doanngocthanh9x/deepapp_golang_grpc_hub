package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			from_client TEXT,
			to_client TEXT,
			channel TEXT,
			content TEXT,
			timestamp DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY,
			status TEXT,
			last_seen DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS workers (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'online',
			metadata TEXT,
			registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS capabilities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			worker_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			input_schema TEXT,
			output_schema TEXT,
			http_method TEXT DEFAULT 'POST',
			accepts_file BOOLEAN DEFAULT 0,
			file_field_name TEXT,
			FOREIGN KEY (worker_id) REFERENCES workers(id) ON DELETE CASCADE,
			UNIQUE(worker_id, name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_capabilities_name ON capabilities(name)`,
		`CREATE INDEX IF NOT EXISTS idx_capabilities_worker ON capabilities(worker_id)`,
		`CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}