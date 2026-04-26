package ports

import (
	"context"

	"komunumo/backend/internal/domain/audit"
)

type AuditRepository interface {
	Append(ctx context.Context, e *audit.Event) error
}
