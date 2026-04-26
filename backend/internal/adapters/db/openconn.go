package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite connection at dsn and applies the PRAGMAs required
// by the Komunumo stack: WAL journal, foreign keys on, busy_timeout 5s.
//
// The returned *sql.DB is safe for concurrent use.
func Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("db.Open: empty dsn")
	}

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.Open: sql.Open: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA synchronous=NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("db.Open: %s: %w", p, err)
		}
	}

	return conn, nil
}
