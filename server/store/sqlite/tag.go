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

type imageTagStore struct {
	db *sql.DB
}

func newImageStore(db *sql.DB) *imageTagStore {
	return &imageTagStore{db: db}
}

func (t *imageTagStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return t.db
}

func (t *imageTagStore) Create(ctx context.Context, registryId, namespaceId, repositoryId, tag string) (id string, err error) {

	q := t.getQuerier(ctx)

	err = q.QueryRowContext(ctx, TagCreateQuery, registryId, namespaceId, repositoryId, tag).Scan(&id)

	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create image tag")
		return "", dberrors.ClassifyError(err, TagCreateQuery)
	}

	return id, nil
}

// Get retrieves a tag by repository + name
func (t *imageTagStore) Get(
	ctx context.Context,
	repositoryId, tag string,
) (*models.ImageTagModel, error) {

	q := t.getQuerier(ctx)

	var createdAt, updatedAt string
	var m models.ImageTagModel

	err := q.QueryRowContext(ctx, TagGetQuery,
		repositoryId, tag,
	).Scan(
		&m.Id,
		&m.RegistryId,
		&m.NamespaceId,
		&m.RepositoryId,
		&m.Tag,
		&m.IsStable,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to get image tag")
		return nil, dberrors.ClassifyError(err, TagGetQuery)
	}

	// timestamps
	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse image tag updated_at")
		return nil, dberrors.ClassifyError(err, TagGetQuery)
	}

	m.CreatedAt = *createdTime

	if updatedAt != "" {
		parsed, err := utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse image tag updated_at")
			return nil, dberrors.ClassifyError(err, TagGetQuery)
		}
		m.UpdatedAt = parsed
	}

	return &m, nil
}

// Delete removes a tag
func (t *imageTagStore) Delete(ctx context.Context, repositoryId, tag string) error {

	q := t.getQuerier(ctx)

	_, err := q.ExecContext(ctx, TagDeleteQuery, repositoryId, tag)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete image tag")
		return dberrors.ClassifyError(err, TagDeleteQuery)
	}
	return nil
}

func (t *imageTagStore) LinkManifest(ctx context.Context, tagId, manifestId string) error {

	q := t.getQuerier(ctx)

	_, err := q.ExecContext(ctx, TagLinkManifestQuery, manifestId, tagId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to link manifest to tag")
		return dberrors.ClassifyError(err, TagLinkManifestQuery)
	}
	return nil
}

func (t *imageTagStore) UpdateManifest(ctx context.Context, tagId, manifestId string) error {

	q := t.getQuerier(ctx)

	_, err := q.ExecContext(ctx, TagUpdateManifestQuery, manifestId, tagId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update manifest link")
		return dberrors.ClassifyError(err, TagUpdateManifestQuery)
	}
	return nil
}

func (t *imageTagStore) UnlinkManifest(ctx context.Context, tagId string) error {

	q := t.getQuerier(ctx)

	_, err := q.ExecContext(ctx, TagUnlinkManifestQuery, tagId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to unlink manifest")
		return dberrors.ClassifyError(err, TagUnlinkManifestQuery)
	}
	return nil
}

func (t *imageTagStore) GetManifestID(ctx context.Context, tagId string) (string, error) {
	q := t.getQuerier(ctx)

	row := q.QueryRowContext(ctx, TagGetManifestID, tagId)

	var manifestId string
	err := row.Scan(&manifestId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get manifest id for tag")
		return "", dberrors.ClassifyError(err, TagGetManifestID)
	}

	return manifestId, nil
}
