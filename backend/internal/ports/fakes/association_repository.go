package fakes

import (
	"context"

	"komunumo/backend/internal/domain/association"
	"komunumo/backend/internal/ports"
)

var _ ports.AssociationRepository = (*AssociationRepository)(nil)

type AssociationRepository struct {
	byAccountID map[string]*association.Association
}

func NewAssociationRepository() *AssociationRepository {
	return &AssociationRepository{byAccountID: make(map[string]*association.Association)}
}

func (r *AssociationRepository) Create(_ context.Context, a *association.Association) error {
	cp := *a
	r.byAccountID[a.AccountID] = &cp
	return nil
}

func (r *AssociationRepository) FindByAccountID(_ context.Context, accountID string) (*association.Association, error) {
	a, ok := r.byAccountID[accountID]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (r *AssociationRepository) Update(_ context.Context, a *association.Association) error {
	if _, ok := r.byAccountID[a.AccountID]; !ok {
		return nil
	}
	cp := *a
	r.byAccountID[a.AccountID] = &cp
	return nil
}
