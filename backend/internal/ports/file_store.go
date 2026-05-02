package ports

import (
	"context"
	"io"
)

type FileStore interface {
	// StoreAvatar stores the original file and returns its path relative to data/uploads/.
	StoreAvatar(ctx context.Context, accountID string, content io.Reader, ext string) (path string, err error)
	// AvatarURL returns the public URL for a given avatar path.
	AvatarURL(path string) string
}
