package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/ports/fakes"
)

func newRegisterMemberService(t *testing.T) (*auth.RegisterMemberService, *fakes.AccountRepository, *fakes.MemberRepository, *fakes.AuditRepository, *fakes.EmailSender) {
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

	svc := auth.NewRegisterMemberService(
		accounts,
		members,
		auditLog,
		emails,
		hasher,
		tokenGen,
		tokens,
		clk,
		rl,
		uow,
	)
	return svc, accounts, members, auditLog, emails
}

func TestRegisterMember_Success(t *testing.T) {
	svc, accounts, members, auditLog, emails := newRegisterMemberService(t)

	input := auth.RegisterMemberInput{
		Email:     "lea@example.com",
		Password:  "SecurePass123!",
		FirstName: "Léa",
		LastName:  "Martin",
		BirthDate: "2000-01-15",
	}

	err := svc.RegisterMember(context.Background(), "1.2.3.4", input)

	require.NoError(t, err)

	// Check account
	acc, err := accounts.FindByEmail(context.Background(), "lea@example.com")
	require.NoError(t, err)
	assert.Equal(t, account.KindMember, acc.Kind)
	assert.Equal(t, account.StatusPendingVerification, acc.Status)

	// Check member
	m, err := members.FindByAccountID(context.Background(), acc.ID)
	require.NoError(t, err)
	assert.Equal(t, "Léa", m.FirstName)
	assert.Equal(t, "Martin", m.LastName)

	// Check audit and email
	assert.True(t, auditLog.ContainsType("account.created"))
	assert.Equal(t, 1, emails.SentCount())
}

func TestRegisterMember_TooYoung(t *testing.T) {
	svc, _, _, _, _ := newRegisterMemberService(t)

	input := auth.RegisterMemberInput{
		Email:     "too-young@example.com",
		Password:  "SecurePass123!",
		FirstName: "Young",
		LastName:  "Léa",
		BirthDate: "2015-01-01",
	}

	err := svc.RegisterMember(context.Background(), "1.2.3.4", input)

	assert.ErrorIs(t, err, member.ErrTooYoung)
}

func TestRegisterMember_EmailTaken(t *testing.T) {
	svc, accounts, _, _, _ := newRegisterMemberService(t)

	// Create existing account
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	existing, _ := account.New("acc-1", "lea@example.com", now)
	_ = accounts.Create(context.Background(), existing)

	input := auth.RegisterMemberInput{
		Email:     "lea@example.com",
		Password:  "SecurePass123!",
		FirstName: "Léa",
		LastName:  "Martin",
		BirthDate: "2000-01-15",
	}

	err := svc.RegisterMember(context.Background(), "1.2.3.4", input)

	assert.ErrorIs(t, err, account.ErrEmailTaken)
}

func TestRegisterMember_WeakPassword(t *testing.T) {
	svc, _, _, _, _ := newRegisterMemberService(t)

	input := auth.RegisterMemberInput{
		Email:     "weak@example.com",
		Password:  "123",
		FirstName: "Léa",
		LastName:  "Martin",
		BirthDate: "2000-01-15",
	}

	err := svc.RegisterMember(context.Background(), "1.2.3.4", input)

	assert.ErrorIs(t, err, account.ErrPasswordTooShort)
}
