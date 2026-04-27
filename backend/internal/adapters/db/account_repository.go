package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"komunumo/backend/internal/adapters/db/sqlc"
	"komunumo/backend/internal/domain/account"
)

type AccountRepository struct {
	q *sqlc.Queries
}

func NewAccountRepository(conn *sql.DB) *AccountRepository {
	return &AccountRepository{q: sqlc.New(conn)}
}

func (r *AccountRepository) withTx(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *AccountRepository) Create(ctx context.Context, a *account.Account) error {
	err := r.withTx(ctx).CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:             a.ID,
		Email:          a.Email,
		EmailCanonical: a.EmailCanonical,
		PasswordHash:   a.PasswordHash,
		Status:         string(a.Status),
		FirstName:      a.FirstName,
		LastName:       a.LastName,
		DateOfBirth:    a.DateOfBirth.Format("2006-01-02"),
		CreatedAt:      encodeTime(a.CreatedAt),
		UpdatedAt:      encodeTime(a.UpdatedAt),
	})
	if err != nil {
		if isUniqueConstraintError(err) {
			return account.ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *AccountRepository) FindByEmailCanonical(ctx context.Context, emailCanonical string) (*account.Account, error) {
	row, err := r.withTx(ctx).GetAccountByEmailCanonical(ctx, emailCanonical)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return decodeAccount(row)
}

func (r *AccountRepository) FindByID(ctx context.Context, id string) (*account.Account, error) {
	row, err := r.withTx(ctx).GetAccountByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return decodeAccount(row)
}

func (r *AccountRepository) UpdateStatus(ctx context.Context, id string, status account.Status, at time.Time) error {
	return r.withTx(ctx).UpdateAccountStatus(ctx, sqlc.UpdateAccountStatusParams{
		Status:    string(status),
		UpdatedAt: encodeTime(at),
		ID:        id,
	})
}

func (r *AccountRepository) UpdatePasswordHash(ctx context.Context, id, hash string, at time.Time) error {
	return r.withTx(ctx).UpdateAccountPasswordHash(ctx, sqlc.UpdateAccountPasswordHashParams{
		PasswordHash: hash,
		UpdatedAt:    encodeTime(at),
		ID:           id,
	})
}

func (r *AccountRepository) TouchLastLogin(ctx context.Context, id string, at time.Time) error {
	return r.withTx(ctx).TouchAccountLastLogin(ctx, sqlc.TouchAccountLastLoginParams{
		LastLoginAt: sql.NullString{String: encodeTime(at), Valid: true},
		ID:          id,
	})
}

func decodeAccount(row sqlc.Account) (*account.Account, error) {
	createdAt, err := decodeTime(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := decodeTime(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	dob, err := time.Parse("2006-01-02", row.DateOfBirth)
	if err != nil {
		return nil, err
	}
	lastLogin, err := decodeNullTime(row.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return &account.Account{
		ID:             row.ID,
		Email:          row.Email,
		EmailCanonical: row.EmailCanonical,
		PasswordHash:   row.PasswordHash,
		Status:         account.Status(row.Status),
		FirstName:      row.FirstName,
		LastName:       row.LastName,
		DateOfBirth:    dob,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		LastLoginAt:    lastLogin,
	}, nil
}
