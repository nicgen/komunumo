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
	"komunumo/backend/internal/ports/fakes"
)

func newLoginHandler(t *testing.T) (*httpadapter.AuthHandler, *fakes.AccountRepository, *fakes.SessionRepository) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	sessions := fakes.NewSessionRepository()
	auditLog := fakes.NewAuditRepository()
	hasher := fakes.NewPasswordHasher()
	tokenGen := fakes.NewTokenGenerator()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	registerSvc := auth.NewRegisterService(accounts, fakes.NewTokenRepository(), auditLog, fakes.NewEmailSender(), hasher, tokenGen, clk, rl, uow)
	verifySvc := auth.NewVerifyEmailService(accounts, fakes.NewTokenRepository(), auditLog, tokenGen, clk, uow)
	resendSvc := auth.NewResendVerificationService(accounts, fakes.NewTokenRepository(), auditLog, fakes.NewEmailSender(), tokenGen, clk, rl, uow)
	loginSvc := auth.NewLoginService(accounts, sessions, auditLog, hasher, tokenGen, clk, rl, uow)
	logoutSvc := auth.NewLogoutService(sessions, auditLog, tokenGen, clk)

	handler := httpadapter.NewAuthHandler(registerSvc, verifySvc, resendSvc, loginSvc, logoutSvc, nil, nil)
	return handler, accounts, sessions
}

func seedVerifiedAccount(t *testing.T, accounts *fakes.AccountRepository, email, password string) {
	t.Helper()
	hasher := fakes.NewPasswordHasher()
	hash, _ := hasher.Hash(password)
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	acc, err := account.New("acc-1", email, "Léa", "Dupont", dob, now)
	require.NoError(t, err)
	acc.PasswordHash = hash
	acc.Status = account.StatusVerified
	require.NoError(t, accounts.Create(t.Context(), acc))
}

// T070 — POST /login with valid credentials returns 200 and sets session cookie.
func TestLoginHandler_Success_200(t *testing.T) {
	handler, accounts, sessions := newLoginHandler(t)
	seedVerifiedAccount(t, accounts, "lea@example.com", "SecurePass123!")

	body, _ := json.Marshal(map[string]any{
		"email":    "lea@example.com",
		"password": "SecurePass123!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, sessions.Count())

	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie, "session_id cookie must be set")
	assert.True(t, sessionCookie.HttpOnly)
}

// T071 — POST /login with wrong password returns 401.
func TestLoginHandler_WrongPassword_401(t *testing.T) {
	handler, accounts, _ := newLoginHandler(t)
	seedVerifiedAccount(t, accounts, "lea@example.com", "SecurePass123!")

	body, _ := json.Marshal(map[string]any{
		"email":    "lea@example.com",
		"password": "WrongPassword!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// T072 — POST /login on unverified account returns 403.
func TestLoginHandler_Unverified_403(t *testing.T) {
	handler, accounts, _ := newLoginHandler(t)

	hasher := fakes.NewPasswordHasher()
	hash, _ := hasher.Hash("SecurePass123!")
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	acc, _ := account.New("acc-pending", "pending@example.com", "Jean", "Dupont", dob, now)
	acc.PasswordHash = hash
	require.NoError(t, accounts.Create(t.Context(), acc))

	body, _ := json.Marshal(map[string]any{
		"email":    "pending@example.com",
		"password": "SecurePass123!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// T073 — POST /logout with valid session cookie returns 200 and clears cookie.
func TestLogoutHandler_Success_200(t *testing.T) {
	handler, accounts, sessions := newLoginHandler(t)
	seedVerifiedAccount(t, accounts, "lea@example.com", "SecurePass123!")

	// Login first to get a session.
	loginBody, _ := json.Marshal(map[string]any{
		"email":    "lea@example.com",
		"password": "SecurePass123!",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)
	require.Equal(t, http.StatusOK, loginRR.Code)

	var sessionID string
	for _, c := range loginRR.Result().Cookies() {
		if c.Name == "session_id" {
			sessionID = c.Value
		}
	}
	require.NotEmpty(t, sessionID)

	// Now logout.
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutReq.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	logoutRR := httptest.NewRecorder()

	handler.Logout(logoutRR, logoutReq)

	assert.Equal(t, http.StatusOK, logoutRR.Code)
	assert.Equal(t, 0, sessions.Count())

	// Cookie must be cleared (MaxAge=-1 or Expires in the past).
	var clearedCookie *http.Cookie
	for _, c := range logoutRR.Result().Cookies() {
		if c.Name == "session_id" {
			clearedCookie = c
		}
	}
	require.NotNil(t, clearedCookie)
	assert.True(t, clearedCookie.MaxAge < 0 || clearedCookie.Value == "")
}
