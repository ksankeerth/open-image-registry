package resource

import (
	"context"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
)

// GetUserAccessByLevels find all accesses of resources the user has.
// Currently, this method only supports upto maximum 2 access levels. If nil or empty access levels
// are passed, it will return all the resource access.
func (m *Manager) GetUserAccessByLevels(ctx context.Context, userId string, page, limit uint,
	accessLevels ...string) (accesses []*models.ResourceAccessView, total int, err error) {

	queryConds := store.ListQueryConditions{
		Page:  page,
		Limit: limit,
		Filters: []store.Filter{
			{
				Field:    constants.FilterFieldUserID,
				Operator: store.OpEqual,
				Values:   []any{userId},
			},
		},
	}

	if len(accessLevels) > 0 {
		queryConds.Filters = append(queryConds.Filters, store.Filter{
			Field:    constants.FilterFieldUserID,
			Operator: store.OpIn,
			Values:   []any{accessLevels},
		})
	}

	return m.store.Access().List(ctx, &queryConds)
}