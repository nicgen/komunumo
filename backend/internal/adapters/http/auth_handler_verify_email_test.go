package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func newVerifyEmailHandler(t *testing.T) (*httpadapter.AuthHandler, *fakes.AccountRepository, *fakes.TokenRepository) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	tokenGen := fakes.NewTokenGenerator()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	uow := fakes.NewUnitOfWork()

	verifySvc := auth.NewVerifyEmailService(accounts, tokens, auditLog, tokenGen, clk, uow)
	handler := httpadapter.NewAuthHandler(nil, verifySvc, nil, nil, nil)
	return handler, accounts, tokens
}

func seedVerifyScenario(t *testing.T, accounts *fakes.AccountRepository, tokens *fakes.TokenRepository, now time.Time) (rawToken string) {
	t.Helper()
	dob := now.AddDate(-20, 0, 0)
	a, err := account.New("acc1", "lea@example.com", "Léa", "Dupont", dob, now)
	require.NoError(t, err)
	require.NoError(t, accounts.Create(context.Background(), a))

	gen := fakes.NewTokenGenerator()
	raw, err := gen.NewRawToken()
	require.NoError(t, err)
	h := gen.HashToken(raw)
	tok := token.New("tok1", a.ID, token.KindEmailVerification, h, now, token.EmailVerificationTTL)
	require.NoError(t, tokens.Create(context.Background(), tok))
	return raw
}

func TestVerifyEmailHandler_JSON_200(t *testing.T) {
	handler, accounts, tokens := newVerifyEmailHandler(t)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	rawToken := seedVerifyScenario(t, accounts, tokens, now)

	body := strings.NewReader(`{"token":"` + rawToken + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.VerifyEmail(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestVerifyEmailHandler_FormEncoded_303(t *testing.T) {
	handler, accounts, tokens := newVerifyEmailHandler(t)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	rawToken := seedVerifyScenario(t, accounts, tokens, now)

	form := url.Values{}
	form.Set("token", rawToken)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.VerifyEmail(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "login")
}

func TestVerifyEmailHandler_InvalidToken_400(t *testing.T) {
	handler, _, _ := newVerifyEmailHandler(t)

	body := strings.NewReader(`{"token":"unknown-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.VerifyEmail(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestVerifyEmailHandler_ExpiredToken_410(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	tokenGen := fakes.NewTokenGenerator()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	// Clock is 25h after token creation
	clk := fakes.NewClock(now.Add(25 * time.Hour))
	uow := fakes.NewUnitOfWork()

	rawToken := seedVerifyScenario(t, accounts, tokens, now)

	verifySvc := auth.NewVerifyEmailService(accounts, tokens, auditLog, tokenGen, clk, uow)
	handler := httpadapter.NewAuthHandler(nil, verifySvc, nil, nil, nil)

	body := strings.NewReader(`{"token":"` + rawToken + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.VerifyEmail(rr, req)

	assert.Equal(t, http.StatusGone, rr.Code)
}
