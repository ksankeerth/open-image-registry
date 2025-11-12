package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type registryCacheStore struct {
	db *sql.DB
}

func newRegistryCacheStore(db *sql.DB) *registryCacheStore {
	return &registryCacheStore{db: db}
}

func (b *registryCacheStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return b.db
}

func (c *registryCacheStore) Create(ctx context.Context, registryId, namespaceId, repositoryId, identifier, digest string, expiresAt time.Time) error {
	q := c.getQuerier(ctx)

	_, err := q.ExecContext(ctx, CacheCreateEntryQuery,
		namespaceId, registryId, repositoryId, identifier, digest, expiresAt,
	)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create registry cache entry")
		return dberrors.ClassifyError(err, CacheCreateEntryQuery)
	}

	return nil
}

func (c *registryCacheStore) Get(ctx context.Context, repositoryId, identifier string) (*models.RegistryCacheModel, error) {
	q := c.getQuerier(ctx)

	row := q.QueryRowContext(ctx, CacheGetEntryQuery, repositoryId, identifier)

	var createdAt, updatedAt string
	var expiresAt string

	var m models.RegistryCacheModel
	err := row.Scan(
		&m.NamespaceID,
		&m.RegistryID,
		&m.RepositoryID,
		&m.Identifier,
		&m.Digest,
		&expiresAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to get registry cache entry")
		return nil, dberrors.ClassifyError(err, CacheGetEntryQuery)
	}

	// Parse timestamps
	exp, err := utils.ParseSqliteTimestamp(expiresAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse expires_at")
		return nil, dberrors.ClassifyError(err, CacheGetEntryQuery)
	}
	m.ExpiresAt = *exp

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse created_at")
		return nil, dberrors.ClassifyError(err, CacheGetEntryQuery)
	}
	m.CreatedAt = *createdTime

	m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse updated_at")
		return nil, dberrors.ClassifyError(err, CacheGetEntryQuery)
	}

	return &m, nil
}

func (c *registryCacheStore) Delete(ctx context.Context, repositoryId, identifier string) (err error) {
	q := c.getQuerier(ctx)

	_, err = q.ExecContext(ctx, CacheDeleteEntryQuery, repositoryId, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete registry cache entry")
		return dberrors.ClassifyError(err, CacheDeleteEntryQuery)
	}

	return nil
}

func (c *registryCacheStore) Refresh(ctx context.Context, repositoryId, identifier string, expiresAt time.Time) error {
	q := c.getQuerier(ctx)

	_, err := q.ExecContext(ctx, CacheRefreshEntryQuery, expiresAt, repositoryId, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to refresh registry cache entry")
		return dberrors.ClassifyError(err, CacheRefreshEntryQuery)
	}

	return nil
}
