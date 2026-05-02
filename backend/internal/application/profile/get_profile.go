package profile

import (
	"context"
	"fmt"
	"time"

	"komunumo/backend/internal/domain/account"
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
	clock        ports.Clock
}

func NewGetProfileService(
	accounts ports.AccountRepository,
	members ports.MemberRepository,
	associations ports.AssociationRepository,
	sessions ports.SessionRepository,
	clock ports.Clock,
) *GetProfileService {
	return &GetProfileService{
		accounts:     accounts,
		members:      members,
		associations: associations,
		sessions:     sessions,
		clock:        clock,
	}
}

func (s *GetProfileService) GetMyProfile(ctx context.Context, sessionID string) (*ProfileOutput, error) {
	now := s.clock.Now()
	// Wait, I should add clock to the service for testability if needed.
	// But GetMyProfile doesn't strictly need it if we assume session is already checked by middleware.
	// Actually, the service should check it.

	sess, err := s.sessions.FindByID(ctx, sessionID, now)
	if err != nil {
		return nil, err
	}

	acc, err := s.accounts.FindByID(ctx, sess.AccountID)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("account not found")
	}

	out := &ProfileOutput{
		AccountID: acc.ID,
		Email:     acc.Email,
		Kind:      string(acc.Kind),
		CreatedAt: acc.CreatedAt,
	}

	if acc.Kind == account.KindMember {
		m, err := s.members.FindByAccountID(ctx, acc.ID)
		if err != nil {
			return nil, err
		}
		if m != nil {
			out.FirstName = m.FirstName
			out.LastName = m.LastName
			out.BirthDate = m.BirthDate
			out.Nickname = m.Nickname
			out.AboutMe = m.AboutMe
			out.AvatarPath = m.AvatarPath
			out.Visibility = string(m.Visibility)
		}
	} else if acc.Kind == account.KindAssociation {
		asso, err := s.associations.FindByAccountID(ctx, acc.ID)
		if err != nil {
			return nil, err
		}
		if asso != nil {
			out.LegalName = asso.LegalName
			out.SIREN = asso.SIREN
			out.RNA = asso.RNA
			out.PostalCode = asso.PostalCode
			out.About = asso.About
			out.LogoPath = asso.LogoPath
			out.Visibility = string(asso.Visibility)
		}
	}

	return out, nil
}
