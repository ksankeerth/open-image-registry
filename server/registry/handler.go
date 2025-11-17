package registry

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"

	"github.com/ksankeerth/open-image-registry/errors/dockererrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/utils"
)

type RegistryHandler struct {
	registryId   string
	registryName string
	svc          *RegistryService
}

func NewRegistryHandler(registryId, registryName string, s store.Store) *RegistryHandler {

	svc := NewRegistryService(registryId, registryName, s)

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
		r.Get("/", rh.dockerV2APISupport)

		//blob
		r.Post("/{namespace}/{repository}/blobs/uploads/", rh.initiateBlobUpload)
		r.Post("/{repository}/blobs/uploads/", rh.initiateBlobUpload)

		r.Head("/{namespace}/{repository}/blobs/{digest}", rh.blobExists)
		r.Head("/{repository}/blobs/{digest}", rh.blobExists)

		r.Put("/{namespace}/{repository}/blobs/uploads/{session_id}", rh.handleBlobUpload)
		r.Put("/{repository}/blobs/uploads/{session_id}", rh.handleBlobUpload)

		r.Patch("/{namespace}/{repository}/blobs/uploads/{session_id}", rh.handleBlobUpload)
		r.Patch("/{repository}/blobs/uploads/{session_id}", rh.handleBlobUpload)

		r.Get("/{namespace}/{repository}/blobs/{digest}", rh.getImageBlob)
		r.Get("/{repository}/blobs/{digest}", rh.getImageBlob)

		r.Head("/{namespace}/{repository}/manifests/{tag_or_digest}", rh.manifestExists)
		r.Head("/{repository}/manifests/{tag_or_digest}", rh.manifestExists)

		r.Put("/{namespace}/{repository}/manifests/{tag}", rh.updateManifest)
		r.Put("/{repository}/manifests/{tag}", rh.updateManifest)

		r.Get("/{namespace}/{repository}/manifests/{tag_or_digest}", rh.getManifest)
		r.Get("/{repository}/manifests/{tag_or_digest}", rh.getManifest)
	})

	return r
}

func (rh *RegistryHandler) dockerV2APISupport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Docker-Distribution-API-Version": "registry/2.0"}`))
}

