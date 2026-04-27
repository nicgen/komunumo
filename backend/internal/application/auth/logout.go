package auth

import (
	"context"

	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports"
)

type LogoutService struct {
	sessions ports.SessionRepository
	audit    ports.AuditRepository
	tokenGen ports.TokenGenerator
	clock    ports.Clock
}

func NewLogoutService(
	sessions ports.SessionRepository,
	audit ports.AuditRepository,
	tokenGen ports.TokenGenerator,
	clock ports.Clock,
) *LogoutService {
	return &LogoutService{sessions: sessions, audit: audit, tokenGen: tokenGen, clock: clock}
}

func (s *LogoutService) Logout(ctx context.Context, sessionID string) error {
	now := s.clock.Now()

	sess, err := s.sessions.FindByID(ctx, sessionID, now)
	if err != nil {
		if err == session.ErrSessionNotFound || err == session.ErrSessionExpired {
			return nil
		}
		return err
	}

	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return err
	}

	return s.audit.Append(ctx, &audit.Event{
		ID:         s.tokenGen.NewID(),
		OccurredAt: now,
		Type:       audit.EventAuthLogout,
		AccountID:  &sess.AccountID,
	})
}
