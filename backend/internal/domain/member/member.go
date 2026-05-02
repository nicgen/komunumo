package member

import (
	"errors"
	"time"
)

type Visibility string

const (
	VisibilityPublic      Visibility = "public"
	VisibilityMembersOnly Visibility = "members_only"
	VisibilityPrivate     Visibility = "private"
)

type Member struct {
	AccountID  string
	FirstName  string
	LastName   string
	BirthDate  string // ISO 8601 YYYY-MM-DD
	Nickname   string
	AboutMe    string
	AvatarPath string
	Visibility Visibility
}

var (
	ErrTooYoung         = errors.New("member: registrant must be at least 18 years old")
	ErrInvalidName      = errors.New("member: first_name and last_name are required")
	ErrInvalidBirthDate = errors.New("member: birth_date must be a valid YYYY-MM-DD date")
	ErrAboutMeTooLong   = errors.New("member: about_me must not exceed 500 characters")
)

// New creates a Member after validating all invariants against the provided time.
func New(accountID, firstName, lastName, birthDate string, now time.Time) (*Member, error) {
	if firstName == "" || lastName == "" {
		return nil, ErrInvalidName
	}

	dob, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return nil, ErrInvalidBirthDate
	}

	now = now.UTC()
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	if age < 18 {
		return nil, ErrTooYoung
	}

	return &Member{
		AccountID:  accountID,
		FirstName:  firstName,
		LastName:   lastName,
		BirthDate:  birthDate,
		Visibility: VisibilityPublic,
	}, nil
}

// SetAboutMe validates and sets the about_me field.
func (m *Member) SetAboutMe(text string, _ time.Time) error {
	if len([]rune(text)) > 500 {
		return ErrAboutMeTooLong
	}
	m.AboutMe = text
	return nil
}

func (m *Member) Validate() error {
	if len([]rune(m.AboutMe)) > 500 {
		return ErrAboutMeTooLong
	}
	switch m.Visibility {
	case VisibilityPublic, VisibilityMembersOnly, VisibilityPrivate:
	default:
		m.Visibility = VisibilityPublic
	}
	return nil
}
