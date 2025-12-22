package user

import (
	"fmt"
	"slices"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/utils"
)

func ValidateCreateUserAccount(req *mgmt.CreateUserAccountRequest) (bool, string) {
	if !utils.IsValidUsername(req.Username) {
		return false, "Invalid username"
	}

	if !utils.IsValidEmail(req.Email) {
		return false, "Invalid email"
	}

	if req.Role == "" {
		return false, "Role is not set"
	}

	if !(req.Role == constants.RoleAdmin || req.Role == constants.RoleMaintainer || req.Role == constants.RoleDeveloper ||
		req.Role == constants.RoleGuest) {
		return false, fmt.Sprintf("Invalid role: %s", req.Role)
	}

	return true, ""
}

func ValidateUpdateUserEmail(req *mgmt.UpdateUserEmailRequest) (bool, string) {
	if !utils.IsValidEmail(req.Email) {
		return false, "Invalid email"
	}
	return true, ""
}

func ValidateUpdateUserAccount(req *mgmt.UpdateUserAccountRequest) (bool, string) {
	if req.DisplayName == "" || len(req.DisplayName) > 255 {
		return false, "Invalid user display name"
	}

	return true, ""
}

func ValidateUserValidateRequest(req *mgmt.UsernameEmailValidationRequest) (bool, string) {
	if req.Email == "" && req.Username == "" {
		return false, "Both email and username cannot be empty"
	}

	return true, ""
}

func ValidateListUserCondition(cond *store.ListQueryConditions) (bool, string) {

	if cond.SortField != "" && !slices.Contains(constants.AllowedUserSortFields, cond.SortField) {
		return false, fmt.Sprintf("Not allowed sort fied: %s", cond.SortField)
	}

	for _, f := range cond.Filters {
		if !slices.Contains(constants.AllowedUserFilterFields, f.Field) {
			return false, fmt.Sprintf("Not allowed filter fied: %s", f.Field)
		}

		if f.Field == "locked" && len(f.Values) != 1 {
			return false, fmt.Sprintf("Boolean value is only accepted for field `locked`: %v", f.Values)
		}
	}

	return true, ""
}

func ValidateAccountSetupCompleteRequest(req *mgmt.AccountSetupCompleteRequest) (bool, string) {

	if req.UserId == "" {
		return false, "Invalid User Id"
	}

	if !utils.IsValidUsername(req.Username) {
		return false, "Invalid username"
	}

	valid, msg := utils.ValidatePassword(req.Password)
	if valid {
		return true, ""
	}
	return false, msg
}

func isValidRole(role string) bool {
	if !(role == constants.RoleAdmin || role == constants.RoleDeveloper ||
		role == constants.RoleGuest || role == constants.RoleMaintainer) {
		return false
	}

	return true
}