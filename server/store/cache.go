package store

import (
	"context"
	"time"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type RegistryCacheStore interface {
	Create(ctx context.Context, registryId, namespaceId, repositoryId, identifier, digest string, expiresAt time.Time) error

	Get(ctx context.Context, repositoryId, identifier string) (*models.RegistryCacheModel, error)

	Delete(ctx context.Context, repositoryId, identifier string) (err error)

	Refresh(ctx context.Context, repositoryId, identifier string, expiresAt time.Time) error
}
