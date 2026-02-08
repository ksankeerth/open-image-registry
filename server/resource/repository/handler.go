package repository

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/errors/httperrors"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/resource/access"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
)

type RepositoryHandler struct {
	svc *repositoryService
}

func NewHandler(s store.Store, accessManager *access.Manager) *RepositoryHandler {
	svc := &repositoryService{
		store:         s,
		accessManager: accessManager,
	}
	return &RepositoryHandler{
		svc,
	}
}

func (h *RepositoryHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.createRepository)
	r.Get("/", h.listRepositories)

	r.Route("/{id}", func(r chi.Router) {
		r.Head("/", h.repositoryExists)
		r.Get("/", h.getRepository)
		r.Put("/", h.updateRepository)
		r.Delete("/", h.deleteRepository)
		r.Patch("/state", h.changeState)
		r.Patch("/visibility", h.changeVisiblity)

		r.Get("/users", h.listUserAccess)
		r.Post("/users", h.grantUserAccess)
		r.Delete("/users/{userID}", h.revokeUserAccess)

		// r.Get("/tags", h.listTags) TODO: after https://github.com/ksankeerth/open-image-registry/issues/24
	})
	return r
}

func (h *RepositoryHandler) createRepository(w http.ResponseWriter, r *http.Request) {
	var req mgmt.CreateRepositoryRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to bad request: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	valid, errMsg := validateCreateRepositoryRequest(&req)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	res, err := h.svc.createRepository(r.Context(), &req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if res == nil {
		// unexpected sceanrio when err=nil
		httperrors.InternalError(w, 500, "Unexpected error occurred. Please contact server administrator")
		return
	}

	if res.conflict {
		httperrors.AlreadyExist(w, 409, res.errMsg)
		return
	}

	if res.creatorNotFound {
		httperrors.UnprocessableEntity(w, 422, res.errMsg)
		return
	}

	if res.visiblityIssue {
		httperrors.UnprocessableEntity(w, 422, res.errMsg)
		return
	}

	if res.invalidNs {
		httperrors.UnprocessableEntity(w, 422, res.errMsg)
		return
	}

	response := mgmt.CreateNamespaceResponse{
		Id: res.id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}
}

func (h *RepositoryHandler) listRepositories(w http.ResponseWriter, r *http.Request) {
	cond := lib.ParseListConditions(r, map[string]store.FilterOperator{})

	ok, errMsg := ValidateListRepositoryCondition(cond)
	if !ok {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	repositories, total, err := h.svc.listRepositories(r.Context(), cond)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	res := mgmt.EntityListResponse[*mgmt.RepositoryViewDTO]{
		Total:    total,
		Page:     int(cond.Page),
		Limit:    int(cond.Limit),
		Entities: make([]*mgmt.RepositoryViewDTO, len(repositories)),
	}

	for index, repo := range repositories {
		repoDto := ToRepositoryViewDTO(repo)
		res.Entities[index] = repoDto
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}

}

func (h *RepositoryHandler) repositoryExists(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Passing namespaceId as empty since repository id is enough to validate existense
	exists, err := h.svc.repositoryExists(r.Context(), id, "")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *RepositoryHandler) getRepository(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	model, err := h.svc.getRepository(r.Context(), id, "")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if model == nil {
		httperrors.NotFound(w, 404, "Not found")
		return
	}

	res := makeGetRepositoryResponse(model)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when writing response :%s", r.RequestURI)
	}
}

func (h *RepositoryHandler) updateRepository(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req mgmt.UpdateRepositoryRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Parsing request body of update repository request failed")
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	notFound, err := h.svc.updateRepsitory(r.Context(), id, &req)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if notFound {
		httperrors.NotFound(w, 404, "")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RepositoryHandler) deleteRepository(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	notFound, err := h.svc.deleteRepository(r.Context(), id, "")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if notFound {
		httperrors.NotFound(w, 404, "")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RepositoryHandler) changeState(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	state := r.URL.Query().Get("state")

	if state == "" {
		log.Logger().Warn().Msg("Changing repository state request was rejected due to empty state")
		httperrors.BadRequest(w, 400, "Missing query param state in request")
		return
	}

	if !(state == constants.ResourceStateActive || state == constants.ResourceStateDeprecated ||
		state == constants.ResourceStateDisabled) {
		log.Logger().Warn().Msgf("Changing repository state request was rejected due to invalid repository state '%s'", state)
		httperrors.BadRequest(w, 400, "Invalid repository state")
		return
	}

	result, err := h.svc.changeState(r.Context(), id, state)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}
	if result.success {
		w.WriteHeader(http.StatusOK)
		return
	}

	httperrors.SendError(w, result.httpStatusCode, result.httpErrorMsg)

	if result.httpStatusCode == http.StatusNotFound {
		httperrors.NotFound(w, 404, result.httpErrorMsg)
		return
	}

	httperrors.NotAllowed(w, 403, result.httpErrorMsg)
}

func (h *RepositoryHandler) changeVisiblity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	visibility := r.URL.Query().Get("public")
	isPublic, err := strconv.ParseBool(visibility)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Changing visibility of namespace(%s) failed due to invalid query param", visibility)
		httperrors.BadRequest(w, 400, "Invalid query param")
		return
	}

	result, err := h.svc.changeVisiblity(r.Context(), id, isPublic)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}
	if result.success {
		w.WriteHeader(http.StatusOK)
		return
	}

	httperrors.SendError(w, result.httpStatusCode, result.httpErrorMsg)
}

