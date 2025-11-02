package user

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/db"
	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/security"

	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/types/query"
)

type userService struct {
	userDao       db.UserDAO
	accessDao     db.ResourceAccessDAO
	ec            *email.EmailClient
	adapter       *UserAdapter
	userIdNameMap sync.Map // for now, we'll sync.Map, later we have to use map with Mutex
}

func (svc *userService) getUsername(userId string) (string, error) {
	val, ok := svc.userIdNameMap.Load(userId)
	if ok {
		return val.(string), nil
	}

	username, err := svc.userDao.GetUsernameById(userId, "")
	if err != nil || username == "" {
		return "", err
	}

	svc.userIdNameMap.Store(userId, username)

	return username, nil
}

func (svc *userService) getUserList(cond *query.ListModelsConditions) (users []*models.UserAccountView, total int, err error) {
	//TODO: https://github.com/ksankeerth/open-image-registry/issues/6
	txKey := fmt.Sprintf("get-user-list-%d", cond.Pagination.Page)

	err = svc.userDao.Begin(txKey)
	if err != nil {
		return nil, -1, err
	}
	defer func() {
		if err != nil {
			svc.userDao.Rollback(txKey)
		} else {
			svc.userDao.Commit(txKey)
		}
	}()

	users, total, err = svc.userDao.ListUserAccounts(cond, txKey)
	if err != nil {
		return nil, -1, err
	}

	return
}

func (svc *userService) createUserAccount(username, email, displayName, role string) (userId string, conflict bool, err error) {
	txKey := fmt.Sprintf("create-user-account-%s-%s", username, email)

	err = svc.userDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when starting db transaction")
		return
	}

	defer func() {
		var err1 error
		if err != nil {
			err1 = svc.userDao.Rollback(txKey)
		} else {
			err1 = svc.userDao.Commit(txKey)
		}
		if err1 != nil {
			log.Logger().Error().Err(err).Msgf("Error occurred when finishing the transaction")
		}
	}()

	if displayName == "" {
		displayName = DisplayNameNotSet
	}

	userId, err = svc.userDao.CreateUserAccount(username, email, displayName, PasswordNotSet, SaltNotSet, txKey)
	if err != nil {
		if yes, _ := db_errors.IsUniqueConstraint(err); yes {
			return "", true, nil
		}
		log.Logger().Error().Err(err).Msgf("Error occurred when creating user account: %s", username)
		return
	}

	locked, err := svc.userDao.LockUserAccount(username, ReasonLockedNewAccountVerficationRequired, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when locking new user account: %s", username)
		return
	}
	if !locked {
		log.Logger().Warn().Msgf("Locking new user account was not successful: %s", username)
		return "", false, fmt.Errorf("locking user account failed")
	}

	err = svc.userDao.AssignUserRole(role, userId, txKey)
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

	err = svc.userDao.PersistPasswordRecovery(userId, recoveryUuid, ReasonPasswordRecoveryNewAccountSetup, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Account initialization of new user acccount: %s has been failed", username)
		return
	}

	return
}

func (svc *userService) updateUserAccount(userId, email, displayName, role string) error {
	txKey := fmt.Sprintf("update-user-account-%s-%s-%s-%s", userId, email, displayName, role)

	err := svc.userDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to start transaction for key: %s", txKey)
		return err
	}
	defer func() {
		if err != nil {
			svc.userDao.Rollback(txKey)
		} else {
			svc.userDao.Commit(txKey)
		}
	}()

	err = svc.userDao.UpdateUserAccount(userId, email, displayName, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Updating user account failed: %s", userId)
		return err
	}

	err = svc.userDao.RemoveUserRoleAssignment(userId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Removing role assignment from user: %s failed", userId)
		return err
	}

	err = svc.userDao.AssignUserRole(role, userId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Assigning role(%s) to user(%s) failed", role, userId)
		return err
	}
	return nil
}

func (svc *userService) validateUser(username, email string) (usernameAvailable, emailAvailable bool, err error) {
	return svc.userDao.ValidateUsernameAndEmail(username, email, "")
}

type changePasswordResult struct {
	invalidId          bool
	expired            bool
	oldPasswordDiff    bool
	invalidUserAccount bool
	changed            bool
}

