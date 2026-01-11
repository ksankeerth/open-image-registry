package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/errors/httperrors"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/middleware"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/user"
)

type AuthAPIHandler struct {
	svc           *authService
	userAdapter   *user.UserAdapter
	authenticator *middleware.Authenticator
}

// NewAuthAPIHandler creates a new auth API handler
func NewAuthAPIHandler(store store.Store, jwtProvider lib.JWTProvider,
	authenticator *middleware.Authenticator) *AuthAPIHandler {
	return &AuthAPIHandler{
		svc: &authService{
			store:            store,
			jwtAuthenticator: jwtProvider,
		},
		authenticator: authenticator,
	}
}

// Login handles POST /api/v1/auth/login
func (h *AuthAPIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest mgmt.AuthLoginRequest

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when parsing login request")
		httperrors.BadRequest(w, 400, "Invalid request body")
		return
	}

	userAgent := r.Header.Get("User-Agent")
	xForwardedFor := r.Header.Get("X-Forwarded-For") // client_ip, lb1_ip, ........
	var clientIp string
	ips := strings.Split(xForwardedFor, ",")
	if len(ips) > 0 {
		clientIp = ips[0]
	} else {
		clientIp = xForwardedFor
	}

	loginResult, err := h.svc.authenticateUser(r.Context(), &loginRequest, userAgent, clientIp)
	authLoginResponse := mgmt.AuthLoginResponse{}

	if err != nil {
		log.Logger().Error().Err(err).Msg("Login failed due to errors")
		httperrors.SendError(w, loginResult.statusCode, loginResult.errorMessage)
		return
	}
	if !loginResult.success {
		httperrors.SendError(w, loginResult.statusCode, loginResult.errorMessage)
		return
	}

	h.setAuthCookie(w, loginResult.jwtToken, loginResult.expiryInSeconds)
	w.WriteHeader(http.StatusOK)

	authLoginResponse.User = mgmt.UserProfileInfo{
		UserId:   loginResult.userID,
		Username: loginResult.username,
		Role:     loginResult.userRole,
	}

	err = json.NewEncoder(w).Encode(authLoginResponse)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when writing login response to client")
	}
}

func (h *AuthAPIHandler) Logout(w http.ResponseWriter, r *http.Request) {
	signatureHash := r.Context().Value(constants.ContextSignatureHash)
	if signatureHash == nil {
		log.Logger().Error().Msg("signature hash is not found in request context")
		httperrors.InternalError(w, 500, "Invalid value for signature hash")
		return
	}

	signatureHashVal, ok := signatureHash.(string)
	if !ok {
		log.Logger().Error().Msg("Signature hash is not found in request context")
		httperrors.InternalError(w, 500, "Invalid value for signature hash")
		return
	}
	if signatureHashVal == "" {
		log.Logger().Error().Msg("Signature hash is empty in request context")
		httperrors.InternalError(w, 500, "Signature hash is not set properly")
		return
	}

	expAt := r.Context().Value(constants.ContextExpAt)
	if expAt == nil {
		log.Logger().Error().Msg("Token expiry is not in request context")
		httperrors.InternalError(w, 500, "Invalid value for expiry")
		return
	}
	expVal, ok := expAt.(int64)
	if !ok {
		log.Logger().Error().Msgf("Invalid value set for token expiry: %T", expAt)
		httperrors.InternalError(w, 500, "Invalid value for expiry")
		return
	}

	iat := r.Context().Value(constants.ContextIssuedAt)
	if iat == nil {
		log.Logger().Error().Msg("token issued time is not set in request context")
		httperrors.InternalError(w, 500, "Invalid value for iat")
		return
	}
	iatVal, ok := iat.(int64)
	if !ok {
		log.Logger().Error().Msgf("Invalid value set for token expiry: %T", expAt)
		httperrors.InternalError(w, 500, "Invalid value for iat")
		return
	}

	userId := r.Context().Value(constants.ContextUserID)
	if userId == nil {
		log.Logger().Error().Msg("user id is not set in request context")
		httperrors.InternalError(w, 500, "Invalid value for userid")
		return
	}
	userIdStr, ok := userId.(string)
	if !ok {
		log.Logger().Error().Msg("invalid value set for userId in request context")
		httperrors.InternalError(w, 500, "Invalid value for sub")
		return
	}

	err := h.svc.revokeToken(r.Context(), signatureHashVal, userIdStr, expVal, iatVal)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Request failed due to errors")
		httperrors.InternalError(w, 500, "Request failed due to errors")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthAPIHandler) Routes() chi.Router {

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.With(h.authenticator.Authenticate).Post("/logout", h.Logout)
	})

	return router
}

func (h *AuthAPIHandler) setAuthCookie(w http.ResponseWriter, token string, expiryInSeconds int) {
	cookie := &http.Cookie{
		Name:     constants.AuthTokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   expiryInSeconds,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	if config.GetDevelopmentConfig().Enable {
		cookie.SameSite = http.SameSiteLaxMode
		cookie.Secure = false
	}

	http.SetCookie(w, cookie)
}