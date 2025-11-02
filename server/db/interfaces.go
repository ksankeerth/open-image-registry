package db

import (
	"time"

	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/types/query"
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

type UserDAO interface {
	Transactional

	CreateUserAccount(username, email, displayName, password, salt string, txkey string) (id string, err error)

	DeleteUserAccount(userId string, txkey string) (deleted bool, err error)

	UpdateUserAccount(userId, email, displayName string, txKey string) error

	UpdateUserEmail(userId, newEmail string, txKey string) (updated bool, err error)

	UpdateUserDisplayName(userId, displayName string, txKey string) (updated bool, err error)

	LockUserAccount(username string, lockedReason int, txKey string) (locked bool, err error)

	UnlockUserAccount(username string, txKey string) (unlocked bool, err error)

	RecordFailedAttempt(username string, txKey string) error

	ValidateUsernameAndEmail(username, email string, txKey string) (usernameAvail, emailAvail bool, err error)

	GetUserAccount(username string, txKey string) (*models.UserAccount, error)

	GetUsernameById(userId string, txKey string) (string, error)

	GetUserAccountById(userId string, txKey string) (*models.UserAccount, error)

	GetUserPasswordAndSaltById(userId string, txKey string) (password, salt string, err error)

	UpdateUserPasswordAndSalt(userId, password, salt string, txKey string) (updated bool, err error)

	ListUserAccounts(conditions *query.ListModelsConditions, txKey string) (users []*models.UserAccountView, total int, err error)

	CountUserAccounts(txKey string) (total int, err error)

	PersistPasswordRecovery(userId, recoveryUuid string, reasonType int, txkey string) error

	RetrivePasswordRecoveryByUserId(userId string, txKey string) (*models.PasswordRecovery, error)

	RetrivePasswordRecovery(uuid string, txKey string) (*models.PasswordRecovery, error)

	DeletePasswordRecovery(userId string, txKey string) (deleted bool, err error)

	// PersistUserRole(roleName, txKey string) error

	// DeleteUserRole(roleName string, txKey string) (deleted bool, err error)

	AssignUserRole(roleName, userId string, txKey string) error

	RemoveUserRoleAssignment(userId string, txKey string) error

	IsUserAssignedToRole(userId, roleName string, txKey string) (bool, error)

	GetUserRole(userId string, txKey string) (roleName string, err error)
}

type ResourceAccessDAO interface {
	Transactional

	GrantAccess(resourceType, resourceId, userId, accessLevel, grantedBy string, txKey string) (granted bool, err error)

	RevokeAccess(resourceId, userId, resourceType string, txKey string) (revoked bool, err error)

	GetUserAccess(resourceType, userId string, txKey string) ([]*models.ResourceAccess, error)

	GetUserNamespaceAccess(userId string, txKey string) ([]*models.NamespaceAccess, error)

	GetUserRepositoryAccess(userId string, txKey string) ([]*models.RepositoryAccess, error)
}

type OAuthDAO interface {
	PersistScope(scopeName, description string, txKey string) error

	PersistScopeRoleBinding(scopeName, roleName string, txKey string) error

	GetAllScopeRoleBindings() ([]*models.ScopeRoleBinding, error)

	PersistAuthSession(session *models.OAuthSession, txKey string) error

	PersistAuthSessionScopeBinding(scopes []string, sessionId string, txKey string) error

	RemoveAuthSession(sessionId string, txKey string) error

	GetAuthSession(scopeHash, userId string, txKey string) (*models.OAuthSession, error)

	UpdateSessionLastAccess(sesssionId string, lastAccessed time.Time, txKey string) error
}