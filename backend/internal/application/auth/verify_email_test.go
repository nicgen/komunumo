package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports/fakes"
)

func seedPendingAccount(t *testing.T, accounts *fakes.AccountRepository) *account.Account {
	t.Helper()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	dob := now.AddDate(-20, 0, 0)
	a, err := account.New("acc1", "lea@example.com", "Léa", "Dupont", dob, now)
	require.NoError(t, err)
	require.NoError(t, accounts.Create(context.Background(), a))
	return a
}

func seedActiveToken(t *testing.T, tokens *fakes.TokenRepository, accountID string, now time.Time) (rawToken string, hash string) {
	t.Helper()
	gen := fakes.NewTokenGenerator()
	raw, err := gen.NewRawToken()
	require.NoError(t, err)
	h := gen.HashToken(raw)
	tok := token.New("tok1", accountID, token.KindEmailVerification, h, now, token.EmailVerificationTTL)
	require.NoError(t, tokens.Create(context.Background(), tok))
	return raw, h
}

func TestVerifyEmail_Nominal(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	clk := fakes.NewClock(now)
	uow := fakes.NewUnitOfWork()
	gen := fakes.NewTokenGenerator()

	a := seedPendingAccount(t, accounts)
	rawToken, _ := seedActiveToken(t, tokens, a.ID, now)

	svc := auth.NewVerifyEmailService(accounts, tokens, auditLog, gen, clk, uow)
	err := svc.VerifyEmail(context.Background(), auth.VerifyEmailInput{RawToken: rawToken})
	require.NoError(t, err)

	updated, _ := accounts.FindByID(context.Background(), a.ID)
	assert.Equal(t, account.StatusVerified, updated.Status)

	ev := auditLog.LastOfType(audit.EventAccountEmailVerified)
	require.NotNil(t, ev)
}

func TestVerifyEmail_ExpiredToken(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	// Clock is 25h after token creation → expired
	clk := fakes.NewClock(now.Add(25 * time.Hour))
	uow := fakes.NewUnitOfWork()
	gen := fakes.NewTokenGenerator()

	a := seedPendingAccount(t, accounts)
	rawToken, _ := seedActiveToken(t, tokens, a.ID, now)

	svc := auth.NewVerifyEmailService(accounts, tokens, auditLog, gen, clk, uow)
	err := svc.VerifyEmail(context.Background(), auth.VerifyEmailInput{RawToken: rawToken})
	require.ErrorIs(t, err, token.ErrTokenExpired)

	// Account status must not have changed
	acc, _ := accounts.FindByID(context.Background(), a.ID)
	assert.Equal(t, account.StatusPendingVerification, acc.Status)
}

func TestVerifyEmail_UnknownToken(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	clk := fakes.NewClock(now)
	uow := fakes.NewUnitOfWork()
	gen := fakes.NewTokenGenerator()

	svc := auth.NewVerifyEmailService(accounts, tokens, auditLog, gen, clk, uow)
	err := svc.VerifyEmail(context.Background(), auth.VerifyEmailInput{RawToken: "nonexistent-token"})
	require.ErrorIs(t, err, token.ErrTokenNotFound)
}

func TestVerifyEmail_AlreadyConsumed(t *testing.T) {
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	clk := fakes.NewClock(now)
	uow := fakes.NewUnitOfWork()
	gen := fakes.NewTokenGenerator()

	a := seedPendingAccount(t, accounts)
	rawToken, _ := seedActiveToken(t, tokens, a.ID, now)

	svc := auth.NewVerifyEmailService(accounts, tokens, auditLog, gen, clk, uow)
	// First use: success
	require.NoError(t, svc.VerifyEmail(context.Background(), auth.VerifyEmailInput{RawToken: rawToken}))
	// Second use: already consumed
	err := svc.VerifyEmail(context.Background(), auth.VerifyEmailInput{RawToken: rawToken})
	require.ErrorIs(t, err, token.ErrTokenNotFound)
}
