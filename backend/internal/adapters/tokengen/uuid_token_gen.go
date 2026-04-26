package tokengen

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

const RawTokenBytes = 32

type UUIDTokenGen struct{}

func New() *UUIDTokenGen { return &UUIDTokenGen{} }

// NewRawToken returns 32 cryptographically random bytes encoded as
// base64 URL-safe (no padding) — ~43 chars, safe in URLs and emails.
func (UUIDTokenGen) NewRawToken() (string, error) {
	buf := make([]byte, RawTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("tokengen.NewRawToken: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashToken returns the SHA-256 hex digest of the raw token. Stored in DB.
func (UUIDTokenGen) HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// NewID returns a UUID v7 (time-ordered, 36-char canonical form).
func (UUIDTokenGen) NewID() string {
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback: v4. NewV7 only fails if rand.Read fails, which means
		// the OS entropy pool is broken — extremely rare; fall back to v4
		// rather than crash the request.
		id = uuid.New()
	}
	return id.String()
}
