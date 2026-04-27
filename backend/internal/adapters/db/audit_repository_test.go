package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/domain/audit"
)

func TestAuditRepository_Append(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAuditRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	e1 := &audit.Event{
		ID:         "evt1",
		OccurredAt: now,
		Type:       audit.EventAccountCreated,
		IP:         "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Metadata: map[string]any{
			"email": "lea@example.com",
		},
	}

	err := repo.Append(context.Background(), e1)
	require.NoError(t, err)

	// Append a second event
	later := now.Add(1 * time.Hour)
	e2 := &audit.Event{
		ID:         "evt2",
		OccurredAt: later,
		Type:       audit.EventAuthLoginSuccess,
		AccountID:  strPtr("acc1"),
		IP:         "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Metadata: map[string]any{
			"session_id": "sess123",
		},
	}

	err = repo.Append(context.Background(), e2)
	require.NoError(t, err)
}

func TestAuditRepository_NoUpdate(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAuditRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	e := &audit.Event{
		ID:         "evt1",
		OccurredAt: now,
		Type:       audit.EventAccountCreated,
		IP:         "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
	}

	err := repo.Append(context.Background(), e)
	require.NoError(t, err)

	// Attempt to update the event in the audit_log table
	result, err := conn.ExecContext(
		context.Background(),
		"UPDATE audit_log SET event_type = ? WHERE id = ?",
		"account.disabled",
		"evt1",
	)
	require.Error(t, err, "UPDATE should fail due to audit_log trigger")

	// Verify no rows were affected
	if err == nil {
		rows, _ := result.RowsAffected()
		assert.Equal(t, int64(0), rows)
	}
}

func TestAuditRepository_NoDelete(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAuditRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	e := &audit.Event{
		ID:         "evt1",
		OccurredAt: now,
		Type:       audit.EventAccountCreated,
		IP:         "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
	}

	err := repo.Append(context.Background(), e)
	require.NoError(t, err)

	// Attempt to delete the event from the audit_log table
	result, err := conn.ExecContext(
		context.Background(),
		"DELETE FROM audit_log WHERE id = ?",
		"evt1",
	)
	require.Error(t, err, "DELETE should fail due to audit_log trigger")

	// Verify no rows were deleted
	if err == nil {
		rows, _ := result.RowsAffected()
		assert.Equal(t, int64(0), rows)
	}
}

// Helper function to create a pointer to a string
func strPtr(s string) *string {
	return &s
}
