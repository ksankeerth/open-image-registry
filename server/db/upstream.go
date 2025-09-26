package db

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/ksankeerth/open-image-registry/errors/db"
	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type upstreamDaoImpl struct {
	db *sql.DB
	*TransactionManager
}

func (u *upstreamDaoImpl) CreateUpstreamRegistry(upstreamReg *models.UpstreamRegistryEntity,
	authConfig *models.UpstreamRegistryAuthConfig,
	accessConfig *models.UpstreamRegistryAccessConfig,
	storageConfig *models.UpstreamRegistryStorageConfig,
	cacheConfig *models.UpstreamRegistryCacheConfig,
	txKey string) (regId string, regName string, err error) {
	var tx *sql.Tx
	// if user provided txKey, We'd use transasaction stored in TransactionManager. Otherwise, We'll obtain
	// new trasaction.
	if txKey != "" {
		tx = u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction associated with `txKey`(%s) was not found", txKey)
			return "", "", db.ErrTxAlreadyClosed
		}
	} else {
		tx, err = u.db.Begin()
		if err != nil {
			log.Logger().Error().Err(err).Msg("Unable to acuquire db transaction to create upstream registry")
			return "", "", db.ClassifyError(err, "BEGIN")
		}
	}

	defer func(tx *sql.Tx) {
		// If tx was created by caller, call will take care of closing it.
		if txKey != "" {
			return
		}
		if err != nil {
			err1 := tx.Rollback()
			if err1 != nil {
				log.Logger().Error().Err(err1).Msg("Error occured when rolling back transaction")
			}
		} else {
			err1 := tx.Commit()
			if err1 != nil {
				log.Logger().Error().Err(err1).Msg("Error occured when commiting transaction")
			}
		}
	}(tx)

	regId = ""
	err = tx.QueryRow(InsertUptreamRegistry, upstreamReg.Name, upstreamReg.Port, upstreamReg.Status, upstreamReg.UpstreamUrl).
		Scan(&regId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UptreamRegistry %s into DB", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUptreamRegistry)
	}

	var credentailsJsonBytes []byte

	credentailsJsonBytes, err = json.Marshal(authConfig.CredentialJson)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when marshaling CredentialJson")
		return "", "", err
	}

	res, err := tx.Exec(InsertUpstreamRegistryAuthConfig, regId, authConfig.AuthType,
		credentailsJsonBytes, authConfig.TokenEndpoint)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamRegistryAuthConfig for upstream registry %s into DB", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryAuthConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamOciRegistryAuthConfig is not successful: %s", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryAuthConfig)
	}

	res, err = tx.Exec(InsertUpstreamRegistryAccessConfig, regId, accessConfig.ProxyEnabled,
		accessConfig.ProxyUrl, accessConfig.ConnectionTimeoutInSeconds, accessConfig.ReadTimeoutInSeconds,
		accessConfig.MaxConnections, accessConfig.MaxRetries, accessConfig.RetryDelayInSeconds)

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamRegistryAccessConfig for upstream registry %s into db", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryAccessConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamRegistryAccessConfig of registry(%s) was not successful", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryAccessConfig)
	}

	res, err = tx.Exec(InsertUpstreamRegistryStorageConfig, regId, storageConfig.StorageLimitInMbs,
		storageConfig.CleanupPolicy, storageConfig.CleanupThreshold)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamRegistryStorageConfig for upstream registry %s into db", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryStorageConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamRegistryStorageConfig of registry(%s) is not successful", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryStorageConfig)
	}

	var cacheEnabled = 0
	if cacheConfig.Enabled {
		cacheEnabled = 1
	}
	res, err = tx.Exec(InsertUpstreamRegistryCacheConfig, regId, cacheEnabled, cacheConfig.TtlInSeconds, cacheConfig.OfflineMode)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamRegistryCacheConfig for upstream registry %s into db", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryCacheConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamRegistryCacheConfig is not successful: %s", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamRegistryCacheConfig)
	}

	return regId, upstreamReg.Name, nil
}

// DeleteUpstreamRegistry implements UpstreamDAO.
func (u *upstreamDaoImpl) DeleteUpstreamRegistry(regId string) (err error) {
	res, err := u.db.Exec(DeleteUpstreamRegistry, regId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to delete UpstreamRegistry by id : %s", regId)
		return db.ClassifyError(err, DeleteUpstreamRegistry)
	}
	if r, err1 := res.RowsAffected(); r <= 0 && err1 != nil {
		log.Logger().Warn().Err(err1).Msgf("Deleting UpstreamRegistry was not successful by id %s", regId)
		return db.ClassifyError(err1, DeleteUpstreamRegistry)
	}
	return nil
}

