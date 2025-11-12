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

type repositoryStore struct {
	db *sql.DB
}

func newRepositoryStore(db *sql.DB) *repositoryStore {
	return &repositoryStore{db: db}
}

func (r *repositoryStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *repositoryStore) Create(ctx context.Context, regId, nsId, name, description string,
	isPublic bool) (string, error) {

	q := r.getQuerier(ctx)

	var id string
	err := q.QueryRowContext(ctx, RepositoryCreateQuery, name, description, isPublic, nsId, regId).Scan(&id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create repository")
		return "", dberrors.ClassifyError(err, RepositoryCreateQuery)
	}

	return id, nil
}

func (r *repositoryStore) Get(ctx context.Context, id string) (*models.RepositoryModel, error) {
	q := r.getQuerier(ctx)

	row := q.QueryRowContext(ctx, RepositoryGetQuery, id)

	var createdAt, updatedAt string

	var m models.RepositoryModel
	err := row.Scan(&m.ID, &m.Name, &m.Description, &m.IsPublic, &m.NamespaceID, &m.RegistryID, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get repository")
		return nil, dberrors.ClassifyError(err, RepositoryGetQuery)
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse repository created_at")
		return nil, dberrors.ClassifyError(err, RepositoryGetQuery)
	}
	m.CreatedAt = *createdTime

	m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse repository updated_at")
		return nil, dberrors.ClassifyError(err, RepositoryGetQuery)
	}

	return &m, nil
}

func (r *repositoryStore) Delete(ctx context.Context, id string) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, RepositoryDeleteQuery, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete repository")
		return dberrors.ClassifyError(err, RepositoryDeleteQuery)
	}

	return nil
}

func (r *repositoryStore) GetID(ctx context.Context, namespaceId, name string) (string, error) {
	q := r.getQuerier(ctx)

	row := q.QueryRowContext(ctx, RepositoryGetIDQuery, namespaceId, name)

	var id string
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get repository id")
		return "", dberrors.ClassifyError(err, RepositoryGetIDQuery)
	}

	return id, nil
}
