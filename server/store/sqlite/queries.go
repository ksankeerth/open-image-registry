package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type queries struct {
	db *sql.DB
}

func newQueries(db *sql.DB) *queries {
	return &queries{db: db}
}

func (q *queries) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return q.db
}

func (q *queries) ValidateMaintainers(ctx context.Context, userIds []string) (bool, error) {
	qr := q.getQuerier(ctx)

	placeHolders := strings.Join(slices.Repeat([]string{"?"}, len(userIds)), ",")
	query := fmt.Sprintf(CountMaintainersByIdsQuery, placeHolders)

	var validMaintainers int
	args := make([]any, len(userIds))
	for i, v := range userIds {
		args[i] = v
	}
	err := qr.QueryRowContext(ctx, query, args...).Scan(&validMaintainers)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Warn().Msg("none of users are valid maintainers")
			return false, nil
		}
		log.Logger().Error().Err(err).Msg("failed to validate maintainers")
		return false, dberrors.ClassifyError(err, query)
	}

	return validMaintainers == len(userIds), nil
}

func (q *queries) CountUsersByIDs(ctx context.Context, userIds []string) (int, error) {
	qr := q.getQuerier(ctx)

	userIdPlaceHolders := strings.Join(slices.Repeat([]string{"?"}, len(userIds)), ",")

	query := fmt.Sprintf(CountMatchingUsersByIDs, userIdPlaceHolders)
	var matchingUsers int

	idArgs := make([]any, len(userIds))
	for i, v := range userIds {
		idArgs[i] = v
	}

	err := qr.QueryRowContext(ctx, query, idArgs...).Scan(&matchingUsers)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Warn().Msg("no matching users by ids")
			return -1, nil
		}
		log.Logger().Error().Err(err).Msg("failed to count matching users by ids")
		return -1, dberrors.ClassifyError(err, query)
	}

	return matchingUsers, nil
}

func (q *queries) CountUsersByIDsAndRoles(ctx context.Context, userIds []string, roles []string) (int,
	error) {
	qr := q.getQuerier(ctx)

	userIdPlaceHolders := strings.Join(slices.Repeat([]string{"?"}, len(userIds)), ",")
	rolePlaceHolders := strings.Join(slices.Repeat([]string{"?"}, len(roles)), ",")

	query := fmt.Sprintf(CountMatchingUsersByIDsAndRoles, userIdPlaceHolders, rolePlaceHolders)
	var matchingUsers int

	idArgs := make([]any, len(userIds))
	for i, v := range userIds {
		idArgs[i] = v
	}

	roleArgs := make([]any, len(roles))
	for i, v := range roles {
		roleArgs[i] = v
	}
	allArgs := make([]any, len(idArgs)+len(roleArgs))
	allArgs = append(idArgs, roleArgs...)
	err := qr.QueryRowContext(ctx, query, allArgs...).Scan(&matchingUsers)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Warn().Msg("no matching users")
			return -1, nil
		}
		log.Logger().Error().Err(err).Msg("failed to count matching users")
		return -1, dberrors.ClassifyError(err, query)
	}

	return matchingUsers, nil
}

func (q *queries) CountUsersByResourceAccess(ctx context.Context, userIds []string, resourceType,
	resourceID string) (int, error) {
	qr := q.getQuerier(ctx)

	userIdPlaceHolders := strings.Join(slices.Repeat([]string{"?"}, len(userIds)), ",")

	query := fmt.Sprintf(CountMatchingUsersHaveAccessToResource, userIdPlaceHolders)
	var matchingUsers int

	idArgs := make([]any, len(userIds))
	for i, v := range userIds {
		idArgs[i] = v
	}

	allArgs := []any{resourceType, resourceID}
	allArgs = append(allArgs, idArgs...)

	err := qr.QueryRowContext(ctx, query, allArgs...).Scan(&matchingUsers)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Warn().Msg("no matching users")
			return -1, nil
		}
		log.Logger().Error().Err(err).Msg("failed to count matching users")
		return -1, dberrors.ClassifyError(err, query)
	}

	return matchingUsers, nil
}

func (q *queries) GetManifestByTag(ctx context.Context, withContent bool, repositoryId,
	tag string) (*models.ImageManifestModel, error) {

	qr := q.getQuerier(ctx)

	var manifest models.ImageManifestModel
	var createdAt, updatedAt string

	var err error
	var query string
	if withContent {
		query = GetManifestWithContentByTagQuery
		err = qr.QueryRowContext(ctx, GetManifestWithContentByTagQuery, repositoryId, tag).Scan(&manifest.ID,
			&manifest.Digest, &manifest.Size, &manifest.MediaType, &manifest.Content, &manifest.NamespaceID,
			&manifest.RegistryID, &manifest.RepositoryID, &manifest.UniqueDigest, &createdAt, &updatedAt)
	} else {
		query = GetManifestByTagQuery
		err = qr.QueryRowContext(ctx, GetManifestByTagQuery, repositoryId, tag).Scan(&manifest.ID,
			&manifest.Digest, &manifest.Size, &manifest.MediaType, &manifest.NamespaceID,
			&manifest.RegistryID, &manifest.RepositoryID, &manifest.UniqueDigest, &createdAt, &updatedAt)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve manifest by tag")
		return nil, dberrors.ClassifyError(err, query)
	}

	if createdAt != "" {
		createdtime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, query)
		}
		manifest.CreatedAt = *createdtime
	}

	if updatedAt != "" {
		manifest.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, query)
		}
	}

	return &manifest, nil
}

func (q *queries) GetRepositoryByNames(ctx context.Context, namespace, repository string) (*models.RepositoryModel,
	error) {
	qr := q.getQuerier(ctx)

	var m models.RepositoryModel
	var createdAt, updatedAt string
	var isPublic int

	err := qr.QueryRowContext(ctx, GetRepositoryByNamesQuery, namespace, repository).Scan(&m.ID, &m.Name, &m.Description,
		&isPublic, &m.State, &m.NamespaceID, &m.RegistryID, &createdAt, &updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		log.Logger().Error().Err(err).Msg("failed to retrieve repository")
		return nil, dberrors.ClassifyError(err, GetRepositoryByNamesQuery)
	}

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, GetRepositoryByNamesQuery)
		}
		m.CreatedAt = *createdTime
	}

	if updatedAt != "" {
		m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, GetRepositoryByNamesQuery)
		}
	}

	if isPublic == 0 {
		m.IsPublic = false
	} else {
		m.IsPublic = true
	}

	return &m, nil
}
