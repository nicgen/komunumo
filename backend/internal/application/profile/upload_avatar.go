package profile

import (
	"context"
	"io"

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
	return "", nil
}
