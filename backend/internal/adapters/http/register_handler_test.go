package http_test

import (
	"bytes"
	"context"
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

func newRegisterHandler(t *testing.T) (*httpadapter.RegisterHandler, *fakes.AccountRepository, *fakes.MemberRepository, *fakes.RateLimiter) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	members := fakes.NewMemberRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	hasher := fakes.NewPasswordHasher()
	tokenGen := fakes.NewTokenGenerator()
	tokens := fakes.NewTokenRepository()
	clk := fakes.NewClock(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	svc := auth.NewRegisterMemberService(accounts, members, auditLog, emails, hasher, tokenGen, tokens, clk, rl, uow)
	handler := httpadapter.NewRegisterHandler(svc)
	return handler, accounts, members, rl
}

func TestRegisterMemberHandler_Success_201(t *testing.T) {
	handler, accounts, _, _ := newRegisterHandler(t)

	body, _ := json.Marshal(map[string]any{
		"email":      "lea@example.com",
		"password":   "SecurePass123!",
		"first_name": "Léa",
		"last_name":  "Martin",
		"birth_date": "2000-01-15",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/member", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	rr := httptest.NewRecorder()

	handler.HandleRegisterMember(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	canonical, _ := account.CanonicalizeEmail("lea@example.com")
	acc, err := accounts.FindByEmailCanonical(context.Background(), canonical)
	require.NoError(t, err)
	assert.NotNil(t, acc)
}

func TestRegisterMemberHandler_TooYoung_422(t *testing.T) {
	handler, _, _, _ := newRegisterHandler(t)

	body, _ := json.Marshal(map[string]any{
		"email":      "too-young@example.com",
		"password":   "SecurePass123!",
		"first_name": "Young",
		"last_name":  "Léa",
		"birth_date": "2015-01-01",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/member", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	rr := httptest.NewRecorder()

	handler.HandleRegisterMember(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestRegisterMemberHandler_RateLimited_429(t *testing.T) {
	handler, _, _, rl := newRegisterHandler(t)
	rl.Block("register:ip:192.0.2.1")

	body, _ := json.Marshal(map[string]any{
		"email":      "lea@example.com",
		"password":   "SecurePass123!",
		"first_name": "Léa",
		"last_name":  "Martin",
		"birth_date": "2000-01-15",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/member", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	rr := httptest.NewRecorder()

	handler.HandleRegisterMember(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}

func TestRegisterMemberHandler_InvalidJSON_400(t *testing.T) {
	handler, _, _, _ := newRegisterHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/member", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	rr := httptest.NewRecorder()

	handler.HandleRegisterMember(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterMemberHandler_EmailTaken_409(t *testing.T) {
	handler, accounts, _, _ := newRegisterHandler(t)

	// Seed existing account
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	acc, _ := account.New("acc-1", "lea@example.com", now)
	_ = accounts.Create(context.Background(), acc)

	body, _ := json.Marshal(map[string]any{
		"email":      "lea@example.com",
		"password":   "SecurePass123!",
		"first_name": "Léa",
		"last_name":  "Martin",
		"birth_date": "2000-01-15",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/member", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "192.0.2.1")
	rr := httptest.NewRecorder()

	handler.HandleRegisterMember(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}
