package registry

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ksankeerth/open-image-registry/client"
	client_errors "github.com/ksankeerth/open-image-registry/errors/client"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/dockerv2"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/oci"
	"github.com/ksankeerth/open-image-registry/types/models"

	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/utils"

	"github.com/ksankeerth/open-image-registry/db"
	"github.com/ksankeerth/open-image-registry/storage"
)

type RegistryService struct {
	registryDao      db.ImageRegistryDAO
	upstreamDao      db.UpstreamDAO
	registryName     string
	registryId       string
	namespaceIdMap   sync.Map
	repositoryIdMap  sync.Map
	upstreamRegistry *UpstreamRegistry
}

type UpstreamRegistry struct {
	reg           *models.UpstreamRegistryEntity
	cacheConfig   *models.UpstreamRegistryCacheConfig
	authConfig    *models.UpstreamRegistryAuthConfig
	accessConfig  *models.UpstreamRegistryAccessConfig
	storageConfig *models.UpstreamRegistryStorageConfig
}

func NewRegistryService(registryId, registryName string,
	registryDao db.ImageRegistryDAO, upstreamDAO db.UpstreamDAO) *RegistryService {
	if registryDao == nil {
		return nil
	}
	return &RegistryService{
		registryId:   registryId,
		registryName: registryName,
		registryDao:  registryDao,
		upstreamDao:  upstreamDAO,
	}
}

func (svc *RegistryService) loadUpstreamConfig() error {
	if svc.registryId == HostedRegistryID {
		return nil
	}

	reg, access, auth, cache, storage, err := svc.upstreamDao.GetUpstreamRegistryWithConfig(svc.registryId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when loading upstream registry")
		return err
	}
	svc.upstreamRegistry = &UpstreamRegistry{
		reg:           reg,
		authConfig:    auth,
		accessConfig:  access,
		cacheConfig:   cache,
		storageConfig: storage,
	}
	return nil
}

func (svc *RegistryService) getImageBlob(namespace, repository,
	digest string) (exists bool, content []byte, err error) {
	return svc.loadImageBlob(namespace, repository, digest, false)
}

func (svc *RegistryService) isImageBlobPresent(namespace, repository, digest string) (bool, error) {
	exits, _, err := svc.loadImageBlob(namespace, repository, digest, false)
	return exits, err
}

func (svc *RegistryService) loadImageBlob(namespace, repository,
	digest string, skipContent bool) (exists bool, content []byte, err error) {
	if svc.registryId == HostedRegistryID {
		storageLocation, size, err := svc.registryDao.GetImageBlobStorageLocationAndSize(digest, namespace, repository, svc.registryId, "")

		if err != nil {
			if db_errors.IsNotFound(err) {
				return false, nil, nil
			}
			log.Logger().Error().Err(err).Msgf("Unable to check Image blob existense for %s:%s:%s@%s from database", namespace, repository, digest)
			return false, nil, err
		}

		if skipContent {
			return true, nil, nil
		}

		content, err = storage.ReadFile(storageLocation)
		if err != nil {
			return false, nil, err
		}
		if len(content) != size {
			log.Logger().Warn().Msgf("Size mismatch for %s; Size in storage: %d, Size in database: %d", storageLocation, len(content), size)
		}
		return true, content, nil
	} else {
		var cacheMiss = false
		if svc.upstreamRegistry.cacheConfig.Enabled {
			storageLocation, size, err := svc.registryDao.GetImageBlobStorageLocationAndSize(digest, namespace, repository, svc.registryId, "")
			if db_errors.IsNotFound(err) {
				cacheMiss = true
			} else {
				if err != nil {
					log.Logger().Error().Err(err).Msgf("Unable to check Image blob existense for %s:%s:%s@%s from database", namespace, repository, digest)
					return false, nil, err
				}

				if skipContent {
					return true, nil, nil
				}

				content, err = storage.ReadFile(storageLocation)
				if err != nil {
					return false, nil, err
				}
				if len(content) != size {
					log.Logger().Warn().Msgf("Size mismatch for %s; Size in storage: %d, Size in database: %d", storageLocation, len(content), size)
				}
				return true, content, nil
			}
		}

		dockerClient := client.NewDockerV2RegistryClient(svc.upstreamRegistry.reg.UpstreamUrl, svc.upstreamRegistry.authConfig.TokenEndpoint)
		dockerClient.SetWireLogging(true)
		if cacheMiss || !skipContent {
			content, err := dockerClient.GetBlob(namespace, repository, digest)
			if err != nil {
				if client_errors.IsNotFound(err) {
					return false, nil, nil
				} else {
					return false, nil, err
				}
			}
			if cacheMiss {
				err = svc.persistImageBlob(namespace, repository, digest, content)
				if err != nil {
					log.Logger().Error().Err(err).Msgf("Error occured when caching image blob(%s:%s:%s@%s) locally", svc.registryName, namespace, repository, digest)
				}
			}
			return true, content, nil
		} else {
			exists, err = dockerClient.HeadBlob(namespace, repository, digest)
			return exists, nil, err
		}
	}
}

