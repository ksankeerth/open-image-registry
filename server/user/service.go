package user

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/ksankeerth/open-image-registry/access/resource"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/security"
	"github.com/ksankeerth/open-image-registry/store"

	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type userService struct {
	accessManager *resource.Manager
	store         store.Store
	ec            *email.EmailClient
	adapter       *UserAdapter
	userIdNameMap sync.Map // for now, we'll sync.Map, later we have to use map with Mutex
}

func (svc *userService) getUserList(cond *store.ListQueryConditions) (users []*models.UserAccountView, total int, err error) {

	ctx := context.Background()
	tx, err := svc.store.Begin(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("error occurred when starting transaction")
		return nil, -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx = store.WithTxContext(ctx, tx)

	users, total, err = svc.store.Users().ListUserAccounts(ctx, cond)
	if err != nil {
		return nil, -1, err
	}

	return
}

// createUserAccount returns passwordRecoveryUUID only if development mode is enabled
func (svc *userService) createUserAccount(username, email, displayName, role string) (userId string, conflict bool,
	passwordRecoveryId string, err error) {

	ctx := context.Background()
	tx, err := svc.store.Begin(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("error occurred when starting transaction")
		return "", false, "", err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if displayName == "" {
		displayName = constants.DisplayNameNotSet
	}

	userId, err = svc.store.Users().Create(ctx, username, email, displayName, constants.PasswordNotSet, constants.SaltNotSet)
	if err != nil {
		if yes, _ := dberrors.IsUniqueConstraint(err); yes {
			return "", true, "", nil
		}
		log.Logger().Error().Err(err).Msgf("Error occurred when creating user account: %s", username)
		return
	}

	err = svc.store.Users().LockAccount(ctx, username, constants.ReasonLockedNewAccountVerficationRequired)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when locking new user account: %s", username)
		return
	}

	err = svc.store.Users().AssignRole(ctx, userId, role)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when assigning role to new user account: %s", username)
		return
	}

	//TODO: dummy link; we have to construct this link to show password set/account setup
	recoveryUuid := uuid.New().String()
	passwordResetLink := "" // TODO: we need to set correct link
	if config.GetDevelopmentConfig().Enable && config.GetDevelopmentConfig().MockEmail {
		passwordResetLink = recoveryUuid
	}

	err = svc.ec.SendAccountSetupEmail(username, email, passwordResetLink)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when sending account setup mail to %s", email)
		return
	}

	err = svc.store.AccountRecovery().Create(ctx, userId, recoveryUuid, constants.ReasonPasswordRecoveryNewAccountSetup)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Account initialization of new user acccount: %s has been failed", username)
		return
	}

	// this will be set only if development config is enabled.
	if config.GetDevelopmentConfig().Enable {
		passwordRecoveryId = recoveryUuid
	}
	return
}

func (svc *userService) updateUserAccount(reqCtx context.Context, userId, displayName string) error {

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

	err = svc.store.Users().UpdateDisplayName(ctx, userId, displayName)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Updating user account failed: %s", userId)
		return err
	}

	return nil
}

func (svc *userService) validateUser(username, email string) (usernameAvailable, emailAvailable bool, err error) {
	return svc.store.Users().CheckAvailability(context.Background(), username, email)
}

type changePasswordResult struct {
	invalidId          bool
	expired            bool
	oldPasswordDiff    bool
	invalidUserAccount bool
	changed            bool
}

func (svc *userService) changePassword(req *mgmt.PasswordChangeRequest) (res *changePasswordResult, err error) {
	ctx := context.Background()
	tx, err := svc.store.Begin(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when starting transaction")
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx = store.WithTxContext(ctx, tx)

	res = &changePasswordResult{
		invalidId:          false,
		expired:            false,
		oldPasswordDiff:    false,
		invalidUserAccount: false,
	}

	pwRecovery, err := svc.store.AccountRecovery().GetByUserID(ctx, req.UserId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred retriving password recovery record with uuid: %s", req.RecoveryId)
		return res, err
	}
	// TODO for now, we'll just verify whether record exists or not. Later, We may have to check the expiry.
	if pwRecovery == nil && err == nil {
		res.invalidId = true
		return res, nil
	}

	currPwHash, salt, err := svc.store.Users().GetPasswordAndSalt(ctx, pwRecovery.UserID)
	if dberrors.IsNotFound(err) {
		res.invalidUserAccount = true
		return
	}
	if err != nil {
		return
	}

	macthed := security.ComparePasswordAndHash(req.OldPassword, salt, currPwHash)
	if !macthed {
		res.oldPasswordDiff = true
		return
	}

	newSalt, err := security.GenerateSalt(16)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when generating salt")
		return
	}
	newPasswordHash := security.GeneratePasswordHash(req.Password, newSalt)

	err = svc.store.Users().UpdatePasswordAndSalt(ctx, pwRecovery.UserID, newPasswordHash, newSalt)
	if err != nil {
		return
	}
	res.changed = true

	err = svc.store.AccountRecovery().DeleteByUserID(ctx, pwRecovery.UserID)
	if err != nil {
		return
	}

	return res, err
}

