package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"komunumo/backend/internal/adapters/db/sqlc"
	"komunumo/backend/internal/domain/token"
)

type TokenRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewTokenRepository(conn *sql.DB) *TokenRepository {
	return &TokenRepository{db: conn, q: sqlc.New(conn)}
}

func (r *TokenRepository) withTx(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *TokenRepository) Create(ctx context.Context, t *token.Token) error {
	q := r.withTx(ctx)
	switch t.Kind {
	case token.KindEmailVerification:
		return q.CreateEmailVerification(ctx, sqlc.CreateEmailVerificationParams{
			ID:        t.ID,
			AccountID: t.AccountID,
			TokenHash: t.TokenHash,
			CreatedAt: encodeTime(t.CreatedAt),
			ExpiresAt: encodeTime(t.ExpiresAt),
		})
	case token.KindPasswordReset:
		return q.CreatePasswordReset(ctx, sqlc.CreatePasswordResetParams{
			ID:        t.ID,
			AccountID: t.AccountID,
			TokenHash: t.TokenHash,
			CreatedAt: encodeTime(t.CreatedAt),
			ExpiresAt: encodeTime(t.ExpiresAt),
		})
	default:
		return token.ErrTokenKindMismatch
	}
}

func (r *TokenRepository) FindActiveByHash(ctx context.Context, kind token.Kind, tokenHash string, now time.Time) (*token.Token, error) {
	q := r.withTx(ctx)
	nowStr := encodeTime(now)
	switch kind {
	case token.KindEmailVerification:
		row, err := q.GetActiveEmailVerificationByHash(ctx, sqlc.GetActiveEmailVerificationByHashParams{
			TokenHash: tokenHash,
			NowTime:   nowStr,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return nil, r.tokenInactiveReason(ctx, "email_verifications", tokenHash)
		}
		if err != nil {
			return nil, err
		}
		return decodeEmailVerification(row)
	case token.KindPasswordReset:
		row, err := q.GetActivePasswordResetByHash(ctx, sqlc.GetActivePasswordResetByHashParams{
			TokenHash: tokenHash,
			NowTime:   nowStr,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return nil, r.tokenInactiveReason(ctx, "password_resets", tokenHash)
		}
		if err != nil {
			return nil, err
		}
		return decodePasswordReset(row)
	default:
		return nil, token.ErrTokenKindMismatch
	}
}

func (r *TokenRepository) Consume(ctx context.Context, kind token.Kind, id string, at time.Time) error {
	q := r.withTx(ctx)
	atStr := encodeTime(at)
	switch kind {
	case token.KindEmailVerification:
		return q.ConsumeEmailVerification(ctx, sqlc.ConsumeEmailVerificationParams{
			ConsumedAt: sql.NullString{String: atStr, Valid: true},
			ID:         id,
		})
	case token.KindPasswordReset:
		return q.ConsumePasswordReset(ctx, sqlc.ConsumePasswordResetParams{
			ConsumedAt: sql.NullString{String: atStr, Valid: true},
			ID:         id,
		})
	default:
		return token.ErrTokenKindMismatch
	}
}

func (r *TokenRepository) RevokeActiveForAccount(ctx context.Context, kind token.Kind, accountID string, at time.Time) error {
	q := r.withTx(ctx)
	atStr := encodeTime(at)
	switch kind {
	case token.KindEmailVerification:
		return q.RevokeActiveEmailVerificationsForAccount(ctx, sqlc.RevokeActiveEmailVerificationsForAccountParams{
			ConsumedAt: sql.NullString{String: atStr, Valid: true},
			AccountID:  accountID,
		})
	case token.KindPasswordReset:
		return q.RevokeActivePasswordResetsForAccount(ctx, sqlc.RevokeActivePasswordResetsForAccountParams{
			ConsumedAt: sql.NullString{String: atStr, Valid: true},
			AccountID:  accountID,
		})
	default:
		return token.ErrTokenKindMismatch
	}
}

// tokenInactiveReason returns ErrTokenExpired if the token exists but is expired,
// ErrTokenNotFound if it exists but consumed, or nil if it simply doesn't exist.
func (r *TokenRepository) tokenInactiveReason(ctx context.Context, table, tokenHash string) error {
	var consumedAt sql.NullString
	var expiresAt string
	q := "SELECT expires_at, consumed_at FROM " + table + " WHERE token_hash=?"
	err := r.db.QueryRowContext(ctx, q, tokenHash).Scan(&expiresAt, &consumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // truly not found
	}
	if err != nil {
		return err
	}
	if consumedAt.Valid {
		return token.ErrTokenNotFound
	}
	return token.ErrTokenExpired
}

func decodeEmailVerification(row sqlc.EmailVerification) (*token.Token, error) {
	createdAt, err := decodeTime(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	expiresAt, err := decodeTime(row.ExpiresAt)
	if err != nil {
		return nil, err
	}
	consumedAt, err := decodeNullTime(row.ConsumedAt)
	if err != nil {
		return nil, err
	}
	return &token.Token{
		ID:         row.ID,
		AccountID:  row.AccountID,
		Kind:       token.KindEmailVerification,
		TokenHash:  row.TokenHash,
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
		ConsumedAt: consumedAt,
	}, nil
}

func decodePasswordReset(row sqlc.PasswordReset) (*token.Token, error) {
	createdAt, err := decodeTime(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	expiresAt, err := decodeTime(row.ExpiresAt)
	if err != nil {
		return nil, err
	}
	consumedAt, err := decodeNullTime(row.ConsumedAt)
	if err != nil {
		return nil, err
	}
	return &token.Token{
		ID:         row.ID,
		AccountID:  row.AccountID,
		Kind:       token.KindPasswordReset,
		TokenHash:  row.TokenHash,
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
		ConsumedAt: consumedAt,
	}, nil
}
