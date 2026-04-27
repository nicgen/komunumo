package fakes

import (
	"context"

	"komunumo/backend/internal/domain/audit"
	"komunumo/backend/internal/ports"
)

var _ ports.AuditRepository = (*AuditRepository)(nil)

type AuditRepository struct {
	Events []*audit.Event
}

func NewAuditRepository() *AuditRepository { return &AuditRepository{} }

func (r *AuditRepository) Append(_ context.Context, e *audit.Event) error {
	cp := *e
	r.Events = append(r.Events, &cp)
	return nil
}

func (r *AuditRepository) LastOfType(t audit.EventType) *audit.Event {
	for i := len(r.Events) - 1; i >= 0; i-- {
		if r.Events[i].Type == t {
			return r.Events[i]
		}
	}
	return nil
}

func (r *AuditRepository) ContainsType(t string) bool {
	for _, e := range r.Events {
		if string(e.Type) == t {
			return true
		}
	}
	return false
}
