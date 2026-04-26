package account

import (
	"errors"
	"time"
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