func (svc *RegistryService) createNamespaceAndRepositoryIfNotExist(namespace, repository string) error {
	_, _, err := svc.registryDao.CreateNamespaceAndRepositoryIfNotExist(svc.registryId, namespace, repository, "")
	return err
}

func (svc *RegistryService) persistImageBlob(namespace, repository, digest string, payload []byte) error {
	txKey := fmt.Sprintf("blob-update-%s-%s/%s@%s", svc.registryId, namespace, repository, digest)
	err := svc.registryDao.Begin(txKey)
	if err != nil {
		return err
	}

	var dbErr error

	defer func() {
		if dbErr != nil {
			svc.registryDao.Rollback(txKey)
		} else {
			svc.registryDao.Commit(txKey)
		}
	}()

	blobNotFound := true
	_, _, dbErr = svc.registryDao.GetImageBlobStorageLocationAndSize(digest, namespace, repository, svc.registryId, txKey)
	if dbErr != nil {
		if db_errors.IsNotFound(dbErr) {
			blobNotFound = true
		} else {
			return dbErr
		}
	} else {
		blobNotFound = false
	}

	if !blobNotFound {
		return nil
	}

	storageLocation := utils.StorageLocation("blobs", svc.registryName, namespace, repository, digest)
	err = svc.storeImageBlob(storageLocation, false, 0, payload)
	if err != nil {
		return err
	}

	nsId, repoId, err := svc.getNameSpaceIdAndRepositoryId(namespace, repository)
	if err != nil {
		return err
	}

	dbErr = svc.registryDao.PersistImageBlobMetaInfo(svc.registryId, nsId, repoId, digest,
		storageLocation, int64(len(payload)), txKey)

	return dbErr
}

func (svc *RegistryService) persistImageBlobFinalChunk(namespace, repository, digest,
	sessionId string) error {
	newLocation := utils.StorageLocation("blobs", svc.registryName, namespace, repository, digest)
	oldLocation := utils.StorageLocation("blobs", svc.registryName, namespace, repository, sessionId)

	err := storage.RenameFile(oldLocation, newLocation)
	if err != nil {
		return err
	}

	size, err := storage.Size(newLocation)

	nsId, repoId, err := svc.getNameSpaceIdAndRepositoryId(namespace, repository)
	if err != nil {
		return err
	}

	err = svc.registryDao.PersistImageBlobMetaInfo(svc.registryId, nsId, repoId, digest,
		newLocation, size, "")
	return err
}

func (svc *RegistryService) storeImageBlob(storageLocation string, isChunked bool,
	offset int64, payload []byte) error {
	var err error
	if isChunked {
		err = storage.PutFileChunk(storageLocation, payload, offset)
	} else {
		err = storage.PutFile(storageLocation, payload)
	}
	return err
}

func (svc *RegistryService) getNameSpaceIdAndRepositoryId(namespace, repository string) (string, string, error) {
	nsId, err := svc.getNamespaceID(namespace)
	if err != nil {
		return "", "", err
	}

	repositoryId, err := svc.getRepositoryID(namespace, repository)
	if err != nil {
		return "", "", err
	}
	return nsId, repositoryId, nil
}

func (svc *RegistryService) getNamespaceID(namespace string) (string, error) {
	val, ok := svc.namespaceIdMap.Load(namespace)
	if ok {
		return val.(string), nil
	}

	nsId, err := svc.registryDao.GetNamespaceId(svc.registryId, namespace, "")
	if err != nil {
		return "", err
	}
	svc.namespaceIdMap.Store(namespace, nsId)
	return nsId, nil
}

