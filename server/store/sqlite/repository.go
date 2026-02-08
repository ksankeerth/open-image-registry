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
	isPublic bool, createdBy string) (string, error) {

	q := r.getQuerier(ctx)

	var id string
	err := q.QueryRowContext(ctx, RepositoryCreateQuery, name, description, isPublic, nsId, regId, createdBy).Scan(&id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create repository")
		return "", dberrors.ClassifyError(err, RepositoryCreateQuery)
	}

	return id, nil
}

func (r *repositoryStore) Exists(ctx context.Context, id string) (bool, error) {
	q := r.getQuerier(ctx)

	var exists int
	err := q.QueryRowContext(ctx, RepositoryExistsQuery, id).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to check repository by identifier")
		return false, dberrors.ClassifyError(err, RepositoryExistsQuery)
	}

	return true, nil
}

func (r *repositoryStore) Get(ctx context.Context, id string) (*models.RepositoryModel, error) {
	q := r.getQuerier(ctx)

	row := q.QueryRowContext(ctx, RepositoryGetQuery, id)

	var createdAt, updatedAt string

	var m models.RepositoryModel
	err := row.Scan(&m.ID, &m.Name, &m.Description, &m.IsPublic, &m.State, &m.NamespaceID, &m.RegistryID, &createdAt, &updatedAt, &m.CreatedBy)
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

func (r *repositoryStore) Update(ctx context.Context, id, description string) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, RepositoryUpdateQuery, description, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update repository")
		return dberrors.ClassifyError(err, RepositoryUpdateQuery)
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

func (r *repositoryStore) SetState(ctx context.Context, id, newState string) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, RepositorySetStateQuery, newState, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to change repository state")
		return dberrors.ClassifyError(err, RepositorySetStateQuery)
	}

	return nil
}

func (r *repositoryStore) SetVisibility(ctx context.Context, id string, public bool) error {
	q := r.getQuerier(ctx)

	var isPublic = 0
	if public {
		isPublic = 1
	}

	_, err := q.ExecContext(ctx, RepositorySetVisiblityQuery, isPublic, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to change repository visiblity")
		return dberrors.ClassifyError(err, RepositorySetVisiblityQuery)
	}
	return nil
}

func (r *repositoryStore) SetStateByNamespaceID(ctx context.Context, namespaceId, state string) (count int64, err error) {
	q := r.getQuerier(ctx)

	res, err := q.ExecContext(ctx, RepositorySetStateByNamespaceQuery, state, namespaceId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update repository state by namespace")
		return -1, dberrors.ClassifyError(err, RepositorySetStateByNamespaceQuery)
	}

	count, err = res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to read sql result")
		return -1, dberrors.ClassifyError(err, RepositorySetStateByNamespaceQuery)
	}
	return count, nil
}

func (r *repositoryStore) SetVisiblityByNamespaceID(ctx context.Context, namespaceId string, public bool) (count int64, err error) {
	q := r.getQuerier(ctx)

	var isPublic = 0
	if public {
		isPublic = 1
	}

	res, err := q.ExecContext(ctx, RepositorySetVisiblityByNamespaceQuery, isPublic, namespaceId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update repository visiblity by namespace")
		return -1, dberrors.ClassifyError(err, RepositorySetVisiblityByNamespaceQuery)
	}

	count, err = res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to read sql result")
		return -1, dberrors.ClassifyError(err, RepositorySetVisiblityByNamespaceQuery)
	}
	return count, nil
}

func (r *repositoryStore) List(ctx context.Context, conditions *store.ListQueryConditions) (repositories []*models.RepositoryView,
	total int, err error) {
	qb := store.NewQueryBuilder(store.DBTypeSqlite).
		WithSearchFields("NAME", "DESCRIPTION").
		WithBooleanField("IS_PUBLIC").
		WithAllowedSortFields("NAME", "TAGS", "CREATED_AT").
		WithAllowedFilterFields("STATE", "IS_PUBLIC", "TAGS", "NAMESPACE_ID")

	listQuery, countQuery, args, err := qb.Build(RepositoryListBaseQuery, RepositoryGetIDQuery, conditions)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to build repository list query")
		return nil, 0, fmt.Errorf("build query: %w", err)
	}

	// Execute count query first (without limit/offset)
	countArgs := args[:len(args)-2]

	q := r.getQuerier(ctx)

	err = q.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to count total repositories")
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// Execute list query
	rows, err := q.QueryContext(ctx, listQuery, args...)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve repositories")
		return nil, 0, fmt.Errorf("query repositories: %w", err)
	}
	defer rows.Close()

	repositories = make([]*models.RepositoryView, 0)
	for rows.Next() {
		repository, err := r.scanRepositoryView(rows)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Failed to scan repository")
			continue
		}
		repositories = append(repositories, repository)
	}

	// Check for row iteration errors
	if err = rows.Err(); err != nil {
		log.Logger().Error().Err(err).Msg("Error during row iteration")
		return nil, 0, fmt.Errorf("iterate rows: %w", err)
	}

	return repositories, total, nil
}

func (r *repositoryStore) scanRepositoryView(rows *sql.Rows) (*models.RepositoryView, error) {
	var createdAt, updatedAt sql.NullString
	var isPublic int

	var repo models.RepositoryView

	err := rows.Scan(&repo.RegistryID, &repo.ID, &repo.Name, &repo.Description, &repo.Namespace, &isPublic, &repo.State,
		&repo.NamespaceID, &createdAt, &updatedAt, &repo.CreatedBy, &repo.TagsCount)
	if err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}

	if createdAt.Valid {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, err
		}
		repo.CreatedAt = *createdTime
	}

	if updatedAt.Valid {
		repo.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, err
		}
	}

	if isPublic == 1 {
		repo.IsPublic = true
	}

	return &repo, nil
}

// identifier can be name or id
func (r *repositoryStore) ExistsByIdentifier(ctx context.Context, namesapceID, identifier string) (bool, error) {
	q := r.getQuerier(ctx)

	var exists int
	err := q.QueryRowContext(ctx, RepositoryExistsByIdentifierQuery, namesapceID, identifier, identifier).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to check repository by identifier")
		return false, dberrors.ClassifyError(err, RepositoryExistsByIdentifierQuery)
	}

	return true, nil
}

// identifier can be name or id
func (r *repositoryStore) DeleteByIdentifier(ctx context.Context, namesapceID, identifier string) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, RepositoryDeleteByIdentifierQuery, namesapceID, identifier, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete repository by identifier")
		return dberrors.ClassifyError(err, RepositoryDeleteByIdentifierQuery)
	}
	return nil
}

// identifier can be name or id
func (r *repositoryStore) GetByIdentifier(ctx context.Context, namesapceID, identifier string) (*models.RepositoryModel,
	error) {
	q := r.getQuerier(ctx)

	row := q.QueryRowContext(ctx, RepositoryGetByIdentifierQuery, namesapceID, identifier, identifier)

	var createdAt, updatedAt string

	var m models.RepositoryModel
	err := row.Scan(&m.ID, &m.Name, &m.Description, &m.IsPublic, &m.NamespaceID, &m.RegistryID, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get repository")
		return nil, dberrors.ClassifyError(err, RepositoryGetByIdentifierQuery)
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse repository created_at")
		return nil, dberrors.ClassifyError(err, RepositoryGetByIdentifierQuery)
	}
	m.CreatedAt = *createdTime

	m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse repository updated_at")
		return nil, dberrors.ClassifyError(err, RepositoryGetByIdentifierQuery)
	}

	return &m, nil
}
