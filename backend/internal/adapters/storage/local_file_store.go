package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"komunumo/backend/internal/ports"
)

var _ ports.FileStore = (*LocalFileStore)(nil)

type LocalFileStore struct {
	baseDir string // absolute path to data/uploads/
	baseURL string // e.g. "/uploads"
}

func NewLocalFileStore(baseDir, baseURL string) *LocalFileStore {
	return &LocalFileStore{baseDir: baseDir, baseURL: baseURL}
}

func (s *LocalFileStore) StoreAvatar(_ context.Context, accountID string, content io.Reader, ext string) (string, error) {
	dir := filepath.Join(s.baseDir, "avatars", accountID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("storage: create avatar dir: %w", err)
	}

	filename := fmt.Sprintf("avatar.%s", ext)
	dest := filepath.Join(dir, filename)

	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("storage: create avatar file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("storage: write avatar: %w", err)
	}

	return filepath.Join("avatars", accountID, filename), nil
}

func (s *LocalFileStore) AvatarURL(path string) string {
	return s.baseURL + "/" + path
}