func (svc *RegistryService) getRepositoryID(namespace, repository string) (string, error) {
	key := fmt.Sprintf("%s/%s", namespace, repository)

	val, ok := svc.repositoryIdMap.Load(key)
	if ok {
		return val.(string), nil
	}

	nsId, err := svc.getNamespaceID(namespace)
	if err != nil {
		return "", err
	}

	repositoryId, err := svc.registryDao.GetRepositoryId(svc.registryId, nsId, repository, "")
	if err != nil {
		return "", err
	}
	svc.repositoryIdMap.Store(key, repositoryId)

	return repositoryId, nil
}

func (svc *RegistryService) getImageManifest(namespace, repository, tagOrDigest string) (exists bool,
	mediaType, digest string, content []byte, err error) {
	exists, mediaType, digest, content, err = svc.loadImageManifest(namespace, repository, tagOrDigest, false)
	return
}

func (svc *RegistryService) isImageManifestPresent(namespace, repository,
	tagOrDigest string) (exists bool, mediaType, digest string, err error) {

	exists, mediaType, digest, _, err = svc.loadImageManifest(namespace, repository, tagOrDigest, true)
	return
}

func (svc *RegistryService) loadImageManifest(namespace, repository,
	tagOrDigest string, skipContent bool) (exists bool, mediaType, digest string, content []byte, err error) {
	if svc.registryId == HostedRegistryID {
		if utils.IsImageDigest(tagOrDigest) {
			exists, content, mediaType, err = svc.loadImageManifestFromDbByDigest(namespace, repository, tagOrDigest,
				skipContent, "")

		} else {
			exists, digest, mediaType, content, err = svc.loadImageManifestFromDbByTag(namespace, repository, tagOrDigest,
				skipContent, "")
		}

		if !exists {
			log.Logger().Warn().Msgf("No manifest found for: %s/%s:%s", namespace, repository, tagOrDigest)
			err = nil
		}

		return exists, mediaType, digest, content, err
	} else {
		var cacheMiss = false

		// load upstream config on demand
		if svc.upstreamRegistry == nil {
			err = svc.loadUpstreamConfig()
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to handle request. Loading Upstream registry configuration is failed")
				return
			}
		}
		if svc.upstreamRegistry.cacheConfig.Enabled {
			exists, digest, mediaType, content, err = svc.loadManifestFromCache(namespace, repository, tagOrDigest, skipContent)
			if exists {
				return
			}
			cacheMiss = true
		}

		if cacheMiss {

			// if it is cache miss, we'll always download manifest from registry and cache locally regardless of `skipContent`
			exists, digest, mediaType, content, err = svc.loadImageManifestFromRegistry(namespace, repository, tagOrDigest, false)
			if err != nil {
				return
			}
			if exists {
				if utils.IsImageDigest(tagOrDigest) {
					err = svc.cacheManifest(namespace, repository, digest, mediaType, content)
				} else {
					err = svc.cacheManifestWithTag(namespace, repository, tagOrDigest, digest, mediaType, content)
				}
				if err != nil {
					log.Logger().Error().Err(err).Msgf("Error occured caching image manifest: (%s/%s/%s@%s)", svc.registryName, namespace, repository, digest)
				}
			}
			return
		}

		exists, digest, mediaType, content, err = svc.loadImageManifestFromRegistry(namespace, repository,
			tagOrDigest, skipContent)
		if !exists {
			log.Logger().Warn().Msgf("No manifest found for: %s/%s/%s:%s", svc.registryName, namespace, repository,
				tagOrDigest)
			err = nil
		}

		return
	}
}

