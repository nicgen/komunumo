package profile

import (
	"context"
	"fmt"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/association"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/ports"
)

type UpdateProfileInput struct {
	// Member specific
	Nickname *string `json:"nickname,omitempty"`
	AboutMe  *string `json:"about_me,omitempty"`

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
	tokenGen     ports.TokenGenerator
}

func NewUpdateProfileService(
	accounts ports.AccountRepository,
	members ports.MemberRepository,
	associations ports.AssociationRepository,
	sessions ports.SessionRepository,
	audit ports.AuditRepository,
	clock ports.Clock,
	tokenGen ports.TokenGenerator,
) *UpdateProfileService {
	return &UpdateProfileService{
		accounts:     accounts,
		members:      members,
		associations: associations,
		sessions:     sessions,
		audit:        audit,
		clock:        clock,
		tokenGen:     tokenGen,
	}
}

func (s *UpdateProfileService) UpdateProfile(ctx context.Context, sessionID string, ip string, in UpdateProfileInput) error {
	now := s.clock.Now()

	sess, err := s.sessions.FindByID(ctx, sessionID, now)
	if err != nil {
		return err
	}

	acc, err := s.accounts.FindByID(ctx, sess.AccountID)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("account not found")
	}

	if acc.Kind == account.KindMember {
		m, err := s.members.FindByAccountID(ctx, acc.ID)
		if err != nil {
			return err
		}
		if m == nil {
			return fmt.Errorf("member profile not found")
		}

		if in.Nickname != nil {
			m.Nickname = *in.Nickname
		}
		if in.AboutMe != nil {
			m.AboutMe = *in.AboutMe
		}
		if in.Visibility != nil {
			m.Visibility = member.Visibility(*in.Visibility)
		}

		if err := m.Validate(); err != nil {
			return err
		}

		if err := s.members.Update(ctx, m); err != nil {
			return err
		}
	} else if acc.Kind == account.KindAssociation {
		asso, err := s.associations.FindByAccountID(ctx, acc.ID)
		if err != nil {
			return err
		}
		if asso == nil {
			return fmt.Errorf("association profile not found")
		}

		if in.About != nil {
			asso.About = *in.About
		}
		if in.PostalCode != nil {
			asso.PostalCode = *in.PostalCode
		}
		if in.Visibility != nil {
			asso.Visibility = association.Visibility(*in.Visibility)
		}

		if err := asso.Validate(); err != nil {
			return err
		}

		if err := s.associations.Update(ctx, asso); err != nil {
			return err
		}
	}

	// Audit log
	return s.audit.Append(ctx, &audit.Event{
		ID:         s.tokenGen.NewID(),
		OccurredAt: now,
		Type:       audit.EventProfileUpdated,
		AccountID:  &acc.ID,
		IP:         ip,
		Metadata:   map[string]any{"kind": string(acc.Kind)},
	})
}
