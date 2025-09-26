package db

import (
	"time"

	"github.com/ksankeerth/open-image-registry/types/models"
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
	CreateUpstreamRegistry(upstreamReg *models.UpstreamRegistryEntity,
		authConfig *models.UpstreamRegistryAuthConfig,
		accessConfig *models.UpstreamRegistryAccessConfig,
		storageConfig *models.UpstreamRegistryStorageConfig,
		cacheConfig *models.UpstreamRegistryCacheConfig,
		txKey string) (regId string, regName string, err error)

	// UpdateUpstreamRegistry updates the registry and its configurations.
	UpdateUpstreamRegistry(regId string, upstreamReg *models.UpstreamRegistryEntity,
		authConfig *models.UpstreamRegistryAuthConfig,
		accessConfig *models.UpstreamRegistryAccessConfig,
		storageConfig *models.UpstreamRegistryStorageConfig,
		cacheConfig *models.UpstreamRegistryCacheConfig) error

	// ListUpstreamRegistries returns all upstream registries with additional metadata.
	ListUpstreamRegistries() ([]*models.UpstreamRegistrySummary, error)

	// DeleteUpstreamRegistry removes a registry by ID.
	DeleteUpstreamRegistry(regId string) error

	// GetUpstreamRegistryWithConfig fetches a registry and its associated configs.
	GetUpstreamRegistryWithConfig(regId string) (*models.UpstreamRegistryEntity,
		*models.UpstreamRegistryAccessConfig,
		*models.UpstreamRegistryAuthConfig,
		*models.UpstreamRegistryCacheConfig,
		*models.UpstreamRegistryStorageConfig,
		error)

	// GetActiveUpstreamAddresses loads all active upstream registry and return addresses of them
	GetActiveUpstreamAddresses() (upstreamAddrs []*models.UpstreamRegistryAddress, err error)
}

// ImageRegistryDAO defines data-access operations for image registries.
// Methods are grouped by entity: Namespace → Repository → Blob → Manifest → Tag → Config.
type ImageRegistryDAO interface {
	Transactional

	// ---- Namespace / Repository ----

	// CreateNamespaceAndRepositoryIfNotExist ensures a namespace and repository exist,
	// create them if necessary.
	CreateNamespaceAndRepositoryIfNotExist(regId, namespace, repository string,
		txKey string) (namespaceId, repositoryId string, err error)

	GetNamespaceId(regId, namespace, txKey string) (string, error)
	GetRepositoryId(regId, namespaceId, repository, txKey string) (string, error)

	// ---- Blob ----

	// GetImageBlobStorageLocationAndSize returns blob storage location and size.
	GetImageBlobStorageLocationAndSize(digest, namespace, repository,
		regId string, txKey string) (storageLocation string, size int, err error)

	// PersistImageBlobMetaInfo stores blob metadata (location, type, size).
	PersistImageBlobMetaInfo(registryId, namespaceId, repositoryId, digest, location string,
		size int64, txKey string) error

	// ---- Manifest ----

	CheckImageManifestExistsByTag(registryId, namespaceId, repositoryId, tag string, txKey string) (exists bool, digest, mediaType string, err error)

	CheckImageManifestExistsByDigest(registryId, namespaceId, repositoryId, digest string, txKey string) (exists bool, mediaType string, err error)

	GetImageManifestByTag(registryId, namespaceId, repositoryId, tag string, txKey string) (exists bool, content []byte, digest, mediaType string, err error)

	GetImageManifestByDigest(registryId, namespaceId, repositoryId, digest string, txKey string) (exists bool, content []byte, mediaType string, err error)

	CheckImageManifestExistsByUniqueDigest(registryId, namespaceId, repositoryId, unqiqueDigest string,
		txKey string) (exists bool, id string, err error)

	// PersistImageManifest stores a new image manifest and its metadata.
	PersistImageManifest(regId, namespaceId, repositoryId, manifestDigest, mediaType, uniqueDigest string, size int64,
		content []byte, txKey string) (id string, err error)

	// ---- Tag ----

	GetImageTagId(regId, namespaceId, repositoryId, tag, txKey string) (tagId string, err error)
	CreateImageTag(regId, namespaceId, repositoryId, tag string, txKey string) (tagId string, err error)

	// ---- Registry Cache ----

	CacheImageManifestReference(regId, namespaceId, repositoryId, identifier string,
		expiresAt time.Time, txKey string) error
	GetCacheImageManifestReference(regId, namespaceId, repositoryId, identifier string,
		txKey string) (cacheMiss bool, digest string, expiresAt time.Time, err error)
	DeleteCacheImageManifestReference(regId, namespaceId, repositoryId, identifier string, txKey string) error
	RefreshCacheImageManifestReference(regId, namespaceId, repositoryId, identifier string, expiresAt time.Time, txKey string) error

	// ---- Tag ↔ Manifest Linking ----

	LinkImageManifestWithTag(tagId, manifestId string, txKey string) error
	UpdateManifestIdForTag(tagId string, newManifestId string, txKey string) error

	GetLinkedManifestByTagId(tagId string, txKey string) (manifestId string, err error)
}