package fakes

import (
	"context"

	"komunumo/backend/internal/ports"
)

var _ ports.MembershipRepository = (*MembershipRepository)(nil)

type MembershipRepository struct {
	memberships []*ports.Membership
}

func NewMembershipRepository() *MembershipRepository {
	return &MembershipRepository{}
}

func (r *MembershipRepository) Create(_ context.Context, m *ports.Membership) error {
	cp := *m
	r.memberships = append(r.memberships, &cp)
	return nil
}

func (r *MembershipRepository) FindByAccountIDs(_ context.Context, memberID, associationID string) (*ports.Membership, error) {
	for _, m := range r.memberships {
		if m.MemberAccountID == memberID && m.AssociationAccountID == associationID {
			cp := *m
			return &cp, nil
		}
	}
	return nil, nil
}
