package registry

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/google/uuid"

	"github.com/ksankeerth/open-image-registry/db"
	"github.com/ksankeerth/open-image-registry/errors/dockererrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/utils"
)

type RegistryHandler struct {
	registryId   string
	registryName string
	svc          *RegistryService
}

func NewRegistryHandler(registryId, registryName string, registryDao db.ImageRegistryDAO,
	upstreamDao db.UpstreamDAO) *RegistryHandler {

	svc := NewRegistryService(registryId, registryName, registryDao, upstreamDao)

	return &RegistryHandler{
		registryId:   registryId,
		registryName: registryName,
		svc:          svc,
	}
}

func (rh *RegistryHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(httplog.NewLogger(fmt.Sprintf("DockerV2API-%s", rh.registryName), httplog.Options{
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	})))

	r.Route("/v2", func(r chi.Router) {
		r.Get("/", rh.GetDockerV2APISupport)

		//blob
		r.Post("/{namespace}/{repository}/blobs/uploads/", rh.InitiateImageBlobUpload)
		r.Post("/{repository}/blobs/uploads/", rh.InitiateImageBlobUpload)

		r.Head("/{namespace}/{repository}/blobs/{digest}", rh.IsImageBlobPresent)
		r.Head("/{repository}/blobs/{digest}", rh.IsImageBlobPresent)

		r.Put("/{namespace}/{repository}/blobs/uploads/{session_id}", rh.HandleImageBlobUpload)
		r.Put("/{repository}/blobs/uploads/{session_id}", rh.HandleImageBlobUpload)

		r.Patch("/{namespace}/{repository}/blobs/uploads/{session_id}", rh.HandleImageBlobUpload)
		r.Patch("/{repository}/blobs/uploads/{session_id}", rh.HandleImageBlobUpload)

		r.Get("/{namespace}/{repository}/blobs/{digest}", rh.GetImageBlob)
		r.Get("/{repository}/blobs/{digest}", rh.GetImageBlob)

		r.Head("/{namespace}/{repository}/manifests/{tag_or_digest}", rh.IsImageManifestPresent)
		r.Head("/{repository}/manifests/{tag_or_digest}", rh.IsImageManifestPresent)

		r.Put("/{namespace}/{repository}/manifests/{tag}", rh.UpdateManifest)
		r.Put("/{repository}/manifests/{tag}", rh.UpdateManifest)

		r.Get("/{namespace}/{repository}/manifests/{tag_or_digest}", rh.GetImageManifest)
		r.Get("/{repository}/manifests/{tag_or_digest}", rh.GetImageManifest)
	})

	return r
}

func (rh *RegistryHandler) GetDockerV2APISupport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Docker-Distribution-API-Version": "registry/2.0"}`))
}