func (svc *RegistryService) loadManifestFromCache(namsespace, repository, tagOrDigest string,
	skipContent bool) (exists bool, digest, mediaType string, content []byte, err error) {
	nsId, repoId, err := svc.getNameSpaceIdAndRepositoryId(namsespace, repository)
	if err != nil {
		return false, "", "", nil, err
	}

	txKey := fmt.Sprintf("load-manifest-from-cache-%s/%s/%s/%s", svc.registryId, nsId, repoId, tagOrDigest)
	err = svc.registryDao.Begin(txKey)
	if err != nil {
		return false, "", "", nil, err
	}

	defer func() {
		if err != nil {
			svc.registryDao.Rollback(txKey)
		} else {
			svc.registryDao.Commit(txKey)
		}
	}()

	cacheMiss, digest, expiresAt, err := svc.registryDao.GetCacheImageManifestReference(svc.registryId, nsId,
		repoId, tagOrDigest, txKey)
	if err != nil {
		return false, "", "", nil, err
	}
	if cacheMiss {
		return false, "", "", nil, nil
	}
	if expiresAt.Before(time.Now()) {
		// We won't delete the cache entry. Even if the time expired, manifest may not be changed in upstream
		// After retriving manifest from upstream proxy, we'll check digest values. if they are same, We'll
		// refresh the cache entry instead of deleting and adding again.
		return false, "", "", nil, nil
	}

	if utils.IsImageDigest(tagOrDigest) {
		exists, content, mediaType, err = svc.loadImageManifestFromDbByDigest(namsespace, repository, tagOrDigest, skipContent, txKey)
	} else {
		exists, _, mediaType, content, err = svc.loadImageManifestFromDbByTag(namsespace, repository, tagOrDigest, skipContent, txKey)
	}

	if !exists {
		log.Logger().Warn().Msgf("Cache reference for manifest: (%s/%s/%s@%s) exists but actual manifest is not available in the database",
			svc.registryName, namsespace, repository, digest)
		err = svc.registryDao.DeleteCacheImageManifestReference(svc.registryId, nsId, repoId, tagOrDigest, txKey)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when cleaning cache manifest referece: (%s/%s/%s@%s)", svc.registryName, namsespace, repository, digest)
		}
	}

	return
}

func (svc *RegistryService) cacheManifestWithTag(namespace, repository, tag, digest, mediaType string, content []byte) error {
	txKey := fmt.Sprintf("cache-manifest-%s:%s:%s:%s", svc.registryName, namespace, repository, tag)
	err := svc.registryDao.Begin(txKey)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			svc.registryDao.Rollback(txKey)
		} else {
			svc.registryDao.Commit(txKey)
		}
	}()

	validTill := time.Now().Add(time.Duration(svc.upstreamRegistry.cacheConfig.TtlInSeconds * time.Now().Second()))

	nsId, repositoryId, err := svc.getNameSpaceIdAndRepositoryId(namespace, repository)
	if err != nil {
		return err
	}

	cacheMiss, oldDigest, _, err := svc.registryDao.GetCacheImageManifestReference(svc.registryId, nsId, repositoryId, tag, txKey)
	if err != nil {
		return err
	}
	var changed = false
	if !cacheMiss {
		if oldDigest == digest {
			err = svc.registryDao.RefreshCacheImageManifestReference(svc.registryId, nsId, repositoryId, tag, validTill, txKey)
			if err != nil {
				return err
			}
			err = svc.registryDao.RefreshCacheImageManifestReference(svc.registryId, nsId, repositoryId, digest, validTill, txKey)
			if err != nil {
				return err
			}
		} else {
			changed = true
		}
	}

	if changed {
		err = svc.registryDao.DeleteCacheImageManifestReference(svc.registryId, nsId, repositoryId, tag, txKey)
		if err != nil {
			return err
		}
		err = svc.registryDao.DeleteCacheImageManifestReference(svc.registryId, nsId, repositoryId, oldDigest, txKey)
		if err != nil {
			return err
		}
	}

	err = svc.registryDao.CacheImageManifestReference(svc.registryId, nsId, repositoryId, tag, validTill, txKey)
	if err != nil {
		return err
	}

	err = svc.registryDao.CacheImageManifestReference(svc.registryId, nsId, repositoryId, digest, validTill, txKey)
	if err != nil {
		return err
	}

	// for upstream manifests, unique-digest = digest
	var manifestId string
	manifestId, err = svc.registryDao.PersistImageManifest(svc.registryId, nsId, repositoryId, digest, mediaType,
		digest, int64(len(content)), content, txKey)
	if err != nil {
		return err
	}

	tagId, err := svc.registryDao.GetImageTagId(repositoryId, nsId, repositoryId, tag, txKey)
	if db_errors.IsNotFound(err) {
		tagId, err = svc.registryDao.CreateImageTag(svc.registryId, nsId, repositoryId, tag, txKey)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	err = svc.registryDao.LinkImageManifestWithTag(tagId, manifestId, txKey)
	return err
}

