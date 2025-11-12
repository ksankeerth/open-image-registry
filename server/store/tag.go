package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type ImageTagStore interface {
	Create(ctx context.Context, registryId, namespaceId, repositoryId, tag string) (id string, err error)

	Get(ctx context.Context, repositoryId, tag string) (*models.ImageTagModel, error)

	Delete(ctx context.Context, repositoryId, tag string) (err error)

	LinkManifest(ctx context.Context, tagId, manifestId string) error

	UpdateManifest(ctx context.Context, tagId, newManifestId string) error

	UnlinkManifest(ctx context.Context, tagId string) error

	GetManifestID(ctx context.Context, tagId string) (string, error)
}
