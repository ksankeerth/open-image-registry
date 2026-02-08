package repository

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/utils"
)

func ValidateListRepositoryCondition(cond *store.ListQueryConditions) (valid bool, errMsg string) {
	if cond.SortField != "" && slices.Contains(constants.AllowedRepositorySortFields, cond.SortField) {
		return false, fmt.Sprintf("Not allowed sort fied: %s", cond.SortField)
	}

	for _, f := range cond.Filters {
		if !slices.Contains(constants.AllowedRepositoryFilterFields, f.Field) {
			return false, fmt.Sprintf("Not allowed filter fied: %s", f.Field)
		}

		if f.Field == constants.FilterFieldIsPublic && len(f.Values) != 1 {
			return false, fmt.Sprintf("Boolean value is only accepted for field `is_public`: %v", f.Values)
		}

		if f.Field == constants.FilterFieldTagCount {
			if len(f.Values) > 2 {
				return false, fmt.Sprintf("`tags` range filter can have only 2 values at maximum: %v", f.Values)
			}
			var gtMatches = 0
			var ltMatches = 0
			for _, v := range f.Values {
				vstr, ok := v.(string)
				if ok && strings.HasPrefix(vstr, ">") {
					gtMatches++
				}
				if ok && strings.HasPrefix(vstr, "<") {
					ltMatches++
				}
			}

			if ltMatches == 2 || gtMatches == 2 || (ltMatches == 0 && gtMatches == 0) {
				return false, fmt.Sprintf("`tags` range filter invalid values: %v", f.Values)
			}
		}
	}

	return true, ""
}

// TODO: We have to write this after multi platform support so we can plan better UI and better APIs
// func validateListTagsCondition(cond *store.ListQueryConditions) (valid bool, errMsg string) {
// 	if cond.SortField != "" && slices.Contains(constants.All, cond.SortField) {
// 		return false, fmt.Sprintf("Not allowed sort fied: %s", cond.SortField)
// 	}

// 	for _, f := range cond.Filters {
// 		if !slices.Contains(constants.AllowedRepositoryFilterFields, f.Field) {
// 			return false, fmt.Sprintf("Not allowed filter fied: %s", f.Field)
// 		}

// 		if f.Field == constants.FilterFieldIsPublic && len(f.Values) != 1 {
// 			return false, fmt.Sprintf("Boolean value is only accepted for field `is_public`: %v", f.Values)
// 		}

// 		if f.Field == constants.FilterFieldTagCount {
// 			if len(f.Values) > 2 {
// 				return false, fmt.Sprintf("`tags` range filter can have only 2 values at maximum: %v", f.Values)
// 			}
// 			var gtMatches = 0
// 			var ltMatches = 0
// 			for _, v := range f.Values {
// 				vstr, ok := v.(string)
// 				if ok && strings.HasPrefix(vstr, ">") {
// 					gtMatches++
// 				}
// 				if ok && strings.HasPrefix(vstr, "<") {
// 					ltMatches++
// 				}
// 			}

// 			if ltMatches == 2 || gtMatches == 2 || (ltMatches == 0 && gtMatches == 0) {
// 				return false, fmt.Sprintf("`tags` range filter invalid values: %v", f.Values)
// 			}
// 		}
// 	}

// 	return true, ""
// }

func validateCreateRepositoryRequest(req *mgmt.CreateRepositoryRequest) (vaild bool, errMsg string) {
	if !utils.IsValidRepository(req.Name) {
		return false, "Invalid Repository name"
	}

	if req.NamespaceId == "" {
		return false, "Invalid Namespace"
	}

	return true, ""
}

func validateRepositoryGrantAccessRequest(req *mgmt.AccessGrantRequest) (valid bool, errMsg string) {
	if req.UserID == "" {
		return false, "Invalid user id"
	}

	if req.GrantedBy == "" {
		return false, "Invalid grantedBy"
	}

	if req.ResourceType != constants.ResourceTypeRepository {
		return false, "Resource is not Repository"
	}

	if req.ResourceID == "" {
		return false, "Invalid ResourceID"
	}

	if !(req.AccessLevel == constants.AccessLevelGuest || req.AccessLevel == constants.AccessLevelDeveloper) {
		return false, "Invalid Access Level"
	}

	return true, ""
}