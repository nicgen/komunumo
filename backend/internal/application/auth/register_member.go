package auth

import (
	"context"
	"fmt"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports"
)

type RegisterMemberInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
}

type RegisterMemberService struct {
	accounts ports.AccountRepository
	members  ports.MemberRepository
	audit    ports.AuditRepository
	email    ports.EmailSender
	hasher   ports.PasswordHasher
	tokenGen ports.TokenGenerator
	tokens   ports.TokenRepository
	clock    ports.Clock
	rl       ports.RateLimiter
	uow      ports.UnitOfWork
}

func NewRegisterMemberService(
	accounts ports.AccountRepository,
	members ports.MemberRepository,
	audit ports.AuditRepository,
	email ports.EmailSender,
	hasher ports.PasswordHasher,
	tokenGen ports.TokenGenerator,
	tokens ports.TokenRepository,
	clock ports.Clock,
	rl ports.RateLimiter,
	uow ports.UnitOfWork,
) *RegisterMemberService {
	return &RegisterMemberService{
		accounts: accounts,
		members:  members,
		audit:    audit,
		email:    email,
		hasher:   hasher,
		tokenGen: tokenGen,
		tokens:   tokens,
		clock:    clock,
		rl:       rl,
		uow:      uow,
	}
}

func (s *RegisterMemberService) RegisterMember(ctx context.Context, ip string, in RegisterMemberInput) error {
	// Rate limiting
	rlKey := "register:ip:" + ip
	if allowed, _ := s.rl.Allow(ctx, rlKey); !allowed {
		return ErrRateLimited
	}

	now := s.clock.Now()

	// 1. Domain validation (Member invariants)
	m, err := member.New("placeholder", in.FirstName, in.LastName, in.BirthDate, now)
	if err != nil {
		return err
	}

	// 2. Account checks
	canonical, err := account.CanonicalizeEmail(in.Email)
	if err != nil {
		return err
	}

	existing, err := s.accounts.FindByEmailCanonical(ctx, canonical)
	if err != nil {
		return err
	}
	if existing != nil {
		return account.ErrEmailTaken
	}

	// 3. Password validation and hashing
	if err := account.ValidatePassword(in.Password); err != nil {
		return err
	}
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return err
	}

	// 4. Create Account
	accountID := s.tokenGen.NewID()
	acc, err := account.New(accountID, in.Email, now)
	if err != nil {
		return err
	}
	acc.PasswordHash = hash
	acc.Kind = account.KindMember

	m.AccountID = accountID

	// 5. Transactional persistence
	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.accounts.Create(ctx, acc); err != nil {
			return err
		}
		if err := s.members.Create(ctx, m); err != nil {
			return err
		}

		// 6. Token generation
		rawToken, err := s.tokenGen.NewRawToken()
		if err != nil {
			return err
		}
		tokenHash := s.tokenGen.HashToken(rawToken)
		tok := token.New(s.tokenGen.NewID(), accountID, token.KindEmailVerification, tokenHash, now, token.EmailVerificationTTL)
		if err := s.tokens.Create(ctx, tok); err != nil {
			return err
		}

		// 7. Audit log
		if err := s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAccountCreated,
			AccountID:  &accountID,
			IP:         ip,
			Metadata:   map[string]any{"kind": string(account.KindMember)},
		}); err != nil {
			return err
		}

		// 8. Send Email
		displayName := fmt.Sprintf("%s %s", in.FirstName, in.LastName)
		return s.email.SendVerification(ctx, acc.Email, displayName, rawToken)
	})
}