func (svc *userService) changePassword(req *mgmt.PasswordChangeRequest) (res *changePasswordResult, err error) {
	txKey := fmt.Sprintf("user-change-password-%s-%s", req.UserId, req.RecoveryId)

	res = &changePasswordResult{
		invalidId:          false,
		expired:            false,
		oldPasswordDiff:    false,
		invalidUserAccount: false,
	}

	err = svc.userDao.Begin(txKey)
	if err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			err1 := svc.userDao.Rollback(txKey)
			log.Logger().Error().Err(err1).Msgf("Rollback failed with errors: %s", txKey)
		} else {
			err1 := svc.userDao.Commit(txKey)
			log.Logger().Error().Err(err1).Msgf("Commit failed with errors: %s", txKey)
		}
	}()

	pwRecovery, err := svc.userDao.RetrivePasswordRecovery(req.UserId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred retriving password recovery record with uuid: %s", req.RecoveryId)
		return res, err
	}
	// TODO for now, we'll just verify whether record exists or not. Later, We may have to check the expiry.
	if pwRecovery == nil && err == nil {
		res.invalidId = true
		return res, nil
	}

	currPwHash, salt, err := svc.userDao.GetUserPasswordAndSaltById(pwRecovery.UserId, txKey)
	if db_errors.IsNotFound(err) {
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

	updated, err := svc.userDao.UpdateUserPasswordAndSalt(pwRecovery.UserId, newPasswordHash, newSalt, txKey)
	if err != nil {
		return
	}

	if updated {
		res.changed = true
	}

	deleted, err := svc.userDao.DeletePasswordRecovery(pwRecovery.UserId, txKey)
	if err != nil {
		return
	}

	if !deleted {
		log.Logger().Warn().Msgf("Password recovery record: %s is not cleared", pwRecovery.RecoveryId)
	}
	return
}

func (svc *userService) getUserAccount(userId string) (*models.UserAccount, error) {
	return svc.userDao.GetUserAccountById(userId, "")
}

func (svc *userService) deleteUserAccount(userId string) (deleted bool, err error) {
	return svc.userDao.DeleteUserAccount(userId, "")
}

// TODO: for we'll just update the email address in database. Later, We have to verify the email
// address change by sending a mail
func (svc *userService) updateUserEmail(userId, email string) (updated bool, err error) {
	return svc.userDao.UpdateUserEmail(userId, email, "")
}

func (svc *userService) updateUserDisplayName(userId, displayName string) (updated bool, err error) {
	return svc.userDao.UpdateUserDisplayName(userId, displayName, "")
}

func (svc *userService) getUserNamespaceAccess(userId string) (username string,
	access []*models.NamespaceAccess, err error) {

	access, err = svc.accessDao.GetUserNamespaceAccess(userId, "")
	if err != nil {
		return "", nil, err
	}

	username, err = svc.getUsername(userId)
	if err != nil {
		return "", nil, err
	}

	return username, access, nil
}

func (svc *userService) getUserRepositoryAccess(userId string) (username string,
	access []*models.RepositoryAccess, err error) {
	access, err = svc.accessDao.GetUserRepositoryAccess(userId, "")
	if err != nil {
		return "", nil, err
	}

	username, err = svc.getUsername(userId)
	if err != nil {
		return "", nil, err
	}

	return username, access, nil
}

func (svc *userService) assignRoleToUser(userId, roleName string) (err error) {
	txKey := fmt.Sprintf("assign-user-role-%s-%s", userId, roleName)

	err = svc.userDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when starting database transaction: %s", txKey)
		return err
	}

	defer func() {
		var err1 error
		if err != nil {
			err1 = svc.userDao.Rollback(txKey)
		} else {
			err1 = svc.userDao.Commit(txKey)
		}
		if err1 != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when committing or rollingback transaction: %s", txKey)
		}
	}()

	err = svc.userDao.RemoveUserRoleAssignment(userId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Removing existing role from user(%s) failed.", userId)
		return err
	}

	err = svc.userDao.AssignUserRole(roleName, userId, txKey)
	if err != nil {
		return err
	}
	return nil
}

// TODO: consider roleName when removing
func (svc *userService) unassignRoleFromUser(userId, roleName string) (err error) {
	err = svc.userDao.RemoveUserRoleAssignment(userId, "")
	if err != nil {
		return err
	}
	return nil
}

type lockResult struct {
	alreadyLocked bool
	success       bool
}

