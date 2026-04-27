package auth

import (
	"context"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports"
)

type MeOutput struct {
	AccountID string
	Email     string
	FirstName string
	LastName  string
	Status    account.Status
}

type MeService struct {
	sessions ports.SessionRepository
	accounts ports.AccountRepository
	clock    ports.Clock
}

func NewMeService(
	sessions ports.SessionRepository,
	accounts ports.AccountRepository,
	clock ports.Clock,
) *MeService {
	return &MeService{sessions: sessions, accounts: accounts, clock: clock}
}

func (s *MeService) Me(ctx context.Context, sessionID string) (MeOutput, error) {
	now := s.clock.Now()

	sess, err := s.sessions.FindByID(ctx, sessionID, now)
	if err != nil {
		if err == session.ErrSessionNotFound || err == session.ErrSessionExpired {
			return MeOutput{}, session.ErrSessionNotFound
		}
		return MeOutput{}, err
	}

	acc, err := s.accounts.FindByID(ctx, sess.AccountID)
	if err != nil {
		return MeOutput{}, err
	}
	if acc == nil {
		return MeOutput{}, account.ErrAccountNotFound
	}

	return MeOutput{
		AccountID: acc.ID,
		Email:     acc.Email,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		Status:    acc.Status,
	}, nil
}
