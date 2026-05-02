package profile

import (
	"context"

	"komunumo/backend/internal/ports"
)

type UpdateProfileInput struct {
	// Member specific
	Nickname  *string `json:"nickname,omitempty"`
	AboutMe   *string `json:"about_me,omitempty"`
	
	// Association specific
	About      *string `json:"about,omitempty"`
	PostalCode *string `json:"postal_code,omitempty"`
	
	// Common
	Visibility *string `json:"visibility,omitempty"`
}

type UpdateProfileService struct {
	accounts     ports.AccountRepository
	members      ports.MemberRepository
	associations ports.AssociationRepository
	sessions     ports.SessionRepository
	audit        ports.AuditRepository
	clock        ports.Clock
}

func NewUpdateProfileService(
	accounts ports.AccountRepository,
	members ports.MemberRepository,
	associations ports.AssociationRepository,
	sessions ports.SessionRepository,
	audit ports.AuditRepository,
	clock ports.Clock,
) *UpdateProfileService {
	return &UpdateProfileService{
		accounts:     accounts,
		members:      members,
		associations: associations,
		sessions:     sessions,
		audit:        audit,
		clock:        clock,
	}
}

func (s *UpdateProfileService) UpdateProfile(ctx context.Context, sessionID string, ip string, in UpdateProfileInput) error {
	return nil
}
