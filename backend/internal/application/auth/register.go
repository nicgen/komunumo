package auth

import (
	"context"
	"errors"
	"time"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports"
)

type RegisterInput struct {
	Email       string
	FirstName   string
	LastName    string
	DateOfBirth time.Time
	Password    string
}

type RegisterService struct {
	accounts ports.AccountRepository
	tokens   ports.TokenRepository
	audit    ports.AuditRepository
	email    ports.EmailSender
	hasher   ports.PasswordHasher
	tokenGen ports.TokenGenerator
	clock    ports.Clock
	rl       ports.RateLimiter
	uow      ports.UnitOfWork
}

func NewRegisterService(
	accounts ports.AccountRepository,
	tokens ports.TokenRepository,
	auditRepo ports.AuditRepository,
	email ports.EmailSender,
	hasher ports.PasswordHasher,
	tokenGen ports.TokenGenerator,
	clock ports.Clock,
	rl ports.RateLimiter,
	uow ports.UnitOfWork,
) *RegisterService {
	return &RegisterService{
		accounts: accounts,
		tokens:   tokens,
		audit:    auditRepo,
		email:    email,
		hasher:   hasher,
		tokenGen: tokenGen,
		clock:    clock,
		rl:       rl,
		uow:      uow,
	}
}

// Register creates a new account. The email is sent before DB writes so that
// a Brevo failure leaves no orphan account (spec: "pas de compte sans email envoyé").
func (s *RegisterService) Register(ctx context.Context, ip string, in RegisterInput) error {
	if err := account.ValidatePassword(in.Password); err != nil {
		return err
	}

	now := s.clock.Now()

	acc, err := account.New(s.tokenGen.NewID(), in.Email, in.FirstName, in.LastName, in.DateOfBirth, now)
	if err != nil {
		return err
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return err
	}
	acc.PasswordHash = hash

	// Check for duplicate email before generating token / sending email.
	existing, err := s.accounts.FindByEmailCanonical(ctx, acc.EmailCanonical)
	if err != nil {
		return err
	}
	if existing != nil {
		displayName := existing.FirstName
		return s.email.SendAccountAlreadyExists(ctx, in.Email, displayName)
	}

	raw, err := s.tokenGen.NewRawToken()
	if err != nil {
		return err
	}
	tokenHash := s.tokenGen.HashToken(raw)
	tok := token.New(s.tokenGen.NewID(), acc.ID, token.KindEmailVerification, tokenHash, now, token.EmailVerificationTTL)

	// Send email first. If it fails, nothing is written to the DB.
	if err := s.email.SendVerification(ctx, in.Email, in.FirstName, raw); err != nil {
		return err
	}

	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.accounts.Create(ctx, acc); err != nil {
			// Race condition: another request created the same email between our check and insert.
			if errors.Is(err, account.ErrEmailTaken) {
				return nil
			}
			return err
		}
		if err := s.tokens.Create(ctx, tok); err != nil {
			return err
		}
		return s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAccountCreated,
			AccountID:  &acc.ID,
			IP:         ip,
		})
	})
}
