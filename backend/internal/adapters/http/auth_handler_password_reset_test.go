package http_test

import (
	"bytes"
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
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports/fakes"
)

func newPasswordResetHandler(t *testing.T) (
	*httpadapter.AuthHandler,
	*fakes.AccountRepository,
	*fakes.TokenRepository,
	*fakes.EmailSender,
) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	sessions := fakes.NewSessionRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	hasher := fakes.NewPasswordHasher()
	tokenGen := fakes.NewTokenGenerator()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	pwResetReqSvc := auth.NewPasswordResetRequestService(accounts, tokens, auditLog, emails, tokenGen, clk, rl, uow)
	pwResetConfSvc := auth.NewPasswordResetConfirmService(accounts, tokens, sessions, auditLog, emails, hasher, tokenGen, clk, uow)

	handler := httpadapter.NewAuthHandler(nil, nil, nil, nil, nil, pwResetReqSvc, pwResetConfSvc, nil)
	return handler, accounts, tokens, emails
}

func seedVerifiedAccountForHandler(t *testing.T, accounts *fakes.AccountRepository) *account.Account {
	t.Helper()
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	acc, err := account.New("acc-pr", "anne@example.com", "Anne", "Dupont", dob, now)
	require.NoError(t, err)
	acc.PasswordHash = "hash:OldPass123!"
	acc.Status = account.StatusVerified
	require.NoError(t, accounts.Create(t.Context(), acc))
	return acc
}

// T081 — POST /password-reset/request with known email returns 200 and sends email.
func TestPasswordResetRequestHandler_KnownEmail_200(t *testing.T) {
	handler, accounts, _, emails := newPasswordResetHandler(t)
	seedVerifiedAccountForHandler(t, accounts)

	body, _ := json.Marshal(map[string]any{"email": "anne@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.PasswordResetRequest(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, emails.Called("SendPasswordReset"))
}

// T082 — POST /password-reset/request with unknown email returns 200 (anti-enumeration).
func TestPasswordResetRequestHandler_UnknownEmail_200(t *testing.T) {
	handler, _, _, emails := newPasswordResetHandler(t)

	body, _ := json.Marshal(map[string]any{"email": "nobody@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/request", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.PasswordResetRequest(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, emails.Called("SendPasswordReset"))
}

// T083 — POST /password-reset/confirm with valid token returns 200.
func TestPasswordResetConfirmHandler_Success_200(t *testing.T) {
	handler, accounts, tokens, emails := newPasswordResetHandler(t)
	acc := seedVerifiedAccountForHandler(t, accounts)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	tokenGen := fakes.NewTokenGenerator()
	raw := "raw-token-1"
	hash := tokenGen.HashToken(raw)
	tok := token.New("tok-1", acc.ID, token.KindPasswordReset, hash, now, token.PasswordResetTTL)
	require.NoError(t, tokens.Create(t.Context(), tok))

	body, _ := json.Marshal(map[string]any{
		"token":        raw,
		"new_password": "NewSecurePass456!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.PasswordResetConfirm(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, emails.Called("SendPasswordChanged"))
}

// T084 — POST /password-reset/confirm with invalid token returns 400.
func TestPasswordResetConfirmHandler_InvalidToken_400(t *testing.T) {
	handler, _, _, _ := newPasswordResetHandler(t)

	body, _ := json.Marshal(map[string]any{
		"token":        "bad-token",
		"new_password": "NewSecurePass456!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.PasswordResetConfirm(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
