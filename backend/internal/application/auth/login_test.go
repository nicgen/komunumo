package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/ports/fakes"
)

func makeVerifiedAccount(t *testing.T, accounts *fakes.AccountRepository, email, password string) {
	t.Helper()
	hasher := fakes.NewPasswordHasher()
	hash, _ := hasher.Hash(password)
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	acc, err := account.New("acc-1", email, "Léa", "Dupont", dob, now)
	require.NoError(t, err)
	acc.PasswordHash = hash
	acc.Status = account.StatusVerified
	require.NoError(t, accounts.Create(context.Background(), acc))
}

func newLoginService(t *testing.T) (*auth.LoginService, *fakes.AccountRepository, *fakes.SessionRepository, *fakes.AuditRepository) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	sessions := fakes.NewSessionRepository()
	auditLog := fakes.NewAuditRepository()
	hasher := fakes.NewPasswordHasher()
	tokenGen := fakes.NewTokenGenerator()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()
	svc := auth.NewLoginService(accounts, sessions, auditLog, hasher, tokenGen, clk, rl, uow)
	return svc, accounts, sessions, auditLog
}

// T063 — Login success creates session and audit event.
func TestLogin_Success(t *testing.T) {
	svc, accounts, sessions, auditLog := newLoginService(t)
	makeVerifiedAccount(t, accounts, "lea@example.com", "SecurePass123!")

	out, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "lea@example.com",
		Password: "SecurePass123!",
		IP:       "1.2.3.4",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, out.SessionID)
	assert.Equal(t, 1, sessions.Count())
	assert.True(t, auditLog.ContainsType("auth.login_success"))
}

// T064 — Wrong password returns ErrInvalidCredentials (no account info leak).
func TestLogin_WrongPassword(t *testing.T) {
	svc, accounts, sessions, auditLog := newLoginService(t)
	makeVerifiedAccount(t, accounts, "lea@example.com", "SecurePass123!")

	_, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "lea@example.com",
		Password: "WrongPassword!",
		IP:       "1.2.3.4",
	})

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Equal(t, 0, sessions.Count())
	assert.True(t, auditLog.ContainsType("auth.login_failed"))
}

// T065 — Unknown email returns same ErrInvalidCredentials (anti-enumeration).
func TestLogin_UnknownEmail(t *testing.T) {
	svc, _, sessions, _ := newLoginService(t)

	_, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "nobody@example.com",
		Password: "SecurePass123!",
		IP:       "1.2.3.4",
	})

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Equal(t, 0, sessions.Count())
}

// T066 — Unverified account returns ErrAccountNotVerified.
func TestLogin_PendingVerification(t *testing.T) {
	svc, accounts, sessions, _ := newLoginService(t)

	hasher := fakes.NewPasswordHasher()
	hash, _ := hasher.Hash("SecurePass123!")
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	acc, _ := account.New("acc-pending", "pending@example.com", "Jean", "Dupont", dob, now)
	acc.PasswordHash = hash
	require.NoError(t, accounts.Create(context.Background(), acc))

	_, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "pending@example.com",
		Password: "SecurePass123!",
		IP:       "1.2.3.4",
	})

	assert.ErrorIs(t, err, account.ErrAccountNotVerified)
	assert.Equal(t, 0, sessions.Count())
}

// T067 — Rate-limited IP returns ErrRateLimited.
func TestLogin_RateLimited(t *testing.T) {
	rl := fakes.NewRateLimiter()
	rl.Block("login:1.2.3.4")
	svc := auth.NewLoginService(
		fakes.NewAccountRepository(),
		fakes.NewSessionRepository(),
		fakes.NewAuditRepository(),
		fakes.NewPasswordHasher(),
		fakes.NewTokenGenerator(),
		fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)),
		rl,
		fakes.NewUnitOfWork(),
	)

	_, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "lea@example.com",
		Password: "SecurePass123!",
		IP:       "1.2.3.4",
	})

	assert.ErrorIs(t, err, auth.ErrRateLimited)
}
