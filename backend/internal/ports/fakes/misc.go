package fakes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"komunumo/backend/internal/ports"
)

// Clock

var _ ports.Clock = (*Clock)(nil)

type Clock struct{ T time.Time }

func NewClock(t time.Time) *Clock  { return &Clock{T: t} }
func (c *Clock) Now() time.Time    { return c.T }
func (c *Clock) Advance(d time.Duration) { c.T = c.T.Add(d) }

// TokenGenerator

var _ ports.TokenGenerator = (*TokenGenerator)(nil)

type TokenGenerator struct {
	counter int
}

func NewTokenGenerator() *TokenGenerator { return &TokenGenerator{} }

func (g *TokenGenerator) NewRawToken() (string, error) {
	g.counter++
	return fmt.Sprintf("raw-token-%d", g.counter), nil
}

func (g *TokenGenerator) HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (g *TokenGenerator) NewID() string {
	g.counter++
	return fmt.Sprintf("fake-id-%d", g.counter)
}

// PasswordHasher

var _ ports.PasswordHasher = (*PasswordHasher)(nil)

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher { return &PasswordHasher{} }

func (h *PasswordHasher) Hash(plaintext string) (string, error) {
	return "hash:" + plaintext, nil
}

func (h *PasswordHasher) Verify(hash, plaintext string) (bool, error) {
	return hash == "hash:"+plaintext, nil
}

// RateLimiter

var _ ports.RateLimiter = (*RateLimiter)(nil)

type RateLimiter struct {
	blocked map[string]bool
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{blocked: make(map[string]bool)}
}

func (r *RateLimiter) Block(key string) { r.blocked[key] = true }

func (r *RateLimiter) Allow(_ context.Context, key string) (bool, time.Duration) {
	if r.blocked[key] {
		return false, 30 * time.Minute
	}
	return true, 0
}

// UnitOfWork (no-op: just calls fn)

var _ ports.UnitOfWork = (*UnitOfWork)(nil)

type UnitOfWork struct{}

func NewUnitOfWork() *UnitOfWork { return &UnitOfWork{} }

func (u *UnitOfWork) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
