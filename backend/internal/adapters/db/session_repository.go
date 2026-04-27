package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"komunumo/backend/internal/adapters/db/sqlc"
	"komunumo/backend/internal/domain/session"
)

type SessionRepository struct {
	q *sqlc.Queries
}

func NewSessionRepository(conn *sql.DB) *SessionRepository {
	return &SessionRepository{q: sqlc.New(conn)}
}

func (r *SessionRepository) withTx(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
	return r.withTx(ctx).CreateSession(ctx, sqlc.CreateSessionParams{
		ID:         s.ID,
		AccountID:  s.AccountID,
		CreatedAt:  encodeTime(s.CreatedAt),
		ExpiresAt:  encodeTime(s.ExpiresAt),
		LastSeenAt: encodeTime(s.LastSeenAt),
		Ip:         sql.NullString{String: s.IP, Valid: s.IP != ""},
		UserAgent:  sql.NullString{String: s.UserAgent, Valid: s.UserAgent != ""},
	})
}

func (r *SessionRepository) FindByID(ctx context.Context, id string, now time.Time) (*session.Session, error) {
	row, err := r.withTx(ctx).GetSessionByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, session.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	s, err := decodeSession(row)
	if err != nil {
		return nil, err
	}
	if s.Expired(now) {
		return nil, session.ErrSessionExpired
	}
	return s, nil
}

func (r *SessionRepository) TouchLastSeen(ctx context.Context, id string, at time.Time) error {
	return r.withTx(ctx).TouchSessionLastSeen(ctx, sqlc.TouchSessionLastSeenParams{
		LastSeenAt: encodeTime(at),
		ID:         id,
	})
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	return r.withTx(ctx).DeleteSession(ctx, id)
}

func (r *SessionRepository) DeleteAllForAccount(ctx context.Context, accountID string) error {
	return r.withTx(ctx).DeleteAllSessionsForAccount(ctx, accountID)
}

func (r *SessionRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	return r.withTx(ctx).DeleteExpiredSessions(ctx, encodeTime(now))
}

func decodeSession(row sqlc.Session) (*session.Session, error) {
	createdAt, err := decodeTime(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	expiresAt, err := decodeTime(row.ExpiresAt)
	if err != nil {
		return nil, err
	}
	lastSeenAt, err := decodeTime(row.LastSeenAt)
	if err != nil {
		return nil, err
	}
	return &session.Session{
		ID:         row.ID,
		AccountID:  row.AccountID,
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
		LastSeenAt: lastSeenAt,
		IP:         row.Ip.String,
		UserAgent:  row.UserAgent.String,
	}, nil
}
