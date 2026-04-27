package auth

import (
	"context"
	"errors"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports"
)

type ResendVerificationInput struct {
	Email string
	IP    string
}

// ErrRateLimited is returned when the rate limiter blocks the request.
var ErrRateLimited = errors.New("auth: rate limited")

type ResendVerificationService struct {
	accounts ports.AccountRepository
	tokens   ports.TokenRepository
	audit    ports.AuditRepository
	email    ports.EmailSender
	tokenGen ports.TokenGenerator
	clock    ports.Clock
	rl       ports.RateLimiter
	uow      ports.UnitOfWork
}

func NewResendVerificationService(
	accounts ports.AccountRepository,
	tokens ports.TokenRepository,
	auditRepo ports.AuditRepository,
	email ports.EmailSender,
	tokenGen ports.TokenGenerator,
	clock ports.Clock,
	rl ports.RateLimiter,
	uow ports.UnitOfWork,
) *ResendVerificationService {
	return &ResendVerificationService{
		accounts: accounts,
		tokens:   tokens,
		audit:    auditRepo,
		email:    email,
		tokenGen: tokenGen,
		clock:    clock,
		rl:       rl,
		uow:      uow,
	}
}

func (s *ResendVerificationService) Resend(ctx context.Context, in ResendVerificationInput) error {
	canonical, err := account.CanonicalizeEmail(in.Email)
	if err != nil {
		return nil // anti-enumeration: invalid email → silent no-op
	}

	rlKey := "resend_verification:email:" + canonical
	if allowed, _ := s.rl.Allow(ctx, rlKey); !allowed {
		return ErrRateLimited
	}

	now := s.clock.Now()

	acc, err := s.accounts.FindByEmailCanonical(ctx, canonical)
	if err != nil {
		return err
	}
	if acc == nil {
		return nil // anti-enumeration: unknown email → silent no-op
	}

	var rawToken string
	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.tokens.RevokeActiveForAccount(ctx, token.KindEmailVerification, acc.ID, now); err != nil {
			return err
		}

		raw, err := s.tokenGen.NewRawToken()
		if err != nil {
			return err
		}
		rawToken = raw
		tokenHash := s.tokenGen.HashToken(raw)
		tok := token.New(s.tokenGen.NewID(), acc.ID, token.KindEmailVerification, tokenHash, now, token.EmailVerificationTTL)
		if err := s.tokens.Create(ctx, tok); err != nil {
			return err
		}

		return s.email.SendVerification(ctx, acc.Email, acc.FirstName, rawToken)
	})
}
