package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type BlobMetaStore interface {
	Get(ctx context.Context, digest, repositoryId string) (*models.ImageBlobMetaModel, error)

	Create(ctx context.Context, registryId, namespaceId, repositoryId, digest, location string, size int64) (err error)
}
