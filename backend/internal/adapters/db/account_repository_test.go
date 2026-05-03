package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/domain/account"
)

func makeAccount(now time.Time) *account.Account {
	return &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		Kind:           account.KindMember,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func TestAccountRepository_Create(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := makeAccount(now)

	require.NoError(t, repo.Create(context.Background(), a))

	found, err := repo.FindByEmailCanonical(context.Background(), "lea@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, a.ID, found.ID)
	assert.Equal(t, a.Email, found.Email)
	assert.Equal(t, a.EmailCanonical, found.EmailCanonical)
	assert.Equal(t, a.PasswordHash, found.PasswordHash)
	assert.Equal(t, a.Status, found.Status)
	assert.Equal(t, a.Kind, found.Kind)
}

func TestAccountRepository_Create_DuplicateEmail(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := makeAccount(now)
	require.NoError(t, repo.Create(context.Background(), a))

	a2 := &account.Account{
		ID:             "acc2",
		Email:          "different@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash2",
		Status:         account.StatusPendingVerification,
		Kind:           account.KindMember,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.ErrorIs(t, repo.Create(context.Background(), a2), account.ErrEmailTaken)
}

func TestAccountRepository_FindByID(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := makeAccount(now)
	require.NoError(t, repo.Create(context.Background(), a))

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, a.ID, found.ID)
	assert.Equal(t, a.Email, found.Email)
}

func TestAccountRepository_FindByEmailCanonical_NotFound(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)
	found, err := repo.FindByEmailCanonical(context.Background(), "nonexistent@example.com")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestAccountRepository_UpdateStatus(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := makeAccount(now)
	require.NoError(t, repo.Create(context.Background(), a))

	later := now.Add(time.Hour)
	require.NoError(t, repo.UpdateStatus(context.Background(), "acc1", account.StatusActive, later))

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	assert.Equal(t, account.StatusActive, found.Status)
}

func TestAccountRepository_UpdatePasswordHash(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := makeAccount(now)
	require.NoError(t, repo.Create(context.Background(), a))

	later := now.Add(time.Hour)
	require.NoError(t, repo.UpdatePasswordHash(context.Background(), "acc1", "newhash", later))

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	assert.Equal(t, "newhash", found.PasswordHash)
}

func TestAccountRepository_TouchLastLogin(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := makeAccount(now)
	require.NoError(t, repo.Create(context.Background(), a))

	later := now.Add(90 * time.Minute)
	require.NoError(t, repo.TouchLastLogin(context.Background(), "acc1", later))

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found.LastLoginAt)
	assert.WithinDuration(t, later, *found.LastLoginAt, 2*time.Second)
}
