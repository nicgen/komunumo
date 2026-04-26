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

// Active returns true if the token is neither expired nor already consumed.
func (t *Token) Active(now time.Time) bool {
	if t.ConsumedAt != nil {
		return false
	}
	return t.ExpiresAt.After(now)
}
