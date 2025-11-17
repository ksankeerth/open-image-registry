package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type BlobMetaStore interface {
	Get(ctx context.Context, digest, repositoryId string) (*models.ImageBlobMetaModel, error)

	Create(ctx context.Context, registryId, namespaceId, repositoryId, digest, location string, size int64) (err error)

	CreateUploadSession(ctx context.Context, sessionID, namespace, repository string) error

	UpdateUploadSession(ctx context.Context, sessionID string, bytesReceived int) error

	DeleteUploadSession(ctx context.Context, sessionID string) error

	GetUploadSession(ctx context.Context, sessionID string) (*models.ImageBlobUploadSessionModel, error)
}
