package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type NamespaceQueries interface {
	ValidateMaintainers(ctx context.Context, userIds []string) (bool, error)
}

type ImageQueries interface {
	GetManifestByTag(ctx context.Context, withContent bool, repositoryId, tag string) (*models.ImageManifestModel, error)

	GetRepositoryByNames(ctx context.Context, namespace, repository string) (*models.RepositoryModel, error)
}
