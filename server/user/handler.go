package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/db"
	"github.com/ksankeerth/open-image-registry/errors/httperrors"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/utils"
)

type UserAPIHandler struct {
	svc     *userService
	adapter *UserAdapter
}

// NewUserAPIHandler creates a new user API handler
func NewUserAPIHandler(userDao db.UserDAO, accessDao db.ResourceAccessDAO,
	emailClient *email.EmailClient) *UserAPIHandler {
	return &UserAPIHandler{
		svc: &userService{
			userDao:   userDao,
			accessDao: accessDao,
			adapter:   &UserAdapter{},
			ec:        emailClient,
		},
	}
}

// ListUsers handles GET /api/v1/users
func (h *UserAPIHandler) ListUsers(w http.ResponseWriter, r *http.Request) {

	cond := lib.ParseListConditions(r)

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

	res := mgmt.ListUsersResponse{
		Total: total,
		Page:  int(cond.Pagination.Page),
		Limit: int(cond.Pagination.Limit),
		Users: make([]*mgmt.UserAccountViewDTO, len(users)),
	}

	for index, user := range users {
		userDto := h.adapter.ToUserAccountViewDTO(user)
		res.Users[index] = userDto
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

	userId, conflict, err := h.svc.createUserAccount(reqBody.Username, reqBody.Email, reqBody.DisplayName, reqBody.Role)
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

	err = h.svc.updateUserAccount(userId, req.Email, req.DisplayName, req.Role)
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

func (h *UserAPIHandler) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	var req mgmt.PasswordValidationRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when parsing request body: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Unable to parse the request")
		return
	}

	valid, msg := utils.ValidatePassword(req.Password)
	response := mgmt.PasswordValidationResponse{
		IsValid: valid,
		Msg:     msg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when writing json response : %s", r.RequestURI)
	}
}

// GetUser handles GET /api/v1/users/{id}
// func (h *UserAPIHandler) GetUser(w http.ResponseWriter, r *http.Request) {
// 	userId := chi.URLParam(r, "id")

// 	userAccount, err := h.svc.getUserAccount(userId)
// 	if userAccount == nil && err == nil {
// 		httperrors.NotFound(w, 404, "No user account found")
// 		return
// 	}

// 	if err != nil {
// 		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
// 		httperrors.InternalError(w, 500, "Request aborted due to database errors")
// 		return
// 	}

