package sqlite

import (
	"context"
	"database/sql"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
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
