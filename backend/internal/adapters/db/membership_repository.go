package db

import (
	"context"
	"database/sql"
	"errors"

	"komunumo/backend/internal/adapters/db/sqlc"
	"komunumo/backend/internal/ports"
)

type MembershipRepository struct {
	q *sqlc.Queries
}

func NewMembershipRepository(conn *sql.DB) *MembershipRepository {
	return &MembershipRepository{q: sqlc.New(conn)}
}

func (r *MembershipRepository) withTx(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *MembershipRepository) Create(ctx context.Context, m *ports.Membership) error {
	return r.withTx(ctx).CreateMembership(ctx, sqlc.CreateMembershipParams{
		ID:                   m.ID,
		MemberAccountID:      m.MemberAccountID,
		AssociationAccountID: m.AssociationAccountID,
		Role:                 m.Role,
		Status:               m.Status,
		JoinedAt:             m.JoinedAt,
	})
}

func (r *MembershipRepository) FindByAccountIDs(ctx context.Context, memberID, associationID string) (*ports.Membership, error) {
	row, err := r.withTx(ctx).GetMembershipByAccountIDs(ctx, sqlc.GetMembershipByAccountIDsParams{
		MemberAccountID:      memberID,
		AssociationAccountID: associationID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ports.Membership{
		ID:                   row.ID,
		MemberAccountID:      row.MemberAccountID,
		AssociationAccountID: row.AssociationAccountID,
		Role:                 row.Role,
		Status:               row.Status,
		JoinedAt:             row.JoinedAt,
	}, nil
}
