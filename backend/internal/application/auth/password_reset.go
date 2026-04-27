package auth

import (
	"context"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports"
)

// --- Request ---

type PasswordResetRequestInput struct {
	Email string
	IP    string
}

type PasswordResetRequestService struct {
	accounts ports.AccountRepository
	tokens   ports.TokenRepository
	audit    ports.AuditRepository
	email    ports.EmailSender
	tokenGen ports.TokenGenerator
	clock    ports.Clock
	rl       ports.RateLimiter
	uow      ports.UnitOfWork
}

func NewPasswordResetRequestService(
	accounts ports.AccountRepository,
	tokens ports.TokenRepository,
	audit ports.AuditRepository,
	email ports.EmailSender,
	tokenGen ports.TokenGenerator,
	clock ports.Clock,
	rl ports.RateLimiter,
	uow ports.UnitOfWork,
) *PasswordResetRequestService {
	return &PasswordResetRequestService{
		accounts: accounts, tokens: tokens, audit: audit, email: email,
		tokenGen: tokenGen, clock: clock, rl: rl, uow: uow,
	}
}

// Request always returns nil (no account existence leak). Side effects only
// occur when the account exists and is not rate-limited.
func (s *PasswordResetRequestService) Request(ctx context.Context, in PasswordResetRequestInput) error {
	if ok, _ := s.rl.Allow(ctx, "pwreset:"+in.IP); !ok {
		return ErrRateLimited
	}

	canonical, err := account.CanonicalizeEmail(in.Email)
	if err != nil {
		return nil
	}

	acc, err := s.accounts.FindByEmailCanonical(ctx, canonical)
	if err != nil {
		return err
	}
	if acc == nil {
		return nil
	}

	now := s.clock.Now()

	raw, err := s.tokenGen.NewRawToken()
	if err != nil {
		return err
	}
	tokenHash := s.tokenGen.HashToken(raw)
	tok := token.New(s.tokenGen.NewID(), acc.ID, token.KindPasswordReset, tokenHash, now, token.PasswordResetTTL)

	if err := s.email.SendPasswordReset(ctx, acc.Email, acc.FirstName, raw); err != nil {
		return err
	}

	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.tokens.RevokeActiveForAccount(ctx, token.KindPasswordReset, acc.ID, now); err != nil {
			return err
		}
		if err := s.tokens.Create(ctx, tok); err != nil {
			return err
		}
		return s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAuthPasswordResetReq,
			AccountID:  &acc.ID,
			IP:         in.IP,
		})
	})
}

// --- Confirm ---

type PasswordResetConfirmInput struct {
	RawToken    string
	NewPassword string
}

type PasswordResetConfirmService struct {
	accounts ports.AccountRepository
	tokens   ports.TokenRepository
	sessions ports.SessionRepository
	audit    ports.AuditRepository
	email    ports.EmailSender
	hasher   ports.PasswordHasher
	tokenGen ports.TokenGenerator
	clock    ports.Clock
	uow      ports.UnitOfWork
}

func NewPasswordResetConfirmService(
	accounts ports.AccountRepository,
	tokens ports.TokenRepository,
	sessions ports.SessionRepository,
	audit ports.AuditRepository,
	email ports.EmailSender,
	hasher ports.PasswordHasher,
	tokenGen ports.TokenGenerator,
	clock ports.Clock,
	uow ports.UnitOfWork,
) *PasswordResetConfirmService {
	return &PasswordResetConfirmService{
		accounts: accounts, tokens: tokens, sessions: sessions, audit: audit,
		email: email, hasher: hasher, tokenGen: tokenGen, clock: clock, uow: uow,
	}
}

func (s *PasswordResetConfirmService) Confirm(ctx context.Context, in PasswordResetConfirmInput) error {
	if err := account.ValidatePassword(in.NewPassword); err != nil {
		return err
	}

	now := s.clock.Now()
	tokenHash := s.tokenGen.HashToken(in.RawToken)

	tok, err := s.tokens.FindActiveByHash(ctx, token.KindPasswordReset, tokenHash, now)
	if err != nil {
		return err
	}
	if tok == nil {
		return token.ErrTokenNotFound
	}

	newHash, err := s.hasher.Hash(in.NewPassword)
	if err != nil {
		return err
	}

	acc, err := s.accounts.FindByID(ctx, tok.AccountID)
	if err != nil {
		return err
	}
	if acc == nil {
		return account.ErrAccountNotFound
	}

	if err := s.email.SendPasswordChanged(ctx, acc.Email, acc.FirstName); err != nil {
		return err
	}

	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.accounts.UpdatePasswordHash(ctx, acc.ID, newHash, now); err != nil {
			return err
		}
		if err := s.sessions.DeleteAllForAccount(ctx, acc.ID); err != nil {
			return err
		}
		if err := s.tokens.Consume(ctx, token.KindPasswordReset, tok.ID, now); err != nil {
			return err
		}
		return s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAuthPasswordChanged,
			AccountID:  &acc.ID,
		})
	})
}
