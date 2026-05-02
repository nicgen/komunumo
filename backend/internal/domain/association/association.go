package association

import (
	"errors"
	"regexp"
	"time"
)

type Visibility string

const (
	VisibilityPublic      Visibility = "public"
	VisibilityMembersOnly Visibility = "members_only"
	VisibilityPrivate     Visibility = "private"
)

type Association struct {
	AccountID  string
	LegalName  string
	SIREN      string
	RNA        string
	PostalCode string
	About      string
	LogoPath   string
	Visibility Visibility
}

var (
	ErrInvalidSIREN    = errors.New("association: siren must be exactly 9 digits")
	ErrInvalidRNA      = errors.New("association: rna must match W followed by 9 digits")
	ErrInvalidLegalName = errors.New("association: legal_name is required")
	ErrInvalidPostalCode = errors.New("association: postal_code is required")
	ErrAboutTooLong    = errors.New("association: about must not exceed 2000 characters")
)

var (
	reSIREN = regexp.MustCompile(`^\d{9}$`)
	reRNA   = regexp.MustCompile(`^W\d{9}$`)
)

// ValidateSIREN validates the SIREN format. Empty string is allowed (optional field).
func ValidateSIREN(siren string) error {
	if siren == "" {
		return nil
	}
	if !reSIREN.MatchString(siren) {
		return ErrInvalidSIREN
	}
	return nil
}

// ValidateRNA validates the RNA format. Empty string is allowed (optional field).
func ValidateRNA(rna string) error {
	if rna == "" {
		return nil
	}
	if !reRNA.MatchString(rna) {
		return ErrInvalidRNA
	}
	return nil
}

// New creates an Association after validating required fields.
func New(accountID, legalName, postalCode string, _ time.Time) (*Association, error) {
	if legalName == "" {
		return nil, ErrInvalidLegalName
	}
	if postalCode == "" {
		return nil, ErrInvalidPostalCode
	}
	return &Association{
		AccountID:  accountID,
		LegalName:  legalName,
		PostalCode: postalCode,
		Visibility: VisibilityPublic,
	}, nil
}

// SetAbout validates and sets the about field.
func (a *Association) SetAbout(text string) error {
	if len([]rune(text)) > 2000 {
		return ErrAboutTooLong
	}
	a.About = text
	return nil
}
