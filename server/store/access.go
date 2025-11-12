package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type ResourceAccessStore interface {
	GrantAccess(ctx context.Context, resourceId, resourceType, userId, accessLevel, grantedBy string) (id string, err error)

	RevokeAccess(ctx context.Context, resourceId, resourceType, userId string) (err error)

	List(ctx context.Context, conditions *ListQueryConditions) (entries []*models.ResourceAccessView, total int, err error)
}
