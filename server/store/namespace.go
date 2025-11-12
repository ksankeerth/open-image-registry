package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type NamespaceStore interface {
	Create(ctx context.Context, regId, name, purpose, description string, isPublic bool) (id string, err error)

	Get(ctx context.Context, id string) (*models.NamespaceModel, error)

	GetByName(ctx context.Context, regId, name string) (*models.NamespaceModel, error)

	GetID(ctx context.Context, registryId, namesapce string) (id string, err error)

	Delete(ctx context.Context, id string) (err error)
}
