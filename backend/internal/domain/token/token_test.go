package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/domain/token"
)

var now = time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)

func TestNew_EmailVerification_TTL(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindEmailVerification, "hash", now, token.EmailVerificationTTL)
	assert.Equal(t, now.Add(token.EmailVerificationTTL), tok.ExpiresAt)
	assert.Equal(t, token.EmailVerificationTTL, 24*time.Hour)
}

func TestNew_PasswordReset_TTL(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindPasswordReset, "hash", now, token.PasswordResetTTL)
	assert.Equal(t, now.Add(token.PasswordResetTTL), tok.ExpiresAt)
	assert.Equal(t, token.PasswordResetTTL, 30*time.Minute)
}

func TestActive_WithinTTL(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindEmailVerification, "hash", now, token.EmailVerificationTTL)
	assert.True(t, tok.Active(now.Add(23*time.Hour)))
	assert.True(t, tok.Active(now))
}

func TestActive_Expired(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindEmailVerification, "hash", now, token.EmailVerificationTTL)
	assert.False(t, tok.Active(now.Add(token.EmailVerificationTTL+time.Second)))
	assert.False(t, tok.Active(now.Add(token.EmailVerificationTTL)))
}

func TestConsume_SetsConsumedAt(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindEmailVerification, "hash", now, token.EmailVerificationTTL)
	consumedAt := now.Add(time.Hour)
	err := tok.Consume(consumedAt)
	require.NoError(t, err)
	require.NotNil(t, tok.ConsumedAt)
	assert.Equal(t, consumedAt, *tok.ConsumedAt)
}

func TestConsume_MakesTokenInactive(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindEmailVerification, "hash", now, token.EmailVerificationTTL)
	_ = tok.Consume(now.Add(time.Hour))
	assert.False(t, tok.Active(now.Add(2*time.Hour)))
}

func TestConsume_AlreadyConsumed_ReturnsError(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindEmailVerification, "hash", now, token.EmailVerificationTTL)
	_ = tok.Consume(now.Add(time.Hour))
	err := tok.Consume(now.Add(2 * time.Hour))
	require.ErrorIs(t, err, token.ErrTokenAlreadyConsumed)
}

func TestConsume_Expired_ReturnsError(t *testing.T) {
	tok := token.New("id1", "acc1", token.KindEmailVerification, "hash", now, token.EmailVerificationTTL)
	err := tok.Consume(now.Add(token.EmailVerificationTTL + time.Second))
	require.ErrorIs(t, err, token.ErrTokenExpired)
}
