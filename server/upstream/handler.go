package upstream

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/db"
	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/errors/httperrors"

	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
)

type UpstreamRegistryHandler struct {
	svc *upstreamService
}

func NewUpstreamRegistryHandler(dao db.UpstreamDAO) *UpstreamRegistryHandler {

	return &UpstreamRegistryHandler{
		svc: &upstreamService{
			upstreamDao: dao,
			adapter:     &UpstreamAdapter{},
		},
	}
}

func (h *UpstreamRegistryHandler) Routes() chi.Router {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Put("/{registry_id}", h.UpdateUpstreamRegistry)
		r.Get("/{registry_id}", h.GetUpstreamRegistry)
		r.Delete("/{registry_id}", h.DeleteUpstreamRegistry)
		r.Post("/", h.CreateUpstreamRegistry)
		r.Get("/", h.ListUpstreamRegistries)
	})
	return router
}

func (h *UpstreamRegistryHandler) CreateUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	var reqBody mgmt.CreateUpstreamRegistryRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Err(err).Msg("Unable to decode JSON request body for POST /api/v1/upstreams")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isValid, failureMsg := ValidateCreateRequest(&reqBody)
	if !isValid {
		httperrors.BadRequest(w, 400, failureMsg)
		return
	}

	registryId, registryName, err := h.svc.createUpstreamRegistry(&reqBody)
	if err != nil {
		if yes, column := db_errors.IsUniqueConstraint(err); yes {
			httperrors.AlreadyExist(w, 409, fmt.Sprintf("Upstream Registry with same %s already exists.", column))
			return
		}
		httperrors.InternalError(w, 500, "Error occured when persisting upstream registry")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(mgmt.CreateUpstreamRegistryResponse{
		RegId:   registryId,
		RegName: registryName,
	})
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response after upstream registry is created: %s", registryId)
	}
}

func (h *UpstreamRegistryHandler) UpdateUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	var reqBody mgmt.UpdateUpstreamRegistryRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Err(err).Msgf("Unable to decode JSON request body for PUT %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Cannot understand request")
		return
	}

	registryId := chi.URLParam(r, "registry_id")
	if registryId != reqBody.RegId {
		log.Logger().Error().Msgf("Updating upstream oci registery failed due to unmatching registery ids in body and path")
		httperrors.BadRequest(w, 400, "Updating upstream oci registery failed due to unmatching registery ids in body and path")
		return
	}

	isValid, failureMsg := ValidateCreateRequest(&reqBody.CreateUpstreamRegistryRequest)
	if !isValid {
		httperrors.BadRequest(w, 400, failureMsg)
		return
	}

	err = h.svc.updateUpstreamRegistry(&reqBody)
	if err != nil {
		httperrors.InternalError(w, 500, fmt.Sprintf("Error occurred while updating upstream registry: %s", registryId))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UpstreamRegistryHandler) DeleteUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	regId := chi.URLParam(r, "registry_id")

	notFound, err := h.svc.deleteUpstreamRegistry(regId)
	if notFound {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		httperrors.InternalError(w, 500, "Error occured when deleting upstream registery")
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *UpstreamRegistryHandler) GetUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	regId := chi.URLParam(r, "registry_id")

	reg, access, auth, cache, storage, err := h.svc.getUpstreamRegistryWithConfig(regId)
	if err != nil {
		if db_errors.IsNotFound(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		httperrors.InternalError(w, 500, "Unable to retrieve upstream registry from database")
		return
	}

	res := mgmt.UpstreamRegistryResponse{
		Id:          reg.Id,
		Name:        reg.Name,
		Port:        reg.Port,
		Status:      reg.Status,
		UpstreamUrl: reg.UpstreamUrl,
		CreatedAt:   reg.CreatedAt,
		UpdatedAt:   reg.UpdatedAt,

		AuthConfig: mgmt.UpstreamAuthConfigResponse{
			UpstreamAuthConfigDTO: mgmt.UpstreamAuthConfigDTO{
				AuthType:       auth.AuthType,
				CredentialJson: auth.CredentialJson,
				TokenEndpoint:  auth.TokenEndpoint,
			},
			CreatedAt: reg.CreatedAt,
			UpdatedAt: auth.UpdatedAt,
		},
		AccessConfig: mgmt.UpstreamAccessConfigResponse{
			UpstreamAccessConfigDTO: mgmt.UpstreamAccessConfigDTO{
				ProxyEnabled:               access.ProxyEnabled,
				ProxyUrl:                   access.ProxyUrl,
				ConnectionTimeoutInSeconds: access.ConnectionTimeoutInSeconds,
				ReadTimeoutInSeconds:       access.ReadTimeoutInSeconds,
				MaxConnections:             access.MaxConnections,
				MaxRetries:                 access.MaxRetries,
				RetryDelayInSeconds:        access.RetryDelayInSeconds,
			},
			CreatedAt: reg.CreatedAt,
			UpdatedAt: access.UpdatedAt,
		},
		StorageConfig: mgmt.UpstreamStorageConfigResponse{
			UpstreamStorageConfigDTO: mgmt.UpstreamStorageConfigDTO{
				StorageLimitInMbs: storage.StorageLimitInMbs,
				CleanupPolicy:     storage.CleanupPolicy,
				CleanupThreshold:  storage.CleanupThreshold,
			},
			CreatedAt: reg.CreatedAt,
			UpdatedAt: storage.UpdatedAt,
		},
		CacheConfig: mgmt.UpstreamCacheConfigResponse{
			UpstreamCacheConfigDTO: mgmt.UpstreamCacheConfigDTO{
				Enabled:      cache.Enabled,
				TtlInSeconds: cache.TtlInSeconds,
				OfflineMode:  cache.OfflineMode,
			},
			CreatedAt: reg.CreatedAt,
			UpdatedAt: *cache.UpdatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when writing response to GET %s", r.RequestURI)
	}
}

// ListUpstreamRegisteries returns all the upstream registeries. In future, We may support filter
// for advanced use-cases. #TODO
func (h *UpstreamRegistryHandler) ListUpstreamRegistries(w http.ResponseWriter, r *http.Request) {
	total, limit, page, registeries, err := h.svc.listUpstreamRegisteries()
	if err != nil {
		if db_errors.IsNotFound(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		httperrors.InternalError(w, 500, "Error occured when loading upstream registeries")
		return
	}

	res := mgmt.ListUpstreamsResponse{
		Total:      total,
		Limit:      limit,
		Page:       page,
		Registries: registeries,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Str("path", r.URL.Path).Msgf("Failed to write JSON response")
	}
}