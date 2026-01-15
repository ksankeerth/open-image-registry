package user

import (
	"github.com/google/uuid"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

// TODO: Several methods in this package are exported unnecessarily.
// Review visibility and unexport methods that don't need to be accessed by other packages.

type UserAdapter struct{}

// Converts CreateUserAccountRequest → UserAccount entity
func (ua *UserAdapter) ToUserEntity(req *mgmt.CreateUserAccountRequest) *models.UserAccount {
	if req == nil {
		return nil
	}

	return &models.UserAccount{
		Id:             uuid.NewString(), // assuming new user creation
		Username:       req.Username,
		Email:          req.Email,
		DisplayName:    req.DisplayName,
		Locked:         false,
		FailedAttempts: 0,
	}
}

func (ua *UserAdapter) ToUserAccountViewDTO(model *models.UserAccountView) *mgmt.UserAccountViewDTO {
	if model == nil {
		return nil
	}

	lockedReason := ""
	switch model.LockedReason {
	case constants.ReasonLockedNewAccountVerficationRequired:
		lockedReason = "New User Account and Verification is required."
	case constants.ReasonLockedFailedLoginAttempts:
		lockedReason = "User Account locked due to multiple failed attempts."
	case constants.ReasonLockedAdminLocked:
		lockedReason = "User Account was locked by Admin priveleged user"
	}

	pwRecoveryReason := ""
	switch model.PasswordRecoveryReason {
	case constants.ReasonPasswordRecoveryNewAccountSetup:
		pwRecoveryReason = "New User Account and Verification is required."
	case constants.ReasonPasswordRecoveryForgotPassowrd:
		pwRecoveryReason = "User forgot password."
	case constants.ReasonPasswordRecoveryResetPassword:
		pwRecoveryReason = "User initiated password reset."
	}

	return &mgmt.UserAccountViewDTO{
		Id:                     model.Id,
		Username:               model.Username,
		Email:                  model.Email,
		DisplayName:            model.DisplayName,
		Locked:                 model.Locked,
		LockedReason:           lockedReason,
		LockedAt:               model.LockedAt,
		FailedAttempts:         model.FailedAttempts,
		PasswordRecoveryId:     model.PasswordRecoveryId,
		PasswordRecoveryReason: pwRecoveryReason,
		PasswordRecoveryAt:     model.PasswordRecoveryCreatedAt,
		LastLoggedInAt:         model.LastLoggedInAt,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
		Role:                   model.Role,
	}
}

func (ua *UserAdapter) ToChangePasswordResposne(result *changePasswordResult) *mgmt.ChangePasswordResponse {
	switch {
	case result.invalidUserAccount:
		return &mgmt.ChangePasswordResponse{
			Status:  "failed",
			Message: "User account not found or inactive.",
		}

	case result.invalidId:
		return &mgmt.ChangePasswordResponse{
			Status:  "failed",
			Message: "Invalid or unknown password recovery link.",
		}

	case result.expired:
		return &mgmt.ChangePasswordResponse{
			Status:  "failed",
			Message: "Password recovery link has expired.",
		}

	case result.oldPasswordDiff:
		return &mgmt.ChangePasswordResponse{
			Status:  "failed",
			Message: "Old password does not match the current password.",
		}

	case result.changed:
		return &mgmt.ChangePasswordResponse{
			Status:  "success",
			Message: "Password changed successfully.",
		}

	default:
		return &mgmt.ChangePasswordResponse{
			Status:  "failed",
			Message: "Password change failed due to an unknown error.",
		}
	}
}

func (ua *UserAdapter) toUserAccountSetupVerficationResponse(result *accountSetupVerficationResult) *mgmt.UserAccountSetupInfoResponse {
	var response mgmt.UserAccountSetupInfoResponse
	response.DisplayName = result.displayName
	response.Email = result.email
	response.ErrorMessage = result.errorMsg
	response.Role = result.role
	response.UserId = result.userId
	response.Username = result.username
	return &response
}

func (ua *UserAdapter) makeGetUserResponse(m *models.UserAccount, role string) *mgmt.GetUserResponse {
	if m == nil {
		return nil
	}

	res := &mgmt.GetUserResponse{
		Id:           m.Id,
		Username:     m.Username,
		Email:        m.Email,
		DisplayName:  m.DisplayName,
		Role:         role,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		LockedAt:     m.LockedAt,
		LockedReason: m.LockedReason,
	}

	return res
}