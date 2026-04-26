package session

import (
	"errors"
	"time"
)

type Session struct {
	ID         string
	AccountID  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
	IP         string
	UserAgent  string
}

var (
	ErrSessionNotFound = errors.New("session: not found")
	ErrSessionExpired  = errors.New("session: expired")
)

// Expired returns true if expiresAt is non-zero and <= now.
func (s *Session) Expired(now time.Time) bool {
	return !s.ExpiresAt.IsZero() && !s.ExpiresAt.After(now)
}