func (svc *userService) deleteUserAccount(userId string) (err error) {
	return svc.store.Users().Delete(context.Background(), userId)
}

// TODO: for we'll just update the email address in database. Later, We have to verify the email
// address change by sending a mail
func (svc *userService) updateUserEmail(userId, email string) (err error) {
	return svc.store.Users().UpdateEmail(context.Background(), userId, email)
}

type lockResult struct {
	alreadyLocked bool
	success       bool
}

func (svc *userService) lockUserAccount(userId string) (*lockResult, error) {
	ctx := context.Background()
	tx, err := svc.store.Begin(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when starting transaction")
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	user, err := svc.store.Users().Get(ctx, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user account(%s)", userId)
		return nil, err
	}

	res := &lockResult{}
	if user.Locked {
		res.alreadyLocked = true
		return res, nil
	}

	err = svc.store.Users().LockAccount(ctx, user.Username, constants.ReasonLockedAdminLocked)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when locking user account: %s", userId)
		return res, err
	}

	res.success = true

	return res, nil
}

type unlockResult struct {
	success    bool
	newAccount bool
}

func (svc *userService) unlockUserAccount(userId string) (*unlockResult, error) {
	ctx := context.Background()
	tx, err := svc.store.Begin(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when starting transaction")
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	user, err := svc.store.Users().Get(ctx, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user account(%s)", userId)
		return nil, err
	}

	res := &unlockResult{}

	if user.Locked && user.LockedReason == constants.ReasonLockedNewAccountVerficationRequired {
		res.newAccount = true
		return res, nil
	}

	err = svc.store.Users().UnlockAccount(ctx, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when unlocking user account: %s", userId)
		return nil, err
	}

	res.success = true

	return res, nil
}

type accountSetupVerficationResult struct {
	id          string
	found       bool
	errorMsg    string
	userId      string
	username    string
	email       string
	role        string
	displayName string
}

func (svc *userService) getAccountSetupInfo(id string) (*accountSetupVerficationResult, error) {
	ctx := context.Background()
	tx, err := svc.store.Begin(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when starting transaction")
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var res accountSetupVerficationResult

	pwRecovery, err := svc.store.AccountRecovery().Get(ctx, id)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error ocurred when retriving password recovery by uuid: %s", id)
		return nil, err
	}

	if pwRecovery == nil {
		res.found = false
		res.errorMsg = "This account setup link is no longer valid. It may have already been used.."
		return &res, nil
	}

	if pwRecovery.ReasonType != constants.ReasonPasswordRecoveryNewAccountSetup {
		res.found = true
		res.errorMsg = "This link is not valid. Please check it again ..."
		return &res, nil
	}
	res.found = true

	user, err := svc.store.Users().Get(ctx, pwRecovery.UserID)
	// user account must exist. We don't check for not-found case. Reason: PaswordRecovery Table has FK reference to UserID
	// and Delete cascade
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user details for account setup verification")
		return nil, err
	}

	if user == nil {
		res.found = false
		res.errorMsg = "This account setup link is no longer valid. It may have already been used.."
		return &res, nil
	}

	res.userId = pwRecovery.UserID
	res.username = user.Username
	res.displayName = user.DisplayName
	res.email = user.Email

	role, err := svc.store.Users().GetRole(ctx, pwRecovery.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user role for account setup verification")
		return nil, err
	}
	res.role = role
	return &res, nil
}

