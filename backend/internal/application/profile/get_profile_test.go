package profile_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/profile"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/domain/association"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports/fakes"
)

func newGetProfileService(t *testing.T) (*profile.GetProfileService, *fakes.AccountRepository, *fakes.MemberRepository, *fakes.AssociationRepository, *fakes.SessionRepository) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	members := fakes.NewMemberRepository()
	associations := fakes.NewAssociationRepository()
	sessions := fakes.NewSessionRepository()
	clk := fakes.NewClock(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))

	svc := profile.NewGetProfileService(accounts, members, associations, sessions, clk)
	return svc, accounts, members, associations, sessions
}

func TestGetMyProfile_Member_Success(t *testing.T) {
	svc, accounts, members, _, sessions := newGetProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Seed data
	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)

	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	_ = members.Create(context.Background(), m)

	sess := &session.Session{
		ID:        "sess-1",
		AccountID: "acc-1",
		ExpiresAt: now.Add(1 * time.Hour),
	}
	_ = sessions.Create(context.Background(), sess)

	out, err := svc.GetMyProfile(context.Background(), "sess-1")

	require.NoError(t, err)
	assert.Equal(t, "acc-1", out.AccountID)
	assert.Equal(t, "lea@example.com", out.Email)
	assert.Equal(t, string(account.KindMember), out.Kind)
	assert.Equal(t, "Léa", out.FirstName)
	assert.Equal(t, "Martin", out.LastName)
}

func TestGetMyProfile_Association_Success(t *testing.T) {
	svc, accounts, _, associations, sessions := newGetProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Seed data
	acc, _ := account.New("acc-2", "asso@example.com", now)
	acc.Kind = account.KindAssociation
	_ = accounts.Create(context.Background(), acc)

	asso, _ := association.New("acc-2", "Les Amis du Code", "75011", now)
	_ = associations.Create(context.Background(), asso)

	sess := &session.Session{
		ID:        "sess-2",
		AccountID: "acc-2",
		ExpiresAt: now.Add(1 * time.Hour),
	}
	_ = sessions.Create(context.Background(), sess)

	out, err := svc.GetMyProfile(context.Background(), "sess-2")

	require.NoError(t, err)
	assert.Equal(t, "acc-2", out.AccountID)
	assert.Equal(t, "asso@example.com", out.Email)
	assert.Equal(t, string(account.KindAssociation), out.Kind)
	assert.Equal(t, "Les Amis du Code", out.LegalName)
}

func TestGetMyProfile_SessionInvalid(t *testing.T) {
	svc, _, _, _, _ := newGetProfileService(t)

	_, err := svc.GetMyProfile(context.Background(), "invalid")

	assert.Error(t, err) // Should be ErrUnauthorized or similar
}
func TestGetPublicProfile_Public(t *testing.T) {
	svc, accounts, members, _, _ := newGetProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)

	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	m.Visibility = member.VisibilityPublic
	_ = members.Create(context.Background(), m)

	out, err := svc.GetPublicProfile(context.Background(), "acc-1", "")

	require.NoError(t, err)
	assert.Equal(t, "acc-1", out.AccountID)
	assert.Empty(t, out.BirthDate) // Protected PII
	assert.Equal(t, "Léa", out.FirstName)
}

func TestGetPublicProfile_Private(t *testing.T) {
	svc, accounts, members, _, _ := newGetProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)

	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	m.Visibility = member.VisibilityPrivate
	_ = members.Create(context.Background(), m)

	_, err := svc.GetPublicProfile(context.Background(), "acc-1", "")

	assert.ErrorIs(t, err, profile.ErrNotFound)
}

func TestGetPublicProfile_MembersOnly_WithoutSession(t *testing.T) {
	svc, accounts, members, _, _ := newGetProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)

	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	m.Visibility = member.VisibilityMembersOnly
	_ = members.Create(context.Background(), m)

	_, err := svc.GetPublicProfile(context.Background(), "acc-1", "")

	assert.ErrorIs(t, err, profile.ErrNotFound)
}

func TestGetPublicProfile_MembersOnly_WithSession(t *testing.T) {
	svc, accounts, members, _, sessions := newGetProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)

	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	m.Visibility = member.VisibilityMembersOnly
	_ = members.Create(context.Background(), m)

	sess := &session.Session{ID: "sess-1", AccountID: "viewer", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	out, err := svc.GetPublicProfile(context.Background(), "acc-1", "sess-1")

	require.NoError(t, err)
	assert.Equal(t, "acc-1", out.AccountID)
	assert.Empty(t, out.BirthDate)
}
