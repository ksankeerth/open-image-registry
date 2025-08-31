package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksankeerth/open-image-registry/db"
	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	storage_errors "github.com/ksankeerth/open-image-registry/errors/storage"
	"github.com/ksankeerth/open-image-registry/handlers/common"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/storage"
	"github.com/ksankeerth/open-image-registry/types"
	"github.com/ksankeerth/open-image-registry/utils"
)

// For Upstream registries, random UUID will be generated.
// For images which are supposed to be hosted in OpenImageRegistry,
// We'll use this ID.
const HostedRegistryID = "1"

// By default, the namespace can be username or organization name.
// If namespace is not provided, We'll use this namespace.
const DefaultNamespace = "library"

const UnknownBlobMediaType = "unknown_media_type"

type DockerV2Handler struct {
	regId     string
	regName   string
	imgRegDao db.ImageRegistryDAO
}

func NewDockerV2Handler(regId string, regName string, imgRegDao db.ImageRegistryDAO) *DockerV2Handler {
	return &DockerV2Handler{
		regId:     regId,
		regName:   regName,
		imgRegDao: imgRegDao,
	}
}

// GetDockerV2APISupport returns HTTP 200.
// New Docker clients use this endpoints to decide which version
// Docker Registry API should be used when pulling/pushing images.
func (dh *DockerV2Handler) GetDockerV2APISupport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Docker-Distribution-API-Version": "registry/2.0"}`))
}

// InitiateImageBlobUpload returns HTTP 202 and send URL in `Location` header
// That url will be used to upload blobs in subsequent API calls.
// The returned URLs will have namespace, repository and session id.
func (dh *DockerV2Handler) InitiateImageBlobUpload(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	repository := chi.URLParam(r, "repository")

	if namespace == "" {
		namespace = DefaultNamespace
	}

	// allowing namespace and repository with invalid characters may cause issues when storing
	// blob in file-system. Therefore, We will not accept this request.
	if !utils.IsValidNamespace(namespace) || !utils.IsValidRepository(repository) {
		common.HandleBadRequest(w, 400, "Namespace or Repository has invalid characters")
		return
	}

	sessionId := uuid.New().String()

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	nsId, repoId, err := dh.imgRegDao.CreateNamespaceAndRepositoryIfNotExist(dh.regId, namespace, repository, "")
	if err != nil {
		ok, code := db_errors.UnwrapDBError(err)
		if ok {
			common.HandleInternalError(w, code, "Unable to create namespace and repository")
		} else {
			common.HandleInternalError(w, 500, "Unable to create namespace and repository")
		}
		return
	}

	err = dh.imgRegDao.CreateNewBlobUploadSession(dh.regId, nsId, repoId, sessionId, "")
	if err != nil {
		ok, code := db_errors.UnwrapDBError(err)
		if ok {
			common.HandleInternalError(w, code, "Unable to create new blob upload session in database")
		} else {
			common.HandleInternalError(w, 500, "Unable to create new blob upload session in database")
		}
		return
	}

	uploadUrl := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s", scheme, r.Host, namespace, repository, sessionId)

	w.Header().Set("Location", uploadUrl)
	w.Header().Set("Docker-Upload-UUID", sessionId)
	w.Header().Set("Range", "0-0")

	w.WriteHeader(http.StatusAccepted)

	w.Write([]byte(`{"Location":"` + uploadUrl + `" }`))
}

