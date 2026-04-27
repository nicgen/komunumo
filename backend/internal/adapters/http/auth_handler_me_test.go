package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpadapter "komunumo/backend/internal/adapters/http"
	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports/fakes"
)

func newMeHandler(t *testing.T) (*httpadapter.AuthHandler, *fakes.SessionRepository, *fakes.AccountRepository) {
	t.Helper()
	sessions := fakes.NewSessionRepository()
	accounts := fakes.NewAccountRepository()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	meSvc := auth.NewMeService(sessions, accounts, clk)
	handler := httpadapter.NewAuthHandler(nil, nil, nil, nil, nil, nil, nil, meSvc)
	return handler, sessions, accounts
}

// T088 — GET /me with valid session cookie returns account JSON.
func TestMeHandler_ValidSession_200(t *testing.T) {
	handler, sessions, accounts := newMeHandler(t)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	acc, err := account.New("acc-1", "anne@example.com", "Anne", "Dupont", dob, now)
	require.NoError(t, err)
	acc.Status = account.StatusVerified
	require.NoError(t, accounts.Create(t.Context(), acc))

	sess := &session.Session{
		ID:         "sess-1",
		AccountID:  acc.ID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(30 * 24 * time.Hour),
		LastSeenAt: now,
	}
	require.NoError(t, sessions.Create(t.Context(), sess))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-1"})
	rr := httptest.NewRecorder()

	handler.Me(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Equal(t, "anne@example.com", body["email"])
	assert.Equal(t, "Anne", body["first_name"])
}

// T089 — GET /me without cookie returns 401.
func TestMeHandler_NoCookie_401(t *testing.T) {
	handler, _, _ := newMeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rr := httptest.NewRecorder()

	handler.Me(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// T090 — GET /me with unknown session cookie returns 401.
func TestMeHandler_UnknownSession_401(t *testing.T) {
	handler, _, _ := newMeHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "nonexistent"})
	rr := httptest.NewRecorder()

	handler.Me(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
