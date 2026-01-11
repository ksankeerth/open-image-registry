package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/security"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type authService struct {
	store            store.Store
	jwtAuthenticator lib.JWTProvider
}

type authLoginResult struct {
	statusCode      int
	success         bool
	errorMessage    string
	userRole        string
	userID          string
	username        string
	jwtToken        string
	expiryInSeconds int
}

func (svc *authService) authenticateUser(reqCtx context.Context, req *mgmt.AuthLoginRequest, userAgent, clientIp string) (*authLoginResult, error) {
	loginRes := &authLoginResult{}

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("error occurred when starting transaction")

		loginRes.success = false
		loginRes.errorMessage = "Opps! Error occured when logging in!"
		loginRes.statusCode = http.StatusInternalServerError
		return loginRes, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	userAccount, err := svc.store.Users().GetByUsername(ctx, req.Username)
	if err != nil {
		loginRes.success = false
		loginRes.errorMessage = "Opps! Error occured when logging in!"
		loginRes.statusCode = http.StatusInternalServerError

		return loginRes, err
	}

	if userAccount == nil {
		log.Logger().Warn().Msgf("Login attempt with non-existing user: %s", req.Username)
		loginRes.success = false
		loginRes.errorMessage = "Invalid username or password!"
		loginRes.statusCode = http.StatusUnauthorized

		return loginRes, nil
	}

	if userAccount.Locked {
		loginRes.success = false
		loginRes.errorMessage = "User account has been locked! Contact system administrator."
		loginRes.statusCode = http.StatusForbidden

		return loginRes, nil
	}

	currentPw, currentSalt, err := svc.store.Users().GetPasswordAndSalt(ctx, userAccount.Id)
	if err != nil {
		loginRes.success = false
		loginRes.errorMessage = "Opps! Error occured when logging in!"
		loginRes.statusCode = http.StatusInternalServerError

		return loginRes, err
	}

	matched := security.ComparePasswordAndHash(req.Password, currentSalt, currentPw)
	if !matched {
		err = svc.store.Users().RecordFailedAttempt(ctx, req.Username)
		if err != nil {
			loginRes.success = false
			loginRes.errorMessage = "Opps! Error occured when logging in!"
			loginRes.statusCode = http.StatusInternalServerError

			return loginRes, err
		}
		if (userAccount.FailedAttempts + 1) > constants.MaxFailedLoginAttempts {
			err = svc.store.Users().LockAccount(ctx, req.Username, constants.ReasonLockedFailedLoginAttempts)
			if err != nil {
				loginRes.success = false
				loginRes.errorMessage = "Opps! Error occured when logging in!"
				loginRes.statusCode = http.StatusInternalServerError

				return loginRes, err
			}

			loginRes.success = false
			loginRes.errorMessage = "User account has been locked! Contact system administrator."
			loginRes.statusCode = http.StatusForbidden

			return loginRes, nil
		}

		loginRes.success = false
		loginRes.errorMessage = "Invalid username or password!"
		loginRes.statusCode = http.StatusUnauthorized

		return loginRes, nil
	}
	err = svc.store.Users().UnlockAccount(ctx, req.Username)
	if err != nil {
		loginRes.success = false
		loginRes.errorMessage = "Opps! Error occured when logging in!"
		loginRes.statusCode = http.StatusInternalServerError

		return loginRes, nil
	}

	roleName, err := svc.store.Users().GetRole(ctx, userAccount.Id)
	if err != nil {
		loginRes.success = false
		loginRes.errorMessage = "Opps! Error occured when logging in!"
		loginRes.statusCode = http.StatusInternalServerError

		return loginRes, err
	}

	loginRes.userRole = roleName

	token, err := svc.jwtAuthenticator.Sign(map[string]any{
		"sub":  req.Username,
		"role": loginRes.userRole,
	})
	if err != nil {
		log.Logger().Error().Err(err).Msgf("User(%s) login failed due to jwt token genaration errors", req.Username)
		loginRes.success = false
		loginRes.statusCode = http.StatusInternalServerError
		loginRes.errorMessage = "Opps! Token gernation failed. Please try again!"
		return loginRes, err
	}

	// Write last accessed time to db
	err = svc.store.Users().RecordLastAccessedTime(ctx, userAccount.Id, time.Now())
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to record user login(%s) time due to errors", req.Username)
		loginRes.statusCode = http.StatusInternalServerError
		loginRes.errorMessage = "Opps! Failed to persist data. Please try again!"
		loginRes.success = false
		return loginRes, err
	}

	loginRes.success = true
	loginRes.jwtToken = token
	loginRes.expiryInSeconds = config.GetAuthTokenConfig().Expiry
	loginRes.userID = userAccount.Id
	loginRes.username = userAccount.Username

	return loginRes, nil
}

func (svc *authService) revokeToken(reqCtx context.Context, signatureHash, userID string, expAt, issuedAt int64) error {

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when starting transaction")
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	// Check the database first
	revokedToken, err := svc.store.Auth().GetRevokedToken(ctx, signatureHash)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when checking revoked tokens")
		return err
	}

	// if already revoked, no action required
	if revokedToken != nil {
		return nil
	}

	revokedToken = &models.RevokedToken{
		SignatureHash: signatureHash,
		UserID:        userID,
		ExpiresAt:     expAt,
		IssuedAt:      issuedAt,
	}

	err = svc.store.Auth().RecordTokenRevocation(ctx, revokedToken)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to revoke token due to errors")
		return err
	}

	return nil
}