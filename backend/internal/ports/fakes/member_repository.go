package fakes

import (
	"context"

	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/ports"
)

var _ ports.MemberRepository = (*MemberRepository)(nil)

type MemberRepository struct {
	byAccountID map[string]*member.Member
}

func NewMemberRepository() *MemberRepository {
	return &MemberRepository{byAccountID: make(map[string]*member.Member)}
}

func (r *MemberRepository) Create(_ context.Context, m *member.Member) error {
	cp := *m
	r.byAccountID[m.AccountID] = &cp
	return nil
}

func (r *MemberRepository) FindByAccountID(_ context.Context, accountID string) (*member.Member, error) {
	m, ok := r.byAccountID[accountID]
	if !ok {
		return nil, nil
	}
	cp := *m
	return &cp, nil
}

func (r *MemberRepository) Update(_ context.Context, m *member.Member) error {
	if _, ok := r.byAccountID[m.AccountID]; !ok {
		return nil
	}
	cp := *m
	r.byAccountID[m.AccountID] = &cp
	return nil
}
