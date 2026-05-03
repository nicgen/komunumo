package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/association"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/ports/fakes"
)

func newRegisterAssociationService(t *testing.T) (*auth.RegisterAssociationService, *fakes.AccountRepository, *fakes.AssociationRepository, *fakes.MemberRepository, *fakes.MembershipRepository, *fakes.AuditRepository, *fakes.EmailSender) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	associations := fakes.NewAssociationRepository()
	members := fakes.NewMemberRepository()
	memberships := fakes.NewMembershipRepository()
	auditLog := fakes.NewAuditRepository()
	emails := fakes.NewEmailSender()
	hasher := fakes.NewPasswordHasher()
	tokenGen := fakes.NewTokenGenerator()
	tokens := fakes.NewTokenRepository()
	clk := fakes.NewClock(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	rl := fakes.NewRateLimiter()
	uow := fakes.NewUnitOfWork()

	svc := auth.NewRegisterAssociationService(
		accounts,
		associations,
		members,
		memberships,
		auditLog,
		emails,
		hasher,
		tokenGen,
		tokens,
		clk,
		rl,
		uow,
	)
	return svc, accounts, associations, members, memberships, auditLog, emails
}

func TestRegisterAssociation_Success(t *testing.T) {
	svc, accounts, associations, members, memberships, auditLog, emails := newRegisterAssociationService(t)

	input := auth.RegisterAssociationInput{
		Email:      "asso@example.com",
		Password:   "SecurePass123!",
		LegalName:  "Les Amis du Code",
		PostalCode: "75011",
		FirstName:  "Anne",
		LastName:   "Dupont",
		BirthDate:  "1985-06-20",
	}

	err := svc.RegisterAssociation(context.Background(), "1.2.3.4", input)

	require.NoError(t, err)

	// Check account
	canonical, _ := account.CanonicalizeEmail("asso@example.com")
	acc, err := accounts.FindByEmailCanonical(context.Background(), canonical)
	require.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, account.KindAssociation, acc.Kind)
	assert.Equal(t, account.StatusPendingVerification, acc.Status)

	// Check association
	asso, err := associations.FindByAccountID(context.Background(), acc.ID)
	require.NoError(t, err)
	assert.NotNil(t, asso)
	assert.Equal(t, "Les Amis du Code", asso.LegalName)
	assert.Equal(t, "75011", asso.PostalCode)

	// Check member (representative)
	m, err := members.FindByAccountID(context.Background(), acc.ID)
	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, "Anne", m.FirstName)
	assert.Equal(t, "Dupont", m.LastName)

	// Check membership (owner)
	ms, err := memberships.FindByAccountIDs(context.Background(), acc.ID, acc.ID)
	require.NoError(t, err)
	assert.NotNil(t, ms)
	assert.Equal(t, "owner", ms.Role)
	assert.Equal(t, "active", ms.Status)

	// Check audit and email
	assert.True(t, auditLog.ContainsType("account.created"))
	assert.Equal(t, 1, len(emails.Calls))
}

func TestRegisterAssociation_InvalidSIREN(t *testing.T) {
	svc, _, _, _, _, _, _ := newRegisterAssociationService(t)

	input := auth.RegisterAssociationInput{
		Email:      "asso@example.com",
		Password:   "SecurePass123!",
		LegalName:  "Les Amis du Code",
		PostalCode: "75011",
		SIREN:      "123", // Too short
		FirstName:  "Anne",
		LastName:   "Dupont",
		BirthDate:  "1985-06-20",
	}

	err := svc.RegisterAssociation(context.Background(), "1.2.3.4", input)

	assert.ErrorIs(t, err, association.ErrInvalidSIREN)
}

func TestRegisterAssociation_InvalidRNA(t *testing.T) {
	svc, _, _, _, _, _, _ := newRegisterAssociationService(t)

	input := auth.RegisterAssociationInput{
		Email:      "asso@example.com",
		Password:   "SecurePass123!",
		LegalName:  "Les Amis du Code",
		PostalCode: "75011",
		RNA:        "123456789", // Missing W
		FirstName:  "Anne",
		LastName:   "Dupont",
		BirthDate:  "1985-06-20",
	}

	err := svc.RegisterAssociation(context.Background(), "1.2.3.4", input)

	assert.ErrorIs(t, err, association.ErrInvalidRNA)
}

func TestRegisterAssociation_TooYoung(t *testing.T) {
	svc, _, _, _, _, _, _ := newRegisterAssociationService(t)

	input := auth.RegisterAssociationInput{
		Email:      "asso@example.com",
		Password:   "SecurePass123!",
		LegalName:  "Les Amis du Code",
		PostalCode: "75011",
		FirstName:  "Young",
		LastName:   "Anne",
		BirthDate:  "2015-01-01",
	}

	err := svc.RegisterAssociation(context.Background(), "1.2.3.4", input)

	assert.ErrorIs(t, err, member.ErrTooYoung)
}

func TestRegisterAssociation_MissingPostalCode(t *testing.T) {
	svc, _, _, _, _, _, _ := newRegisterAssociationService(t)

	input := auth.RegisterAssociationInput{
		Email:      "asso@example.com",
		Password:   "SecurePass123!",
		LegalName:  "Les Amis du Code",
		PostalCode: "",
		FirstName:  "Anne",
		LastName:   "Dupont",
		BirthDate:  "1985-06-20",
	}

	err := svc.RegisterAssociation(context.Background(), "1.2.3.4", input)

	assert.ErrorIs(t, err, association.ErrInvalidPostalCode)
}
