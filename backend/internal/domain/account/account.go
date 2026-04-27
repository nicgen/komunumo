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
	StatusVerified            Status = "verified"
	StatusDisabled            Status = "disabled"
)

type Account struct {
	ID             string
	Email          string
	EmailCanonical string
	PasswordHash   string
	Status         Status
	FirstName      string
	LastName       string
	DateOfBirth    time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastLoginAt    *time.Time
}

var (
	ErrEmailTaken         = errors.New("account: email already in use")
	ErrAgeBelow16         = errors.New("account: registrant must be at least 16")
	ErrInvalidStatus      = errors.New("account: invalid status")
	ErrInvalidTransition  = errors.New("account: invalid status transition")
	ErrPasswordTooShort   = errors.New("account: password too short")
	ErrPasswordTooWeak    = errors.New("account: password lacks required character classes")
	ErrEmailMalformed     = errors.New("account: email is malformed")
	ErrAccountNotFound    = errors.New("account: not found")
	ErrAccountNotVerified = errors.New("account: not verified")
	ErrAccountDisabled    = errors.New("account: disabled")
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

// New creates an Account in pending_verification status. Returns an error if
// email is malformed or the registrant is under 16.
func New(id, email, firstName, lastName string, dob, now time.Time) (*Account, error) {
	canonical, err := CanonicalizeEmail(email)
	if err != nil {
		return nil, err
	}
	age := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		age--
	}
	if age < 16 {
		return nil, ErrAgeBelow16
	}
	return &Account{
		ID:             id,
		Email:          email,
		EmailCanonical: canonical,
		Status:         StatusPendingVerification,
		FirstName:      firstName,
		LastName:       lastName,
		DateOfBirth:    dob,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Verify transitions the account from pending_verification to verified.
func (a *Account) Verify(at time.Time) error {
	if a.Status != StatusPendingVerification {
		return ErrInvalidTransition
	}
	a.Status = StatusVerified
	a.UpdatedAt = at
	return nil
}

// Disable transitions the account to disabled from any non-disabled status.
func (a *Account) Disable(at time.Time) error {
	if a.Status == StatusDisabled {
		return ErrInvalidTransition
	}
	a.Status = StatusDisabled
	a.UpdatedAt = at
	return nil
}