func (svc *userService) completeAccountSetup(req *mgmt.AccountSetupCompleteRequest) error {
	ctx := context.Background()
	tx, err := svc.store.Begin(ctx)
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

	// TODO:  Related issue: https://github.com/ksankeerth/open-image-registry/issues/16

	err = svc.store.Users().UpdateDisplayName(ctx, req.UserId, req.DisplayName)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when  completing account setup")
		return err
	}

	salt, err := security.GenerateSalt(16)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when generating salt")
		return err
	}
	passwordHash := security.GeneratePasswordHash(req.Password, salt)

	err = svc.store.Users().UpdatePasswordAndSalt(ctx, req.UserId, passwordHash, salt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when setting new password for user: %s", req.UserId)
		return err
	}

	err = svc.store.Users().UnlockAccount(ctx, req.Username)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when unlocking user account during account setup: %s", req.UserId)
		return err
	}

	err = svc.store.AccountRecovery().DeleteByUserID(ctx, req.UserId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when removing password recovery reference: %s", req.Uuid)
		return err
	}

	// TODO: We may send a mail here. Idea: Send instructions to use OpenImageRegistry
	return nil
}

// changeRole changes the role if possible. If not it will send an `errMsg` with details.
// If any other errors occured, it will return `error`.
func (svc *userService) changeRole(reqCtx context.Context, newRole, userId string) (errMsg string, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when starting transaction")
		return "", err
	}

	ctx := store.WithTxContext(reqCtx, tx)

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	role, err := svc.store.Users().GetRole(ctx, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to change role due to database errors")
		return "", err
	}

	if role == newRole {
		// no change so let's pretend like we've have changed
		return "", nil
	}

	// conditions
	// 1. if a user have access to resources and any of the current access level exceeds max access
	// level allowed by the new role, it wil be not allowed to change roles.
	// first admin must revoke/change access levels before chaning roles
	// 2. if a user have admin role, we cannot change the role of that user; reason to avoid
	// accidental changes or attackers revoking access to admins

	if role == constants.RoleAdmin {
		return "Not allowed to change role of Admins", nil
	}

	if !isRolePromotion(role, newRole) {
		var overPriveleges int
		if role == constants.RoleMaintainer {
			if newRole == constants.RoleDeveloper {
				_, overPriveleges, err = svc.accessManager.GetUserAccessByLevels(ctx, userId, 1, 1, constants.AccessLevelMaintainer)
			}
			if newRole == constants.RoleGuest {
				_, overPriveleges, err = svc.accessManager.GetUserAccessByLevels(ctx, userId, 1, 1, constants.AccessLevelMaintainer,
					constants.AccessLevelDeveloper)
			}
		}

		if role == constants.RoleDeveloper {
			// if the role is developer, only possiblity for new role is guest
			_, overPriveleges, err = svc.accessManager.GetUserAccessByLevels(ctx, userId, 1, 1, constants.AccessLevelDeveloper)
		}
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to change role due to database errors")
			return "", err
		}

		if overPriveleges > 0 {
			return fmt.Sprintf("User has %d resource(s) with access levels exceeding the new role (%s). Revoke these first.",
				overPriveleges, newRole), nil
		}
	}

	err = svc.store.Users().UnAssignRole(ctx, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change role due to database errors")
		return "", err
	}

	err = svc.store.Users().AssignRole(ctx, userId, newRole)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change role due to database errors")
		return "", err
	}

	return "", nil
}

func isRolePromotion(currentRole, newRole string) bool {
	if currentRole == newRole {
		return false
	}

	if currentRole == constants.RoleGuest && (newRole == constants.RoleDeveloper ||
		newRole == constants.RoleMaintainer || newRole == constants.RoleAdmin) {
		return true
	}

	if currentRole == constants.RoleDeveloper && (newRole == constants.RoleMaintainer ||
		newRole == constants.RoleAdmin) {
		return true
	}

	if currentRole == constants.RoleMaintainer && newRole == constants.RoleAdmin {
		return true
	}
	return false
}

func (svc *userService) getUser(reqCtx context.Context, identifier string) (user *models.UserAccount, role string, err error) {

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when starting transaction")
		return nil, "", err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	user, err = svc.store.Users().Get(ctx, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user account(%s) from database", identifier)
		return nil, "", err
	}

	if user == nil {
		log.Logger().Warn().Msgf("No user accounts found with %s", identifier)
		return nil, "", nil
	}

	role, err = svc.store.Users().GetRole(ctx, user.Id)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to retrieve role of user(%s)", identifier)
		return nil, "", err
	}

	return
}
