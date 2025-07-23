package db

import "github.com/ksankeerth/open-image-registry/types"

// UpstreamDAO is an interface for data-access of Upstream Registeries.
type UpstreamDAO interface {
	CreateUpstreamRegistry(upstreamReg *types.UpstreamOCIRegEntity,
		authConfig *types.UpstreamOCIRegAuthConfig,
		accessConfig *types.UpstreamOCIRegAccessConfig,
		storageConfig *types.UpstreamOCIRegStorageConfig,
		cacheConfig *types.UpstreamOCIRegCacheConfig) (regId string, regName string, err error)
	UpdateUpstreamRegistry(regId string, upstreamReg *types.UpstreamOCIRegEntity,
		authConfig *types.UpstreamOCIRegAuthConfig,
		accessConfig *types.UpstreamOCIRegAccessConfig,
		storageConfig *types.UpstreamOCIRegStorageConfig,
		cacheConfig *types.UpstreamOCIRegCacheConfig) (err error)
	ListUpstreamRegistries() (registeries []*types.UpstreamOCIRegEntityWithAdditionalInfo, err error)
	DeleteUpstreamRegistry(regId string) (err error)
	GetUpstreamRegistry(regId string) (*types.UpstreamOCIRegEntity, error)

	GetUpstreamRegistryWithConfig(regId string) (*types.UpstreamOCIRegEntity,
		*types.UpstreamOCIRegAccessConfig, *types.UpstreamOCIRegAuthConfig,
		*types.UpstreamOCIRegCacheConfig, *types.UpstreamOCIRegStorageConfig, error)
}