func (u *upstreamDaoImpl) GetUpstreamRegistryWithConfig(regId string) (reg *models.UpstreamRegistryEntity,
	accessConfig *models.UpstreamRegistryAccessConfig,
	authConfig *models.UpstreamRegistryAuthConfig,
	cacheConfig *models.UpstreamRegistryCacheConfig,
	storageConfig *models.UpstreamRegistryStorageConfig,
	err error) {

	var credentialJsonBytes []byte

	var regCreatedAt string
	var regUpdatedAt string
	var authConfigUpdatedAt string
	var accessConfigUpdatedAt string
	var storageConfigUpdatedAt string
	var cacheConfigUpdatedAt string

	reg = &models.UpstreamRegistryEntity{}
	authConfig = &models.UpstreamRegistryAuthConfig{}
	accessConfig = &models.UpstreamRegistryAccessConfig{}
	storageConfig = &models.UpstreamRegistryStorageConfig{}
	cacheConfig = &models.UpstreamRegistryCacheConfig{}

	err = u.db.QueryRow(GetUpstreamRegistryWithConfig, regId).Scan(
		&reg.Name, &reg.Port, &reg.UpstreamUrl, &regCreatedAt, &regUpdatedAt,
		&authConfig.AuthType, &credentialJsonBytes, &authConfig.TokenEndpoint, &authConfigUpdatedAt,
		&accessConfig.ProxyEnabled, &accessConfig.ProxyUrl, &accessConfig.ConnectionTimeoutInSeconds, &accessConfig.ReadTimeoutInSeconds,
		&accessConfig.MaxConnections, &accessConfig.MaxRetries, &accessConfig.RetryDelayInSeconds, &accessConfigUpdatedAt,
		&storageConfig.StorageLimitInMbs, &storageConfig.CleanupPolicy, &storageConfig.CleanupThreshold, &storageConfigUpdatedAt,
		&cacheConfig.TtlInSeconds, &cacheConfig.OfflineMode, &cacheConfigUpdatedAt)

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrive upstream registery with config: %s", regId)
		return nil, nil, nil, nil, nil, db.ClassifyError(err, GetUpstreamRegistryWithConfig)
	}

	reg.CreatedAt, err = utils.ParseSqliteTimestamp(regCreatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when parsing sqlite-timestamp: %s", regCreatedAt)
	}
	reg.UpdatedAt, err = utils.ParseSqliteTimestamp(regUpdatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when parsing sqlite-timestamp: %s", regUpdatedAt)
	}
	authConfig.UpdatedAt, err = utils.ParseSqliteTimestamp(authConfigUpdatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when parsing sqlite-timestamp: %s", authConfigUpdatedAt)
	}
	accessConfig.UpdatedAt, err = utils.ParseSqliteTimestamp(accessConfigUpdatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when parsing sqlite-timestamp: %s", accessConfigUpdatedAt)
	}
	storageConfig.UpdatedAt, err = utils.ParseSqliteTimestamp(storageConfigUpdatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when parsing sqlite-timestamp: %s", storageConfigUpdatedAt)
	}
	cacheConfig.UpdatedAt, err = utils.ParseSqliteTimestamp(cacheConfigUpdatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when parsing sqlite-timestamp: %s", cacheConfigUpdatedAt)
	}

	err = json.Unmarshal(credentialJsonBytes, &authConfig.CredentialJson)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when unmarshalling auth.CredentialJson")
		return nil, nil, nil, nil, nil, err
	}

	return
}

// ListUpstreamRegistries implements UpstreamDAO.
func (u *upstreamDaoImpl) ListUpstreamRegistries() (registeries []*models.UpstreamRegistrySummary, err error) {

	rows, err := u.db.Query(ListUpstreamRegistrySummary)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving upstream registeries from db.")
		return nil, db.ClassifyError(err, ListUpstreamRegistrySummary)
	}
	defer rows.Close()

	var createdAt string
	var updatedAt string

	var rowCount = 0
	for rows.Next() {
		var uptreamReg models.UpstreamRegistrySummary

		err = rows.Scan(&uptreamReg.Id, &uptreamReg.Name, &uptreamReg.Port,
			&uptreamReg.Status, &uptreamReg.UpstreamUrl, &createdAt,
			&updatedAt, &uptreamReg.CachedImagesCount)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when reading resultset at %d. Moving to next line.", rowCount)
			continue
		}

		// we just ignore the timestamp parse errors
		uptreamReg.CreatedAt, err = utils.ParseSqliteTimestamp(createdAt)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when parsing timestamp: %s", createdAt)
		}
		uptreamReg.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when parsing timestamp: %s", updatedAt)
		}

		registeries = append(registeries, &uptreamReg)
		rowCount++
	}

	return registeries, nil
}

