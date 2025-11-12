package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type ManifestStore interface {
	Create(ctx context.Context, registryId, namespaceId, repositoryId, digest, mediaType, uniqueDigest string,
		size int64, content []byte) (id string, err error)

	GetByUniqueDigest(ctx context.Context, withContent bool, repositoryId, digest string) (*models.ImageManifestModel, error)

	GetByDigest(ctx context.Context, withContent bool, epositoryId, digest string) (*models.ImageManifestModel, error)
}
