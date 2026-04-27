package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	httpadapter "komunumo/backend/internal/adapters/http"
	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/ports/fakes"
)

func newRegisterHandler(t *testing.T) (*httpadapter.AuthHandler, *fakes.AccountRepository, *fakes.EmailSender) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	hasher := fakes.NewPasswordHasher()
	tokenGen := fakes.NewTokenGenerator()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	registerSvc := auth.NewRegisterService(accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow)
	verifySvc := auth.NewVerifyEmailService(accounts, tokens, auditLog, tokenGen, clk, uow)
	resendSvc := auth.NewResendVerificationService(accounts, tokens, auditLog, emails, tokenGen, clk, rl, uow)

	handler := httpadapter.NewAuthHandler(registerSvc, verifySvc, resendSvc, nil, nil)
	return handler, accounts, emails
}

func TestRegisterHandler_JSON_201(t *testing.T) {
	handler, accounts, emails := newRegisterHandler(t)

	body, _ := json.Marshal(map[string]any{
		"email":         "lea@example.com",
		"first_name":    "Léa",
		"last_name":     "Dupont",
		"date_of_birth": "2000-01-15",
		"password":      "SecurePass123!",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, 1, accounts.Count())
	assert.True(t, emails.Called("SendVerification"))
}

func TestRegisterHandler_FormEncoded_303(t *testing.T) {
	handler, _, _ := newRegisterHandler(t)

	form := url.Values{}
	form.Set("email", "lea@example.com")
	form.Set("first_name", "Léa")
	form.Set("last_name", "Dupont")
	form.Set("date_of_birth", "2000-01-15")
	form.Set("password", "SecurePass123!")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "sent")
}

func TestRegisterHandler_ValidationError_400(t *testing.T) {
	handler, _, _ := newRegisterHandler(t)

	// Missing required fields
	body, _ := json.Marshal(map[string]any{"email": "not-an-email"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterHandler_DuplicateEmail_AntiEnumeration(t *testing.T) {
	handler, _, emails := newRegisterHandler(t)

	payload := map[string]any{
		"email":         "lea@example.com",
		"first_name":    "Léa",
		"last_name":     "Dupont",
		"date_of_birth": "2000-01-15",
		"password":      "SecurePass123!",
	}

	doRequest := func() int {
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.Register(rr, req)
		return rr.Code
	}

	first := doRequest()
	assert.Equal(t, http.StatusCreated, first)
	emails.Calls = nil

	// Second registration with same email: must return same 201 (anti-enumeration)
	second := doRequest()
	assert.Equal(t, http.StatusCreated, second)
	assert.True(t, emails.Called("SendAccountAlreadyExists"))
}
