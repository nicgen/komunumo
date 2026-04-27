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
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports/fakes"
)

func sessionOf(accountID string, now time.Time) *session.Session {
	return &session.Session{
		ID:         "sess-" + accountID,
		AccountID:  accountID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(30 * 24 * time.Hour),
		LastSeenAt: now,
	}
}

func newPasswordResetRequestService(t *testing.T) (
	*auth.PasswordResetRequestService,
	*fakes.AccountRepository,
	*fakes.TokenRepository,
	*fakes.EmailSender,
	*fakes.AuditRepository,
) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	tokenGen := fakes.NewTokenGenerator()
	clk := fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC))
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()
	svc := auth.NewPasswordResetRequestService(accounts, tokens, auditLog, emails, tokenGen, clk, rl, uow)
	return svc, accounts, tokens, emails, auditLog
}

func seedVerifiedAccountForReset(t *testing.T, accounts *fakes.AccountRepository) *account.Account {
	t.Helper()
	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	acc, err := account.New("acc-1", "anne@example.com", "Anne", "Dupont", dob, now)
	require.NoError(t, err)
	acc.PasswordHash = "hash:OldPassword123!"
	acc.Status = account.StatusVerified
	require.NoError(t, accounts.Create(context.Background(), acc))
	return acc
}

// T074 — Request with known email sends reset email and creates token.
func TestPasswordResetRequest_KnownEmail(t *testing.T) {
	svc, accounts, tokens, emails, auditLog := newPasswordResetRequestService(t)
	seedVerifiedAccountForReset(t, accounts)

	err := svc.Request(context.Background(), auth.PasswordResetRequestInput{
		Email: "anne@example.com",
		IP:    "1.2.3.4",
	})

	require.NoError(t, err)
	assert.True(t, emails.Called("SendPasswordReset"))
	assert.Equal(t, 1, tokens.CountByKind(token.KindPasswordReset))
	assert.True(t, auditLog.ContainsType("auth.password_reset_requested"))
}

// T075 — Request with unknown email returns nil (anti-enumeration, no email sent).
func TestPasswordResetRequest_UnknownEmail(t *testing.T) {
	svc, _, tokens, emails, _ := newPasswordResetRequestService(t)

	err := svc.Request(context.Background(), auth.PasswordResetRequestInput{
		Email: "nobody@example.com",
		IP:    "1.2.3.4",
	})

	require.NoError(t, err)
	assert.False(t, emails.Called("SendPasswordReset"))
	assert.Equal(t, 0, tokens.CountByKind(token.KindPasswordReset))
}

// T076 — Request revokes previous active token before creating a new one.
func TestPasswordResetRequest_RevokesOldToken(t *testing.T) {
	svc, accounts, tokens, _, _ := newPasswordResetRequestService(t)
	seedVerifiedAccountForReset(t, accounts)

	require.NoError(t, svc.Request(context.Background(), auth.PasswordResetRequestInput{Email: "anne@example.com", IP: "1.2.3.4"}))
	require.NoError(t, svc.Request(context.Background(), auth.PasswordResetRequestInput{Email: "anne@example.com", IP: "1.2.3.4"}))

	assert.Equal(t, 1, tokens.CountActiveByKind(token.KindPasswordReset))
}

// T077 — Rate-limited IP returns ErrRateLimited.
func TestPasswordResetRequest_RateLimited(t *testing.T) {
	rl := fakes.NewRateLimiter()
	rl.Block("pwreset:1.2.3.4")
	svc := auth.NewPasswordResetRequestService(
		fakes.NewAccountRepository(), fakes.NewTokenRepository(), fakes.NewAuditRepository(),
		fakes.NewEmailSender(), fakes.NewTokenGenerator(),
		fakes.NewClock(time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)),
		rl, fakes.NewUnitOfWork(),
	)

	err := svc.Request(context.Background(), auth.PasswordResetRequestInput{Email: "anne@example.com", IP: "1.2.3.4"})
	assert.ErrorIs(t, err, auth.ErrRateLimited)
}

// --- Confirm tests ---

func newPasswordResetConfirmService(t *testing.T) (
	*auth.PasswordResetConfirmService,
	*fakes.AccountRepository,
	*fakes.TokenRepository,
	*fakes.SessionRepository,
	*fakes.EmailSender,
	*fakes.AuditRepository,
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
	uow := fakes.NewUnitOfWork()
	svc := auth.NewPasswordResetConfirmService(accounts, tokens, sessions, auditLog, emails, hasher, tokenGen, clk, uow)
	return svc, accounts, tokens, sessions, emails, auditLog
}

// T078 — Confirm with valid token changes password, invalidates sessions, sends email.
func TestPasswordResetConfirm_Success(t *testing.T) {
	svc, accounts, tokens, sessions, emails, auditLog := newPasswordResetConfirmService(t)
	acc := seedVerifiedAccountForReset(t, accounts)

	// Seed an active reset token.
	tokenGen := fakes.NewTokenGenerator()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	raw := "raw-token-1"
	hash := tokenGen.HashToken(raw)
	tok := token.New("tok-1", acc.ID, token.KindPasswordReset, hash, now, token.PasswordResetTTL)
	require.NoError(t, tokens.Create(context.Background(), tok))

	// Seed an active session.
	sess := sessionOf(acc.ID, now)
	require.NoError(t, sessions.Create(context.Background(), sess))

	err := svc.Confirm(context.Background(), auth.PasswordResetConfirmInput{
		RawToken:    raw,
		NewPassword: "NewSecurePass456!",
	})

	require.NoError(t, err)
	assert.Equal(t, 0, sessions.Count())
	assert.True(t, emails.Called("SendPasswordChanged"))
	assert.True(t, auditLog.ContainsType("auth.password_changed"))

	updated, _ := accounts.FindByID(context.Background(), acc.ID)
	assert.Equal(t, "hash:NewSecurePass456!", updated.PasswordHash)
}

// T079 — Confirm with unknown/expired token returns ErrTokenNotFound.
func TestPasswordResetConfirm_InvalidToken(t *testing.T) {
	svc, _, _, _, _, _ := newPasswordResetConfirmService(t)

	err := svc.Confirm(context.Background(), auth.PasswordResetConfirmInput{
		RawToken:    "nonexistent-token",
		NewPassword: "NewSecurePass456!",
	})

	assert.ErrorIs(t, err, token.ErrTokenNotFound)
}

// T080 — Confirm with weak password returns validation error.
func TestPasswordResetConfirm_WeakPassword(t *testing.T) {
	svc, _, _, _, _, _ := newPasswordResetConfirmService(t)

	err := svc.Confirm(context.Background(), auth.PasswordResetConfirmInput{
		RawToken:    "any-token",
		NewPassword: "short",
	})

	assert.Error(t, err)
	assert.NotErrorIs(t, err, token.ErrTokenNotFound)
}
