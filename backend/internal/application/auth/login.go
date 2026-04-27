package auth

import (
	"context"
	"errors"
	"time"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports"
)

const SessionTTL = 30 * 24 * time.Hour

var ErrInvalidCredentials = errors.New("auth: invalid credentials")

type LoginInput struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
}

type LoginOutput struct {
	SessionID string
}

type LoginService struct {
	accounts  ports.AccountRepository
	sessions  ports.SessionRepository
	audit     ports.AuditRepository
	hasher    ports.PasswordHasher
	tokenGen  ports.TokenGenerator
	clock     ports.Clock
	rl        ports.RateLimiter
	uow       ports.UnitOfWork
}

func NewLoginService(
	accounts ports.AccountRepository,
	sessions ports.SessionRepository,
	audit ports.AuditRepository,
	hasher ports.PasswordHasher,
	tokenGen ports.TokenGenerator,
	clock ports.Clock,
	rl ports.RateLimiter,
	uow ports.UnitOfWork,
) *LoginService {
	return &LoginService{
		accounts: accounts,
		sessions: sessions,
		audit:    audit,
		hasher:   hasher,
		tokenGen: tokenGen,
		clock:    clock,
		rl:       rl,
		uow:      uow,
	}
}

func (s *LoginService) Login(ctx context.Context, in LoginInput) (LoginOutput, error) {
	if ok, _ := s.rl.Allow(ctx, "login:"+in.IP); !ok {
		return LoginOutput{}, ErrRateLimited
	}

	canonical, err := account.CanonicalizeEmail(in.Email)
	if err != nil {
		return LoginOutput{}, ErrInvalidCredentials
	}

	now := s.clock.Now()

	acc, err := s.accounts.FindByEmailCanonical(ctx, canonical)
	if err != nil {
		return LoginOutput{}, err
	}
	if acc == nil {
		_ = s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAuthLoginFailed,
			IP:         in.IP,
			UserAgent:  in.UserAgent,
		})
		return LoginOutput{}, ErrInvalidCredentials
	}

	if acc.Status == account.StatusDisabled {
		return LoginOutput{}, account.ErrAccountDisabled
	}

	if acc.Status == account.StatusPendingVerification {
		return LoginOutput{}, account.ErrAccountNotVerified
	}

	ok, err := s.hasher.Verify(acc.PasswordHash, in.Password)
	if err != nil {
		return LoginOutput{}, err
	}
	if !ok {
		_ = s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAuthLoginFailed,
			AccountID:  &acc.ID,
			IP:         in.IP,
			UserAgent:  in.UserAgent,
		})
		return LoginOutput{}, ErrInvalidCredentials
	}

	sess := &session.Session{
		ID:         s.tokenGen.NewID(),
		AccountID:  acc.ID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(SessionTTL),
		LastSeenAt: now,
		IP:         in.IP,
		UserAgent:  in.UserAgent,
	}

	if err := s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.sessions.Create(ctx, sess); err != nil {
			return err
		}
		if err := s.accounts.TouchLastLogin(ctx, acc.ID, now); err != nil {
			return err
		}
		return s.audit.Append(ctx, &audit.Event{
			ID:         s.tokenGen.NewID(),
			OccurredAt: now,
			Type:       audit.EventAuthLoginSuccess,
			AccountID:  &acc.ID,
			IP:         in.IP,
			UserAgent:  in.UserAgent,
		})
	}); err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{SessionID: sess.ID}, nil
}