// CheckImageBlobExistence returns HTTP 200 if the given digest available in DB.
// Otherwise, it returns HTTP 404.
func (dh *DockerV2Handler) CheckImageBlobExistence(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository := chi.URLParam(r, "repository")
	digest := chi.URLParam(r, "digest")

	exists, err := dh.imgRegDao.CheckImageBlobExists(digest, namespace, repository, dh.regId, "")
	if err != nil {
		ok, code := db_errors.UnwrapDBError(err)
		if ok {
			common.HandleInternalError(w, code, "Error occured when checking image blob in database")
		} else {
			common.HandleInternalError(w, 500, "Error occured when checking image blob in database")
		}
		return
	}

	if !exists {
		common.HandleNotFound(w, 404, "Image blob not found.")
		return
	}

	w.Header().Set("Content-Length", "0")
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

func (dh *DockerV2Handler) HandleImageBlobUpload(w http.ResponseWriter, r *http.Request) {
	sessionId := chi.URLParam(r, "session_id")
	namespace := chi.URLParam(r, "namespace")
	repository := chi.URLParam(r, "repository")
	if namespace == "" {
		namespace = DefaultNamespace
	}

	// allowing namespace and repository with invalid characters may cause issues when storing
	// blob in file-system. Therefore, We will not accept this request.
	if !utils.IsValidNamespace(namespace) || !utils.IsValidRepository(repository) {
		common.HandleBadRequest(w, 400, "Namespace or Repository has invalid characters")
		return
	}

	sessionNotFound := false
	_, isChunked, _, err := dh.imgRegDao.GetBlobUploadSession(sessionId, "")
	if db_errors.IsNotFound(err) {
		sessionNotFound = true
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	if sessionNotFound {
		log.Logger().Warn().Msgf("Image blob upload session: %s is not found in database. Probably unknown session or removed due to timeout.", sessionId)
		common.HandleInternalError(w, 500, "Session not found.")
		return
	}

	storageLocationToCleanup := ""

	defer func() {
		if r.Method == http.MethodPut { // monolithic blob upload or final chunk upload
			log.Logger().Debug().Msgf("Removing blob upload session from database ......")

			err := dh.imgRegDao.DeleteBlobUploadSession(sessionId, "")
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Deleting blob upload session: %s failed due to errors.", sessionId)
			}

			if storageLocationToCleanup != "" {
				err = storage.DeleteFile(storageLocationToCleanup)
				if err != nil {
					log.Logger().Error().Err(err).Msgf("Deleting partial blob: %s from storage failed due to errors.", storageLocationToCleanup)
				}
			}
		}
	}()

	// chunked uploload
	if r.Method == http.MethodPatch {
		location := utils.StorageLocation("blobs", dh.regName, namespace, repository, sessionId)

		// it is not marked already, this is the first chunk
		// so let's mark the session as chunked
		if !isChunked {
			err := dh.imgRegDao.MarkBlobUploadSessionAsChunked(sessionId, "")
			if err != nil {
				log.Logger().Err(err).Msgf("Unable to update the session(%s) in the database. Hence, first chunk of blob upload failed.", sessionId)
				common.HandleInternalError(w, 500, "Unable to upload chunk")
				return
			}
		}

		contentRange := r.Header.Get("Content-Range")
		// Content-Range cannot empty from second chunks.
		if contentRange == "" && isChunked {
			log.Logger().Warn().Msg("Content-Range header is empty. Unable to find chunk boundary.")
			common.HandleBadRequest(w, 400, "Content-Range header is missing.")
			return
		}
		start, end, err := utils.ParseImageBlobContentRangeFromRequest(contentRange)

		chunk, err := io.ReadAll(r.Body)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to read chunk from request: %s", r.RequestURI)
			common.HandleInternalError(w, 500, "Failed to read chunk from request")
			return
		}

		// For the first, start and end could be zero. Therefore, we avoid checking content size validation
		// on first chunk
		if isChunked && len(chunk) != int(end-start) {
			log.Logger().Warn().Msgf("Partial chunk was received.")
			common.HandleBadRequest(w, 400, "Partial chunk was received.")
			return
		}
		err = storage.PutFileChunk(location, chunk, start)
		if err != nil {
			ok, code := storage_errors.UnwrapStorageError(err)
			if ok {
				common.HandleInternalError(w, code, "Unable to store chunk in storage")
			} else {
				common.HandleInternalError(w, 500, "Unable to store chunk in storage")
			}
			return
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		uploadUrl := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s", scheme, r.Host, namespace, repository, sessionId)

		if start == end && start == 0 {
			end = int64(len(chunk))
		}
		w.Header().Add("Location", uploadUrl)
		w.Header().Add("Range", fmt.Sprintf("%d-%d", start, end))
		w.Header().Add("Docker-Upload-UUID", sessionId)
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusCreated)
		return
	}

	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// final chunk upload
	if isChunked {
		blobDigest := r.URL.Query().Get("digest")
		if blobDigest == "" {
			log.Logger().Warn().Msg("Blob upload failed due to missing query param: digest")
			common.HandleBadRequest(w, 400, "Query param `digest` was not available.")
			return
		}

		oldLocation := utils.StorageLocation("blobs", dh.regName, namespace, repository, sessionId)
		newLocation := utils.StorageLocation("blobs", dh.regName, namespace, repository, blobDigest)

		err = storage.RenameFile(oldLocation, newLocation)
		if err != nil {
			storageLocationToCleanup = oldLocation
			log.Logger().Warn().Msgf("Renaming %s to %s failed due to storage errors.", oldLocation, newLocation)
			common.HandleInternalError(w, 500, "Request aborted due to storage errors")
			return
		}
		size, err := storage.Size(newLocation)
		if err != nil {
			log.Logger().Warn().Msgf("Unable to retrive `size` of %s failed due to storage errors.", newLocation)
			storageLocationToCleanup = newLocation
			common.HandleInternalError(w, 500, "Request aborted due to storage errors")
			return
		}
		nsId, err := dh.imgRegDao.GetNamespaceId(dh.regId, namespace, "")
		if err != nil {
			storageLocationToCleanup = newLocation
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors: %s, %s", dh.regId, namespace)
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}

		repoId, err := dh.imgRegDao.GetRepositoryId(dh.regId, nsId, repository, "")
		if err != nil {
			storageLocationToCleanup = newLocation
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors: %s, %s, %s", dh.regId, namespace, repository)
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}

		exists, err := dh.imgRegDao.CheckImageBlobExistsByIds(blobDigest, nsId, repoId, dh.regId, "")
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}

		if !exists {
			// At this point, We don't know the media type.
			err = dh.imgRegDao.PersistImageBlobMetaInfo(dh.regId, nsId, repoId, blobDigest, newLocation, UnknownBlobMediaType, size, "")
			if err != nil {
				storageLocationToCleanup = newLocation
				log.Logger().Error().Err(err).Msgf("Request aborted due to database errors: %s, %s, %s, %s, %s", dh.regId, nsId, repoId, blobDigest, newLocation)
				common.HandleInternalError(w, 500, "Request aborted due to database errors")
				return
			}
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		uploadUrl := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s", scheme, r.Host, namespace, repository, sessionId)

		w.Header().Add("Location", uploadUrl)
		w.Header().Add("Range", fmt.Sprintf("bytes=0-%d", size))
		w.Header().Add("Docker-Upload-UUID", sessionId)
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// monolithic blob upload
	blobDigest := r.URL.Query().Get("digest")
	if blobDigest == "" {
		log.Logger().Warn().Msgf("Request aborted due to missing query param `digest` : %s", r.RequestURI)
		common.HandleBadRequest(w, 400, "Request aborted due to missing query param `digest`")
		return
	}

	storageLocation := utils.StorageLocation("blobs", dh.regName, namespace, repository, blobDigest)

	blobBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to read request body: %s", r.RequestURI)
		common.HandleBadRequest(w, 400, "Unable to read payload from request body")
		return
	}

	err = storage.PutFile(storageLocation, blobBytes)
	if err != nil {
		log.Logger().Warn().Msgf("Request aborted due to storage errors: %s", storageLocation)
		common.HandleInternalError(w, 500, "Request aborted due to storage errors")
		return
	}

	nsId, err := dh.imgRegDao.GetNamespaceId(dh.regId, namespace, "")
	if err != nil {
		storageLocationToCleanup = storageLocation
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors: %s, %s", dh.regId, namespace)
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	repoId, err := dh.imgRegDao.GetRepositoryId(dh.regId, nsId, repository, "")
	if err != nil {
		storageLocationToCleanup = storageLocation
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors: %s, %s, %s", dh.regId, namespace, repository)
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	err = dh.imgRegDao.PersistImageBlobMetaInfo(dh.regId, nsId, repoId, blobDigest,
		storageLocation, UnknownBlobMediaType, int64(len(blobBytes)), "")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors: %s, %s, %s, %s, %s",
			dh.regId, namespace, repository, blobDigest, storageLocation)
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	uploadUrl := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s", scheme, r.Host, namespace, repository, sessionId)

	w.Header().Set("Location", uploadUrl)
	w.Header().Set("Range", fmt.Sprintf("bytes=0-%d", len(blobBytes)))
	w.Header().Set("Docker-Upload-UUID", sessionId)
	w.WriteHeader(http.StatusAccepted)
}

func (dh *DockerV2Handler) GetImageBlob(w http.ResponseWriter, r *http.Request) {
	digest := chi.URLParam(r, "digest")
	namespace := chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository := chi.URLParam(r, "repository")

	location, size, err := dh.imgRegDao.GetImageBlobStorageLocationAndSize(digest,
		namespace, repository, dh.regId, "")
	if err != nil {
		ok, code := db_errors.UnwrapDBError(err)
		if ok {
			common.HandleInternalError(w, code, "Unable to retrieve blob meta data from database")
		} else {
			common.HandleInternalError(w, 500, "Unable to retrieve blob meta data from database")
		}
		return
	}
	data, err := storage.ReadFile(location)
	if err != nil {
		ok, code := storage_errors.UnwrapStorageError(err)
		if ok {
			common.HandleInternalError(w, code, "Unable to read blob from storage")
		} else {
			common.HandleInternalError(w, 500, "Unable to read blob from storage")
		}
		return
	}

	if size != len(data) {
		log.Logger().Warn().Msgf("Retrieved blob's size does not match with metadata.")
		common.HandleInternalError(w, 500, "Retrieved blob's size does not match with metadata.")
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(size))
	w.Header().Set("Docker-Content-Digest", digest)

	w.Write(data)
}

func (dh *DockerV2Handler) CheckImageManifestExistence(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository := chi.URLParam(r, "repository")
	tagOrDigest := chi.URLParam(r, "tag_or_digest")

	manifestExists := false
	var err error
	if utils.IsImageDigest(tagOrDigest) {
		manifestExists, err = dh.imgRegDao.CheckImageManifestExists(tagOrDigest,
			namespace, repository, dh.regId, "")
	} else {
		// TODO: for one tag, We could have many manifests stored for different platforms. Therefore,
		// here, we cannot decide whether we can return 200 or not. Do further research later
		// manifestExists, err = dh.imgRegDao.CheckImageManifestExistsByTag(tagOrDigest,
		// 	namespace, repository, dh.regId, "")

		exist, err := dh.imgRegDao.CheckImageManifestsExistByTag(dh.regId, namespace, repository, tagOrDigest, "")
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}
		if exist {
			w.WriteHeader(http.StatusOK)
			return
		}

		common.HandleNotFound(w, 404, "Not found")
		return
	}
	if err != nil {
		ok, code := db_errors.UnwrapDBError(err)
		if ok {
			common.HandleInternalError(w, code, "Error occured when checking image manifest in database.")
		} else {
			common.HandleInternalError(w, 500, "Error occured when checking image manifest in database.")
		}
		return
	}

	if manifestExists {
		w.WriteHeader(http.StatusOK)
	} else {
		common.HandleNotFound(w, 404, "Image manifest not found")
	}
}

