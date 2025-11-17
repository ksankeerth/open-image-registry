package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	up "github.com/ksankeerth/open-image-registry/client/upstream"
	"github.com/ksankeerth/open-image-registry/client/upstream/docker"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"

	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/utils"

	"github.com/ksankeerth/open-image-registry/storage"
)

type upstreamInfo struct {
	cacheEnabled bool
	cacheTTL     int
}

type RegistryService struct {
	store           store.Store
	registryName    string
	registryId      string
	namespaceIdMap  sync.Map
	repositoryIdMap sync.Map
	upstream        *upstreamInfo
	client          up.UpstreamClient
}

func NewRegistryService(registryID, registryName string, store store.Store) *RegistryService {

	var upstream upstreamInfo
	var client up.UpstreamClient

	if registryID != HostedRegistryID {
		registryModel, err := store.Upstreams().GetRegistry(context.Background(), registryID)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Registry Service Initialization failed due to database errors")
			return nil
		}

		if registryModel.Vendor != RegistryVendorDockerHub {
			// For now, we only support docker-hub
			// TODO: add support for other upstream registeries
			log.Logger().Warn().Msg("OpenImageRegistry currently supports  DockerHub only")
			return nil
		}

		cacheModel, err := store.Upstreams().GetRegistryCacheConfig(context.Background(), registryID)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Registry Service Initialization failed due to database errors")
			return nil
		}
		upstream.cacheEnabled = cacheModel.CacheEnabled
		upstream.cacheTTL = cacheModel.TTLSeconds

		networkConfig, err := store.Upstreams().GetRegistryNetworkConfig(context.Background(), registryID)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Registry Service Initialization failed due to database errors")
			return nil
		}

		if networkConfig == nil {
			log.Logger().Warn().Str("registry", registryName).
				Msg("Upstream Registry exists without network config")
			return nil
		}

		authConfig, err := store.Upstreams().GetRegistryAuthConfig(context.Background(), registryID)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Registry Service Initialization failed due to database errors")
			return nil
		}

		if authConfig == nil {
			return nil
		}

		var dockerAuthConfig docker.AuthConfig

		err = json.Unmarshal(authConfig.ConfigJSON, &dockerAuthConfig)
		if err != nil {
			log.Logger().Warn().Str("registry", registryName).
				Msg("Upstream Registry exists without auth config")
			return nil
		}

		cfg := docker.Config{}
		cfg.RegistryURL = registryModel.UpstreamURL
		cfg.TokenURL = dockerAuthConfig.TokenEndpoint
		cfg.Username = dockerAuthConfig.Username
		cfg.Password = dockerAuthConfig.Credential

		cfg.ConnectionTimeout = time.Duration(networkConfig.ConnectionTimeout)
		cfg.RequestTimeout = time.Duration(networkConfig.ReadTimeout)
		cfg.MaxConnections = networkConfig.MaxConnections
		cfg.MaxIdleConnections = networkConfig.MaxIdleConnections
		cfg.MaxRetries = networkConfig.MaxRetries
		cfg.RetryBackOffMultiplier = networkConfig.RetryBackOffMultiplier

		client = docker.NewClient(&cfg)
	}

	return &RegistryService{
		registryId:   registryID,
		registryName: registryName,
		store:        store,
		upstream:     &upstream,
		client:       client,
	}
}

