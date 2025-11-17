package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type blobMetaStore struct {
	db *sql.DB
}

func newBlobMetaStore(db *sql.DB) *blobMetaStore {
	return &blobMetaStore{db: db}
}

func (b *blobMetaStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return b.db
}

func (b *blobMetaStore) Get(ctx context.Context, digest, repositoryId string) (*models.ImageBlobMetaModel, error) {
	row := b.db.QueryRowContext(ctx, BlobMetaGetQuery, repositoryId, digest)

	var m models.ImageBlobMetaModel
	err := row.Scan(
		&m.NamespaceID,
		&m.RegistryID,
		&m.RepositoryID,
		&m.Digest,
		&m.Size,
		&m.Location,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrivie image blob meta")
		return nil, dberrors.ClassifyError(err, BlobMetaGetQuery)
	}

	return &m, nil
}

func (b *blobMetaStore) Create(ctx context.Context, registryId, namespaceId, repositoryId, digest, location string, size int64) (err error) {
	_, err = b.db.ExecContext(ctx, BlobMetaCreateQuery,
		namespaceId, registryId, repositoryId, digest, size, location,
	)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create image blob meta")
		return dberrors.ClassifyError(err, BlobMetaCreateQuery)
	}
	return nil
}

func (b *blobMetaStore) CreateUploadSession(ctx context.Context, sessionID, namespace, repository string) error {
	q := b.getQuerier(ctx)

	_, err := q.ExecContext(ctx, BlobSessionCreateQuery, sessionID, namespace, repository)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to persist image blob upload session")
		return dberrors.ClassifyError(err, BlobSessionCreateQuery)
	}

	return nil
}

func (b *blobMetaStore) UpdateUploadSession(ctx context.Context, sessionID string, bytesReceived int) error {
	q := b.getQuerier(ctx)

	_, err := q.ExecContext(ctx, BlobSessionUpdateQuery, bytesReceived, sessionID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update image blob upload session")
		return dberrors.ClassifyError(err, BlobSessionUpdateQuery)
	}

	return nil
}

func (b *blobMetaStore) DeleteUploadSession(ctx context.Context, sessionID string) error {
	q := b.getQuerier(ctx)

	_, err := q.ExecContext(ctx, BlobSessionDeleteQuery, sessionID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete image blob upload session")
		return dberrors.ClassifyError(err, BlobSessionDeleteQuery)
	}

	return nil
}

func (b *blobMetaStore) GetUploadSession(ctx context.Context, sessionID string) (*models.ImageBlobUploadSessionModel,
	error) {
	q := b.getQuerier(ctx)
	var createdAt, updatedAt string

	var session models.ImageBlobUploadSessionModel
	err := q.QueryRowContext(ctx, BlobSessionGetQuery, sessionID).Scan(&session.SessionID, &session.Namespace,
		&session.Repository, &session.BytesReceived, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, dberrors.ClassifyError(err, BlobSessionGetQuery)
	}

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, BlobSessionGetQuery)
		}
		session.CreatedAt = *createdTime
	}

	if updatedAt != "" {
		session.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, BlobSessionGetQuery)
		}
	}

	return &session, nil
}
