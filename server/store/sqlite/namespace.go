package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ksankeerth/open-image-registry/config"
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
	var publicNamespace = 0
	if isPublic {
		publicNamespace = 1
	}
	err := q.QueryRowContext(ctx, NamespaceCreateQuery, regId, name, description, purpose, publicNamespace).Scan(&id)
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
	var publicNamespace int
	err := row.Scan(&m.RegistryId, &m.Name, &m.Description, &m.Purpose, &publicNamespace, &m.State, &createdAt, &updatedAt)
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

	if publicNamespace == 1 {
		m.IsPublic = true
	}

	m.Id = id

	return &m, nil
}

func (n *namespaceStore) GetByName(ctx context.Context, regId, name string) (*models.NamespaceModel, error) {
	q := n.getQuerier(ctx)

	row := q.QueryRowContext(ctx, NamespaceGetByNameQuery, regId, name)

	var createdAt, updatedAt string

	var m models.NamespaceModel
	err := row.Scan(&m.RegistryId, &m.Name, &m.Description, &m.Purpose, &m.IsPublic, &m.State, &createdAt, &updatedAt)
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

func (n *namespaceStore) DeleteByIdentifier(ctx context.Context, registryId, identifier string) error {
	q := n.getQuerier(ctx)

	_, err := q.ExecContext(ctx, NamespaceDeleteByIdentifierQuery, registryId, identifier, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete namespace")
		return dberrors.ClassifyError(err, NamespaceDeleteByIdentifierQuery)
	}

	return nil
}

func (n *namespaceStore) ExistsByIdentifier(ctx context.Context, registryId, identifier string) (bool, error) {
	q := n.getQuerier(ctx)

	var exists int
	err := q.QueryRowContext(ctx, NamespaceExistsByIdentifierQuery, registryId, identifier, identifier).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to check namespace existence")
		return false, dberrors.ClassifyError(err, NamespaceExistsByIdentifierQuery)
	}
	return true, nil
}

func (n *namespaceStore) GetByIdentifier(ctx context.Context, registryId, identifier string) (*models.NamespaceModel, error) {
	q := n.getQuerier(ctx)

	m := models.NamespaceModel{}
	var createdAt, updatedAt string
	var isPublic int

	err := q.QueryRowContext(ctx, NamespaceGetByIdentifierQuery, registryId, identifier, identifier).Scan(&m.Id,
		&m.RegistryId, &m.Name, &m.Description, &m.Purpose, &m.IsPublic, &m.State, &createdAt, &updatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve namespace")
		return nil, dberrors.ClassifyError(err, NamespaceGetByIdentifierQuery)
	}

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, NamespaceGetByIdentifierQuery)
		}
		m.CreatedAt = *createdTime
	}

	if updatedAt != "" {
		m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, NamespaceGetByIdentifierQuery)
		}
	}

	if isPublic == 1 {
		m.IsPublic = true
	}

	return &m, nil
}

func (n *namespaceStore) Update(ctx context.Context, id, description, purpose string) error {
	q := n.getQuerier(ctx)

	_, err := q.ExecContext(ctx, NamespaceUpdateQuery, description, purpose, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update namespace")
		return dberrors.ClassifyError(err, NamespaceUpdateQuery)
	}

	return nil
}

func (n *namespaceStore) SetStateByID(ctx context.Context, id, state string) error {
	q := n.getQuerier(ctx)

	_, err := q.ExecContext(ctx, NamespaceSetStateQuery, state, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update state")
		return dberrors.ClassifyError(err, NamespaceSetStateQuery)
	}

	return nil
}

func (n *namespaceStore) SetVisiblityByID(ctx context.Context, id string, public bool) error {
	q := n.getQuerier(ctx)
	var isPublic = 0
	if public {
		isPublic = 1
	}
	_, err := q.ExecContext(ctx, NamespaceSetVisiblityQuery, isPublic, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update namespace visiblity")
		return dberrors.ClassifyError(err, NamespaceSetVisiblityQuery)
	}

	return nil
}

func (n *namespaceStore) List(ctx context.Context, conditions *store.ListQueryConditions) (namespaces []*models.NamespaceView,
	total int, err error) {
	qb := store.NewQueryBuilder(store.DBTypeSqlite).
		WithSearchFields("NAME", "DESCRIPTION").
		WithBooleanField("IS_PUBLIC").
		WithAllowedFilterFields("PURPOSE", "STATE", "IS_PUBLIC").
		WithAllowedSortFields("NAME", "CREATED_AT")

	listQuery, countQuery, args, err := qb.Build(NamespaceListBaseQuery, NamespaceCountBaseQuery, conditions)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to build namespace list query")
		return nil, 0, fmt.Errorf("build query: %w", err)
	}

	// Execute count query first (without limit/offset)
	countArgs := args[:len(args)-2]

	q := n.getQuerier(ctx)

	err = q.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to count total namespaces")
		return nil, 0, fmt.Errorf("count namespaces: %w", err)
	}

	// Execute list query
	rows, err := q.QueryContext(ctx, listQuery, args...)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve namespaces")
		return nil, 0, fmt.Errorf("query namespaces: %w", err)
	}
	defer rows.Close()

	// Scan results
	namespaces = make([]*models.NamespaceView, 0)
	for rows.Next() {
		namespace, err := n.scanNamespaceView(rows)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Failed to scan namespace")
			continue
		}
		namespaces = append(namespaces, namespace)
	}

	if err = rows.Err(); err != nil {
		log.Logger().Error().Err(err).Msg("Error during row iteration")
		return nil, 0, fmt.Errorf("iterate rows: %w", err)
	}

	return namespaces, total, nil
}

func (n *namespaceStore) scanNamespaceView(rows *sql.Rows) (*models.NamespaceView, error) {
	var ns models.NamespaceView
	var createdAt, updatedAt, maintainers, developers sql.NullString
	var isPublic int

	err := rows.Scan(&ns.RegistryID,
		&ns.ID,
		&ns.Name,
		&ns.Description,
		&ns.Purpose,
		&isPublic,
		&ns.State,
		&createdAt,
		&updatedAt,
		&maintainers,
		&developers,
	)
	if err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}

	if createdAt.Valid {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, fmt.Errorf("parse time: %w", err)
		}
		ns.CreatedAt = *createdTime
	}

	if updatedAt.Valid {
		ns.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, fmt.Errorf("parse time: %w", err)
		}
	}

	if maintainers.Valid && maintainers.String != "" {
		ns.Maintainers = strings.Split(maintainers.String, ",")
	}

	if developers.Valid && developers.String != "" {
		ns.Developers = strings.Split(developers.String, ",")
	}

	if isPublic == 1 {
		ns.IsPublic = true
	}

	return &ns, nil
}

func (n *namespaceStore) DeleteAll(ctx context.Context) error {
	if !config.GetTestingConfig().AllowDeleteAll {
		return fmt.Errorf("Not allowed to delete all namespaces.")
	}
	q := n.getQuerier(ctx)

	_, err := q.ExecContext(ctx, NamespaceDeleteAll)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete all namespace")
		return dberrors.ClassifyError(err, NamespaceDeleteAll)
	}

	return nil
}
