package fakes

import (
	"context"
	"time"

	"komunumo/backend/internal/domain/token"
	"komunumo/backend/internal/ports"
)

var _ ports.TokenRepository = (*TokenRepository)(nil)

type TokenRepository struct {
	byID   map[string]*token.Token
	byHash map[string]*token.Token
}

func NewTokenRepository() *TokenRepository {
	return &TokenRepository{
		byID:   make(map[string]*token.Token),
		byHash: make(map[string]*token.Token),
	}
}

func (r *TokenRepository) Create(_ context.Context, t *token.Token) error {
	cp := *t
	r.byID[t.ID] = &cp
	r.byHash[t.TokenHash] = &cp
	return nil
}

func (r *TokenRepository) FindActiveByHash(_ context.Context, kind token.Kind, tokenHash string, now time.Time) (*token.Token, error) {
	t, ok := r.byHash[tokenHash]
	if !ok || t.Kind != kind {
		return nil, nil
	}
	if t.ConsumedAt != nil {
		return nil, token.ErrTokenNotFound
	}
	if !t.ExpiresAt.After(now) {
		return nil, token.ErrTokenExpired
	}
	cp := *t
	return &cp, nil
}

func (r *TokenRepository) Consume(_ context.Context, _ token.Kind, id string, at time.Time) error {
	t, ok := r.byID[id]
	if !ok {
		return token.ErrTokenNotFound
	}
	if err := t.Consume(at); err != nil {
		return err
	}
	r.byHash[t.TokenHash] = t
	return nil
}

func (r *TokenRepository) RevokeActiveForAccount(_ context.Context, kind token.Kind, accountID string, at time.Time) error {
	for _, t := range r.byID {
		if t.AccountID == accountID && t.Kind == kind && t.ConsumedAt == nil {
			_ = t.Consume(at)
			r.byHash[t.TokenHash] = t
		}
	}
	return nil
}

func (r *TokenRepository) CountActive(accountID string, kind token.Kind, now time.Time) int {
	n := 0
	for _, t := range r.byID {
		if t.AccountID == accountID && t.Kind == kind && t.Active(now) {
			n++
		}
	}
	return n
}
