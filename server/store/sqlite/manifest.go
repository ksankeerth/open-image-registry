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

type manifestStore struct {
	db *sql.DB
}

func newManifestStore(db *sql.DB) *manifestStore {
	return &manifestStore{db: db}
}

func (m *manifestStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return m.db
}

func (m *manifestStore) Create(ctx context.Context, registryId, namespaceId, repositoryId, digest, mediaType,
	uniqueDigest string, size int64, content []byte) (string, error) {

	q := m.getQuerier(ctx)

	var id string
	err := q.QueryRowContext(ctx, ManifestCreateQuery,
		digest, size, mediaType, content, namespaceId, registryId, repositoryId, uniqueDigest,
	).Scan(&id)

	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create manifest")
		return "", dberrors.ClassifyError(err, ManifestCreateQuery)
	}

	return id, nil
}

func (m *manifestStore) GetByUniqueDigest(ctx context.Context, withContent bool, repositoryId,
	digest string) (*models.ImageManifestModel, error) {

	q := m.getQuerier(ctx)

	var row *sql.Row
	if withContent {
		row = q.QueryRowContext(ctx, ManifestGetbyUniqueDigestWithContentQuery, digest)
	} else {
		row = q.QueryRowContext(ctx, ManifestGetbyUniqueDigestQuery, digest)
	}

	var createdAt, updatedAt string
	var model models.ImageManifestModel

	if withContent {
		err := row.Scan(&model.ID, &model.Digest, &model.Size, &model.MediaType, &model.Content,
			&model.NamespaceID, &model.RegistryID, &model.RepositoryID, &model.UniqueDigest,
			&createdAt, &updatedAt)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			log.Logger().Error().Err(err).Msg("failed to get manifest by unique digest")
			return nil, dberrors.ClassifyError(err, ManifestGetbyUniqueDigestWithContentQuery)
		}
	} else {
		err := row.Scan(&model.ID, &model.Digest, &model.Size, &model.MediaType,
			&model.NamespaceID, &model.RegistryID, &model.RepositoryID, &model.UniqueDigest,
			&createdAt, &updatedAt)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			log.Logger().Error().Err(err).Msg("failed to get manifest by unique digest")
			return nil, dberrors.ClassifyError(err, ManifestGetbyUniqueDigestQuery)
		}
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse created_at timestamp")
		return nil, dberrors.ClassifyError(err, ManifestGetbyUniqueDigestQuery)
	}
	model.CreatedAt = *createdTime

	model.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse updated_at timestamp")
		return nil, dberrors.ClassifyError(err, ManifestGetbyUniqueDigestQuery)
	}

	return &model, nil
}

func (m *manifestStore) GetByDigest(ctx context.Context, withContent bool, repositoryId,
	digest string) (*models.ImageManifestModel, error) {

	q := m.getQuerier(ctx)

	var row *sql.Row
	if withContent {
		row = q.QueryRowContext(ctx, ManifestGetbyDigestWithContentQuery, digest)
	} else {
		row = q.QueryRowContext(ctx, ManifestGetbyDigestQuery, digest)
	}

	var createdAt, updatedAt string
	var model models.ImageManifestModel

	if withContent {
		err := row.Scan(&model.ID, &model.Digest, &model.Size, &model.MediaType, &model.Content,
			&model.NamespaceID, &model.RegistryID, &model.RepositoryID, &model.UniqueDigest,
			&createdAt, &updatedAt)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			log.Logger().Error().Err(err).Msg("failed to get manifest by digest")
			return nil, dberrors.ClassifyError(err, ManifestGetbyDigestWithContentQuery)
		}
	} else {
		err := row.Scan(&model.ID, &model.Digest, &model.Size, &model.MediaType,
			&model.NamespaceID, &model.RegistryID, &model.RepositoryID, &model.UniqueDigest,
			&createdAt, &updatedAt)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			log.Logger().Error().Err(err).Msg("failed to get manifest by digest")
			return nil, dberrors.ClassifyError(err, ManifestGetbyDigestQuery)
		}
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse created_at timestamp")
		return nil, dberrors.ClassifyError(err, ManifestGetbyDigestQuery)
	}
	model.CreatedAt = *createdTime

	model.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse updated_at timestamp")
		return nil, dberrors.ClassifyError(err, ManifestGetbyDigestQuery)
	}

	return &model, nil
}

func (m *manifestStore) DeleteByDigest(ctx context.Context, repositoryId, digest string) error {
	q := m.getQuerier(ctx)

	_, err := q.ExecContext(ctx, ManifestDeleteByDigest, repositoryId, digest)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete manifest")
		return dberrors.ClassifyError(err, ManifestDeleteByDigest)
	}

	return nil
}