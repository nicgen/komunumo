package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports/fakes"
)

func newMeService(t *testing.T) (*auth.MeService, *fakes.SessionRepository, *fakes.AccountRepository, *fakes.Clock) {
	t.Helper()
	sessions := fakes.NewSessionRepository()
	accounts := fakes.NewAccountRepository()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	svc := auth.NewMeService(sessions, accounts, clk)
	return svc, sessions, accounts, clk
}

// T085 — Me returns account info for a valid session.
func TestMe_ValidSession(t *testing.T) {
	svc, sessions, accounts, clk := newMeService(t)

	now := clk.Now()
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	acc, err := account.New("acc-1", "anne@example.com", "Anne", "Dupont", dob, now)
	require.NoError(t, err)
	acc.Status = account.StatusVerified
	require.NoError(t, accounts.Create(context.Background(), acc))

	sess := &session.Session{
		ID:         "sess-1",
		AccountID:  acc.ID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(30 * 24 * time.Hour),
		LastSeenAt: now,
	}
	require.NoError(t, sessions.Create(context.Background(), sess))

	out, err := svc.Me(context.Background(), "sess-1")

	require.NoError(t, err)
	assert.Equal(t, acc.ID, out.AccountID)
	assert.Equal(t, "anne@example.com", out.Email)
	assert.Equal(t, "Anne", out.FirstName)
	assert.Equal(t, account.StatusVerified, out.Status)
}

// T086 — Me returns ErrSessionNotFound for unknown session.
func TestMe_UnknownSession(t *testing.T) {
	svc, _, _, _ := newMeService(t)

	_, err := svc.Me(context.Background(), "nonexistent")

	assert.ErrorIs(t, err, session.ErrSessionNotFound)
}

// T087 — Me returns ErrSessionNotFound for expired session.
func TestMe_ExpiredSession(t *testing.T) {
	svc, sessions, _, clk := newMeService(t)

	now := clk.Now()
	sess := &session.Session{
		ID:         "sess-expired",
		AccountID:  "acc-1",
		CreatedAt:  now.Add(-48 * time.Hour),
		ExpiresAt:  now.Add(-1 * time.Hour),
		LastSeenAt: now.Add(-48 * time.Hour),
	}
	require.NoError(t, sessions.Create(context.Background(), sess))

	_, err := svc.Me(context.Background(), "sess-expired")

	assert.ErrorIs(t, err, session.ErrSessionNotFound)
}