func (dh *DockerV2Handler) GetImageManifest(w http.ResponseWriter, r *http.Request) {
	tagOrDigest := chi.URLParam(r, "tag_or_digest")
	namespace := chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository := chi.URLParam(r, "repository")

	if utils.IsImageDigest(tagOrDigest) {
		content, mediaType, _, err := dh.imgRegDao.GetImageManifestByDigest(tagOrDigest,
			namespace, repository, dh.regId, "")
		if err != nil {
			ok, code := db_errors.UnwrapDBError(err)
			if ok {
				common.HandleInternalError(w, code, "Unable to retrieve Image manifest from database")
			} else {
				common.HandleInternalError(w, 500, "Unable to retrieve Image manifest from database")
			}
			return
		}

		w.Header().Set("Content-Type", mediaType)
		w.Header().Set("docker-content-digest", tagOrDigest)
		w.Header().Set("docker-distribution-api-version", "registry/2.0")

		w.Write(content)
		return

	}
	manifests, err := dh.imgRegDao.GetImageManifestsByTag(tagOrDigest, namespace, repository, dh.regId, "")
	if err != nil {
		ok, code := db_errors.UnwrapDBError(err)
		if ok {
			common.HandleInternalError(w, code, "Unable to retrieve Image manifest from database")
		} else {
			common.HandleInternalError(w, 500, "Unable to retrieve Image manifest from database")
		}
		return
	}

	fatManifests := types.DockerV2ManifestList{
		SchemaVersion: 2,
		MediaType:     "application/vnd.docker.distribution.manifest.list.v2+json",
		Manifests:     make([]types.DockerV2ManifestEntry, 0),
	}

	for _, manifest := range manifests {
		fatManifests.Manifests = append(fatManifests.Manifests, types.DockerV2ManifestEntry{
			MediaType: manifest.MediaType,
			Size:      manifest.Size,
			Digest:    manifest.Digest,
			Platform: types.DockerV2PlatformEntry{
				OS:           manifest.OS,
				Architecture: manifest.Arch,
			},
		})
	}

	var buf bytes.Buffer

	err = json.NewEncoder(&buf).Encode(fatManifests)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to marshall fatmanifests: %s", r.RequestURI)
		common.HandleInternalError(w, 500, "Failed to marshall response")
		return
	}

	responseDigest := utils.CalcuateDigest(buf.Bytes())

	w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.list.v2+json")
	w.Header().Set("docker-content-digest", responseDigest)
	w.Header().Set("docker-distribution-api-version", "registry/2.0")

	w.Write(buf.Bytes())
}

