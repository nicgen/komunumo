package token

import (
	"errors"
	"time"
)

type Kind string

const (
	KindEmailVerification Kind = "email_verification"
	KindPasswordReset     Kind = "password_reset"
)

const (
	EmailVerificationTTL = 24 * time.Hour
	PasswordResetTTL     = 30 * time.Minute
)

type Token struct {
	ID         string
	AccountID  string
	Kind       Kind
	TokenHash  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	ConsumedAt *time.Time
}

var (
	ErrTokenNotFound        = errors.New("token: not found")
	ErrTokenExpired         = errors.New("token: expired")
	ErrTokenAlreadyConsumed = errors.New("token: already consumed")
	ErrTokenKindMismatch    = errors.New("token: kind mismatch")
)

// New creates a Token with a TTL starting from now.
func New(id, accountID string, kind Kind, tokenHash string, now time.Time, ttl time.Duration) *Token {
	return &Token{
		ID:        id,
		AccountID: accountID,
		Kind:      kind,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
}

// Active returns true if the token is neither expired nor already consumed.
func (t *Token) Active(now time.Time) bool {
	if t.ConsumedAt != nil {
		return false
	}
	return t.ExpiresAt.After(now)
}

// Consume marks the token as consumed. Returns ErrTokenExpired if expired,
// ErrTokenAlreadyConsumed if already consumed.
func (t *Token) Consume(now time.Time) error {
	if t.ConsumedAt != nil {
		return ErrTokenAlreadyConsumed
	}
	if !t.ExpiresAt.After(now) {
		return ErrTokenExpired
	}
	t.ConsumedAt = &now
	return nil
}
