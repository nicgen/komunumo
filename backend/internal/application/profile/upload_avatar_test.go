package profile_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/application/profile"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/member"
	"komunumo/backend/internal/domain/session"
	"komunumo/backend/internal/ports/fakes"
)

func newUploadAvatarService(t *testing.T) (*profile.UploadAvatarService, *fakes.AccountRepository, *fakes.MemberRepository, *fakes.SessionRepository, *fakes.FileStore) {
	t.Helper()
	accounts := fakes.NewAccountRepository()
	members := fakes.NewMemberRepository()
	sessions := fakes.NewSessionRepository()
	storage := fakes.NewFileStore()
	clk := fakes.NewClock(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))

	svc := profile.NewUploadAvatarService(accounts, members, sessions, storage, clk)
	return svc, accounts, members, sessions, storage
}

func TestUploadAvatar_Success(t *testing.T) {
	svc, accounts, members, sessions, storage := newUploadAvatarService(t)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Seed data
	acc, _ := account.New("acc-1", "lea@example.com", now)
	_ = accounts.Create(context.Background(), acc)
	m, _ := member.New("acc-1", "Léa", "Martin", "2000-01-15", now)
	_ = members.Create(context.Background(), m)
	sess := &session.Session{ID: "sess-1", AccountID: "acc-1", ExpiresAt: now.Add(1 * time.Hour)}
	_ = sessions.Create(context.Background(), sess)

	content := []byte("fake image content")
	reader := bytes.NewReader(content)
	
	path, err := svc.UploadAvatar(context.Background(), "sess-1", reader, int64(len(content)), "image/png")

	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.True(t, storage.FileExists(path))

	// Check member updated
	mUpdated, _ := members.FindByAccountID(context.Background(), "acc-1")
	assert.Equal(t, path, mUpdated.AvatarPath)
}

func TestUploadAvatar_TooLarge(t *testing.T) {
	svc, _, _, _, _ := newUploadAvatarService(t)

	content := make([]byte, 2*1024*1024+1) // 2MB + 1B
	reader := bytes.NewReader(content)
	
	_, err := svc.UploadAvatar(context.Background(), "sess-1", reader, int64(len(content)), "image/png")

	assert.Error(t, err)
}

func TestUploadAvatar_InvalidMime(t *testing.T) {
	svc, _, _, _, _ := newUploadAvatarService(t)

	content := []byte("fake text content")
	reader := bytes.NewReader(content)
	
	_, err := svc.UploadAvatar(context.Background(), "sess-1", reader, int64(len(content)), "text/plain")

	assert.Error(t, err)
}
