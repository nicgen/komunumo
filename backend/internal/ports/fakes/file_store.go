package fakes

import (
	"context"
	"fmt"
	"io"

	"komunumo/backend/internal/ports"
)

var _ ports.FileStore = (*FileStore)(nil)

type FileStore struct {
	stored map[string][]byte
}

func NewFileStore() *FileStore {
	return &FileStore{stored: make(map[string][]byte)}
}

func (s *FileStore) StoreAvatar(_ context.Context, accountID string, content io.Reader, ext string) (string, error) {
	data, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("avatars/%s/avatar.%s", accountID, ext)
	s.stored[path] = data
	return path, nil
}

func (s *FileStore) AvatarURL(path string) string {
	return "/uploads/" + path
}

func (s *FileStore) FileExists(path string) bool {
	_, ok := s.stored[path]
	return ok
}
