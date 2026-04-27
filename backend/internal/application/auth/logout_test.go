package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports/fakes"
)

func newLogoutService(t *testing.T) (*auth.LogoutService, *fakes.SessionRepository, *fakes.AuditRepository) {
	t.Helper()
	sessions := fakes.NewSessionRepository()
	auditLog := fakes.NewAuditRepository()
	tokenGen := fakes.NewTokenGenerator()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	svc := auth.NewLogoutService(sessions, auditLog, tokenGen, clk)
	return svc, sessions, auditLog
}

// T068 — Logout deletes session and logs audit event.
func TestLogout_Success(t *testing.T) {
	svc, sessions, auditLog := newLogoutService(t)

	accID := "acc-1"
	sess := &session.Session{
		ID:         "sess-1",
		AccountID:  accID,
		CreatedAt:  time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC),
		ExpiresAt:  time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC),
		LastSeenAt: time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC),
	}
	require.NoError(t, sessions.Create(context.Background(), sess))

	err := svc.Logout(context.Background(), "sess-1")

	require.NoError(t, err)
	assert.Equal(t, 0, sessions.Count())
	assert.True(t, auditLog.ContainsType("auth.logout"))
}

// T069 — Logout with unknown/expired session is a no-op (idempotent).
func TestLogout_UnknownSession(t *testing.T) {
	svc, _, _ := newLogoutService(t)

	err := svc.Logout(context.Background(), "nonexistent-session")

	assert.NoError(t, err)
}