func (svc *userService) lockUserAccount(userId string) (*lockResult, error) {
	txKey := fmt.Sprintf("lock-user-account-%s", userId)

	err := svc.userDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when starting transaction: %s", txKey)
		return nil, err
	}

	defer func() {
		if err != nil {
			svc.userDao.Rollback(txKey)
		} else {
			svc.userDao.Commit(txKey)
		}
	}()

	user, err := svc.userDao.GetUserAccountById(userId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user account(%s)", userId)
		return nil, err
	}

	res := &lockResult{}
	if user.Locked {
		res.alreadyLocked = true
		return res, nil
	}

	locked, err := svc.userDao.LockUserAccount(user.Username, ReasonLockedAdminLocked, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when locking user account: %s", userId)
		return res, err
	}
	if !locked {
		log.Logger().Error().Err(err).Msgf("Locking user account failed : %s", userId)
		return nil, fmt.Errorf("locking user account failed")
	}

	res.success = true

	return res, nil
}

type unlockResult struct {
	success    bool
	newAccount bool
}

func (svc *userService) unlockUserAccount(userId string) (*unlockResult, error) {

	txKey := fmt.Sprintf("unlock-user-account-%s", userId)

	err := svc.userDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when starting transaction: %s", txKey)
		return nil, err
	}

	defer func() {
		if err != nil {
			svc.userDao.Rollback(txKey)
		} else {
			svc.userDao.Commit(txKey)
		}
	}()

	user, err := svc.userDao.GetUserAccountById(userId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user account(%s)", userId)
		return nil, err
	}

	res := &unlockResult{}

	if user.Locked && user.LockedReason == ReasonLockedNewAccountVerficationRequired {
		res.newAccount = true
		return res, nil
	}

	unlocked, err := svc.userDao.UnlockUserAccount(userId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when unlocking user account: %s", userId)
		return nil, err
	}

	if !unlocked {
		log.Logger().Error().Err(err).Msgf("Unlocking user account failed: %s", userId)
		res.success = false
		return res, nil
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

	txKey := fmt.Sprintf("get-account-setup-info-%s", id)

	var res accountSetupVerficationResult

	err := svc.userDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when creating transaction to database: %s", txKey)
		return nil, err
	}

	defer func() {
		if err != nil {
			svc.userDao.Rollback(txKey)
		} else {
			svc.userDao.Commit(txKey)
		}
	}()

	pwRecovery, err := svc.userDao.RetrivePasswordRecovery(id, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error ocurred when retriving password recovery by uuid: %s", id)
		return nil, err
	}
	if pwRecovery == nil {
		res.found = false
		res.errorMsg = "This account setup link is no longer valid. It may have already been used.."
		return &res, nil
	}

	if pwRecovery.ReasonType != ReasonPasswordRecoveryNewAccountSetup {
		res.found = true
		res.errorMsg = "This link is not valid. Please check it again ..."
		return &res, nil
	}
	res.found = true

	user, err := svc.userDao.GetUserAccountById(pwRecovery.UserId, txKey)
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

	res.userId = pwRecovery.UserId
	res.username = user.Username
	res.displayName = user.DisplayName
	res.email = user.Email

	role, err := svc.userDao.GetUserRole(pwRecovery.UserId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user role for account setup verification")
		return nil, err
	}
	res.role = role
	return &res, nil
}

func (svc *userService) completeAccountSetup(req *mgmt.AccountSetupCompleteRequest) error {
	txKey := fmt.Sprintf("complete-account-setup-%s", req.Uuid)

	err := svc.userDao.Begin(txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when starting transaction: %s", txKey)
		return err
	}

	defer func() {
		if err != nil {
			svc.userDao.Rollback(txKey)
		} else {
			svc.userDao.Commit(txKey)
		}
	}()

	// TODO:  Related issue: https://github.com/ksankeerth/open-image-registry/issues/16
	_, err = svc.userDao.UpdateUserDisplayName(req.UserId, req.DisplayName, txKey)
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

	updated, err := svc.userDao.UpdateUserPasswordAndSalt(req.UserId, passwordHash, salt, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when setting new password for user: %s", req.UserId)
		return err
	}
	if !updated {
		log.Logger().Warn().Msg("No entry was updated with new passowd")
		return fmt.Errorf("account setup failed")
	}

	unlocked, err := svc.userDao.UnlockUserAccount(req.Username, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when unlocking user account during account setup: %s", req.UserId)
		return err
	}

	if !unlocked {
		log.Logger().Warn().Msg("No user account was unlocked")
		err = fmt.Errorf("account setup failed")
		return err
	}

	deleted, err := svc.userDao.DeletePasswordRecovery(req.UserId, txKey)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when removing password recovery reference: %s", req.Uuid)
		return err
	}

	if !deleted {
		log.Logger().Warn().Msgf("No password recovery reference was deleted for user: %s", req.UserId)
		err = fmt.Errorf("account setup failed")
		return err
	}
	// TODO: We may send a mail here. Idea: Send instructions to use OpenImageRegistry
	return nil
}