func (svc *RegistryService) initiateBlobUpload(reqCtx context.Context, namespace, repository string) (sessionID string,
	err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to initiate blob upload due to database transaction errors")
		return "", err
	}
	ctx := store.WithTxContext(reqCtx, tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	cfg := config.GetImageRegistryConfig()

	// namespace does not exist
	namespaceID, err := svc.getNamespaceID(ctx, namespace)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve namespace from database")
		return "", err
	}
	if namespaceID == "" && cfg.CreateNamespaceOnPush {
		namespaceID, err = svc.store.Namespaces().Create(ctx, svc.registryId, namespace, "", "", false)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Failed to create namespace on image push")
			return "", err
		}
	}

	repositoryID, err := svc.getRepositoryID(ctx, namespace, repository)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve repository from database")
		return "", err
	}
	if repositoryID == "" && cfg.CreateRepositoryOnPush {
		repositoryID, err = svc.store.Repositories().Create(ctx, svc.registryId, namespaceID, repository, "", false)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Failed to create repository on image push")
			return "", err
		}
	}

	if namespaceID == "" || repositoryID == "" {
		return "", nil
	}

	sessionID = uuid.New().String()

	err = svc.store.Blobs().CreateUploadSession(ctx, sessionID, namespace, repository)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to persist image blob upload session")
		return "", err
	}

	return sessionID, nil
}

func (svc *RegistryService) blobExists(reqCtx context.Context, namespace, repository,
	digest string) (bool, error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to check blob existense due to database transaction errors")
		return false, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	exists, _, err := svc.loadImageBlob(ctx, namespace, repository, digest, false)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve image blob due to database errors")
	}
	return exists, err
}

func (svc *RegistryService) verifyNamespaceRepositorySession(ctx context.Context, namespace, repository,
	sessionID string) (bool, *models.ImageBlobUploadSessionModel, error) {
	nsID, repoID, err := svc.getNameSpaceIdAndRepositoryId(ctx, namespace, repository)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to verify namespace and repository from database")
		return false, nil, err
	}
	if nsID == "" || repoID == "" {
		return false, nil, nil
	}

	session, err := svc.store.Blobs().GetUploadSession(ctx, sessionID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to verify blob upload session from database")
		return false, nil, err
	}

	if session == nil {
		return false, nil, nil
	}

	return true, session, nil
}

type blobUploadResult struct {
	invalid       bool // if namespace or repository or session doesn't exist, it will be true
	partialUpload bool // true if partial blob upload detected
}

