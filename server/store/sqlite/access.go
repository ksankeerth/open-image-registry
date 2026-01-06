package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type resourceAccessStore struct {
	db *sql.DB
}

func newResourceAccessStore(db *sql.DB) *resourceAccessStore {
	return &resourceAccessStore{db: db}
}

func (a *resourceAccessStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return a.db
}

func (a *resourceAccessStore) GrantAccess(ctx context.Context, resourceId, resourceType, userId, accessLevel,
	grantedBy string) (id string, err error) {
	q := a.getQuerier(ctx)

	err = q.QueryRowContext(ctx, ResourceAccessGrantQuery, resourceType, resourceId, userId, accessLevel, grantedBy).Scan(&id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to grant access to resource")
		return "", dberrors.ClassifyError(err, ResourceAccessGrantQuery)
	}

	return id, nil
}

func (a *resourceAccessStore) RevokeAccess(ctx context.Context, resourceId, resourceType,
	userId string) (err error) {
	q := a.getQuerier(ctx)

	_, err = q.ExecContext(ctx, ResourceAccessRevokeQuery, resourceId, resourceType, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to revoke resource access")
		return dberrors.ClassifyError(err, ResourceAccessRevokeQuery)
	}

	return nil
}

func (a *resourceAccessStore) List(ctx context.Context, conditions *store.ListQueryConditions) (entries []*models.ResourceAccessView,
	total int, err error) {

	qb := store.NewQueryBuilder(store.DBTypeSqlite).
		WithSearchFields("user").
		WithAllowedFilterFields("access_level", "resource_type", "resource_id").
		WithFieldTransformation("resource_id", "ra.RESOURCE_ID").
		WithFieldTransformation("resource_type", "ra.RESOURCE_TYPE").
		WithAllowedSortFields("user", "granted_user", "granted_at").
		WithFieldTransformation("granted_at", "CREATED_AT").
		WithFieldTransformation("user_id", "ra.USER_ID").
		WithFieldTransformation("user", "ua.USERNAME")

	listQuery, countQuery, args, err := qb.Build(ResourceAccessListBaseQuery, ResourceAccessCountBaseQuery, conditions)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to build resource access list query")
		return nil, 0, fmt.Errorf("build query: %w", err)
	}

	// Execute count query first (without limit/offset)
	countArgs := args[:len(args)-2]

	q := a.getQuerier(ctx)

	err = q.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to count total resource accesses")
		return nil, 0, fmt.Errorf("count resource accesses: %w", err)
	}

	// Execute list query
	rows, err := q.QueryContext(ctx, listQuery, args...)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve user accounts")
		return nil, 0, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	entries = make([]*models.ResourceAccessView, 0)
	for rows.Next() {
		entry, err := a.scanResourceAccessView(rows)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to read resource access entries")
			return nil, -1, dberrors.ClassifyError(err, listQuery)
		}
		entries = append(entries, entry)
	}

	return entries, total, nil
}

func (a *resourceAccessStore) scanResourceAccessView(rows *sql.Rows) (*models.ResourceAccessView, error) {
	var entry models.ResourceAccessView
	var createdAt string
	var namespace sql.NullString
	var repository sql.NullString

	err := rows.Scan(&entry.ID, &entry.ResourceType, &namespace, &repository, &entry.ResourceID, &entry.UserId,
		&entry.AccessLevel, &entry.Username, &entry.GrantedBy, &entry.GrantedUser, &createdAt)
	if err != nil {
		return nil, err
	}

	if namespace.Valid && namespace.String != "" {
		entry.ResourceName = namespace.String
	}

	if repository.Valid && repository.String != "" {
		entry.ResourceName = repository.String
	}

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("error occurre when parsing sqlite timestamp")
			return nil, err
		}
		entry.GrantedAt = *createdTime
	}

	return &entry, nil
}

func (a *resourceAccessStore) GetUserAccess(ctx context.Context, resourceId, resourceType, userId string) (*models.ResourceAccess,
	error) {
	q := a.getQuerier(ctx)

	m := models.ResourceAccess{}
	var createdAt string

	err := q.QueryRowContext(ctx, ResourceAccessGetByUserAndResourceQuery, userId, resourceType, resourceId).Scan(&m.Id,
		&m.AccessLevel, &m.GrantedBy, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retreive user resource access")
		return nil, dberrors.ClassifyError(err, ResourceAccessGetByUserAndResourceQuery)
	}

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, ResourceAccessGetByUserAndResourceQuery)
		}
		m.CreatedAt = *createdTime
	}

	m.UserId = userId
	m.ResourceType = resourceType
	m.ResourceId = resourceId

	return &m, nil
}
