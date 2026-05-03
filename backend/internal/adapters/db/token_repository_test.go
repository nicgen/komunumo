package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/token"
)

func seedTokenAccount(t *testing.T, accountRepo *db.AccountRepository, now time.Time) {
	t.Helper()
	a := &account.Account{
		ID:             "acc1",
		Email:          "lea@example.com",
		EmailCanonical: "lea@example.com",
		PasswordHash:   "hash",
		Status:         account.StatusPendingVerification,
		Kind:           account.KindMember,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, accountRepo.Create(context.Background(), a))
}

func TestTokenRepository_Create_FindActiveByHash(t *testing.T) {
	conn := openTestDB(t)
	accountRepo := db.NewAccountRepository(conn)
	tokenRepo := db.NewTokenRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	seedTokenAccount(t, accountRepo, now)

	// Create a token
	tok := &token.Token{
		ID:        "tok1",
		AccountID: "acc1",
		Kind:      token.KindEmailVerification,
		TokenHash: "hash123",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err := tokenRepo.Create(context.Background(), tok)
	require.NoError(t, err)

	// Find active by hash
	found, err := tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "hash123", now)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, tok.ID, found.ID)
	assert.Equal(t, tok.AccountID, found.AccountID)
	assert.Equal(t, tok.TokenHash, found.TokenHash)
	assert.Nil(t, found.ConsumedAt)
}

func TestTokenRepository_FindActiveByHash_NotFound(t *testing.T) {
	conn := openTestDB(t)
	tokenRepo := db.NewTokenRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)

	found, err := tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "unknown", now)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestTokenRepository_FindActiveByHash_Expired(t *testing.T) {
	conn := openTestDB(t)
	accountRepo := db.NewAccountRepository(conn)
	tokenRepo := db.NewTokenRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	seedTokenAccount(t, accountRepo, now)

	// Create an expired token
	tok := &token.Token{
		ID:        "tok1",
		AccountID: "acc1",
		Kind:      token.KindEmailVerification,
		TokenHash: "hash123",
		CreatedAt: now.Add(-48 * time.Hour),
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	err := tokenRepo.Create(context.Background(), tok)
	require.NoError(t, err)

	// Find active by hash — expired token returns ErrTokenExpired, not nil
	found, err := tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "hash123", now)
	require.ErrorIs(t, err, token.ErrTokenExpired)
	assert.Nil(t, found)
}

func TestTokenRepository_Consume(t *testing.T) {
	conn := openTestDB(t)
	accountRepo := db.NewAccountRepository(conn)
	tokenRepo := db.NewTokenRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	seedTokenAccount(t, accountRepo, now)

	// Create a token
	tok := &token.Token{
		ID:        "tok1",
		AccountID: "acc1",
		Kind:      token.KindEmailVerification,
		TokenHash: "hash123",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err := tokenRepo.Create(context.Background(), tok)
	require.NoError(t, err)

	// Consume the token
	laterTime := now.Add(1 * time.Hour)
	err = tokenRepo.Consume(context.Background(), token.KindEmailVerification, "tok1", laterTime)
	require.NoError(t, err)

	// Find active by hash — consumed token returns ErrTokenNotFound
	found, err := tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "hash123", laterTime)
	require.ErrorIs(t, err, token.ErrTokenNotFound)
	assert.Nil(t, found)
}

func TestTokenRepository_RevokeActiveForAccount(t *testing.T) {
	conn := openTestDB(t)
	accountRepo := db.NewAccountRepository(conn)
	tokenRepo := db.NewTokenRepository(conn)

	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	seedTokenAccount(t, accountRepo, now)

	// Create two tokens for the same account
	tok1 := &token.Token{
		ID:        "tok1",
		AccountID: "acc1",
		Kind:      token.KindEmailVerification,
		TokenHash: "hash123",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err := tokenRepo.Create(context.Background(), tok1)
	require.NoError(t, err)

	tok2 := &token.Token{
		ID:        "tok2",
		AccountID: "acc1",
		Kind:      token.KindEmailVerification,
		TokenHash: "hash456",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	err = tokenRepo.Create(context.Background(), tok2)
	require.NoError(t, err)

	// Verify both are active
	found1, err := tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "hash123", now)
	require.NoError(t, err)
	assert.NotNil(t, found1)

	found2, err := tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "hash456", now)
	require.NoError(t, err)
	assert.NotNil(t, found2)

	// Revoke all active tokens for this account
	laterTime := now.Add(1 * time.Hour)
	err = tokenRepo.RevokeActiveForAccount(context.Background(), token.KindEmailVerification, "acc1", laterTime)
	require.NoError(t, err)

	// Verify both are now revoked — consumed tokens return ErrTokenNotFound
	found1, err = tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "hash123", laterTime)
	require.ErrorIs(t, err, token.ErrTokenNotFound)
	assert.Nil(t, found1)

	found2, err = tokenRepo.FindActiveByHash(context.Background(), token.KindEmailVerification, "hash456", laterTime)
	require.ErrorIs(t, err, token.ErrTokenNotFound)
	assert.Nil(t, found2)
}