func (svc *RegistryService) handleLastBlobChunk(reqCtx context.Context, namespace, repository, digest,
	sessionID string) (result *blobUploadResult, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to upload blob due to database transaction errors")
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	result = &blobUploadResult{}

	ok, session, err := svc.verifyNamespaceRepositorySession(ctx, namespace, repository, sessionID)
	if err != nil {
		return nil, err
	}
	if !ok {
		result.invalid = true
		return result, nil
	}

	newLocation := utils.StorageLocation("blobs", svc.registryName, namespace, repository, digest)
	oldLocation := utils.StorageLocation("blobs", svc.registryName, namespace, repository, sessionID)

	size, err := storage.Size(oldLocation)
	if err != nil {
		return nil, err
	}

	if size != int64(session.BytesReceived) {
		log.Logger().Warn().
			Str("namespace", namespace).
			Str("repository", repository).
			Str("blob digest", digest).
			Msg("Partial blob upload detected")
		result.partialUpload = true
		return result, nil
	}

	err = storage.RenameFile(oldLocation, newLocation)
	if err != nil {
		return nil, err
	}

	nsId, repoId, err := svc.getNameSpaceIdAndRepositoryId(ctx, namespace, repository)
	if err != nil {
		return nil, err
	}

	err = svc.store.Blobs().Create(ctx, svc.registryId, nsId, repoId, digest, newLocation, size)
	if err != nil {
		return nil, err
	}

	err = svc.store.Blobs().DeleteUploadSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (svc *RegistryService) uploadBlobChunk(reqCtx context.Context, namespace, repository,
	sessionID string, offset int64, payload []byte) (result *blobUploadResult, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to upload blob due to database transaction errors")
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	result = &blobUploadResult{}

	ok, session, err := svc.verifyNamespaceRepositorySession(ctx, namespace, repository, sessionID)
	if err != nil {
		return nil, err
	}
	if !ok {
		result.invalid = true
		return result, nil
	}

	location := utils.StorageLocation("blobs", svc.registryName, namespace, repository, sessionID)

	if session.BytesReceived != int(offset) {
		log.Logger().Error().Err(err).
			Str("namespace", namespace).
			Str("repository", repository).
			Str("session", sessionID).
			Msg("Corrupted blob upload detected.")
		result.partialUpload = true

		err = storage.DeleteFile(location)
		if err != nil {
			log.Logger().Error().Err(err).Str("location", location).
				Msg("Cleaning corrupted blob failed")
			return nil, err
		}

		err = svc.store.Blobs().DeleteUploadSession(ctx, sessionID)
		if err != nil {
			log.Logger().Error().Err(err).Str("sessionID", sessionID).
				Msg("Cleaning corrupted blob failed")
			return nil, err
		}
		return result, nil
	}

	err = storage.PutFileChunk(location, payload, offset)
	if err != nil {
		return nil, err
	}

	err = svc.store.Blobs().UpdateUploadSession(ctx, sessionID, int(offset)+len(payload))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (svc *RegistryService) uploadBlobWhole(reqCtx context.Context, namespace, repository,
	sessionID, digest string, payload []byte) (result *blobUploadResult, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to upload blob due to database transaction errors")
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	result = &blobUploadResult{}

	ok, _, err := svc.verifyNamespaceRepositorySession(ctx, namespace, repository, sessionID)
	if err != nil {
		return nil, err
	}
	if !ok {
		result.invalid = true
		return result, nil
	}

	location := utils.StorageLocation("blobs", svc.registryName, namespace, repository, digest)

	err = storage.PutFile(location, payload)
	if err != nil {
		return nil, err
	}

	err = svc.store.Blobs().DeleteUploadSession(ctx, sessionID)
	return result, err
}

func (svc *RegistryService) getImageBlob(reqCtx context.Context, namespace, repository,
	digest string) (exists bool, content []byte, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to initiate blob upload due to database transaction errors")
		return false, nil, err
	}
	ctx := store.WithTxContext(reqCtx, tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	return svc.loadImageBlob(ctx, namespace, repository, digest, false)
}

func (svc *RegistryService) pullBlobFromUpstream(_ context.Context, namespace, repository, digest string) (content []byte,
	err error) {
	return svc.client.GetBlob(namespace, repository, digest)
}

func (svc *RegistryService) loadImageBlob(ctx context.Context, namespace, repository,
	digest string, skipContent bool) (exists bool, content []byte, err error) {
	if svc.registryId == HostedRegistryID {
		return svc.loadImageBlobFromRegistry(ctx, namespace, repository, digest, skipContent)
	} else {
		return svc.loadImageBlobFromUpstream(ctx, namespace, repository, digest, skipContent)
	}
}

func (svc *RegistryService) loadImageBlobFromRegistry(ctx context.Context, namespace, repository,
	digest string, skipContent bool) (exists bool, content []byte, err error) {

	repositoryID, err := svc.getRepositoryID(ctx, namespace, repository)
	if err != nil {
		return false, nil, err
	}

	blobMeta, err := svc.store.Blobs().Get(ctx, digest, repositoryID)
	if err != nil {
		return false, nil, err
	}
	if blobMeta == nil {
		return false, nil, nil
	}

	if skipContent {
		return true, nil, nil
	}

	content, err = storage.ReadFile(blobMeta.Location)
	if err != nil {
		return false, nil, err
	}
	if len(content) != blobMeta.Size {
		log.Logger().Warn().Msgf("Size mismatch for blob %s; Size in storage: %d, Size in database: %d",
			blobMeta.Location, len(content), blobMeta.Size)
	}
	return true, content, nil
}

func (svc *RegistryService) loadImageBlobFromUpstream(ctx context.Context, namespace, repository,
	digest string, skipContent bool) (exists bool, content []byte, err error) {

	if svc.upstream.cacheEnabled {
		repositoryID, err := svc.getRepositoryID(ctx, namespace, repository)
		if err != nil {
			return false, nil, err
		}

		blobMeta, err := svc.store.Blobs().Get(ctx, digest, repositoryID)
		if err != nil {
			return false, nil, err
		}

		if blobMeta == nil {
			// not found in cache, load from cache and store it in cache
			content, err = svc.pullBlobFromUpstream(ctx, namespace, repository, digest)
			if err != nil {
				return false, nil, err
			}

			err = svc.cacheBlob(ctx, namespace, repository, digest, content)
			if err != nil {
				return false, nil, err
			}
			return true, content, nil
		}

		if skipContent {
			return true, nil, nil
		}

		content, err = storage.ReadFile(blobMeta.Location)
		if err != nil {
			return false, nil, err
		}

		if len(content) != blobMeta.Size {
			log.Logger().Warn().Msgf("Size mismatch for %s; Size in storage: %d, Size in database: %d",
				blobMeta.Location, len(content), blobMeta.Size)
		}
		return true, content, nil
	} else {
		if skipContent {
			exists, err = svc.client.HeadBlob(namespace, repository, digest)
			if err != nil {
				return false, nil, err
			}
			return true, nil, nil
		} else {
			content, err = svc.client.GetBlob(namespace, repository, digest)
			if err != nil {
				return false, nil, err
			}
			return true, content, nil
		}
	}
}

// cacheBlob stores the blob fetched from upstream registry and add meta information in database
func (svc *RegistryService) cacheBlob(ctx context.Context, namespace, repository, digest string,
	payload []byte) error {
	location := utils.StorageLocation("blobs", svc.registryName, namespace, repository, digest)
	err := svc.storeImageBlob(ctx, location, false, 0, payload)
	if err != nil {
		return err
	}

	nsId, repoId, err := svc.getNameSpaceIdAndRepositoryId(ctx, namespace, repository)
	if err != nil {
		return err
	}

	err = svc.store.Blobs().Create(ctx, svc.registryId, nsId, repoId, digest, location, int64(len(payload)))
	if err != nil {
		return err
	}
	return nil
}

func (svc *RegistryService) storeImageBlob(ctx context.Context, storageLocation string, isChunked bool,
	offset int64, payload []byte) error {
	var err error
	if isChunked {
		err = storage.PutFileChunk(storageLocation, payload, offset)
	} else {
		err = storage.PutFile(storageLocation, payload)
	}
	return err
}

func (svc *RegistryService) getNameSpaceIdAndRepositoryId(ctx context.Context, namespace,
	repository string) (string, string, error) {
	nsId, err := svc.getNamespaceID(ctx, namespace)
	if err != nil {
		return "", "", err
	}

	repositoryId, err := svc.getRepositoryID(ctx, namespace, repository)
	if err != nil {
		return "", "", err
	}
	return nsId, repositoryId, nil
}

func (svc *RegistryService) getNamespaceID(ctx context.Context, namespace string) (string, error) {
	val, ok := svc.namespaceIdMap.Load(namespace)
	if ok {
		return val.(string), nil
	}

	nsId, err := svc.store.Namespaces().GetID(ctx, svc.registryId, namespace)
	if err != nil {
		return "", err
	}

	svc.namespaceIdMap.Store(namespace, nsId)
	return nsId, nil
}

func (svc *RegistryService) getRepositoryID(ctx context.Context, namespace, repository string) (string, error) {
	key := fmt.Sprintf("%s/%s", namespace, repository)

	val, ok := svc.repositoryIdMap.Load(key)
	if ok {
		return val.(string), nil
	}

	nsId, err := svc.getNamespaceID(ctx, namespace)
	if err != nil {
		return "", err
	}

	repositoryId, err := svc.store.Repositories().GetID(ctx, nsId, repository)
	if err != nil {
		return "", err
	}
	svc.repositoryIdMap.Store(key, repositoryId)

	return repositoryId, nil
}

func (svc *RegistryService) getImageManifest(reqCtx context.Context, namespace, repository,
	tagOrDigest string) (exists bool, mediaType, digest string, content []byte, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve manifest due to database transaction errors")
		return false, "", "", nil, err
	}
	ctx := store.WithTxContext(reqCtx, tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	exists, mediaType, digest, content, err = svc.loadImageManifest(ctx, namespace, repository,
		tagOrDigest, false)
	return
}

func (svc *RegistryService) manifestExists(reqCtx context.Context, namespace, repository,
	tagOrDigest string) (exists bool, mediaType, digest string, err error) {

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to check manifest due to database transaction errors")
		return false, "", "", err
	}
	ctx := store.WithTxContext(reqCtx, tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	exists, mediaType, digest, _, err = svc.loadImageManifest(ctx, namespace, repository, tagOrDigest, true)
	return
}

func (svc *RegistryService) loadImageManifest(ctx context.Context, namespace, repository,
	tagOrDigest string, skipContent bool) (exists bool, mediaType, digest string, content []byte, err error) {
	if svc.registryId == HostedRegistryID {
		if utils.IsImageDigest(tagOrDigest) {
			exists, content, mediaType, err = svc.loadManifestByDigest(ctx, namespace, repository, tagOrDigest,
				skipContent)

		} else {
			exists, digest, mediaType, content, err = svc.loadManifestByTag(ctx, namespace, repository, tagOrDigest,
				skipContent)
		}

		if !exists {
			log.Logger().Warn().Msgf("No manifest found for: %s/%s:%s", namespace, repository, tagOrDigest)
			err = nil
		}

		return exists, mediaType, digest, content, err
	} else {
		if svc.upstream.cacheEnabled {
			exists, digest, mediaType, content, err = svc.loadManifestFromCache(ctx, namespace, repository,
				tagOrDigest, skipContent)
			if err != nil {
				return false, "", "", nil, err
			}
			if exists {
				return true, mediaType, digest, content, nil
			}
			content, mediaType, err := svc.client.GetManifest(namespace, repository, tagOrDigest)
			if err != nil {
				return false, "", "", nil, err
			}

			if utils.IsImageDigest(tagOrDigest) {
				digest = tagOrDigest
			} else {
				digest = utils.CalcuateDigest(content)
			}
			err = svc.cacheManifest(ctx, namespace, repository, tagOrDigest, digest, mediaType, content)
			if err != nil {
				return false, "", "", nil, err
			}

			return true, mediaType, digest, content, nil
		} else {
			if skipContent {
				exists, err = svc.client.HeadManifest(namespace, repository, tagOrDigest)
			} else {
				content, mediaType, err = svc.client.GetManifest(namespace, repository, tagOrDigest)
				exists = true
			}
			if err != nil {
				return false, "", "", nil, err
			}
			return exists, mediaType, utils.CalcuateDigest(content), content, nil
		}
	}
}

func (svc *RegistryService) loadManifestFromCache(ctx context.Context, namsespace, repository,
	tagOrDigest string,
	skipContent bool) (exists bool, digest, mediaType string, content []byte, err error) {
	_, repoId, err := svc.getNameSpaceIdAndRepositoryId(ctx, namsespace, repository)
	if err != nil {
		return false, "", "", nil, err
	}

	cacheModel, err := svc.store.Cache().Get(ctx, repoId, tagOrDigest)
	if err != nil {
		return false, "", "", nil, err
	}

	if cacheModel == nil {
		return false, "", "", nil, nil
	}

	// TODO stable tag
	if cacheModel.ExpiresAt.Before(time.Now()) {
		// We won't delete the cache entry. Even if the time expired, manifest may not be changed in upstream
		// After retriving manifest from upstream proxy, we'll check digest values. if they are same, We'll
		// refresh the cache entry instead of deleting and adding again.
		return false, "", "", nil, nil
	}

	if utils.IsImageDigest(tagOrDigest) {
		exists, content, mediaType, err = svc.loadManifestByDigest(ctx, namsespace, repository, tagOrDigest, skipContent)
	} else {
		exists, _, mediaType, content, err = svc.loadManifestByTag(ctx, namsespace, repository, tagOrDigest, skipContent)
	}

	if !exists {
		log.Logger().Warn().Msgf("Cache reference for manifest: (%s/%s/%s@%s) exists but actual manifest is not available in the database",
			svc.registryName, namsespace, repository, digest)

		err = svc.store.Cache().Delete(ctx, repoId, tagOrDigest)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when cleaning cache manifest referece: (%s/%s/%s@%s)", svc.registryName, namsespace, repository, digest)
			return false, "", "", nil, err
		}
	}

	return
}

// cacheManifest stores a  manifest reference in cache table and actual manifest will be stored
func (svc *RegistryService) cacheManifest(ctx context.Context, namespace, repository, identifier, digest,
	mediaType string, content []byte) error {

	validTill := time.Now().Add(time.Duration(svc.upstream.cacheTTL) * time.Second)
	nsId, repositoryId, err := svc.getNameSpaceIdAndRepositoryId(ctx, namespace, repository)
	if err != nil {
		return err
	}

	cacheEntry, err := svc.store.Cache().Get(ctx, repositoryId, digest)
	if err != nil {
		return err
	}

	var cacheMiss bool
	if cacheEntry == nil { // no cache entry
		err = svc.store.Cache().Create(ctx, svc.registryId, nsId, repositoryId, identifier, digest, validTill)
		cacheMiss = true
	} else { // cache entry exists
		if cacheEntry.Digest == digest { // digest is same
			err = svc.store.Cache().Refresh(ctx, repositoryId, identifier, validTill)
		} else { // digest has changed. so let's add new manifest, identifier is a tag
			cacheMiss = true
			err = svc.store.Cache().Delete(ctx, repositoryId, identifier)
			if err != nil {
				return err
			}
			err = svc.store.Manifests().DeleteByDigest(ctx, repositoryId, digest)
			if err != nil {
				return err
			}
			err = svc.store.Cache().Create(ctx, svc.registryId, nsId, repositoryId, identifier, digest, validTill)
		}
	}

	if err != nil {
		return err
	}

	// for upstream manifests, unique-digest = digest
	if cacheMiss {
		manifestID, err := svc.store.Manifests().Create(ctx, svc.registryId, nsId, repositoryId, digest, mediaType, digest,
			int64(len(content)), content)
		if err != nil {
			return err
		}

		// if identifier is tag, link the tag and manifest
		if !utils.IsImageDigest(identifier) {
			tag, err := svc.store.Tags().Get(ctx, repositoryId, identifier)
			if err != nil {
				return err
			}

			if tag == nil {
				tagID, err := svc.store.Tags().Create(ctx, svc.registryId, nsId, repositoryId, identifier)
				if err != nil {
					return err
				}

				err = svc.store.Tags().LinkManifest(ctx, tagID, manifestID)
				if err != nil {
					return err
				}
			} else {
				err = svc.store.Tags().UnlinkManifest(ctx, tag.Id)
				if err != nil {
					return err
				}
				err = svc.store.Tags().LinkManifest(ctx, tag.Id, manifestID)
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func (svc *RegistryService) loadManifestByTag(ctx context.Context, namespace, repository, tag string,
	skipContent bool) (exists bool, digest, mediaType string, content []byte, err error) {
	_, repositoryId, err := svc.getNameSpaceIdAndRepositoryId(ctx, namespace, repository)
	if err != nil {
		return false, "", "", nil, err
	}

	manifest, err := svc.store.ImageQueries().GetManifestByTag(ctx, !skipContent, repositoryId, tag)
	if err != nil {
		return false, "", "", nil, err
	}
	if manifest == nil {
		return false, "", "", nil, nil
	}

	return true, manifest.Digest, manifest.MediaType, []byte(manifest.Content), nil
}

func (svc *RegistryService) loadManifestByDigest(ctx context.Context, namespace, repository, digest string,
	skipContent bool) (exists bool, content []byte, mediaType string, err error) {
	_, repositoryId, err := svc.getNameSpaceIdAndRepositoryId(ctx, namespace, repository)
	if err != nil {
		return false, nil, "", err
	}

	manifest, err := svc.store.Manifests().GetByDigest(ctx, !skipContent, repositoryId, digest)
	if err != nil {
		return false, nil, "", err
	}
	if manifest == nil {
		return false, nil, "", nil
	}

	return true, []byte(manifest.Content), manifest.MediaType, nil
}

type ManifestScanResult struct {
	TagExists              bool
	ManifestExists         bool
	TagManifestLinkExists  bool
	TagManifestLinkChanged bool
	UniqueDigest           string
	TagId                  string
	ManifestId             string
	NamespaceId            string
	RepositoryId           string
}

func (svc *RegistryService) updateManifest(reqCtx context.Context, namespace, repository, tag,
	mediaType string, content []byte) (digest string, err error) {

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to update manifest due to database transaction errors")
		return "", err
	}
	ctx := store.WithTxContext(reqCtx, tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	uniqueDigest, err := UniqueDigest(mediaType, content)
	if err != nil {
		return "", err
	}

	res, err := svc.scanManifest(ctx, uniqueDigest, namespace, repository, tag)
	if err != nil {
		return "", err
	}

	if !res.TagExists {
		tagId, err := svc.store.Tags().Create(ctx, svc.registryId, res.NamespaceId, res.RepositoryId, tag)
		if err != nil {
			return "", err
		}
		res.TagId = tagId
	}

	manifestDigest := utils.CalcuateDigest(content)

	if !res.ManifestExists {
		manifestId, err := svc.store.Manifests().Create(ctx, svc.registryId, res.NamespaceId, res.RepositoryId,
			manifestDigest, mediaType, res.UniqueDigest, int64(len(content)), content)
		if err != nil {
			return "", err
		}
		res.ManifestId = manifestId

		svc.store.Tags().LinkManifest(ctx, res.TagId, res.ManifestId)
		if err != nil {
			return "", err
		}
	}

	if res.TagManifestLinkChanged {
		err = svc.store.Tags().UpdateManifest(ctx, res.TagId, res.ManifestId)
		if err != nil {
			return "", err
		}
	}
	return manifestDigest, nil
}

func (svc *RegistryService) scanManifest(ctx context.Context, uniqueDigest, namespace, repository,
	tag string) (*ManifestScanResult, error) {

	nsId, repoId, err := svc.getNameSpaceIdAndRepositoryId(ctx, namespace, repository)
	if err != nil {
		return nil, err
	}

	// Check manifest existence
	manifestModel, err := svc.store.Manifests().GetByUniqueDigest(ctx, false, repoId, uniqueDigest)
	if err != nil {
		return nil, err
	}

	if manifestModel == nil {
		return &ManifestScanResult{ManifestExists: false}, nil
	}

	result := &ManifestScanResult{
		ManifestExists: true,
		UniqueDigest:   uniqueDigest,
		ManifestId:     manifestModel.ID,
		NamespaceId:    nsId,
		RepositoryId:   repoId,
	}

	// Check tag existence
	tagModel, err := svc.store.Tags().Get(ctx, repoId, tag)
	if err != nil {
		return nil, err
	}

	if tagModel == nil {
		result.TagExists = false
	} else {
		result.TagExists = true
		result.TagId = tagModel.Id
		// Check tag->manifest link
		oldManifestId, err := svc.store.Tags().GetManifestID(ctx, tagModel.Id)
		if err != nil {
			return nil, err
		}
		if oldManifestId == "" {
			result.TagManifestLinkExists = false
		}
	}

	return result, nil
}
