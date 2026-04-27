package fakes

import (
	"context"
	"time"

	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports"
)

var _ ports.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	byID map[string]*session.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{byID: make(map[string]*session.Session)}
}

func (r *SessionRepository) Create(_ context.Context, s *session.Session) error {
	cp := *s
	r.byID[s.ID] = &cp
	return nil
}

func (r *SessionRepository) FindByID(_ context.Context, id string, now time.Time) (*session.Session, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, session.ErrSessionNotFound
	}
	if s.Expired(now) {
		return nil, session.ErrSessionExpired
	}
	cp := *s
	return &cp, nil
}

func (r *SessionRepository) TouchLastSeen(_ context.Context, id string, at time.Time) error {
	s, ok := r.byID[id]
	if !ok {
		return session.ErrSessionNotFound
	}
	s.LastSeenAt = at
	return nil
}

func (r *SessionRepository) Delete(_ context.Context, id string) error {
	delete(r.byID, id)
	return nil
}

func (r *SessionRepository) DeleteAllForAccount(_ context.Context, accountID string) error {
	for id, s := range r.byID {
		if s.AccountID == accountID {
			delete(r.byID, id)
		}
	}
	return nil
}

func (r *SessionRepository) DeleteExpired(_ context.Context, now time.Time) (int64, error) {
	var count int64
	for id, s := range r.byID {
		if s.Expired(now) {
			delete(r.byID, id)
			count++
		}
	}
	return count, nil
}

func (r *SessionRepository) Count() int { return len(r.byID) }
