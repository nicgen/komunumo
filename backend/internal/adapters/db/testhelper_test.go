package db_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"komunumo/backend/internal/adapters/db"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")
	conn, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("openTestDB: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	schema, err := os.ReadFile("migrations/0001_init_auth.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := conn.ExecContext(context.Background(), string(schema)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	return conn
}
