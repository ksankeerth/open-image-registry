package db

import (
	"time"

	"github.com/ksankeerth/open-image-registry/types"
)

// Transactional defines the ability to execute queries inside transactions.
//
// For each unique key, a database transaction will be created and tracked internally.
// The caller must ensure calling either Commit(key) or Rollback(key) to release
// the transaction created by Begin(key).
//
// ⚠️ The caller must also ensure that keys are unique and won't collide with existing ones.
type Transactional interface {
	Begin(key string) error
	Commit(key string) error
	Rollback(key string) error
}

// UpstreamDAO defines data-access operations for upstream registries.
// Only some operations support transactions.
type UpstreamDAO interface {
	Transactional

	// CreateUpstreamRegistry persists a new upstream registry along with its configurations.
	CreateUpstreamRegistry(upstreamReg *types.UpstreamOCIRegEntity,
		authConfig *types.UpstreamOCIRegAuthConfig,
		accessConfig *types.UpstreamOCIRegAccessConfig,
		storageConfig *types.UpstreamOCIRegStorageConfig,
		cacheConfig *types.UpstreamOCIRegCacheConfig,
		txKey string) (regId string, regName string, err error)

	// UpdateUpstreamRegistry updates the registry and its configurations.
	UpdateUpstreamRegistry(regId string, upstreamReg *types.UpstreamOCIRegEntity,
		authConfig *types.UpstreamOCIRegAuthConfig,
		accessConfig *types.UpstreamOCIRegAccessConfig,
		storageConfig *types.UpstreamOCIRegStorageConfig,
		cacheConfig *types.UpstreamOCIRegCacheConfig) error

	// ListUpstreamRegistries returns all upstream registries with additional metadata.
	ListUpstreamRegistries() ([]*types.UpstreamOCIRegEntityWithAdditionalInfo, error)

	// DeleteUpstreamRegistry removes a registry by ID.
	DeleteUpstreamRegistry(regId string) error

	// GetUpstreamRegistry fetches a registry by ID without configs.
	GetUpstreamRegistry(regId string) (*types.UpstreamOCIRegEntity, error)

	// GetUpstreamRegistryWithConfig fetches a registry and its associated configs.
	GetUpstreamRegistryWithConfig(regId string) (*types.UpstreamOCIRegEntity,
		*types.UpstreamOCIRegAccessConfig,
		*types.UpstreamOCIRegAuthConfig,
		*types.UpstreamOCIRegCacheConfig,
		*types.UpstreamOCIRegStorageConfig,
		error)
}

// ImageRegistryDAO defines data-access operations for image registries.
// Methods are grouped by entity: Namespace → Repository → Blob → Manifest → Tag → Config.
type ImageRegistryDAO interface {
	Transactional

	// ---- Namespace / Repository ----

	// CreateNamespaceAndRepositoryIfNotExist ensures a namespace and repository exist,
	// creating them if necessary.
	CreateNamespaceAndRepositoryIfNotExist(regId, namespace, repository string,
		txKey string) (namespaceId, repositoryId string, err error)

	GetNamespaceId(regId, namespace, txKey string) (string, error)
	GetRepositoryId(regId, namespaceId, repository, txKey string) (string, error)

	// ---- Blob ----

	// CheckImageBlobExists returns true if the blob exists, false otherwise.
	CheckImageBlobExists(digest, namespace, repository, regId string, txKey string) (bool, error)

	// CheckImageBlobExistsByIds is an optimized variant using namespace/repository IDs.
	CheckImageBlobExistsByIds(digest, namespaceId, repositoryId, regId string, txKey string) (bool, error)

	// GetImageBlobStorageLocationAndSize returns blob storage location and size.
	GetImageBlobStorageLocationAndSize(digest, namespace, repository,
		regId string, txKey string) (storageLocation string, size int, err error)

	// GetImageBlobStorageLocationAndSizeByIds is an optimized variant using IDs.
	GetImageBlobStorageLocationAndSizeByIds(digest, namespaceId, repositoryId,
		regId string, txKey string) (storageLocation string, size int, err error)

	// PersistImageBlobMetaInfo stores blob metadata (location, type, size).
	PersistImageBlobMetaInfo(registryId, namespaceId, repositoryId, digest, location, mediaType string,
		size int64, txKey string) error

	// ---- Blob Upload Sessions ----

	GetBlobUploadSession(sessionId string, txKey string) (id string, isChunked bool, lastUpdated time.Time, err error)
	CreateNewBlobUploadSession(registryId, namespaceId, repositoryId, sessionId string, txKey string) error
	DeleteBlobUploadSession(sessionId string, txKey string) error
	MarkBlobUploadSessionAsChunked(sessionId string, txKey string) error

	// ---- Manifest ----

	// CheckImageManifestExists returns true if the manifest exists, false otherwise.
	CheckImageManifestExists(manifestUniqDigest, namespace, repository, regId string, txKey string) (bool, error)

	// GetImageManifestsByTag returns manifests for a given tag including platform info.
	GetImageManifestsByTag(tag, namespace, repository, regId string,
		txKey string) ([]*types.ImageManifestWithPlatform, error)

	// CheckImageManifestsExistByTag returns true if manifests exist for the given tag.
	CheckImageManifestsExistByTag(regId, namespace, repository, tag, txKey string) (bool, error)

	// GetImageManifestByDigest fetches manifest content, type, and size.
	GetImageManifestByDigest(digest, namespace, repository, regId string, txKey string) (
		content []byte, mediaType string, size int, err error)

	// PersistImageManifest stores a new image manifest and its metadata.
	PersistImageManifest(regId, namespaceId, repositoryId, manifestDigest,
		imageConfigDigest, mediaType, uniqueDigest string, size int64,
		content []byte, txKey string) error

	// ---- Tag ----

	GetImageTagId(regId, namespaceId, repositoryId, tag, txKey string) (tagId string, err error)
	CreateImageTag(regId, namespaceId, repositoryId, tag string, txKey string) (tagId string, err error)

	// ---- Config ----

	// MarkImageBlobAsConfig marks a blob as an image config.
	MarkImageBlobAsConfig(regId, namespaceId, repositoryId, blobDigest, txKey string) error

	// CheckImageTagAndConfigDigestMapping checks whether a config is already linked to a tag.
	CheckImageTagAndConfigDigestMapping(configDigest, tagId, txKey string) (exists bool, err error)

	// LinkImageConfigWithTag links a config digest to a tag.
	LinkImageConfigWithTag(configDigest, tagId, txKey string) error

	// PersistImagePlatformConfig stores platform-specific metadata for an image config.
	PersistImagePlatformConfig(regId, namespaceId, repositoryId, tagId,
		configDigest, os, arch string, props []byte, txKey string) error

	// CheckImagePlatformConfigExistence verifies if a platform config already exists.
	CheckImagePlatformConfigExistence(tagId, configDigest string, txKey string) (bool, error)

	// ---- Tag ↔ Layer ----

	// LinkImageTagAndLayer links a tag with a layer digest and config.
	LinkImageTagAndLayer(regId, namespaceId, repositoryId, tagId, layerDigest,
		configDigest string, index int, txKey string) error

	// CheckImageTagAndLayerMappingExistence checks if a tag-layer mapping already exists.
	CheckImageTagAndLayerMappingExistence(tagId, layerDigest string, layerIndex int, txKey string) (bool, error)
}