package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adapter "komunumo/backend/internal/adapters/http"
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

func TestHandleGetMyProfile_Success(t *testing.T) {
	h, accounts, members, _, sessions := newProfileHandler(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Seed
	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	_ = members.Create(context.Background(), m)
	sess := &session.Session{ID: "sess-1", AccountID: "acc-1", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	req := httptest.NewRequest("GET", "/api/v1/me/profile", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-1"})
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

	// Seed
	acc, _ := account.New("acc-1", "lea@example.com", now)
	acc.Kind = account.KindMember
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	_ = members.Create(context.Background(), m)
	sess := &session.Session{ID: "sess-1", AccountID: "acc-1", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	nickname := "lea42"
	body, _ := json.Marshal(profile.UpdateProfileInput{Nickname: &nickname})
	req := httptest.NewRequest("PATCH", "/api/v1/me/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-1"})
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