func (h *RepositoryHandler) listUserAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cond := lib.ParseListConditions(r, map[string]store.FilterOperator{})

	ok, errMsg := access.ValidateListUserAccessCondition(cond)
	if !ok {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	userAccesses, total, err := h.svc.listUsers(r.Context(), id, cond)

	res := mgmt.EntityListResponse[*mgmt.ResourceAccessViewDTO]{
		Total:    total,
		Page:     int(cond.Page),
		Limit:    int(cond.Limit),
		Entities: make([]*mgmt.ResourceAccessViewDTO, len(userAccesses)),
	}

	for index, acs := range userAccesses {
		dto := access.ToResourceAccessViewDTO(acs)
		res.Entities[index] = dto
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}

}

// TODO: after implementing https://github.com/ksankeerth/open-image-registry/issues/24

// func (h *RepositoryHandler) listTags(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "id")

// 	cond := lib.ParseListConditions(r, map[string]store.FilterOperator{})

// 	ok, errMsg := resource.ValidateListUserAccessCondition(cond)
// 	if !ok {
// 		httperrors.BadRequest(w, 400, errMsg)
// 		return
// 	}
// }

func (h *RepositoryHandler) grantUserAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var request mgmt.AccessGrantRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Parsing request failed : %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	// Validate URL and body ID match
	if id != request.ResourceID {
		log.Logger().Warn().Msg("'identifier' in URL and body don't match")
		httperrors.BadRequest(w, 400, "'identifier' in URL and body don't match")
		return
	}

	valid, errMsg := validateRepositoryGrantAccessRequest(&request)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	granter := r.Context().Value(constants.ContextUsername).(string)
	if granter != request.GrantedBy {
		httperrors.NotAllowed(w, 403, "Grantor must be same as authenticated user!")
		return
	}

	result, err := h.svc.grantAccess(r.Context(), &request)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if result.statusCode != http.StatusOK {
		http.Error(w, result.errMsg, result.statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *RepositoryHandler) revokeUserAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userID")

	request := mgmt.AccessRevokeRequest{
		ResourceType: constants.ResourceTypeRepository,
		ResourceID:   id,
		UserID:       userID,
	}

	statusCode, msg, err := h.svc.revokeAccess(r.Context(), &request)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if statusCode == http.StatusOK {
		w.WriteHeader(http.StatusOK)
		return
	}

	httperrors.SendError(w, statusCode, msg)
}