// 	userDto := h.adapter.ToUserDTO(userAccount)

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	err = json.NewEncoder(w).Encode(userDto)
// 	if err != nil {
// 		log.Logger().Error().Err(err).Msg("Error occurred when writing response to get user account request")
// 	}
// }

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

	updated, err := h.svc.updateUserEmail(userId, reqBody.Email)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if updated {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// UpdateUserDisplayName handles PUT /api/v1/users/{id}/display-name
func (h *UserAPIHandler) UpdateUserDisplayName(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	var reqBody mgmt.UpdateUserDisplayNameRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Parsing request body caused errors")
		httperrors.BadRequest(w, 400, "Invalid payload")
		return
	}

	if userId != reqBody.UserId {
		log.Logger().Error().Msgf("Update user display name request with two different user ids were detected; Path: %s, Body: %s", userId, reqBody.UserId)
		httperrors.BadRequest(w, 400, "Invalid body and path parameters")
		return
	}

	valid, errMsg := ValidateUpdateDisplayName(&reqBody)
	if !valid {
		httperrors.BadRequest(w, 400, errMsg)
		return
	}

	updated, err := h.svc.updateUserDisplayName(userId, reqBody.DisplayName)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if updated {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (h *UserAPIHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	deleted, err := h.svc.deleteUserAccount(userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors")
		httperrors.InternalError(w, 500, "Request aborted due to database errors")
		return
	}
	if deleted {
		w.WriteHeader(http.StatusOK)
	} else {
		httperrors.InternalError(w, 500, "Unexpected errors occurred.")
	}
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

// GetUserNamespaces handles GET /api/v1/users/{id}/namespaces
func (h *UserAPIHandler) GetUserNamespaces(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	username, namespaceAccess, err := h.svc.getUserNamespaceAccess(userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if username == "" && namespaceAccess == nil {
		httperrors.NotFound(w, 404, "No namespace was found")
		return
	}

	accessList := make([]*mgmt.NamespaceAccess, len(namespaceAccess))

	for _, access := range namespaceAccess {
		dto := h.adapter.ToNamespaceAccess(access)
		accessList = append(accessList, dto)
	}

	res := mgmt.UserNamespaceAccessResponse{
		Username:   username,
		AccessList: accessList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when writing response to get user namespace request")
	}
}

// GetUserRepositories handles GET /api/v1/users/{id}/repositories
func (h *UserAPIHandler) GetUserRepositories(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	username, repositoryAccess, err := h.svc.getUserRepositoryAccess(userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	if username == "" && repositoryAccess == nil {
		httperrors.NotFound(w, 404, "No namespace was found")
		return
	}

	accessList := make([]*mgmt.RepositoryAccess, len(repositoryAccess))

	for _, access := range repositoryAccess {
		dto := h.adapter.ToRepositoryAccess(access)
		accessList = append(accessList, dto)
	}

	res := mgmt.UserRepositoryAccessResponse{
		Username:   username,
		AccessList: accessList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when writing response to get user repository request")
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

// AssignRole handles POST /api/v1/users/{id}/role
func (h *UserAPIHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var reqBody mgmt.AssignRoleRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when parsing request: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Invalid body")
		return
	}

	if reqBody.UserId == "" {
		httperrors.BadRequest(w, 400, "Invalid user_id")
		return
	}
	if reqBody.RoleName == "" {
		httperrors.BadRequest(w, 400, "Invalid role_name")
		return
	}

	err = h.svc.assignRoleToUser(reqBody.UserId, reqBody.RoleName)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RemoveRole handles DELETE /api/v1/users/{id}/role/{role}
func (h *UserAPIHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	var reqBody mgmt.UnassignRoleRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when parsing the request: %s", r.RequestURI)
		httperrors.BadRequest(w, 400, "Invalid body")
		return
	}

	if reqBody.UserId == "" {
		httperrors.BadRequest(w, 400, "Invalid user_id")
		return
	}
	if reqBody.RoleName == "" {
		httperrors.BadRequest(w, 400, "Invalid role_name")
		return
	}

	err = h.svc.unassignRoleFromUser(reqBody.UserId, reqBody.RoleName)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Request aborted due to errors: %s", r.RequestURI)
		httperrors.InternalError(w, 500, "Request aborted due to errors")
		return
	}

	w.WriteHeader(http.StatusOK)
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

	if !res.found {
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

func (h *UserAPIHandler) Routes() chi.Router {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Put("/{id}/email", h.UpdateUserEmail)
		r.Put("/{id}/display-name", h.UpdateUserDisplayName)
		r.Put("/{id}/password", h.ChangePassword)
		r.Put("/{id}", h.UpdateUser)
		r.Put("/{id}/lock", h.LockUser)
		r.Put("/{id}/unlock", h.UnlockUser)
		r.Get("/{id}/namespaces", h.GetUserNamespaces)
		r.Get("/{id}/repositories", h.GetUserRepositories)
		r.Get("/me", h.GetCurrentUser)
		r.Put("/me", h.UpdateCurrentUser)
		r.Put("/{id}/role", h.AssignRole)
		r.Delete("/{id}/role/{role}", h.RemoveRole)
		// r.Get("/{id}/role", h.GetUserRole)
		// r.Get("/{id}", h.GetUser) TODO: We'll think about this response type later
		r.Delete("/{id}", h.DeleteUser)
		r.Post("/", h.CreateUser)
		r.Get("/", h.ListUsers)
		r.Post("/validate", h.ValidateUser)
		r.Post("/validate-password", h.ValidatePassword)
		r.Get("/account-setup/{uuid}", h.GetUserAccountSetupInfo)
		r.Post("/account-setup/{uuid}/complete", h.CompleteUserAccountSetup)
	})
	return router
}