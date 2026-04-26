package ports

import (
	"context"
	"time"

	"komunumo/backend/internal/domain/token"
)

type TokenRepository interface {
	Create(ctx context.Context, t *token.Token) error
	FindActiveByHash(ctx context.Context, kind token.Kind, tokenHash string, now time.Time) (*token.Token, error)
	Consume(ctx context.Context, kind token.Kind, id string, at time.Time) error
	RevokeActiveForAccount(ctx context.Context, kind token.Kind, accountID string, at time.Time) error
}
