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

func TestAccountRepository_Create(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		FirstName:      "Léa",
		LastName:       "Dupont",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(context.Background(), a)
	require.NoError(t, err)

	found, err := repo.FindByEmailCanonical(context.Background(), "lea@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, a.ID, found.ID)
	assert.Equal(t, a.Email, found.Email)
	assert.Equal(t, a.EmailCanonical, found.EmailCanonical)
	assert.Equal(t, a.PasswordHash, found.PasswordHash)
	assert.Equal(t, a.Status, found.Status)
	assert.Equal(t, a.FirstName, found.FirstName)
	assert.Equal(t, a.LastName, found.LastName)
	assert.Equal(t, a.DateOfBirth, found.DateOfBirth)
	assert.NotNil(t, found.CreatedAt)
	assert.NotNil(t, found.UpdatedAt)
}

func TestAccountRepository_Create_DuplicateEmail(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		FirstName:      "Léa",
		LastName:       "Dupont",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(context.Background(), a)
	require.NoError(t, err)

	a2 := &account.Account{
		ID:             "acc2",
		Email:          "different@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash2",
		Status:         account.StatusPendingVerification,
		FirstName:      "Another",
		LastName:       "User",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = repo.Create(context.Background(), a2)
	require.Error(t, err)
	assert.Equal(t, account.ErrEmailTaken, err)
}

func TestAccountRepository_FindByID(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		FirstName:      "Léa",
		LastName:       "Dupont",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(context.Background(), a)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, a.ID, found.ID)
	assert.Equal(t, a.Email, found.Email)
	assert.Equal(t, a.EmailCanonical, found.EmailCanonical)
	assert.Equal(t, a.FirstName, found.FirstName)
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
	a := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		FirstName:      "Léa",
		LastName:       "Dupont",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(context.Background(), a)
	require.NoError(t, err)

	later := time.Date(2026, 4, 27, 13, 0, 0, 0, time.UTC)
	err = repo.UpdateStatus(context.Background(), "acc1", account.StatusVerified, later)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, account.StatusVerified, found.Status)
}

func TestAccountRepository_UpdatePasswordHash(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		FirstName:      "Léa",
		LastName:       "Dupont",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(context.Background(), a)
	require.NoError(t, err)

	later := time.Date(2026, 4, 27, 13, 0, 0, 0, time.UTC)
	newHash := "newhash123"
	err = repo.UpdatePasswordHash(context.Background(), "acc1", newHash, later)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, newHash, found.PasswordHash)
}

func TestAccountRepository_TouchLastLogin(t *testing.T) {
	conn := openTestDB(t)
	repo := db.NewAccountRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	a := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		FirstName:      "Léa",
		LastName:       "Dupont",
		DateOfBirth:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(context.Background(), a)
	require.NoError(t, err)

	later := time.Date(2026, 4, 27, 13, 30, 0, 0, time.UTC)
	err = repo.TouchLastLogin(context.Background(), "acc1", later)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.NotNil(t, found.LastLoginAt)
	assert.WithinDuration(t, later, *found.LastLoginAt, 2*time.Second)
}
