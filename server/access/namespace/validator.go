package namespace

import (
	"fmt"
	"slices"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/utils"
)

func validateCreateNamespaceRequest(req *mgmt.CreateNamespaceRequest) (vaild bool, errMsg string) {
	if !utils.IsValidNamespace(req.Name) {
		return false, "Invalid Namespace name"
	}

	if !(req.Purpose == constants.NamespacePurposeProject || req.Purpose == constants.NamespacePurposeTeam) {
		return false, "Namespace purpose not provided"
	}

	if len(req.Maintainers) == 0 {
		return false, "Namespace should have atleast one maintainer"
	}

	return true, ""
}

func validateUpdateNamespaceRequest(req *mgmt.UpdateNamespaceRequest) (valid bool, errMsg string) {
	if req.ID == "" {
		return false, "Invalid Namespace ID in body"
	}

	if !(req.Purpose == constants.NamespacePurposeProject || req.Purpose == constants.NamespacePurposeTeam) {
		return false, "Invalid value Namespace purpose"
	}

	return true, ""
}

func validateNamespaceGrantAccessRequest(req *mgmt.AccessGrantRequest) (valid bool, errMsg string) {
	if req.UserID == "" {
		return false, "Invalid user id"
	}

	if req.GrantedBy == "" {
		return false, "Invalid grantedBy"
	}

	if req.ResourceType != constants.ResourceTypeNamespace {
		return false, "Resource is not Namespace"
	}

	if req.ResourceID == "" {
		return false, "Invalid ResourceID"
	}

	if !(req.AccessLevel == constants.AccessLevelGuest || req.AccessLevel == constants.AccessLevelDeveloper ||
		req.AccessLevel == constants.AccessLevelMaintainer) {
		return false, "Invalid Access Level"
	}

	return true, ""
}

func validateNamesapceRevokeRequest(req *mgmt.AccessRevokeRequest) (valid bool, errMsg string) {
	if req.UserID == "" {
		return false, "Invalid user id"
	}

	if req.ResourceType != constants.ResourceTypeNamespace {
		return false, "Resource is not Namespace"
	}

	if req.ResourceID == "" {
		return false, "Invalid ResourceID"
	}
	return true, ""
}

func validateListNamespaceCondition(cond *store.ListQueryConditions) (bool, string) {
	if cond.SortField != "" && !slices.Contains(constants.AllowedNamespaceSortFields, cond.SortField) {
		return false, fmt.Sprintf("Not allowed sort field: %s", cond.SortField)
	}

	for _, f := range cond.Filters {
		if !slices.Contains(constants.AllowedNamespaceFilterFields, f.Field) {
			return false, fmt.Sprintf("Not allowed filter field: %s", f.Field)
		}

		if f.Field == "is_public" && len(f.Values) != 1 {
			return false, fmt.Sprintf("Boolean value is only accepted for field `is_public`: %v", f.Values)
		}
	}

	return true, ""
}