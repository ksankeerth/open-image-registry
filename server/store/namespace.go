package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type NamespaceStore interface {
	Create(ctx context.Context, regId, name, purpose, description string, isPublic bool) (id string, err error)

	Get(ctx context.Context, id string) (*models.NamespaceModel, error)

	GetByName(ctx context.Context, regId, name string) (*models.NamespaceModel, error)

	GetID(ctx context.Context, registryId, namesapce string) (id string, err error)

	Delete(ctx context.Context, id string) (err error)

	Update(ctx context.Context, id, description, purpose string) error

	// identifier can be name or id
	ExistsByIdentifier(ctx context.Context, registryId, identifier string) (bool, error)

	// identifier can be name or id
	DeleteByIdentifier(ctx context.Context, registryId, identifier string) error

	// identifier can be name or id
	GetByIdentifier(ctx context.Context, registryId, identifier string) (*models.NamespaceModel, error)

	SetStateByID(ctx context.Context, id, state string) error

	SetVisiblityByID(ctx context.Context, id string, public bool) error

	List(ctx context.Context, conditions *ListQueryConditions) (users []*models.NamespaceView, total int, err error)

	// DeleteAll deletes all the records from table. The implementation should only this method only
	// when `testing.allow_delete_all` is set to true.
	DeleteAll(ctx context.Context) error
}