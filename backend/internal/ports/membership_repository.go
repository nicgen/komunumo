package ports

import (
	"context"
)

// Membership represents a member's role in an association.
type Membership struct {
	ID                   string
	MemberAccountID      string
	AssociationAccountID string
	Role                 string // owner | admin | member
	Status               string // pending | active | left
	JoinedAt             string // ISO 8601
}

type MembershipRepository interface {
	Create(ctx context.Context, m *Membership) error
	FindByAccountIDs(ctx context.Context, memberID, associationID string) (*Membership, error)
}
