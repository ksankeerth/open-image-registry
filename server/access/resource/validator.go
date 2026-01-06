package resource

import (
	"fmt"
	"slices"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/store"
)

func ValidateListUserAccessCondition(cond *store.ListQueryConditions) (bool, string) {
	if cond.SortField != "" && !slices.Contains(constants.AllowedResourceAccessSortFields, cond.SortField) {
		return false, fmt.Sprintf("Not allowed sort field: %s", cond.SortField)
	}

	for _, f := range cond.Filters {
		if !slices.Contains(constants.AllowedResourceAccessFilterFields, f.Field) {
			return false, fmt.Sprintf("Not allowed filter field: %s", f.Field)
		}

		if f.Field == "is_public" && len(f.Values) != 1 {
			return false, fmt.Sprintf("Boolean value is only accepted for field `is_public`: %v", f.Values)
		}
	}

	return true, ""
}