func (svc *RegistryService) cacheManifest(namespace, repository, digest, mediaType string, content []byte) error {
	txKey := fmt.Sprintf("cache-manifest-%s:%s:%s:%s", svc.registryName, namespace, repository, digest)
	err := svc.registryDao.Begin(txKey)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			svc.registryDao.Rollback(txKey)
		} else {
			svc.registryDao.Commit(txKey)
		}
	}()

	validTill := time.Now().Add(time.Duration(svc.upstreamRegistry.cacheConfig.TtlInSeconds * time.Now().Second()))

	nsId, repositoryId, err := svc.getNameSpaceIdAndRepositoryId(namespace, repository)
	if err != nil {
		return err
	}

	cacheMiss, _, _, err := svc.registryDao.GetCacheImageManifestReference(svc.registryId, nsId, repositoryId, digest, txKey)
	if err != nil {
		return err
	}
	if !cacheMiss {
		svc.registryDao.RefreshCacheImageManifestReference(svc.registryId, nsId, repositoryId, digest, validTill, txKey)
		// no need to update the existing manifest
		return nil
	}

	err = svc.registryDao.CacheImageManifestReference(svc.registryId, nsId, repositoryId, digest, validTill, txKey)
	if err != nil {
		return err
	}

	// for upstream manifests, unique-digest = digest
	_, err = svc.registryDao.PersistImageManifest(svc.registryId, nsId, repositoryId, digest, mediaType,
		digest, int64(len(content)), content, txKey)
	return err

}

func (svc *RegistryService) loadImageManifestFromDbByTag(namespace, repository, tag string,
	skipContent bool, txKey string) (exists bool, digest, mediaType string, content []byte, err error) {
	nsId, repositoryId, err := svc.getNameSpaceIdAndRepositoryId(namespace, repository)
	if err != nil {
		return false, "", "", nil, err
	}

	if skipContent {
		exists, digest, mediaType, err = svc.registryDao.CheckImageManifestExistsByTag(svc.registryId, nsId, repositoryId, tag, txKey)
	} else {
		exists, content, digest, mediaType, err = svc.registryDao.GetImageManifestByTag(svc.registryId, nsId, repositoryId, tag, txKey)
	}
	return
}

func (svc *RegistryService) loadImageManifestFromDbByDigest(namespace, repository, digest string,
	skipContent bool, txKey string) (exists bool, content []byte, mediaType string, err error) {
	nsId, repositoryId, err := svc.getNameSpaceIdAndRepositoryId(namespace, repository)
	if err != nil {
		return false, nil, "", err
	}

	if skipContent {
		exists, mediaType, err = svc.registryDao.CheckImageManifestExistsByDigest(svc.registryId, nsId, repositoryId, digest, txKey)
	} else {
		exists, content, mediaType, err = svc.registryDao.GetImageManifestByDigest(svc.registryId, nsId, repositoryId, digest, txKey)
	}
	return
}

func (svc *RegistryService) loadImageManifestFromRegistry(namespace, repository, tagOrDigest string,
	skipContent bool) (exists bool, digest, mediaType string, content []byte, err error) {

	dockerClient := client.NewDockerV2RegistryClient(svc.upstreamRegistry.reg.UpstreamUrl, svc.upstreamRegistry.authConfig.TokenEndpoint)
	dockerClient.SetWireLogging(true)

	if skipContent {
		exists, digest, mediaType, err = dockerClient.HeadManifest(namespace, repository, tagOrDigest)
	} else {
		content, mediaType, err = dockerClient.GetManifest(namespace, repository, tagOrDigest)
		if client_errors.IsNotFound(err) {
			exists = false
		} else if err != nil {
			return false, "", "", nil, err
		}
		exists = true
	}
	return
}

