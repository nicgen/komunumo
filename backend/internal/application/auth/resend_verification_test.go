package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports/fakes"
)

func TestResendVerification_Nominal(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	gen := fakes.NewTokenGenerator()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	clk := fakes.NewClock(now)
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	seedPendingAccount(t, accounts)

	svc := auth.NewResendVerificationService(accounts, tokens, auditLog, emails, gen, clk, rl, uow)
	err := svc.Resend(context.Background(), auth.ResendVerificationInput{
		Email: "lea@example.com",
		IP:    "1.2.3.4",
	})
	require.NoError(t, err)
	assert.True(t, emails.Called("SendVerification"))
}

func TestResendVerification_RevokesOldTokens(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	gen := fakes.NewTokenGenerator()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	clk := fakes.NewClock(now)
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	a := seedPendingAccount(t, accounts)
	_, _ = seedActiveToken(t, tokens, a.ID, now)

	svc := auth.NewResendVerificationService(accounts, tokens, auditLog, emails, gen, clk, rl, uow)
	require.NoError(t, svc.Resend(context.Background(), auth.ResendVerificationInput{
		Email: "lea@example.com",
		IP:    "1.2.3.4",
	}))

	// Old token should be revoked; new one created
	assert.Equal(t, 1, tokens.CountActive(a.ID, token.KindEmailVerification, clk.Now()))
}

func TestResendVerification_RateLimit(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	gen := fakes.NewTokenGenerator()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	clk := fakes.NewClock(now)
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	seedPendingAccount(t, accounts)
	rl.Block("resend_verification:email:lea@example.com")

	svc := auth.NewResendVerificationService(accounts, tokens, auditLog, emails, gen, clk, rl, uow)
	err := svc.Resend(context.Background(), auth.ResendVerificationInput{
		Email: "lea@example.com",
		IP:    "1.2.3.4",
	})
	require.Error(t, err)
	assert.False(t, emails.Called("SendVerification"))
}

func TestResendVerification_UnknownEmail_NoOp(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	gen := fakes.NewTokenGenerator()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	clk := fakes.NewClock(now)
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	svc := auth.NewResendVerificationService(accounts, tokens, auditLog, emails, gen, clk, rl, uow)
	// Should silently succeed (anti-enumeration: don't reveal account existence)
	err := svc.Resend(context.Background(), auth.ResendVerificationInput{
		Email: "unknown@example.com",
		IP:    "1.2.3.4",
	})
	require.NoError(t, err)
	assert.False(t, emails.Called("SendVerification"))
}
