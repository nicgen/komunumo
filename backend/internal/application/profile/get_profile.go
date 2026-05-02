package profile

import (
	"context"
	"time"

	"komunumo/backend/internal/ports"
)

type ProfileOutput struct {
	AccountID   string    `json:"account_id"`
	Email       string    `json:"email"`
	Kind        string    `json:"kind"`
	FirstName   string    `json:"first_name,omitempty"`
	LastName    string    `json:"last_name,omitempty"`
	BirthDate   string    `json:"birth_date,omitempty"`
	Nickname    string    `json:"nickname,omitempty"`
	AboutMe     string    `json:"about_me,omitempty"`
	AvatarPath  string    `json:"avatar_path,omitempty"`
	LegalName   string    `json:"legal_name,omitempty"`
	SIREN       string    `json:"siren,omitempty"`
	RNA         string    `json:"rna,omitempty"`
	PostalCode  string    `json:"postal_code,omitempty"`
	About       string    `json:"about,omitempty"`
	LogoPath    string    `json:"logo_path,omitempty"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetProfileService struct {
	accounts     ports.AccountRepository
	members      ports.MemberRepository
	associations ports.AssociationRepository
	sessions     ports.SessionRepository
}

func NewGetProfileService(
	accounts ports.AccountRepository,
	members ports.MemberRepository,
	associations ports.AssociationRepository,
	sessions ports.SessionRepository,
) *GetProfileService {
	return &GetProfileService{
		accounts:     accounts,
		members:      members,
		associations: associations,
		sessions:     sessions,
	}
}

func (s *GetProfileService) GetMyProfile(ctx context.Context, sessionID string) (*ProfileOutput, error) {
	return nil, nil
}
