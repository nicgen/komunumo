package ports

import (
	"context"

	"komunumo/backend/internal/domain/member"
)

type MemberRepository interface {
	Create(ctx context.Context, m *member.Member) error
	FindByAccountID(ctx context.Context, accountID string) (*member.Member, error)
	Update(ctx context.Context, m *member.Member) error
}