type ManifestCheckResult struct {
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

func (svc *RegistryService) updateManifest(namespace, repository, tag,
	mediaType string, content []byte) (digest string, err error) {

	txKey := fmt.Sprintf("update-manifest-%s/%s/%s/%s", svc.registryName, namespace, repository, tag)
	err = svc.registryDao.Begin(txKey)
	if err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			svc.registryDao.Rollback(txKey)
		} else {
			svc.registryDao.Commit(txKey)
		}
	}()

	res, err := svc.checkManifestInDatabase(namespace, repository, tag, mediaType, content, txKey)
	if err != nil {
		return "", err
	}

	if !res.TagExists {
		tagId, err := svc.registryDao.CreateImageTag(svc.registryId, res.NamespaceId, res.RepositoryId, tag, txKey)
		if err != nil {
			return "", err
		}
		res.TagId = tagId
	}
	manifestDigest := utils.CalcuateDigest(content)

	if !res.ManifestExists {
		manifestId, err := svc.registryDao.PersistImageManifest(svc.registryId, res.NamespaceId, res.RepositoryId,
			manifestDigest, mediaType, res.UniqueDigest, int64(len(content)), content, txKey)
		if err != nil {
			return "", err
		}
		res.ManifestId = manifestId

		err = svc.registryDao.LinkImageManifestWithTag(res.TagId, res.ManifestId, txKey)
		if err != nil {
			return "", err
		}
	}

	if res.TagManifestLinkChanged {
		err = svc.registryDao.UpdateManifestIdForTag(res.TagId, res.ManifestId, txKey)
		if err != nil {
			return "", err
		}
	}
	return manifestDigest, nil
}

func (svc *RegistryService) checkManifestInDatabase(
	namespace, repository, tag, mediaType string, content []byte, txKey string,
) (*ManifestCheckResult, error) {

	var uniqueDigest string

	switch mediaType {
	case "application/vnd.docker.distribution.manifest.v2+json":
		var manifest dockerv2.ManifestV2
		if err := json.Unmarshal(content, &manifest); err != nil {
			return nil, err
		}
		inputs := []string{manifest.Config.Digest}
		for _, entry := range manifest.Layers {
			inputs = append(inputs, entry.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(inputs...)

	case "application/vnd.docker.distribution.manifest.list.v2+json":
		var manifest dockerv2.ManifestListV2
		if err := json.Unmarshal(content, &manifest); err != nil {
			return nil, err
		}
		inputs := []string{}
		for _, entry := range manifest.Manifests {
			inputs = append(inputs, entry.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(inputs...)

	case "application/vnd.oci.image.manifest.v1+json":
		var manifest oci.OCIImageManifest
		if err := json.Unmarshal(content, &manifest); err != nil {
			return nil, err
		}
		inputs := []string{manifest.Config.Digest}
		for _, layer := range manifest.Layers {
			inputs = append(inputs, layer.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(inputs...)

	case "application/vnd.oci.image.index.v1+json":
		var manifest oci.OCIImageIndex
		if err := json.Unmarshal(content, &manifest); err != nil {
			return nil, err
		}
		if manifest.Manifests == nil {
			return nil, fmt.Errorf("no manifests found in OCI index")
		}
		manifestDigests := []string{}
		for _, m := range manifest.Manifests {
			manifestDigests = append(manifestDigests, m.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(manifestDigests...)

	default:
		return nil, fmt.Errorf("unsupported mediaType: %s", mediaType)
	}

	nsId, repoId, err := svc.getNameSpaceIdAndRepositoryId(namespace, repository)
	if err != nil {
		return nil, err
	}

	// Check manifest existence
	manifestExists, manifestId, err := svc.registryDao.CheckImageManifestExistsByUniqueDigest(
		svc.registryId, nsId, repoId, uniqueDigest, txKey,
	)
	if err != nil {
		return nil, err
	}

	result := &ManifestCheckResult{
		ManifestExists: manifestExists,
		UniqueDigest:   uniqueDigest,
		ManifestId:     manifestId,
		NamespaceId:    nsId,
		RepositoryId:   repoId,
	}

	// Check tag existence
	tagId, err := svc.registryDao.GetImageTagId(svc.registryId, nsId, repoId, tag, txKey)
	if db_errors.IsNotFound(err) {
		result.TagExists = false
	} else if err != nil {
		return nil, err
	} else {
		result.TagExists = true
		result.TagId = tagId
	}

	// Check tag->manifest link
	oldManifestId, err := svc.registryDao.GetLinkedManifestByTagId(tagId, txKey)
	if db_errors.IsNotFound(err) {
		result.TagManifestLinkExists = false
	} else if err != nil {
		return nil, err
	} else {
		result.TagManifestLinkExists = true
		if oldManifestId != manifestId {
			result.TagManifestLinkChanged = true
		}
	}

	return result, nil
}