func (rh *RegistryHandler) InitiateImageBlobUpload(w http.ResponseWriter, r *http.Request) {
	if rh.registryId != HostedRegistryID {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	namespace, repository := extractNamespaceAndRepository(r)

	// allowing namespace and repository with invalid characters may cause issues when storing
	// blob in file-system. Therefore, We will not accept this request.
	if !utils.IsValidNamespace(namespace) || !utils.IsValidRepository(repository) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId := uuid.New().String()

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	err := rh.svc.createNamespaceAndRepositoryIfNotExist(namespace, repository)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uploadUrl := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s", scheme, r.Host, namespace, repository, sessionId)

	w.Header().Set("Location", uploadUrl)
	w.Header().Set("Docker-Upload-UUID", sessionId)
	w.Header().Set("Range", "0-0")

	w.WriteHeader(http.StatusAccepted)

	w.Write([]byte(`{"Location":"` + uploadUrl + `" }`))
}

func (rh *RegistryHandler) IsImageBlobPresent(w http.ResponseWriter, r *http.Request) {
	namespace, repository, digest := extractNamespaceRepositoryAndDigest(r)

	exists, err := rh.svc.isImageBlobPresent(namespace, repository, digest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else if exists {
		writeBlobExistsResponse(w, digest)
	} else {
		dockererrors.WriteBlobNotFound(w)
	}
}

func (rh *RegistryHandler) HandleImageBlobUpload(w http.ResponseWriter, r *http.Request) {
	if rh.registryId != HostedRegistryID {
		// For the proxy registries, not allowed to upload blobs
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	namespace, repository, sessionId := extractNamespaceRepositoryAndSessionId(r)

	// Prevent invalid namespace/repository names (storage safety)
	if !utils.IsValidNamespace(namespace) || !utils.IsValidRepository(repository) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to read chunk from request: %s", r.RequestURI)
		dockererrors.WriteBlobUploadInvalid(w)
		return
	}

	switch {
	case len(payload) == 0 && r.Method == http.MethodPut:
		rh.handleFinalChunkUpload(w, r, namespace, repository, sessionId)

	case r.Method == http.MethodPatch:
		rh.handleChunkedUpload(w, r, namespace, repository, sessionId, payload)

	case r.Method == http.MethodPut:
		rh.handleMonolithicUpload(w, r, namespace, repository, sessionId, payload)
	}
}

func (rh *RegistryHandler) handleFinalChunkUpload(
	w http.ResponseWriter, r *http.Request,
	namespace, repository, sessionId string,
) {
	blobDigest := r.URL.Query().Get("digest")
	if blobDigest == "" {
		log.Logger().Warn().Msgf("Request aborted due to missing query param `digest` : %s", r.RequestURI)
		dockererrors.WriteInvalidDigest(w, blobDigest)
		return
	}

	err := rh.svc.persistImageBlobFinalChunk(namespace, repository, blobDigest, sessionId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeBlobUploadSuccessResponse(w, r.URL.String(), sessionId, 0, 0)
}

func (rh *RegistryHandler) handleChunkedUpload(
	w http.ResponseWriter, r *http.Request,
	namespace, repository, sessionId string, payload []byte,
) {
	firstChunk := false
	var start int64 = 0
	var end int64 = 0
	contentRange := r.Header.Get("Content-Range")

	// Content-Range cannot be empty for second+ chunks
	if contentRange == "" {
		firstChunk = true
		// correcting end
		end = int64(len(payload))
	}
	if !firstChunk {
		var err error
		start, end, err = utils.ParseImageBlobContentRangeFromRequest(contentRange)
		if err != nil {
			log.Logger().Warn().Msg("Unable to parse Content-Range header")
			dockererrors.WriteInvalidRange(w)
			return
		}
		if len(payload) != int(end-start) {
			log.Logger().Warn().Msgf("Partial chunk was received.")
			dockererrors.WriteInvalidRange(w)
			return
		}
	} else {
		start = 0
	}

	location := utils.StorageLocation("blobs", rh.registryName, namespace, repository, sessionId)
	err := rh.svc.storeImageBlob(location, true, start, payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeBlobUploadSuccessResponse(w, r.URL.String(), sessionId, int(start), int(end))
}

func (rh *RegistryHandler) handleMonolithicUpload(
	w http.ResponseWriter, r *http.Request,
	namespace, repository, sessionId string, payload []byte,
) {
	blobDigest := r.URL.Query().Get("digest")
	if blobDigest == "" {
		log.Logger().Warn().Msgf("Request aborted due to missing query param `digest` : %s", r.RequestURI)
		dockererrors.WriteInvalidDigest(w, "")
		return
	}

	err := rh.svc.persistImageBlob(namespace, repository, blobDigest, payload)
	if err != nil {
		log.Logger().Warn().Msgf("Request aborted due to errors")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeBlobUploadSuccessResponse(w, r.URL.Path, sessionId, 0, len(payload))
}

func (rh *RegistryHandler) GetImageBlob(w http.ResponseWriter, r *http.Request) {
	namespace, repository, digest := extractNamespaceRepositoryAndDigest(r)

	exists, content, err := rh.svc.getImageBlob(namespace, repository, digest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		dockererrors.WriteBlobNotFound(w)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		w.Header().Set("Docker-Content-Digest", digest)

		w.Write(content)
	}
}

func (rh *RegistryHandler) IsImageManifestPresent(w http.ResponseWriter, r *http.Request) {
	namespace, repository, tagOrDigest := extractNamespaceRepositoryAndTagOrDigest(r)

	exists, mediaType, digest, err := rh.svc.isImageManifestPresent(namespace, repository, tagOrDigest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exists {
		dockererrors.WriteManifestNotFound(w)
		return
	}

	w.Header().Set("Content-Length", "0")
	w.Header().Set("Docker-Content-Digest", digest)
	w.Header().Set("Content-Type", mediaType)
	w.WriteHeader(http.StatusOK)
}

func (rh *RegistryHandler) GetImageManifest(w http.ResponseWriter, r *http.Request) {
	namespace, repository, tagOrDigest := extractNamespaceRepositoryAndTagOrDigest(r)

	exists, mediaType, digest, content, err := rh.svc.getImageManifest(namespace, repository, tagOrDigest)

	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		dockererrors.WriteManifestNotFound(w)
		return
	}

	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.Header().Set("Docker-Content-Digest", digest)
	w.Header().Set("Content-Type", mediaType)
	w.WriteHeader(http.StatusOK)

	w.Write(content)
}

func (rh *RegistryHandler) UpdateManifest(w http.ResponseWriter, r *http.Request) {
	if rh.registryId != HostedRegistryID {
		dockererrors.WriteUnsupported(w)
		return
	}

	namespace, repository, tag := extractNamespaceRepositoryAndTag(r)

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		dockererrors.WriteManifestInvalid(w, "Content-Type is not avaible in the request")
		return
	}

	content, err := io.ReadAll(r.Body)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading manifest from request: %s", r.RequestURI)
		dockererrors.WriteManifestInvalid(w, "Manifest read failed")
		return
	}

	digest, err := rh.svc.updateManifest(namespace, repository, tag, contentType, content)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when updating manifest for request: %s", r.RequestURI)
		return
	}
	w.Header().Set("Content-Length", "0")
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}
