package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/access/resource"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/errors/httperrors"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
)

type UserAPIHandler struct {
	svc     *userService
	adapter *UserAdapter
}

// NewUserAPIHandler creates a new user API handler
func NewUserAPIHandler(store store.Store, emailClient *email.EmailClient) *UserAPIHandler {
	return &UserAPIHandler{
		svc: &userService{
			store:         store,
			adapter:       &UserAdapter{},
			ec:            emailClient,
			accessManager: resource.NewManager(store),
		},
	}
}

// ListUsers handles GET /api/v1/users
func (h *UserAPIHandler) ListUsers(w http.ResponseWriter, r *http.Request) {

	cond := lib.ParseListConditions(r, map[string]store.FilterOperator{})

	ok, errMsg := ValidateListUserCondition(cond)
	if !ok {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	users, total, err := h.svc.getUserList(cond)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Database errors occured")
		return
	}

	res := mgmt.EntityListResponse[*mgmt.UserAccountViewDTO]{
		Total:    total,
		Page:     int(cond.Page),
		Limit:    int(cond.Limit),
		Entities: make([]*mgmt.UserAccountViewDTO, len(users)),
	}

	for index, user := range users {
		userDto := h.adapter.ToUserAccountViewDTO(user)
		res.Entities[index] = userDto
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}
}

// CreateUser handles POST /api/v1/users
func (h *UserAPIHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var reqBody mgmt.CreateUserAccountRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when reading json request body: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Invalid payload")
		return
	}

	valid, errrMsg := ValidateCreateUserAccount(&reqBody)
	if !valid {
		httperrors.BadRequest(w, 400, errrMsg)
		return
	}

	userId, conflict, recoveryUuid, err := h.svc.createUserAccount(reqBody.Username, reqBody.Email, reqBody.DisplayName, reqBody.Role)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors")
		httperrors.InternalError(w, 500, "")
		return
	}

	if conflict {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		return
	}

	res := mgmt.CreateUserAccountResponse{
		Username: reqBody.Username,
		UserId:   userId,
	}

	w.Header().Set("Content-Type", "application/json")
	if recoveryUuid != "" {
		w.Header().Set("Account-Setup-Id", recoveryUuid)
	}
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when writing response to create user account request")
	}
}

