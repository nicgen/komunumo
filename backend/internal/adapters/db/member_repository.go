package db

import (
	"context"
	"database/sql"
	"errors"

	"komunumo/backend/internal/adapters/db/sqlc"
	"komunumo/backend/internal/domain/member"
)

type MemberRepository struct {
	q *sqlc.Queries
}

func NewMemberRepository(conn *sql.DB) *MemberRepository {
	return &MemberRepository{q: sqlc.New(conn)}
}

func (r *MemberRepository) withTx(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *MemberRepository) Create(ctx context.Context, m *member.Member) error {
	return r.withTx(ctx).CreateMember(ctx, sqlc.CreateMemberParams{
		AccountID: m.AccountID,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		BirthDate: m.BirthDate,
		Nickname:  sql.NullString{String: m.Nickname, Valid: m.Nickname != ""},
		AboutMe:   sql.NullString{String: m.AboutMe, Valid: m.AboutMe != ""},
		AvatarPath: sql.NullString{String: m.AvatarPath, Valid: m.AvatarPath != ""},
		Visibility: string(m.Visibility),
	})
}

func (r *MemberRepository) FindByAccountID(ctx context.Context, accountID string) (*member.Member, error) {
	row, err := r.withTx(ctx).GetMemberByAccountID(ctx, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return decodeMember(row), nil
}

func (r *MemberRepository) Update(ctx context.Context, m *member.Member) error {
	return r.withTx(ctx).UpdateMember(ctx, sqlc.UpdateMemberParams{
		AccountID:  m.AccountID,
		Nickname:   sql.NullString{String: m.Nickname, Valid: m.Nickname != ""},
		AboutMe:    sql.NullString{String: m.AboutMe, Valid: m.AboutMe != ""},
		AvatarPath: sql.NullString{String: m.AvatarPath, Valid: m.AvatarPath != ""},
		Visibility: string(m.Visibility),
	})
}

func decodeMember(row sqlc.Member) *member.Member {
	return &member.Member{
		AccountID:  row.AccountID,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		BirthDate:  row.BirthDate,
		Nickname:   row.Nickname.String,
		AboutMe:    row.AboutMe.String,
		AvatarPath: row.AvatarPath.String,
		Visibility: member.Visibility(row.Visibility),
	}
}
