package sqlite

import (
	"context"
	"database/sql"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type namespaceStore struct {
	db *sql.DB
}

func newNamespaceStore(db *sql.DB) *namespaceStore {
	return &namespaceStore{db: db}
}

func (n *namespaceStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return n.db
}

func (n *namespaceStore) Create(ctx context.Context, regId, name, purpose, description string, isPublic bool) (string, error) {
	q := n.getQuerier(ctx)

	var id string
	err := q.QueryRowContext(ctx, NamespaceCreateQuery, regId, name, description, purpose, isPublic).Scan(&id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create namespace")
		return "", dberrors.ClassifyError(err, NamespaceCreateQuery)
	}

	return id, nil
}

func (n *namespaceStore) Get(ctx context.Context, id string) (*models.NamespaceModel, error) {
	q := n.getQuerier(ctx)

	row := q.QueryRowContext(ctx, NamespaceGetQuery, id)

	var createdAt, updatedAt string

	var m models.NamespaceModel
	err := row.Scan(&m.RegistryId, &m.Name, &m.Description, &m.Purpose, &m.IsPublic, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get namespace")
		return nil, dberrors.ClassifyError(err, NamespaceGetQuery)
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to get namespace")
		return nil, dberrors.ClassifyError(err, NamespaceGetQuery)
	}
	m.CreatedAt = *createdTime

	m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to get namespace")
		return nil, dberrors.ClassifyError(err, NamespaceGetQuery)
	}

	return &m, nil
}

func (n *namespaceStore) GetByName(ctx context.Context, regId, name string) (*models.NamespaceModel, error) {
	q := n.getQuerier(ctx)

	row := q.QueryRowContext(ctx, NamespaceGetByNameQuery, regId, name)

	var createdAt, updatedAt string

	var m models.NamespaceModel
	err := row.Scan(&m.RegistryId, &m.Name, &m.Description, &m.Purpose, &m.IsPublic, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get namespace")
		return nil, dberrors.ClassifyError(err, NamespaceGetByNameQuery)
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to get namespace")
		return nil, dberrors.ClassifyError(err, NamespaceGetByNameQuery)
	}
	m.CreatedAt = *createdTime

	m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to get namespace")
		return nil, dberrors.ClassifyError(err, NamespaceGetByNameQuery)
	}

	return &m, nil
}

func (n *namespaceStore) GetID(ctx context.Context, registryId, namespace string) (string, error) {
	q := n.getQuerier(ctx)

	row := q.QueryRowContext(ctx, NamespaceGetIDQuery, registryId, namespace)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get namespace id")
		return "", dberrors.ClassifyError(err, NamespaceGetIDQuery)
	}

	return id, nil
}

func (n *namespaceStore) Delete(ctx context.Context, id string) error {
	q := n.getQuerier(ctx)

	_, err := q.ExecContext(ctx, NamespaceDeleteQuery, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete namespace")
		return dberrors.ClassifyError(err, NamespaceDeleteQuery)
	}

	return nil
}
