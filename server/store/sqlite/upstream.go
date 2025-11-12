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

type upstreamStore struct {
	db *sql.DB
}

func newUpstreamStore(db *sql.DB) *upstreamStore {
	return &upstreamStore{db: db}
}

func (u *upstreamStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return u.db
}

func (u *upstreamStore) CreateRegistry(ctx context.Context, m *models.UpstreamRegistry) (id string, err error) {
	q := u.getQuerier(ctx)

	err = q.QueryRowContext(ctx, UpstreamCreateQuery, m.Name, m.Description, m.Vendor, m.State, m.Port, m.UpstreamURL).Scan(&id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create upstream registry")
		return "", dberrors.ClassifyError(err, UpstreamCreateQuery)
	}

	return id, nil
}

func (u *upstreamStore) UpdateRegistry(ctx context.Context, m *models.UpstreamRegistry) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UpstreamUpdateQuery, m.Description, m.State, m.Port, m.UpstreamURL, m.ID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update upstream registry")
		return dberrors.ClassifyError(err, UpstreamUpdateQuery)
	}

	return nil
}

func (u *upstreamStore) GetRegistry(ctx context.Context, registryID string) (*models.UpstreamRegistry, error) {
	q := u.getQuerier(ctx)

	var m models.UpstreamRegistry

	var createdAt, updatedAt string
	err := q.QueryRowContext(ctx, UpstreamGetQuery, registryID).Scan(&m.ID, &m.Name, &m.Description, &m.Vendor, &m.State, &m.Port, &m.UpstreamURL,
		&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve upstream registry")
		return nil, dberrors.ClassifyError(err, UpstreamGetQuery)
	}

	if createdAt == "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetQuery)
		}
		m.CreatedAt = *createdTime
	}

	if updatedAt == "" {
		m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetQuery)
		}
	}

	return &m, nil
}

func (u *upstreamStore) DeleteRegistry(ctx context.Context, registryID string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UpstreamDeleteQuery, registryID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete upstream registry")
		return dberrors.ClassifyError(err, UpstreamDeleteQuery)
	}
	return nil
}

func (u *upstreamStore) ChangeRegistryState(ctx context.Context, registryID, state string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UpstreamChangeStateQuery, state, registryID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to change state of upstream registry")
		return dberrors.ClassifyError(err, UpstreamChangeStateQuery)
	}
	return nil
}

func (u *upstreamStore) PersistRegistryAuthConfig(ctx context.Context, m *models.UpstreamRegistryAuthConfig) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UpstreamPersistAuthConfigQuery, m.AuthType, m.ConfigJSON, m.RegistryID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to persist upstream registry auth config")
		return dberrors.ClassifyError(err, UpstreamPersistAuthConfigQuery)
	}

	return nil
}

func (u *upstreamStore) UpdateRegistryAuthConfig(ctx context.Context, m *models.UpstreamRegistryAuthConfig) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UpstreamGetAuthConfigQuery, m.AuthType, m.ConfigJSON, m.RegistryID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update upstream registry auth config")
		return dberrors.ClassifyError(err, UpstreamGetAuthConfigQuery)
	}

	return nil
}

func (u *upstreamStore) GetRegistryAuthConfig(ctx context.Context, registryID string) (*models.UpstreamRegistryAuthConfig, error) {
	q := u.getQuerier(ctx)

	var m models.UpstreamRegistryAuthConfig

	var createdAt, updatedAt string

	err := q.QueryRowContext(ctx, UpstreamGetAuthConfigQuery, registryID).
		Scan(&m.AuthType, &m.ConfigJSON, &createdAt, &updatedAt)

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetCacheConfigQuery)
		}
		m.CreatedAt = *createdTime
	}

	if updatedAt != "" {
		m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetCacheConfigQuery)
		}
	}

	return &m, nil
}

func (u *upstreamStore) PersistRegistryCacheConfig(ctx context.Context, m *models.UpstreamRegistryCacheStoreConfig) error {
	q := u.getQuerier(ctx)

	var cacheEnabled int
	if m.CacheEnabled {
		cacheEnabled = 1
	}
	_, err := q.ExecContext(ctx, UpstreamPersistCacheConfigQuery, m.RegistryID, cacheEnabled, m.TTLSeconds, m.StorageLimit, m.CleanupThreshold)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to persist upstream registry cache config")
		return dberrors.ClassifyError(err, UpstreamPersistCacheConfigQuery)
	}

	return nil
}

