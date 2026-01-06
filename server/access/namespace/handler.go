package namespace

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/access/repository"
	"github.com/ksankeerth/open-image-registry/access/resource"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/errors/httperrors"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
)

type NamespaceHandler struct {
	svc *namespaceService
}

func NewHandler(s store.Store, accessManager *resource.Manager) *NamespaceHandler {
	svc := &namespaceService{
		store:         s,
		accessManager: accessManager,
	}
	return &NamespaceHandler{
		svc,
	}
}

func (h *NamespaceHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.createNamespace)
	r.Get("/", h.listNamespaces)
	// identifier can be uuid value or name.
	r.Route("/{identifier}", func(r chi.Router) {
		r.Head("/", h.namespaceExists)
		r.Get("/", h.getNamespace)
		r.Put("/", h.updateNamespace)
		r.Delete("/", h.deleteNamespace)
		r.Patch("/state", h.changeState)
		r.Patch("/visibility", h.changeVisiblity)

		r.Get("/users", h.listUserAccess)
		r.Get("/repositories", h.listRepositories)

		r.Post("/users", h.grantUserAccess)             // strictly use id instead of name
		r.Delete("/users/{userID}", h.revokeUserAccess) // strictly use id instead of name
	})

	return r
}

func (h *NamespaceHandler) listNamespaces(w http.ResponseWriter, r *http.Request) {
	cond := lib.ParseListConditions(r, map[string]store.FilterOperator{})

	ok, errMsg := validateListNamespaceCondition(cond)
	if !ok {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	namespaces, total, err := h.svc.listNamespaces(r.Context(), cond)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	res := mgmt.ListNamespacesResponse{
		Total:      total,
		Page:       int(cond.Page),
		Limit:      int(cond.Limit),
		Namespaces: make([]*mgmt.NamespaceViewDTO, len(namespaces)),
	}

	for index, ns := range namespaces {
		nsDto := toNamespaceViewDTO(ns)
		res.Namespaces[index] = nsDto
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}
}

func (h *NamespaceHandler) createNamespace(w http.ResponseWriter, r *http.Request) {
	var req mgmt.CreateNamespaceRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to bad request: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	valid, errMsg := validateCreateNamespaceRequest(&req)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	res, err := h.svc.createNamespace(r.Context(), &req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if res.statusCode == http.StatusCreated {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		response := mgmt.CreateNamespaceResponse{
			Id: res.nsId,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occurred when writing response :%s", r.RequestURI)
		}
		return
	}

	httperrors.SendError(w, res.statusCode, res.errMsg)
}

func (h *NamespaceHandler) namespaceExists(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")

	exists, err := h.svc.namespaceExists(r.Context(), identifier)
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

func (h *NamespaceHandler) deleteNamespace(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")

	notFound, err := h.svc.deleteNamespace(r.Context(), identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if notFound {
		httperrors.NotFound(w, 404, "Namespace not found")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *NamespaceHandler) getNamespace(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")

	ns, err := h.svc.getNamespace(r.Context(), identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if ns == nil {
		httperrors.NotFound(w, 404, "Not found")
		return
	}

	res := makeGetNamespaceResponse(ns)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when writing response :%s", r.RequestURI)
	}
}

func (h *NamespaceHandler) updateNamespace(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")

	var request mgmt.UpdateNamespaceRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Parsing request body of update namespace request failed")
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	valid, errMsg := validateUpdateNamespaceRequest(&request)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	notFound, err := h.svc.updateNamespace(r.Context(), identifier, &request)
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

func (h *NamespaceHandler) changeState(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")
	state := r.URL.Query().Get("state")

	if state == "" {
		log.Logger().Warn().Msg("Changing namespace state request was rejected due to empty state")
		httperrors.BadRequest(w, 400, "Missing query param state in request")
		return
	}

	if !(state == constants.ResourceStateActive || state == constants.ResourceStateDeprecated ||
		state == constants.ResourceStateDisabled) {
		log.Logger().Warn().Msgf("Changing namespace state request was rejected due to invalid namespace state '%s'", state)
		httperrors.BadRequest(w, 400, "Invalid namespace state")
		return
	}

	result, err := h.svc.changeState(r.Context(), identifier, state)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}
	if result.success {
		w.WriteHeader(http.StatusOK)
		return
	}

	if result.httpStatusCode == http.StatusNotFound {
		httperrors.NotFound(w, 404, result.httpErrorMsg)
		return
	}

	httperrors.NotAllowed(w, 403, result.httpErrorMsg)
}

func (h *NamespaceHandler) changeVisiblity(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")
	visibility := r.URL.Query().Get("public")

	isPublic, err := strconv.ParseBool(visibility)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Changing visibility of namespace(%s) failed due to invalid query param", visibility)
		httperrors.BadRequest(w, 400, "Invalid query param")
		return
	}

	result, err := h.svc.changeVisiblity(r.Context(), identifier, isPublic)

	if result.success {
		w.WriteHeader(http.StatusOK)
		return
	}

	if result.httpStatusCode == http.StatusNotFound {
		httperrors.NotFound(w, 404, result.httpErrorMsg)
		return
	}

	httperrors.NotAllowed(w, 403, result.httpErrorMsg)
}

func (h *NamespaceHandler) grantUserAccess(w http.ResponseWriter, r *http.Request) {
	var request mgmt.AccessGrantRequest
	id := chi.URLParam(r, "identifier")

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Parsing request failed : %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	valid, errMsg := validateNamespaceGrantAccessRequest(&request)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	if id != request.ResourceID {
		log.Logger().Warn().Msgf("Resource ID in request body does not match the ID in the URL path")
		httperrors.BadRequest(w, 400, "Resource ID in request body does not match the ID in the URL path")
		return
	}

	result, err := h.svc.grantAccess(r.Context(), &request)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if result.statusCode != http.StatusOK {
		httperrors.SendError(w, result.statusCode, result.errMsg)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *NamespaceHandler) revokeUserAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "identifier")

	var request mgmt.AccessRevokeRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Parsing request failed : %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	valid, errMsg := validateNamesapceRevokeRequest(&request)
	if !valid {
		httperrors.BadRequest(w, 404, errMsg)
		return
	}

	if id != request.ResourceID {
		log.Logger().Warn().Msgf("Resource ID in request body does not match the ID in the URL path")
		httperrors.BadRequest(w, 400, "Resource ID in request body does not match the ID in the URL path")
		return
	}

	statusCode, msg, err := h.svc.revokeAccess(r.Context(), &request)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if statusCode == 200 {
		w.WriteHeader(http.StatusOK)
		return
	}

	httperrors.SendError(w, statusCode, msg)
}

func (h *NamespaceHandler) listRepositories(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")

	cond := lib.ParseListConditions(r, map[string]store.FilterOperator{})

	ok, errMsg := repository.ValidateListRepositoryCondition(cond)
	if !ok {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	repositories, total, err := h.svc.listRepositories(r.Context(), identifier, cond)

	res := mgmt.ListRepositoriesResponse{
		Total:        total,
		Page:         int(cond.Page),
		Limit:        int(cond.Limit),
		Repositories: make([]*mgmt.RepositoryViewDTO, len(repositories)),
	}

	for index, repo := range repositories {
		repoDto := repository.ToRepositoryViewDTO(repo)
		res.Repositories[index] = repoDto
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}
}

func (h *NamespaceHandler) listUserAccess(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")

	cond := lib.ParseListConditions(r, map[string]store.FilterOperator{})

	ok, errMsg := resource.ValidateListUserAccessCondition(cond)
	if !ok {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	userAccesses, total, err := h.svc.listUsers(r.Context(), identifier, cond)

	res := mgmt.ListResourceAccessResponse{
		Total:    total,
		Page:     int(cond.Page),
		Limit:    int(cond.Limit),
		Accesses: make([]*mgmt.ResourceAccessViewDTO, len(userAccesses)),
	}

	for index, access := range userAccesses {
		dto := resource.ToResourceAccessViewDTO(access)
		res.Accesses[index] = dto
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}
}
