package user

import (
	"fmt"
	"slices"

	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/query"
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

	if !(req.Role == RoleAdmin || req.Role == RoleMaintainer || req.Role == RoleDeveloper || req.Role == RoleGuest) {
		return false, fmt.Sprintf("Invalid role: %s", req.Role)
	}

	return true, ""
}

func ValidateUpdateUserAccount(req *mgmt.UpdateUserAccountRequest) (bool, string) {
	if req.Email == "" {
		return false, "Invalid email"
	}
	if req.Role == "" {
		return false, "Invalid Role"
	}

	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		return false, "Invalid email"
	}

	if req.DisplayName != "" && len(req.DisplayName) > 255 {
		return false, "Display name is too long"
	}

	if req.Role != "" {
		if !(req.Role == RoleAdmin || req.Role == RoleDeveloper || req.Role == RoleGuest || req.Role == RoleMaintainer) {
			return false, "Invalid Role"
		}
	}

	return true, ""
}

func ValidateUpdateUserEmail(req *mgmt.UpdateUserEmailRequest) (bool, string) {
	if !utils.IsValidEmail(req.Email) {
		return false, "Invalid email"
	}
	return true, ""
}

func ValidateUpdateDisplayName(req *mgmt.UpdateUserDisplayNameRequest) (bool, string) {
	if req.DisplayName != "" {
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

func ValidateListUserCondition(cond *query.ListModelsConditions) (bool, string) {

	if cond.Sort.Field != "" && !slices.Contains(AllowedUserSortFields, cond.Sort.Field) {
		return false, fmt.Sprintf("Not allowed sort fied: %s", cond.Sort.Field)
	}

	for _, f := range cond.Filters {
		if !slices.Contains(AllowedUserFilterFields, f.Field) {
			return false, fmt.Sprintf("Not allowed filter fied: %s", f.Field)
		}

		if f.Field == "locked" && len(f.Values) != 1 {
			return false, fmt.Sprintf("Boolean value is only accepted for field `locked`: %v", f.Values...)
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