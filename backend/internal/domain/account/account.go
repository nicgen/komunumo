package account

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
)

type Status string

const (
	StatusPendingVerification Status = "pending_verification"
	StatusActive              Status = "active"
	StatusSuspended           Status = "suspended"
	StatusDeleted             Status = "deleted"

	// Deprecated aliases kept for existing code compatibility.
	StatusVerified Status = StatusActive
	StatusDisabled Status = StatusSuspended
)

type Kind string

const (
	KindMember      Kind = "member"
	KindAssociation Kind = "association"
)

type Account struct {
	ID             string
	Email          string
	EmailCanonical string
	PasswordHash   string
	Status         Status
	Kind           Kind
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastLoginAt    *time.Time
}

var (
	ErrEmailTaken         = errors.New("account: email already in use")
	ErrInvalidStatus      = errors.New("account: invalid status")
	ErrInvalidTransition  = errors.New("account: invalid status transition")
	ErrPasswordTooShort   = errors.New("account: password too short")
	ErrPasswordTooWeak    = errors.New("account: password lacks required character classes")
	ErrEmailMalformed     = errors.New("account: email is malformed")
	ErrAccountNotFound    = errors.New("account: not found")
	ErrAccountNotVerified = errors.New("account: not verified")
	ErrAccountDisabled    = errors.New("account: disabled")
	ErrAccountSuspended   = errors.New("account: suspended")
)

// CanonicalizeEmail applies NFKC normalization and lowercases the email.
func CanonicalizeEmail(email string) (string, error) {
	normalized := norm.NFKC.String(email)
	at := strings.LastIndex(normalized, "@")
	if at <= 0 || at == len(normalized)-1 {
		return "", ErrEmailMalformed
	}
	local := normalized[:at]
	domain := normalized[at+1:]
	if local == "" || domain == "" || strings.Contains(domain, ".") == false {
		return "", ErrEmailMalformed
	}
	return strings.ToLower(normalized), nil
}

// New creates an Account in pending_verification status.
func New(id, email string, now time.Time) (*Account, error) {
	canonical, err := CanonicalizeEmail(email)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:             id,
		Email:          email,
		EmailCanonical: canonical,
		Status:         StatusPendingVerification,
		Kind:           KindMember,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Verify transitions the account from pending_verification to active.
func (a *Account) Verify(at time.Time) error {
	if a.Status != StatusPendingVerification {
		return ErrInvalidTransition
	}
	a.Status = StatusActive
	a.UpdatedAt = at
	return nil
}

// Disable transitions the account to suspended from any non-suspended status.
func (a *Account) Disable(at time.Time) error {
	if a.Status == StatusSuspended {
		return ErrInvalidTransition
	}
	a.Status = StatusSuspended
	a.UpdatedAt = at
	return nil
}
