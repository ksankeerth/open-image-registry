package api

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/db"
	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/handlers/common"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types"
)

type UpstreamRegistryHandler struct {
	dao db.UpstreamDAO
}

func NewUpstreamRegistryHandler(dao *db.UpstreamDAO) *UpstreamRegistryHandler {
	return &UpstreamRegistryHandler{
		dao: *dao,
	}
}

func (h *UpstreamRegistryHandler) CreateUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	var reqBody types.CreateUpstreamRegRequestMsg

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Err(err).Msg("Unable to decode JSON request body for POST /api/v1/upstreams")
		common.HandleBadRequest(w, http.StatusBadRequest, "Cannot undestrand request")
		return
	}

	isValid, failureMsg := validateCreateUpstreamRegistryRequest(&reqBody)
	if !isValid {
		common.HandleBadRequest(w, http.StatusBadRequest, failureMsg)
		return
	}

	regId, regName, err := h.dao.CreateUpstreamRegistry(&reqBody.UpstreamOCIRegEntity, &reqBody.AuthConfig,
		&reqBody.AccessConfig, &reqBody.StorageConfig, &reqBody.CacheConfig)

	if err != nil {
		if yes, column := db_errors.IsUniqueConstraint(err); yes {
			common.HandleAlreadyExist(w, 409, fmt.Sprintf("OCI Upstream Registry with same %s already exists.", column))
			return
		}
		common.HandleInternalError(w, http.StatusInternalServerError, "Error occured while persisting data")
		return
	}
	
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{
		"reg_id": %s,
		"reg_name": %s
	}`, regId, regName)))
}

func (h *UpstreamRegistryHandler) UpdateUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	var reqBody types.UpdateUpstreamRegRequestMsg

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Err(err).Msgf("Unable to decode JSON request body for PUT %s", r.RequestURI)
		common.HandleBadRequest(w, http.StatusBadRequest, "Cannot undestrand request")
		return
	}

	registryId := chi.URLParam(r, "registry_id")
	if registryId != reqBody.Id {
		log.Logger().Error().Msgf("Updating upstream oci registery failed due to unmatching registery ids in body and path")
		common.HandleBadRequest(w, http.StatusBadRequest, "Invalid request")
	}

	isValid, failureMsg := validateCreateUpstreamRegistryRequest(&reqBody)
	if !isValid {
		common.HandleBadRequest(w, http.StatusBadRequest, failureMsg)
		return
	}

	err = h.dao.UpdateUpstreamRegistry(registryId, &reqBody.UpstreamOCIRegEntity, &reqBody.AuthConfig,
		&reqBody.AccessConfig, &reqBody.StorageConfig, &reqBody.CacheConfig)
	if err != nil {
		common.HandleInternalError(w, http.StatusInternalServerError, fmt.Sprintf("Error occured while updating upstream registry: %s", registryId))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UpstreamRegistryHandler) DeleteUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	regId := chi.URLParam(r, "registry_id")
	err := h.dao.DeleteUpstreamRegistry(regId)
	if err != nil {
		if db_errors.IsNotFound(err) {
			common.HandleNotFound(w, http.StatusNotFound, "")
			return
		}
		common.HandleInternalError(w, 500, "Error occured when deleting upstream registery")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UpstreamRegistryHandler) GetUpstreamRegistry(w http.ResponseWriter, r *http.Request) {
	regId := chi.URLParam(r, "registry_id")
	reg, access, auth, cache, storage, err := h.dao.GetUpstreamRegistryWithConfig(regId)
	if err != nil {
		if db_errors.IsNotFound(err) {
			common.HandleNotFound(w, 404, fmt.Sprintf("No registery found with : %s", regId))
			return
		}

		common.HandleInternalError(w, 500, "Unable to fetch Upstream OCI registery")
		return
	}
	res := types.UpstreamOCIRegResMsg{
		UpstreamOCIRegEntity: *reg,
		AuthConfig:           *auth,
		AccessConfig:         *access,
		StorageConfig:        *storage,
		CacheConfig:          *cache,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response to GET %s", r.RequestURI)
	}
}

// ListUpstreamRegisteries returns all the upstream registeries. In future, We may support filter
// for advanced use-cases. #TODO
func (h *UpstreamRegistryHandler) ListUpstreamRegisteries(w http.ResponseWriter, r *http.Request) {
	registeries, err := h.dao.ListUpstreamRegistries()
	if err != nil {
		if db_errors.IsNotFound(err) {
			common.HandleNotFound(w, http.StatusNotFound, "No registeries found")
			return
		}
		ok, errCode := db_errors.UnwrapDBError(err)
		if ok {
			common.HandleInternalError(w, errCode, "Error occured in DB. Check the logs")
			return
		}
		common.HandleInternalError(w, http.StatusInternalServerError, "Error occured. Check the logs")
		return
	}

	res := types.ListUpstreamRegistriesResponseMsg{
		Total:       len(registeries),
		Limit:       len(registeries),
		Page:        1,
		Registeries: registeries,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Str("path", r.URL.Path).Msgf("Failed to write JSON response")
	}
}

func validateCreateUpstreamRegistryRequest(req *types.CreateUpstreamRegRequestMsg) (bool, string) {
	if req.Name == "" {
		return false, "Registry Name is not provided"
	}

	if req.Port <= 1024 && req.Port > 65535 {
		return false, "Not allowed to choose port outside the allowed range[1025-65535]"
	}

	if req.UpstreamUrl == "" {
		return false, "Upstream Registry URL is not provided"
	}

	// TODO: support Oauth2 later
	if !(req.AuthConfig.AuthType == "anonymous" || req.AuthConfig.AuthType == "basic" || req.AuthConfig.AuthType == "bearer" ||
		req.AuthConfig.AuthType == "mtls") {

	}

	if req.AuthConfig.AuthType == "anonymous" && (req.AuthConfig.CredentialJson != nil ||
		len(req.AuthConfig.CredentialJson) != 0) {
		return false, "Unnecessary credentials are provided for annonymous access"
	}

	if req.AuthConfig.AuthType == "basic" && !(req.AuthConfig.CredentialJson != nil &&
		req.AuthConfig.CredentialJson["username"] != "" && req.AuthConfig.CredentialJson["password"] != "") {
		return false, "Username and password are mandatory for Basic auth access"
	}

	if req.AuthConfig.AuthType == "bearer" && (req.AuthConfig.CredentialJson != nil &&
		req.AuthConfig.CredentialJson["token"] == "") {
		return false, "Token is mandatory for Bearer auth access"
	}
	if req.AuthConfig.AuthType == "mtls" && (req.AuthConfig.Certificate == "" || !isValidCertificate(req.AuthConfig.Certificate)) {
		return false, "Valid Certificate is mandatory for mtls"
	}

	if req.AccessConfig.ProxyEnabled && !isValidUrl(req.AccessConfig.ProxyUrl) {
		return false, "Proxy is enabled but proxy server URL is not provided"
	}

	if !isInRange(req.AccessConfig.ConnectionTimeoutInSeconds, 1, 300) {
		return false, "Given Connection timeout is not allowed. Allowed range is 1 - 300 seconds."
	}

	if !isInRange(req.AccessConfig.ReadTimeoutInSeconds, 1, 600) {
		return false, "Given Read timeout is not allowed. Allowed range is 1 - 600 seconds."
	}

	if !isInRange(req.AccessConfig.MaxRetries, 0, 10) {
		return false, "Given Max Retries is not allowed. Allowed range is 0 - 10."
	}

	if !isInRange(req.AccessConfig.RetryDelayInSeconds, 1, 60) {
		return false, "Given Retry delay is not allowed. Allowed range is 1 - 60 seconds."
	}

	if !isInRangeFloat(req.StorageConfig.StorageLimitInMbs, 100, 102400) {
		return false, "Given Storage Limit is not allowed. Allowed limit is 100MB - 100GB"
	}

	if !isOneOf(req.StorageConfig.CleanupPolicy, []string{"lru_1m", "lru_3m", "lp"}) {
		return false, "Incorrect Cleanup policy."
	}

	if !isInRange(req.CacheConfig.TtlInSeconds, 600, 31536000) {
		return false, "Given Cache TTL is not allowed. Allowed range is 600 - 31536000 seconds."
	}

	return true, ""
}

// TODO: Move this function to different files
func isValidCertificate(certificate string) bool {
	if certificate == "" {
		return false
	}

	certificate = strings.TrimSpace(certificate)

	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return false
	}

	if block.Type != "CERTIFICATE" {
		return false
	}

	_, err := x509.ParseCertificate(block.Bytes)
	return err == nil
}

func isValidUrl(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return false
	}

	return isOneOf(strings.ToLower(parsedUrl.Scheme), []string{"http", "https"})
}

func isInRange(value int, min int, max int) bool {
	return value >= min && value <= max
}

func isInRangeFloat(value float32, min float32, max float32) bool {
	return value >= min && value <= max
}

func isOneOf(value string, allowedValues []string) bool {
	for _, av := range allowedValues {
		if av == value {
			return true
		}
	}
	return false
}
