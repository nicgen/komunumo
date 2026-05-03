package ports

import (
	"context"

	"komunumo/backend/internal/domain/association"
)

type AssociationRepository interface {
	Create(ctx context.Context, a *association.Association) error
	FindByAccountID(ctx context.Context, accountID string) (*association.Association, error)
	Update(ctx context.Context, a *association.Association) error
}
