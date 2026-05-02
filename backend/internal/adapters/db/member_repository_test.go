package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/member"
)

func TestMemberRepository_CreateAndFind(t *testing.T) {
	conn := openTestDB(t)
	accRepo := db.NewAccountRepository(conn)
	memberRepo := db.NewMemberRepository(conn)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Prerequisite: Account
	acc := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		Kind:           account.KindMember,
		Status:         account.StatusPendingVerification,
	}
	require.NoError(t, accRepo.Create(context.Background(), acc))

	m, _ := member.New("acc1", "Léa", "Martin", "2000-01-15", now)
	m.Nickname = "lea42"
	m.AboutMe = "Passionnée de code"
	m.Visibility = member.VisibilityPublic

	require.NoError(t, memberRepo.Create(context.Background(), m))

	found, err := memberRepo.FindByAccountID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, m.FirstName, found.FirstName)
	assert.Equal(t, m.LastName, found.LastName)
	assert.Equal(t, m.BirthDate, found.BirthDate)
	assert.Equal(t, m.Nickname, found.Nickname)
	assert.Equal(t, m.AboutMe, found.AboutMe)
	assert.Equal(t, m.Visibility, found.Visibility)
}

func TestMemberRepository_Update(t *testing.T) {
	conn := openTestDB(t)
	accRepo := db.NewAccountRepository(conn)
	memberRepo := db.NewMemberRepository(conn)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc := &account.Account{ID: "acc1", Email: "lea@example.com", EmailCanonical: "lea@example.com", Kind: account.KindMember, Status: account.StatusPendingVerification}
	require.NoError(t, accRepo.Create(context.Background(), acc))

	m, _ := member.New("acc1", "Léa", "Martin", "2000-01-15", now)
	require.NoError(t, memberRepo.Create(context.Background(), m))

	m.Nickname = "new-nick"
	m.Visibility = member.VisibilityPrivate
	require.NoError(t, memberRepo.Update(context.Background(), m))

	found, err := memberRepo.FindByAccountID(context.Background(), "acc1")
	require.NoError(t, err)
	assert.Equal(t, "new-nick", found.Nickname)
	assert.Equal(t, member.VisibilityPrivate, found.Visibility)
}
