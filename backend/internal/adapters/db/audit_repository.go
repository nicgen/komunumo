package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"komunumo/backend/internal/adapters/db/sqlc"
	"komunumo/backend/internal/domain/audit"
)

type AuditRepository struct {
	q *sqlc.Queries
}

func NewAuditRepository(conn *sql.DB) *AuditRepository {
	return &AuditRepository{q: sqlc.New(conn)}
}

func (r *AuditRepository) withTx(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *AuditRepository) Append(ctx context.Context, e *audit.Event) error {
	var metadata sql.NullString
	if len(e.Metadata) > 0 {
		b, err := json.Marshal(e.Metadata)
		if err != nil {
			return err
		}
		metadata = sql.NullString{String: string(b), Valid: true}
	}

	nullStr := func(s string) sql.NullString {
		if s == "" {
			return sql.NullString{}
		}
		return sql.NullString{String: s, Valid: true}
	}

	nullPtr := func(s *string) sql.NullString {
		if s == nil {
			return sql.NullString{}
		}
		return sql.NullString{String: *s, Valid: true}
	}

	return r.withTx(ctx).AppendAuditEvent(ctx, sqlc.AppendAuditEventParams{
		ID:         e.ID,
		OccurredAt: encodeTime(e.OccurredAt),
		EventType:  string(e.Type),
		AccountID:  nullPtr(e.AccountID),
		EmailHash:  nullPtr(e.EmailHash),
		Ip:         nullStr(e.IP),
		UserAgent:  nullStr(e.UserAgent),
		Metadata:   metadata,
	})
}
