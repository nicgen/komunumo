package db

import (
	"context"
	"database/sql"
	"errors"

	"komunumo/backend/internal/adapters/db/sqlc"
	"komunumo/backend/internal/domain/association"
)

type AssociationRepository struct {
	q *sqlc.Queries
}

func NewAssociationRepository(conn *sql.DB) *AssociationRepository {
	return &AssociationRepository{q: sqlc.New(conn)}
}

func (r *AssociationRepository) withTx(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *AssociationRepository) Create(ctx context.Context, a *association.Association) error {
	return r.withTx(ctx).CreateAssociation(ctx, sqlc.CreateAssociationParams{
		AccountID:  a.AccountID,
		LegalName:  a.LegalName,
		Siren:      sql.NullString{String: a.SIREN, Valid: a.SIREN != ""},
		Rna:        sql.NullString{String: a.RNA, Valid: a.RNA != ""},
		PostalCode: a.PostalCode,
		About:      sql.NullString{String: a.About, Valid: a.About != ""},
		LogoPath:   sql.NullString{String: a.LogoPath, Valid: a.LogoPath != ""},
		Visibility: string(a.Visibility),
	})
}

func (r *AssociationRepository) FindByAccountID(ctx context.Context, accountID string) (*association.Association, error) {
	row, err := r.withTx(ctx).GetAssociationByAccountID(ctx, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return decodeAssociation(row), nil
}

func (r *AssociationRepository) Update(ctx context.Context, a *association.Association) error {
	return r.withTx(ctx).UpdateAssociation(ctx, sqlc.UpdateAssociationParams{
		AccountID:  a.AccountID,
		About:      sql.NullString{String: a.About, Valid: a.About != ""},
		LogoPath:   sql.NullString{String: a.LogoPath, Valid: a.LogoPath != ""},
		PostalCode: a.PostalCode,
		Visibility: string(a.Visibility),
	})
}

func decodeAssociation(row sqlc.Association) *association.Association {
	return &association.Association{
		AccountID:  row.AccountID,
		LegalName:  row.LegalName,
		SIREN:      row.Siren.String,
		RNA:        row.Rna.String,
		PostalCode: row.PostalCode,
		About:      row.About.String,
		LogoPath:   row.LogoPath.String,
		Visibility: association.Visibility(row.Visibility),
	}
}
