package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adapter "komunumo/backend/internal/adapters/http"
	"komunumo/backend/internal/adapters/http/middleware"
	"komunumo/backend/internal/application/profile"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports/fakes"
)

func newProfileHandler(t *testing.T) (*adapter.ProfileHandler, *fakes.AccountRepository, *fakes.MemberRepository, *fakes.AssociationRepository, *fakes.SessionRepository) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	members := fakes.NewMemberRepository()
	associations := fakes.NewAssociationRepository()
	sessions := fakes.NewSessionRepository()
	audit := fakes.NewAuditRepository()
	clk := fakes.NewClock(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	tokenGen := fakes.NewTokenGenerator()
	storage := fakes.NewFileStore()

	getSvc := profile.NewGetProfileService(accounts, members, associations, sessions, clk)
	updateSvc := profile.NewUpdateProfileService(accounts, members, associations, sessions, audit, clk, tokenGen)
	uploadSvc := profile.NewUploadAvatarService(accounts, members, sessions, storage, clk)

	h := adapter.NewProfileHandler(getSvc, updateSvc, uploadSvc)
	return h, accounts, members, associations, sessions
}

// withSession injects a session ID into the request context, simulating the Auth middleware.
func withSession(r *http.Request, sessionID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.SessionIDKey, sessionID)
	return r.WithContext(ctx)
}

func TestHandleGetMyProfile_Success(t *testing.T) {
	h, accounts, members, _, sessions := newProfileHandler(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	_ = members.Create(context.Background(), m)
	sess := &session.Session{ID: "sess-1", AccountID: "acc-1", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	req := withSession(httptest.NewRequest("GET", "/api/v1/me/profile", nil), "sess-1")
	rr := httptest.NewRecorder()

	h.HandleGetMyProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp profile.ProfileOutput
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Léa", resp.FirstName)
}

func TestHandleUpdateMyProfile_Success(t *testing.T) {
	h, accounts, members, _, sessions := newProfileHandler(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	_ = members.Create(context.Background(), m)
	sess := &session.Session{ID: "sess-1", AccountID: "acc-1", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	nickname := "lea42"
	body, _ := json.Marshal(profile.UpdateProfileInput{Nickname: &nickname})
	req := withSession(httptest.NewRequest("PATCH", "/api/v1/me/profile", bytes.NewReader(body)), "sess-1")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.HandleUpdateMyProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mUpdated, _ := members.FindByAccountID(context.Background(), "acc-1")
	assert.Equal(t, "lea42", mUpdated.Nickname)
}

func TestHandleGetMyProfile_Unauthorized(t *testing.T) {
	h, _, _, _, _ := newProfileHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/me/profile", nil)
	rr := httptest.NewRecorder()

	h.HandleGetMyProfile(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
func TestHandleGetPublicProfile_Public(t *testing.T) {
	h, accounts, members, _, _ := newProfileHandler(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	m.Visibility = member.VisibilityPublic
	_ = members.Create(context.Background(), m)

	req := httptest.NewRequest("GET", "/api/v1/accounts/acc-1/profile", nil)
	// Manual injection of chi URL param for testing without full router
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "acc-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	h.HandleGetPublicProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp profile.ProfileOutput
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, "Léa", resp.FirstName)
	assert.Empty(t, resp.BirthDate)
}

func TestHandleGetPublicProfile_Private(t *testing.T) {
	h, accounts, members, _, _ := newProfileHandler(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	m.Visibility = member.VisibilityPrivate
	_ = members.Create(context.Background(), m)

	req := httptest.NewRequest("GET", "/api/v1/accounts/acc-1/profile", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("accountId", "acc-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	h.HandleGetPublicProfile(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