// UpdateUser handles PUT /api/v1/users/{id}
func (h *UserAPIHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req mgmt.UpdateUserAccountRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when reading user account update request: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Invalid request body")
		return
	}

	ok, errMsg := ValidateUpdateUserAccount(&req)
	if !ok {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	userId := chi.URLParam(r, "id")

	err = h.svc.updateUserAccount(r.Context(), userId, req.DisplayName)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to database errors")
		httperrors.InternalError(w, 500, "Request aborted due to database errors")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ValidateUser checks the given username and email POST /api/v1/users/validate
func (h *UserAPIHandler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	var req mgmt.UsernameEmailValidationRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when validating username and email")
		httperrors.BadRequest(w, 400, "Bad request")
		return
	}

	valid, errMsg := ValidateUserValidateRequest(&req)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	res := mgmt.UsernameEmailValidationResponse{}

	res.UsernameAvailable, res.EmailAvailable, err = h.svc.validateUser(req.Username, req.Email)
	if err != nil {
		httperrors.InternalError(w, 500, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when writing response to validate username and email request")
	}
}

// UpdateUserEmail handles PUT /api/v1/users/{id}/email
func (h *UserAPIHandler) UpdateUserEmail(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	var reqBody mgmt.UpdateUserEmailRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Parsing request body caused errors")
		httperrors.BadRequest(w, 400, "Invalid payload")
		return
	}

	if userId != reqBody.UserId {
		log.Logger().Error().Msgf("Update user email request with two different user ids were detected; Path: %s, Body: %s", userId, reqBody.UserId)
		httperrors.BadRequest(w, 400, "Invalid body and path parameters")
		return
	}

	valid, errMsg := ValidateUpdateUserEmail(&reqBody)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	err = h.svc.updateUserEmail(userId, reqBody.Email)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateUserDisplayName handles PUT /api/v1/users/{id}/role
func (h *UserAPIHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	var reqBody mgmt.ChangeUserRoleRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Parsing request body caused errors")
		httperrors.BadRequest(w, 400, "Invalid payload")
		return
	}

	if !isValidRole(reqBody.Role) {
		log.Logger().Error().Err(err).Msgf("Changing role request failed due to invalid role: %s", reqBody.Role)
		httperrors.BadRequest(w, 400, "Invalid role")
		return
	}

	errMsg, err := h.svc.changeRole(r.Context(), reqBody.Role, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if errMsg != "" {
		httperrors.NotAllowed(w, 403, errMsg)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (h *UserAPIHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	err := h.svc.deleteUserAccount(userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to database errors")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ChangePassword handles PUT /api/v1/users/{id}/password
func (h *UserAPIHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var body mgmt.PasswordChangeRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when pasing request body: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Request aborted due to parsing errors")
		return
	}

	userId := chi.URLParam(r, "id")
	if userId != body.UserId {
		log.Logger().Error().Msgf("Password change request with two different user ids were detected; Path: %s, Body: %s", userId, body.UserId)
		httperrors.BadRequest(w, 400, "Invalid body and path parameter")
		return
	}

	res, err := h.svc.changePassword(&body)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	response := h.adapter.ToChangePasswordResposne(res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing response: %s", r.RequestURI)
	}
}

// GetCurrentUser handles GET /api/v1/users/me
func (h *UserAPIHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Implementation needed
}

// UpdateCurrentUser handles PUT /api/v1/users/me
func (h *UserAPIHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Implementation needed
}

// LockUser handles PUT /api/v1/users/{id}/lock
func (h *UserAPIHandler) LockUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	res, err := h.svc.lockUserAccount(userId)
	if err != nil {
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if res.alreadyLocked {
		httperrors.AlreadyExist(w, 409, "Already locked")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UnlockUser handles PUT /api/v1/users/{id}/unlock
func (h *UserAPIHandler) UnlockUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	res, err := h.svc.unlockUserAccount(userId)
	if err != nil {
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if res.newAccount {
		httperrors.AlreadyExist(w, 409, "Not allowed to unlock new user account")
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (h *UserAPIHandler) GetUserAccountSetupInfo(w http.ResponseWriter, r *http.Request) {
	recoveryId := chi.URLParam(r, "uuid")
	res, err := h.svc.getAccountSetupInfo(recoveryId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	// Even if the given recovery id is found in db with a different reason,
	// We'll respond to user with 404. To avoid leaking unneccessary data to
	// annonymous users
	if !res.found || res.errorMsg != "" {
		httperrors.NotFound(w, 404, res.errorMsg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := h.adapter.toUserAccountSetupVerficationResponse(res)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Logger().Err(err).Msgf("Error occurred when writing response to request: %s", r.RequestURI)
	}
}

func (h *UserAPIHandler) CompleteUserAccountSetup(w http.ResponseWriter, r *http.Request) {
	var req mgmt.AccountSetupCompleteRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error in parsing request: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Invalid request body")
		return
	}

	id := chi.URLParam(r, "uuid")
	if req.Uuid != id {
		log.Logger().Error().Msgf("Request UUID doesn't match with request body")
		httperrors.BadRequest(w, 400, "Request UUID doesn't match with request body")
		return
	}

	valid, errMsg := ValidateAccountSetupCompleteRequest(&req)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	err = h.svc.completeAccountSetup(&req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors:")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserAPIHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, role, err := h.svc.getUser(r.Context(), id)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if user == nil {
		httperrors.NotFound(w, 404, "")
		return
	}

	response := h.adapter.makeGetUserResponse(user, role)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Logger().Err(err).Msgf("Error occurred when writing response to request: %s", r.RequestURI)
	}

}

func (h *UserAPIHandler) OnboardingRoutes() chi.Router {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Get("/{uuid}", h.GetUserAccountSetupInfo)
		r.Post("/{uuid}/complete", h.CompleteUserAccountSetup)
	})

	return router
}

func (h *UserAPIHandler) Routes() chi.Router {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/me", h.GetCurrentUser)
		r.Put("/me", h.UpdateCurrentUser)
		r.Post("/validate", h.ValidateUser)

		r.Put("/{id}/email", h.UpdateUserEmail)
		r.Put("/{id}/role", h.ChangeRole)
		r.Put("/{id}/password", h.ChangePassword)
		r.Put("/{id}", h.UpdateUser)
		r.Put("/{id}/lock", h.LockUser)
		r.Put("/{id}/unlock", h.UnlockUser)

		r.Delete("/{id}", h.DeleteUser)
		r.Get("/{id}", h.GetUser)
		r.Post("/", h.CreateUser)
		r.Get("/", h.ListUsers)

	})
	return router
}