func (dh *DockerV2Handler) UpdateManifest(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/vnd.docker.distribution.manifest.v2+json" {
		common.HandleBadRequest(w, 400, "Content type: application/vnd.docker.distribution.manifest.v1+json is not supported.")
		return
	}

	tag := chi.URLParam(r, "tag")
	namespace := chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository := chi.URLParam(r, "repository")

	var buf bytes.Buffer

	_, err := io.Copy(&buf, r.Body)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors in reading request body: %s", r.RequestURI)
		common.HandleBadRequest(w, 400, "Request aborted due to errors in reading request body")
		return
	}

	var manifest types.V2ImageManifest
	err = json.Unmarshal(buf.Bytes(), &manifest)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors in unmarshalling payload: %s", r.RequestURI)
		common.HandleBadRequest(w, 400, "Request aborted due to errors in unmarshalling payload")
		return
	}

	if manifest.Config.MediaType != "application/vnd.docker.container.image.v1+json" {
		log.Logger().Warn().Msgf("Unsupport container image schema: %s", manifest.Config.MediaType)
		common.HandleBadRequest(w, 400, "Request aborted due to unsupported container image schema")
		return
	}

	argsForUniqDigestCalculations := []string{manifest.Config.Digest}
	for _, layer := range manifest.Layers {
		argsForUniqDigestCalculations = append(argsForUniqDigestCalculations, layer.Digest)
	}

	manifestUniqDigest := utils.CombineAndCalculateSHA256Digest(argsForUniqDigestCalculations...)

	manifestDigest := utils.CalcuateDigest(buf.Bytes())

	exists, err := dh.imgRegDao.CheckImageManifestExists(manifestUniqDigest, namespace, repository, dh.regId, "")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors.")
		common.HandleInternalError(w, 500, "Request aborted due to database errors.")
		return
	}
	if exists {
		w.Header().Add("Docker-Content-Digest", manifestDigest)
		w.WriteHeader(http.StatusCreated)
		return
	}

	txKey := fmt.Sprintf("%s-%s-%s-%s", dh.regId, namespace, repository, manifestDigest)

	err = dh.imgRegDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	defer func() {
		if err != nil {
			dh.imgRegDao.Rollback(txKey)
		} else {
			dh.imgRegDao.Commit(txKey)
		}
	}()

	nsId, err := dh.imgRegDao.GetNamespaceId(dh.regId, namespace, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	repoId, err := dh.imgRegDao.GetRepositoryId(dh.regId, nsId, repository, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	var tagId string = ""

	tagId, err = dh.imgRegDao.GetImageTagId(dh.regId, nsId, repoId, tag, txKey)
	if err != nil {
		if db_errors.IsNotFound(err) {
			tagId, err = dh.imgRegDao.CreateImageTag(dh.regId, nsId, repoId, tag, txKey)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
				common.HandleInternalError(w, 500, "Request aborted due to database errors")
				return
			}

			err = dh.imgRegDao.LinkImageConfigWithTag(manifest.Config.Digest, tagId, txKey)

		} else {
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}
	} else {
		exists, err = dh.imgRegDao.CheckImageTagAndConfigDigestMapping(manifest.Config.Digest, tagId, txKey)
		if err != nil && !db_errors.IsNotFound(err) {
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}
		if !exists {
			err = dh.imgRegDao.LinkImageConfigWithTag(manifest.Config.Digest, tag, txKey)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
				common.HandleInternalError(w, 500, "Request aborted due to database errors")
				return
			}
		}

	}

	err = dh.imgRegDao.MarkImageBlobAsConfig(dh.regId, nsId, repoId, manifest.Config.Digest, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	// We observed that if the same image is pushed twice, the manifest JSON can have a different size for layers,
	// which results in a different manifest digest. However, both images share the same image configuration digest.
	// This issue has been observed by others as well, as described here:
	// https://stackoverflow.com/questions/76524662/docker-two-images-share-the-same-image-hash-but-different-digests-how

	for index, layer := range manifest.Layers {
		exists, err = dh.imgRegDao.CheckImageTagAndLayerMappingExistence(tagId, layer.Digest, index, txKey)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}
		if exists {
			continue
		}
		err = dh.imgRegDao.LinkImageTagAndLayer(dh.regId, nsId, repoId, tagId, layer.Digest,
			manifest.Config.Digest, index, txKey)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}
	}

	exists, err = dh.imgRegDao.CheckImagePlatformConfigExistence(tagId, manifest.Config.Digest, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}
	if !exists {
		imageConfigStorageLocation, size, err := dh.imgRegDao.GetImageBlobStorageLocationAndSizeByIds(
			manifest.Config.Digest, nsId, repoId, dh.regId, txKey)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
			common.HandleInternalError(w, 500, "Request aborted due to database errors")
			return
		}

		blobBytes, err := storage.ReadFile(imageConfigStorageLocation)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Request aborted due to storage errors")
			common.HandleInternalError(w, 500, "Request aborted due to storage errors")
			return
		}

		if size != len(blobBytes) {
			log.Logger().Warn().Err(err).Msgf("Size mismatch was observed for Image Config(%s) in database(%d) and storage(%d)", imageConfigStorageLocation, size, len(blobBytes))
			common.HandleInternalError(w, 500, "Size mismatch was observed for Image Config in database and storage")
			return
		}
		var imageConfig types.ImageConfig
		err = json.Unmarshal(blobBytes, &imageConfig)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unmarshalling image config from storage failed")
			common.HandleInternalError(w, 500, "Request aborted due to errors in unmarshalling image config")
			return
		}

		if imageConfig.OS != "" && imageConfig.Architecture != "" {

			err = dh.imgRegDao.PersistImagePlatformConfig(dh.regId, nsId, repoId, tagId,
				manifest.Config.Digest, imageConfig.OS, imageConfig.Architecture,
				[]byte{}, txKey)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
				common.HandleInternalError(w, 500, "Request aborted due to database errors")
				return
			}
		}
	}

	err = dh.imgRegDao.PersistImageManifest(dh.regId, nsId, repoId, manifestDigest,
		manifest.Config.Digest, contentType, manifestUniqDigest, int64(len(buf.Bytes())), buf.Bytes(), txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to database errors")
		common.HandleInternalError(w, 500, "Request aborted due to database errors")
		return
	}

	w.Header().Add("Docker-Content-Digest", manifestDigest)
	w.WriteHeader(http.StatusCreated)
}