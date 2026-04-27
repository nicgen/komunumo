package ports

import (
	"context"
	"time"

	"komunumo/backend/internal/domain/session"
)

type SessionRepository interface {
	Create(ctx context.Context, s *session.Session) error
	FindByID(ctx context.Context, id string, now time.Time) (*session.Session, error)
	TouchLastSeen(ctx context.Context, id string, at time.Time) error
	Delete(ctx context.Context, id string) error
	DeleteAllForAccount(ctx context.Context, accountID string) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}
