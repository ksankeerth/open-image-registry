package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type RepositoryStore interface {
	Create(ctx context.Context, regId, nsId, name, description string, isPublic bool) (id string, err error)

	Get(ctx context.Context, id string) (*models.RepositoryModel, error)

	Delete(ctx context.Context, id string) (err error)

	GetID(ctx context.Context, namespaceId, name string) (id string, err error)
}
