package ports

import (
	"context"
	"time"

	"komunumo/backend/internal/domain/token"
)

type TokenRepository interface {
	Create(ctx context.Context, t *token.Token) error
	// FindActiveByHash returns the token if active (not expired, not consumed).
	// Returns (nil, ErrTokenExpired) if the token exists but is expired or consumed.
	// Returns (nil, nil) if the token does not exist at all.
	FindActiveByHash(ctx context.Context, kind token.Kind, tokenHash string, now time.Time) (*token.Token, error)
	Consume(ctx context.Context, kind token.Kind, id string, at time.Time) error
	RevokeActiveForAccount(ctx context.Context, kind token.Kind, accountID string, at time.Time) error
}