func (rh *RegistryHandler) initiateBlobUpload(w http.ResponseWriter, r *http.Request) {
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

	sessionID, err := rh.svc.initiateBlobUpload(r.Context(), namespace, repository)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		dockererrors.WriteBlobUploadInvalid(w)
		return
	}

	// namespace or repository do not exist
	if sessionID == "" {
		dockererrors.WriteRepositoryNotFound(w)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	uploadUrl := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s", scheme, r.Host, namespace, repository, sessionID)

	w.Header().Set("Location", uploadUrl)
	w.Header().Set("Docker-Upload-UUID", sessionID)
	w.Header().Set("Range", "0-0")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"Location":"` + uploadUrl + `" }`))
}

func (rh *RegistryHandler) blobExists(w http.ResponseWriter, r *http.Request) {
	namespace, repository, digest := extractNamespaceRepositoryAndDigest(r)

	exists, err := rh.svc.blobExists(r.Context(), namespace, repository, digest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else if exists {
		writeBlobExistsResponse(w, digest)
	} else {
		dockererrors.WriteBlobNotFound(w)
	}
}

func (rh *RegistryHandler) handleBlobUpload(w http.ResponseWriter, r *http.Request) {
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
		log.Logger().Error().Err(err).Msgf("Unable to read blob payload from request: %s", r.RequestURI)
		dockererrors.WriteBlobUploadInvalid(w)
		return
	}

	switch {
	case len(payload) == 0 && r.Method == http.MethodPut:
		rh.handleLastBlobChunk(w, r, namespace, repository, sessionId)

	case r.Method == http.MethodPatch:
		rh.handleBlobChunk(w, r, namespace, repository, sessionId, payload)

	case r.Method == http.MethodPut:
		rh.handleBlob(w, r, namespace, repository, sessionId, payload)
	}
}

func (rh *RegistryHandler) handleLastBlobChunk(w http.ResponseWriter, r *http.Request,
	namespace, repository, sessionID string) {
	blobDigest := r.URL.Query().Get("digest")
	if blobDigest == "" {
		log.Logger().Warn().Msgf("Request aborted due to missing query param `digest` : %s", r.RequestURI)
		dockererrors.WriteInvalidDigest(w, blobDigest)
		return
	}

	result, err := rh.svc.handleLastBlobChunk(r.Context(), namespace, repository, blobDigest, sessionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if result.invalid {
		dockererrors.WriteBlobUploadInvalid(w)
		return
	}

	if result.partialUpload {
		dockererrors.WriteBlobUploadNotFound(w)
		return
	}

	writeBlobUploadSuccess(w, r.URL.String(), sessionID, 0, 0)
}

func (rh *RegistryHandler) handleBlobChunk(w http.ResponseWriter, r *http.Request,
	namespace, repository, sessionID string, payload []byte) {
	var start int64 = 0
	var end int64 = 0
	contentRange := r.Header.Get("Content-Range")

	// Content-Range cannot be empty for second+ chunks
	if contentRange == "" {
		// correcting end
		end = int64(len(payload))
	} else {
		var err error
		start, end, err = utils.ParseImageBlobContentRangeFromRequest(contentRange)
		if err != nil {
			log.Logger().Warn().Msg("Unable to parse Content-Range header")
			dockererrors.WriteInvalidRange(w)
			return
		}
	}

	result, err := rh.svc.uploadBlobChunk(r.Context(), namespace, repository, sessionID, start, payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if result.invalid {
		dockererrors.WriteBlobUploadInvalid(w)
		return
	}

	if result.partialUpload {
		dockererrors.WriteBlobUploadNotFound(w)
		return
	}
	writeBlobUploadSuccess(w, r.URL.String(), sessionID, int(start), int(end))
}

func (rh *RegistryHandler) handleBlob(w http.ResponseWriter, r *http.Request,
	namespace, repository, sessionID string, payload []byte) {
	blobDigest := r.URL.Query().Get("digest")
	if blobDigest == "" {
		log.Logger().Warn().Msgf("Request aborted due to missing query param `digest` : %s", r.RequestURI)
		dockererrors.WriteInvalidDigest(w, "")
		return
	}

	res, err := rh.svc.uploadBlobWhole(r.Context(), namespace, repository, sessionID, blobDigest, payload)
	if err != nil {
		log.Logger().Warn().Msgf("Request aborted due to errors")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if res.invalid {
		dockererrors.WriteBlobUploadInvalid(w)
		return
	}

	writeBlobUploadSuccess(w, r.URL.Path, sessionID, 0, len(payload))
}

func (rh *RegistryHandler) getImageBlob(w http.ResponseWriter, r *http.Request) {
	namespace, repository, digest := extractNamespaceRepositoryAndDigest(r)

	exists, content, err := rh.svc.getImageBlob(r.Context(), namespace, repository, digest)
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

func (rh *RegistryHandler) manifestExists(w http.ResponseWriter, r *http.Request) {
	namespace, repository, tagOrDigest := extractNamespaceRepositoryAndTagOrDigest(r)

	exists, mediaType, digest, err := rh.svc.manifestExists(r.Context(), namespace, repository, tagOrDigest)
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

func (rh *RegistryHandler) getManifest(w http.ResponseWriter, r *http.Request) {
	namespace, repository, tagOrDigest := extractNamespaceRepositoryAndTagOrDigest(r)

	exists, mediaType, digest, content, err := rh.svc.getImageManifest(r.Context(), namespace, repository, tagOrDigest)

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

func (rh *RegistryHandler) updateManifest(w http.ResponseWriter, r *http.Request) {
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

	digest, err := rh.svc.updateManifest(r.Context(), namespace, repository, tag, contentType, content)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when updating manifest for request: %s", r.RequestURI)
		return
	}
	w.Header().Set("Content-Length", "0")
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusCreated)
}