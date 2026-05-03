package auth

import (
	"context"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports"
)

type VerifyEmailInput struct {
	RawToken string
}

type VerifyEmailService struct {
	accounts ports.AccountRepository
	tokens   ports.TokenRepository
	audit    ports.AuditRepository
	tokenGen ports.TokenGenerator
	clock    ports.Clock
	uow      ports.UnitOfWork
}

func NewVerifyEmailService(
	accounts ports.AccountRepository,
	tokens ports.TokenRepository,
	auditRepo ports.AuditRepository,
	tokenGen ports.TokenGenerator,
	clock ports.Clock,
	uow ports.UnitOfWork,
) *VerifyEmailService {
	return &VerifyEmailService{
		accounts: accounts,
		tokens:   tokens,
		audit:    auditRepo,
		tokenGen: tokenGen,
		clock:    clock,
		uow:      uow,
	}
}

func (s *VerifyEmailService) VerifyEmail(ctx context.Context, in VerifyEmailInput) error {
	now := s.clock.Now()
	tokenHash := s.tokenGen.HashToken(in.RawToken)

	tok, err := s.tokens.FindActiveByHash(ctx, token.KindEmailVerification, tokenHash, now)
	if err != nil {
		return err
	}
	if tok == nil {
		// Could be expired or unknown; check if it exists but expired
		return token.ErrTokenNotFound
	}

	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.tokens.Consume(ctx, token.KindEmailVerification, tok.ID, now); err != nil {
			return err
		}

		if err := s.accounts.UpdateStatus(ctx, tok.AccountID, account.StatusActive, now); err != nil {
			return err
		}

		return s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAccountEmailVerified,
			AccountID:  &tok.AccountID,
		})
	})
}
