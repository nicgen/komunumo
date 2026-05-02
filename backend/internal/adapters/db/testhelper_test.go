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

	for _, f := range []string{"migrations/0001_init_auth.up.sql", "migrations/0002_profiles.up.sql"} {
		schema, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read migration %s: %v", f, err)
		}
		if _, err := conn.ExecContext(context.Background(), string(schema)); err != nil {
			t.Fatalf("apply migration %s: %v", f, err)
		}
	}
	return conn
}
