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

func newUpdateProfileService(t *testing.T) (*profile.UpdateProfileService, *fakes.AccountRepository, *fakes.MemberRepository, *fakes.AssociationRepository, *fakes.SessionRepository, *fakes.AuditRepository) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	members := fakes.NewMemberRepository()
	associations := fakes.NewAssociationRepository()
	sessions := fakes.NewSessionRepository()
	audit := fakes.NewAuditRepository()
	clk := fakes.NewClock(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	tokenGen := fakes.NewTokenGenerator()

	svc := profile.NewUpdateProfileService(accounts, members, associations, sessions, audit, clk, tokenGen)
	return svc, accounts, members, associations, sessions, audit
}

func TestUpdateProfile_Member_Success(t *testing.T) {
	svc, accounts, members, _, sessions, audit := newUpdateProfileService(t)
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

	nickname := "lea42"
	aboutMe := "Passionnée de code"
	err := svc.UpdateProfile(context.Background(), "sess-1", "1.2.3.4", profile.UpdateProfileInput{
		Nickname: &nickname,
		AboutMe:  &aboutMe,
	})

	require.NoError(t, err)

	// Check member
	mUpdated, _ := members.FindByAccountID(context.Background(), "acc-1")
	assert.Equal(t, "lea42", mUpdated.Nickname)
	assert.Equal(t, "Passionnée de code", mUpdated.AboutMe)

	// Check audit
	assert.True(t, audit.ContainsType("profile.updated"))
}

func TestUpdateProfile_Member_TooLong(t *testing.T) {
	svc, accounts, members, _, sessions, _ := newUpdateProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Seed data
	acc, _ := account.New("acc-1", "lea@example.com", now)
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	_ = members.Create(context.Background(), m)
	sess := &session.Session{ID: "sess-1", AccountID: "acc-1", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	tooLong := ""
	for i := 0; i < 501; i++ {
		tooLong += "a"
	}

	err := svc.UpdateProfile(context.Background(), "sess-1", "1.2.3.4", profile.UpdateProfileInput{
		AboutMe: &tooLong,
	})

	assert.Error(t, err)
}

func TestUpdateProfile_Association_Success(t *testing.T) {
	svc, accounts, _, associations, sessions, audit := newUpdateProfileService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Seed data
	acc, _ := account.New("acc-2", "asso@example.com", now)
	acc.Kind = account.KindAssociation
	_ = accounts.Create(context.Background(), acc)

	asso, _ := association.New("acc-2", "Les Amis du Code", "75011", now)
	_ = associations.Create(context.Background(), asso)

	sess := &session.Session{ID: "sess-2", AccountID: "acc-2", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	about := "Association de codeurs"
	postalCode := "75012"
	err := svc.UpdateProfile(context.Background(), "sess-2", "1.2.3.4", profile.UpdateProfileInput{
		About:      &about,
		PostalCode: &postalCode,
	})

	require.NoError(t, err)

	// Check association
	assoUpdated, _ := associations.FindByAccountID(context.Background(), "acc-2")
	assert.Equal(t, "Association de codeurs", assoUpdated.About)
	assert.Equal(t, "75012", assoUpdated.PostalCode)

	// Check audit
	assert.True(t, audit.ContainsType("profile.updated"))
}
