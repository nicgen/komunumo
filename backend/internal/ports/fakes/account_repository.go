package fakes

import (
	"context"
	"time"

	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/ports"
)

var _ ports.AccountRepository = (*AccountRepository)(nil)

type AccountRepository struct {
	byID    map[string]*account.Account
	byEmail map[string]*account.Account
}

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{
		byID:    make(map[string]*account.Account),
		byEmail: make(map[string]*account.Account),
	}
}

func (r *AccountRepository) Create(_ context.Context, a *account.Account) error {
	if _, exists := r.byEmail[a.EmailCanonical]; exists {
		return account.ErrEmailTaken
	}
	cp := *a
	r.byID[a.ID] = &cp
	r.byEmail[a.EmailCanonical] = &cp
	return nil
}

func (r *AccountRepository) FindByEmailCanonical(_ context.Context, emailCanonical string) (*account.Account, error) {
	a, ok := r.byEmail[emailCanonical]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (r *AccountRepository) FindByID(_ context.Context, id string) (*account.Account, error) {
	a, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (r *AccountRepository) UpdateStatus(_ context.Context, id string, status account.Status, at time.Time) error {
	a, ok := r.byID[id]
	if !ok {
		return account.ErrAccountNotFound
	}
	a.Status = status
	a.UpdatedAt = at
	r.byEmail[a.EmailCanonical] = a
	return nil
}

func (r *AccountRepository) UpdatePasswordHash(_ context.Context, id, hash string, at time.Time) error {
	a, ok := r.byID[id]
	if !ok {
		return account.ErrAccountNotFound
	}
	a.PasswordHash = hash
	a.UpdatedAt = at
	r.byEmail[a.EmailCanonical] = a
	return nil
}

func (r *AccountRepository) TouchLastLogin(_ context.Context, id string, at time.Time) error {
	a, ok := r.byID[id]
	if !ok {
		return account.ErrAccountNotFound
	}
	a.LastLoginAt = &at
	r.byEmail[a.EmailCanonical] = a
	return nil
}

func (r *AccountRepository) Count() int { return len(r.byID) }
