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

func newRegisterDeps() (
	*fakes.AccountRepository,
	*fakes.TokenRepository,
	*fakes.AuditRepository,
	*fakes.EmailSender,
	*fakes.PasswordHasher,
	*fakes.TokenGenerator,
	*fakes.Clock,
	*fakes.RateLimiter,
	*fakes.UnitOfWork,
) {
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	return fakes.NewAccountRepository(),
		fakes.NewTokenRepository(),
		fakes.NewAuditRepository(),
		fakes.NewEmailSender(),
		fakes.NewPasswordHasher(),
		fakes.NewTokenGenerator(),
		fakes.NewClock(now),
		fakes.NewRateLimiter(),
		fakes.NewUnitOfWork()
}

var validInput = auth.RegisterInput{
	Email:       "lea@example.com",
	FirstName:   "Léa",
	LastName:    "Dupont",
	DateOfBirth: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	Password:    "SecurePass123!",
}

func TestRegister_Nominal(t *testing.T) {
	accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow := newRegisterDeps()
	svc := auth.NewRegisterService(accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow)

	err := svc.Register(context.Background(), "1.2.3.4", validInput)
	require.NoError(t, err)

	assert.Equal(t, 1, accounts.Count(), "one account created")
	assert.True(t, emails.Called("SendVerification"), "verification email sent")
	assert.False(t, emails.Called("SendAccountAlreadyExists"), "no duplicate email sent")

	a, _ := accounts.FindByEmailCanonical(context.Background(), "lea@example.com")
	require.NotNil(t, a)
	assert.Equal(t, account.StatusPendingVerification, a.Status)

	ev := auditLog.LastOfType("account.created")
	require.NotNil(t, ev, "account.created audit event expected")
}

func TestRegister_EmailAlreadyTaken_AntiEnumeration(t *testing.T) {
	accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow := newRegisterDeps()
	svc := auth.NewRegisterService(accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow)

	// Register once (nominal)
	require.NoError(t, svc.Register(context.Background(), "1.2.3.4", validInput))
	emails.Calls = nil // reset

	// Register again with same email — must not return error (anti-enumeration)
	err := svc.Register(context.Background(), "1.2.3.4", validInput)
	require.NoError(t, err)

	// No new account should be created
	assert.Equal(t, 1, accounts.Count(), "still one account")

	// A different email is sent
	assert.True(t, emails.Called("SendAccountAlreadyExists"), "account-exists email sent")
	assert.False(t, emails.Called("SendVerification"), "no new verification email")
}

func TestRegister_AgeTooYoung(t *testing.T) {
	_, _, _, _, hasher, tokenGen, clk, rl, uow := newRegisterDeps()
	accounts := fakes.NewAccountRepository()
	tokens := fakes.NewTokenRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	svc := auth.NewRegisterService(accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow)

	input := validInput
	// 10 years old
	input.DateOfBirth = clk.Now().AddDate(-10, 0, 0)

	err := svc.Register(context.Background(), "1.2.3.4", input)
	require.ErrorIs(t, err, account.ErrAgeBelow16)
	assert.Equal(t, 0, accounts.Count())
}

func TestRegister_PasswordTooWeak(t *testing.T) {
	accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow := newRegisterDeps()
	svc := auth.NewRegisterService(accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow)

	input := validInput
	input.Password = "weak"

	err := svc.Register(context.Background(), "1.2.3.4", input)
	require.Error(t, err)
	assert.Equal(t, 0, accounts.Count())
}

func TestRegister_EmailFailure_Transactional(t *testing.T) {
	accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow := newRegisterDeps()
	emails.FailOn("SendVerification", fakes.ErrEmailFailed)
	svc := auth.NewRegisterService(accounts, tokens, auditLog, emails, hasher, tokenGen, clk, rl, uow)

	err := svc.Register(context.Background(), "1.2.3.4", validInput)
	require.ErrorIs(t, err, fakes.ErrEmailFailed)

	// Transactional: no account must be persisted
	assert.Equal(t, 0, accounts.Count(), "account rolled back on email failure")
}
