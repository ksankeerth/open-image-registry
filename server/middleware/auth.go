package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/errors/httperrors"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/utils"
)

type Authenticator struct {
	store       store.Store
	jwtProvider lib.JWTProvider
}

func NewAuthenticator(store store.Store, jwtProvider lib.JWTProvider) *Authenticator {
	return &Authenticator{
		store:       store,
		jwtProvider: jwtProvider,
	}
}

func (a *Authenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		c, err := r.Cookie(constants.AuthTokenCookie)
		if errors.Is(err, http.ErrNoCookie) || c == nil || c.Value == "" {
			httperrors.Unauthorized(w, 401, "cookie not found")
			return
		}

		token := c.Value

		claims, err := a.jwtProvider.Verify(token)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Token verification failed")
			httperrors.Unauthorized(w, 401, "invalid token")
			return
		}

		userId, ok := claims[constants.ClaimSubject]
		if !ok || userId == "" {
			httperrors.Unauthorized(w, 401, "user is not found in token")
			return
		}

		role, ok := claims[constants.ClaimRole]
		if !ok || role == "" {
			httperrors.Unauthorized(w, 401, "role is not found in token")
			return
		}

		tokenParts := strings.Split(token, ".")
		signature := tokenParts[2]

		signatureHash := utils.CalcuateDigest([]byte(signature))

		revokedToken, err := a.store.Auth().GetRevokedToken(r.Context(), signatureHash)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Failed to check whether token was already revoked")
			httperrors.InternalError(w, 500, "unable to check revoked tokens")
			return
		}
		if revokedToken != nil {
			httperrors.Unauthorized(w, 401, "invalid token")
			return
		}

		user, err := a.store.Users().Get(r.Context(), userId.(string))
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to verify login due to user retrieval errors: %s",
				userId)
			httperrors.InternalError(w, 500, "unable to verify user due to errors")
			return
		}

		if user == nil {
			httperrors.Unauthorized(w, 401, "user not found")
			return
		}

		expiresAt, _ := claims[lib.ClaimExp]
		issuedAt, _ := claims[lib.ClaimIat]

		ctx := r.Context()
		ctx = context.WithValue(ctx, constants.ContextUserID, userId)
		ctx = context.WithValue(ctx, constants.ContextRole, role)
		ctx = context.WithValue(ctx, constants.ContextSignatureHash, signatureHash)
		ctx = context.WithValue(ctx, constants.ContextExpAt, expiresAt)
		ctx = context.WithValue(ctx, constants.ContextIssuedAt, issuedAt)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}