package profile

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"komunumo/backend/internal/ports"
)

type UploadAvatarService struct {
	accounts ports.AccountRepository
	members  ports.MemberRepository
	sessions ports.SessionRepository
	storage  ports.FileStore
	clock    ports.Clock
}

func NewUploadAvatarService(
	accounts ports.AccountRepository,
	members ports.MemberRepository,
	sessions ports.SessionRepository,
	storage ports.FileStore,
	clock ports.Clock,
) *UploadAvatarService {
	return &UploadAvatarService{
		accounts: accounts,
		members:  members,
		sessions: sessions,
		storage:  storage,
		clock:    clock,
	}
}

func (s *UploadAvatarService) UploadAvatar(ctx context.Context, sessionID string, r io.Reader, size int64, mimeType string) (string, error) {
	const maxBytes = 2 * 1024 * 1024 // 2MB
	if size > maxBytes {
		return "", fmt.Errorf("file too large")
	}

	allowedMimes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	ext, ok := allowedMimes[mimeType]
	if !ok {
		return "", fmt.Errorf("invalid mime type")
	}

	now := s.clock.Now()
	sess, err := s.sessions.FindByID(ctx, sessionID, now)
	if err != nil {
		return "", err
	}

	m, err := s.members.FindByAccountID(ctx, sess.AccountID)
	if err != nil {
		return "", err
	}
	if m == nil {
		return "", fmt.Errorf("member profile not found")
	}

	// Store file
	path, err := s.storage.StoreAvatar(ctx, sess.AccountID, r, filepath.Clean(ext))
	if err != nil {
		return "", err
	}

	// Update member profile
	m.AvatarPath = path
	if err := s.members.Update(ctx, m); err != nil {
		return "", err
	}

	return path, nil
}
