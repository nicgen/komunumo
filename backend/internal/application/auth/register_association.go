package auth

import (
	"context"
	"fmt"
	"time"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/association"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports"
)

type RegisterAssociationInput struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	LegalName  string `json:"legal_name"`
	PostalCode string `json:"postal_code"`
	SIREN      string `json:"siren,omitempty"`
	RNA        string `json:"rna,omitempty"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
}

type RegisterAssociationService struct {
	accounts     ports.AccountRepository
	associations ports.AssociationRepository
	members      ports.MemberRepository
	memberships  ports.MembershipRepository
	audit        ports.AuditRepository
	emails       ports.EmailSender
	hasher       ports.PasswordHasher
	tokenGen     ports.TokenGenerator
	tokens       ports.TokenRepository
	clock        ports.Clock
	rl           ports.RateLimiter
	uow          ports.UnitOfWork
}

func NewRegisterAssociationService(
	accounts ports.AccountRepository,
	associations ports.AssociationRepository,
	members ports.MemberRepository,
	memberships ports.MembershipRepository,
	audit ports.AuditRepository,
	emails ports.EmailSender,
	hasher ports.PasswordHasher,
	tokenGen ports.TokenGenerator,
	tokens ports.TokenRepository,
	clock ports.Clock,
	rl ports.RateLimiter,
	uow ports.UnitOfWork,
) *RegisterAssociationService {
	return &RegisterAssociationService{
		accounts:     accounts,
		associations: associations,
		members:      members,
		memberships:  memberships,
		audit:        audit,
		emails:       emails,
		hasher:       hasher,
		tokenGen:     tokenGen,
		tokens:       tokens,
		clock:        clock,
		rl:           rl,
		uow:          uow,
	}
}

func (s *RegisterAssociationService) RegisterAssociation(ctx context.Context, ip string, in RegisterAssociationInput) error {
	now := s.clock.Now()

	// Rate limiting
	rlKey := "register:ip:" + ip
	if allowed, _ := s.rl.Allow(ctx, rlKey); !allowed {
		return ErrRateLimited
	}

	// Validate password
	if err := account.ValidatePassword(in.Password); err != nil {
		return err
	}

	// Email canonicalization
	canonical, err := account.CanonicalizeEmail(in.Email)
	if err != nil {
		return err
	}

	// Check if email already taken
	existing, err := s.accounts.FindByEmailCanonical(ctx, canonical)
	if err != nil {
		return err
	}
	if existing != nil {
		return account.ErrEmailTaken
	}

	// Validate domain entities
	// Association
	if err := association.ValidateSIREN(in.SIREN); err != nil {
		return err
	}
	if err := association.ValidateRNA(in.RNA); err != nil {
		return err
	}
	asso, err := association.New(s.tokenGen.NewID(), in.LegalName, in.PostalCode, now)
	if err != nil {
		return err
	}
	asso.SIREN = in.SIREN
	asso.RNA = in.RNA

	// Member (representative)
	m, err := member.New(asso.AccountID, in.FirstName, in.LastName, in.BirthDate, now)
	if err != nil {
		return err
	}

	// Hash password
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return err
	}

	// Create account
	accountID := s.tokenGen.NewID()
	acc, err := account.New(accountID, in.Email, now)
	if err != nil {
		return err
	}
	acc.PasswordHash = hash
	acc.Kind = account.KindAssociation

	asso.AccountID = accountID
	m.AccountID = accountID

	// Transaction
	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.accounts.Create(ctx, acc); err != nil {
			return err
		}

		if err := s.associations.Create(ctx, asso); err != nil {
			return err
		}

		if err := s.members.Create(ctx, m); err != nil {
			return err
		}

		// Create membership (owner)
		membership := &ports.Membership{
			ID:                   s.tokenGen.NewID(),
			MemberAccountID:      acc.ID,
			AssociationAccountID: acc.ID,
			Role:                 "owner",
			Status:               "active",
			JoinedAt:             now.Format(time.RFC3339),
		}
		if err := s.memberships.Create(ctx, membership); err != nil {
			return err
		}

		// Create email verification token
		rawToken, err := s.tokenGen.NewRawToken()
		if err != nil {
			return err
		}
		tokenHash := s.tokenGen.HashToken(rawToken)
		tok := token.New(s.tokenGen.NewID(), acc.ID, token.KindEmailVerification, tokenHash, now, token.EmailVerificationTTL)
		if err := s.tokens.Create(ctx, tok); err != nil {
			return err
		}

		// Audit log
		if err := s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAccountCreated,
			AccountID:  &acc.ID,
			IP:         ip,
			Metadata:   map[string]any{"kind": string(account.KindAssociation), "legal_name": in.LegalName},
		}); err != nil {
			return err
		}

		// Send verification email
		displayName := fmt.Sprintf("%s %s", in.FirstName, in.LastName)
		return s.emails.SendVerification(ctx, acc.Email, displayName, rawToken)
	})
}
