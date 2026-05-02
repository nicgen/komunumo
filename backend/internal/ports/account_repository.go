package ports

import (
	"context"
	"time"

	"komunumo/backend/internal/domain/account"
)

type AccountRepository interface {
	Create(ctx context.Context, a *account.Account) error
	FindByEmailCanonical(ctx context.Context, emailCanonical string) (*account.Account, error)
	FindByID(ctx context.Context, id string) (*account.Account, error)
	UpdateStatus(ctx context.Context, id string, status account.Status, at time.Time) error
	UpdateKindAndStatus(ctx context.Context, id string, kind account.Kind, status account.Status, at time.Time) error
	UpdatePasswordHash(ctx context.Context, id, hash string, at time.Time) error
	TouchLastLogin(ctx context.Context, id string, at time.Time) error
}
