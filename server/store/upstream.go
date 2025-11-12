package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type UpstreamRegistyStore interface {
	CreateRegistry(ctx context.Context, m *models.UpstreamRegistry) (id string, err error)

	UpdateRegistry(ctx context.Context, m *models.UpstreamRegistry) error

	GetRegistry(ctx context.Context, registryID string) (*models.UpstreamRegistry, error)

	DeleteRegistry(ctx context.Context, registryID string) error

	ChangeRegistryState(ctx context.Context, registryID, state string) error

	PersistRegistryAuthConfig(ctx context.Context, m *models.UpstreamRegistryAuthConfig) error

	UpdateRegistryAuthConfig(ctx context.Context, m *models.UpstreamRegistryAuthConfig) error

	GetRegistryAuthConfig(ctx context.Context, registryID string) (*models.UpstreamRegistryAuthConfig, error)

	PersistRegistryCacheConfig(ctx context.Context, m *models.UpstreamRegistryCacheStoreConfig) error

	UpdateRegistryCacheConfig(ctx context.Context, m *models.UpstreamRegistryCacheStoreConfig) error

	GetRegistryCacheConfig(ctx context.Context, registryID string) (*models.UpstreamRegistryCacheStoreConfig, error)

	PersistRegistryNetworkConfig(ctx context.Context, m *models.UpstreamRegistryNetworkConfig) error

	UpdateRegistryNetworkConfig(ctx context.Context, m *models.UpstreamRegistryNetworkConfig) error

	GetRegistryNetworkConfig(ctx context.Context, registryID string) (*models.UpstreamRegistryNetworkConfig, error)

	GetAllUpstreamRegistryAddresses(ctx context.Context) (addresses []*models.UpstreamAddressView, err error)
}