// UpdateUpstreamRegistry implements UpstreamDAO.
func (u *upstreamDaoImpl) UpdateUpstreamRegistry(regId string, upstreamReg *models.UpstreamRegistryEntity,
	authConfig *models.UpstreamRegistryAuthConfig,
	accessConfig *models.UpstreamRegistryAccessConfig,
	storageConfig *models.UpstreamRegistryStorageConfig,
	cacheConfig *models.UpstreamRegistryCacheConfig) (err error) {
	tx, err := u.db.Begin()
	if err != nil {
		log.Logger().Error().Err(err).Msg("Unable to acquire db transaction to update upstream registry")
		return db.ClassifyError(err, "BEGIN TRANSACTION")
	}
	defer func(tx *sql.Tx) {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}(tx)

	if upstreamReg != nil {
		res, err := tx.Exec(UpdateUpstreamRegistry, upstreamReg.Name, upstreamReg.Port, upstreamReg.UpstreamUrl, regId)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamRegistry: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistry)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamRegistry was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistry)
		}
	}

	if accessConfig != nil {
		res, err := tx.Exec(UpdateUpstreamRegistryAccessConfig, accessConfig.ProxyEnabled,
			accessConfig.ProxyUrl, accessConfig.ConnectionTimeoutInSeconds, accessConfig.ReadTimeoutInSeconds,
			accessConfig.MaxConnections, accessConfig.MaxRetries, accessConfig.RetryDelayInSeconds, regId)

		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamRegistryAccessConfig: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryAccessConfig)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamRegistryAccessConfig was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryAccessConfig)
		}
	}

	if authConfig != nil {

		var credentailsJsonBytes []byte

		credentailsJsonBytes, err = json.Marshal(authConfig.CredentialJson)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when marshaling CredentialJson")
			return err
		}

		res, err := tx.Exec(UpdateUpstreamRegistryAuthConfig, authConfig.AuthType,
			credentailsJsonBytes, authConfig.TokenEndpoint, regId)

		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamRegistryAuthConfig: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryAuthConfig)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamRegistryAuthConfig was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryAuthConfig)
		}
	}

	if storageConfig != nil {
		res, err := tx.Exec(UpdateUpstreamRegistryStorageConfig, storageConfig.StorageLimitInMbs,
			storageConfig.CleanupPolicy, storageConfig.CleanupThreshold, regId)

		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamRegistryStorageConfig: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryStorageConfig)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamRegistryStorageConfig was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryStorageConfig)
		}
	}

	if cacheConfig != nil {
		res, err := tx.Exec(UpdateUpstreamRegistryCacheConfig, cacheConfig.Enabled, cacheConfig.TtlInSeconds, cacheConfig.OfflineMode, regId)

		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamRegistryCacheConfig: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryCacheConfig)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamRegistryCacheConfig was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamRegistryCacheConfig)
		}
	}
	return nil
}

func (u *upstreamDaoImpl) GetActiveUpstreamAddresses() (upstreamAddrs []*models.UpstreamRegistryAddress, err error) {

	rows, err := u.db.Query(GetActiveUpstreamAddresses)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Logger().Log().Err(err).Msg("Unable to load upstream addresses")
		return nil, db_errors.ClassifyError(err, GetActiveUpstreamAddresses)
	}
	defer rows.Close()

	for rows.Next() {
		var upstreamAddr models.UpstreamRegistryAddress

		err = rows.Scan(&upstreamAddr.Id, &upstreamAddr.Name, &upstreamAddr.Port, &upstreamAddr.UpstreamUrl)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to read results of upstream registry addresses")
			return nil, db_errors.ClassifyError(err, GetActiveUpstreamAddresses)
		}
		upstreamAddrs = append(upstreamAddrs, &upstreamAddr)
	}
	return upstreamAddrs, nil
}