func (u *upstreamStore) UpdateRegistryCacheConfig(ctx context.Context, m *models.UpstreamRegistryCacheStoreConfig) error {
	q := u.getQuerier(ctx)

	var cacheEnabled int
	if m.CacheEnabled {
		cacheEnabled = 1
	}
	_, err := q.ExecContext(ctx, UpstreamUpdateCacheConfigQuery, cacheEnabled, m.TTLSeconds, m.StorageLimit,
		m.CleanupThreshold, m.RegistryID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update upstream registry cache config")
		return dberrors.ClassifyError(err, UpstreamUpdateCacheConfigQuery)
	}

	return nil
}

func (u *upstreamStore) GetRegistryCacheConfig(ctx context.Context, registryID string) (*models.UpstreamRegistryCacheStoreConfig, error) {
	q := u.getQuerier(ctx)

	var m models.UpstreamRegistryCacheStoreConfig
	var createdAt, updatedAt string
	var cacheEnabled int

	err := q.QueryRowContext(ctx, UpstreamGetCacheConfigQuery, registryID).
		Scan(&cacheEnabled, &m.TTLSeconds, &m.StorageLimit, &m.CleanupThreshold, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve registry cache config")
		return nil, dberrors.ClassifyError(err, UpstreamGetCacheConfigQuery)
	}

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetCacheConfigQuery)
		}
		m.CreatedAt = *createdTime
	}

	if updatedAt != "" {
		m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetCacheConfigQuery)
		}
	}

	if cacheEnabled == 1 {
		m.CacheEnabled = true
	}

	return &m, nil
}

func (u *upstreamStore) PersistRegistryNetworkConfig(ctx context.Context, m *models.UpstreamRegistryNetworkConfig) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UpstreamPersistNetworkConfigQuery, m.RegistryID, m.ConnectionTimeout, m.ReadTimeout,
		m.WriteTimeout, m.MaxConnections, m.MaxIdleConnections, m.MaxRetries, m.RetryDelay, m.RetryBackOffMultiplier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to persist upstream registry network config")
		return dberrors.ClassifyError(err, UpstreamPersistNetworkConfigQuery)
	}

	return nil
}

func (u *upstreamStore) UpdateRegistryNetworkConfig(ctx context.Context, m *models.UpstreamRegistryNetworkConfig) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UpstreamUpdateNetworkConfigQuery, m.ConnectionTimeout, m.ReadTimeout, m.WriteTimeout,
		m.MaxConnections, m.MaxIdleConnections, m.MaxRetries, m.RetryDelay, m.RetryBackOffMultiplier, m.RegistryID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update upstream registry network config")
		return dberrors.ClassifyError(err, UpstreamUpdateNetworkConfigQuery)
	}

	return nil
}

func (u *upstreamStore) GetRegistryNetworkConfig(ctx context.Context, registryID string) (*models.UpstreamRegistryNetworkConfig, error) {
	q := u.getQuerier(ctx)
	var m models.UpstreamRegistryNetworkConfig
	var createdAt, updatedAt string

	err := q.QueryRowContext(ctx, UpstreamGetNetworkConfigQuery, registryID).Scan(&m.ConnectionTimeout, &m.ReadTimeout, &m.WriteTimeout, &m.MaxConnections,
		&m.MaxIdleConnections, &m.MaxRetries, &m.RetryDelay, &m.RetryBackOffMultiplier, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve network config")
		return nil, dberrors.ClassifyError(err, UpstreamGetNetworkConfigQuery)
	}

	if createdAt != "" {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetNetworkConfigQuery)
		}
		m.CreatedAt = *createdTime
	}

	if updatedAt != "" {
		m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, UpstreamGetNetworkConfigQuery)
		}
	}

	return &m, nil
}

func (u *upstreamStore) GetAllUpstreamRegistryAddresses(ctx context.Context) (addresses []*models.UpstreamAddressView, err error) {
	q := u.getQuerier(ctx)

	rows, err := q.QueryContext(ctx, UpstreamGetAllAddresses)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve upstream addresses")
		return nil, dberrors.ClassifyError(err, UpstreamGetAllAddresses)
	}
	defer rows.Close()

	addresses = make([]*models.UpstreamAddressView, 0)

	for rows.Next() {
		var addr models.UpstreamAddressView
		err = rows.Scan(&addr.ID, &addr.Name, &addr.Port, &addr.UpstreamUrl)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to read upstream addresses")
			return nil, dberrors.ClassifyError(err, UpstreamGetAllAddresses)
		}
	}
	return addresses, nil
}
