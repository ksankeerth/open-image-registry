package db

import (
	"database/sql"
	"encoding/json"

	"github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types"
)

type upstreamDaoImpl struct {
	db *sql.DB
	*TransactionManager
}

func (u *upstreamDaoImpl) CreateUpstreamRegistry(upstreamReg *types.UpstreamOCIRegEntity,
	authConfig *types.UpstreamOCIRegAuthConfig, accessConfig *types.UpstreamOCIRegAccessConfig,
	storageConfig *types.UpstreamOCIRegStorageConfig,
	cacheConfig *types.UpstreamOCIRegCacheConfig, txKey string) (regId string, regName string, err error) {
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
			log.Logger().Error().Err(err).Msg("Unable to acuquire db transaction to create upstream OCI registry")
			return "", "", db.ClassifyError(err, "BEGIN")
		}
	}

	defer func(tx *sql.Tx) {
		// If tx was created by caller, call will take care of closing it.
		if txKey != "" {
			return
		}
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Logger().Error().Err(err).Msg("Error occured when rolling back transaction")
			}
		} else {
			err = tx.Commit()
			if err != nil {
				log.Logger().Error().Err(err).Msg("Error occured when commiting transaction")
			}
		}
	}(tx)

	regId = ""
	err = tx.QueryRow(InsertUptreamOciRegistry, upstreamReg.Name, upstreamReg.Port, upstreamReg.UpstreamUrl).
		Scan(&regId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UptreamOciRegistry %s into DB", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUptreamOciRegistry)
	}

	var credentailsJsonBytes []byte

	credentailsJsonBytes, err = json.Marshal(authConfig.CredentialJson)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when marshaling CredentialJson")
		return "", "", err
	}

	res, err := tx.Exec(InsertUpstreamOciRegistryAuthConfig, regId, authConfig.AuthType,
		credentailsJsonBytes, authConfig.TokenEndpoint, authConfig.Certificate)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamOciRegistryAuthConfig for upstream registry %s into DB", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryAuthConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamOciRegistryAuthConfig is not successful: %s", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryAuthConfig)
	}

	res, err = tx.Exec(InsertUpstreamOciRegistryAccessConfig, regId, accessConfig.ProxyEnabled,
		accessConfig.ProxyUrl, accessConfig.ConnectionTimeoutInSeconds, accessConfig.ReadTimeoutInSeconds,
		accessConfig.MaxConnections, accessConfig.MaxRetries, accessConfig.RetryDelayInSeconds)

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamOciRegistryAccessConfig for upstream registry %s into DB", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryAccessConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamOciRegistryAccessConfigUpstreamOciRegistryAuthConfig of registry(%s) is not successful", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryAccessConfig)
	}

	res, err = tx.Exec(InsertUpstreamOciRegistryStorageConfig, regId, storageConfig.StorageLimitInMbs,
		storageConfig.CleanupPolicy, storageConfig.CleanupThreshold)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamOciRegistryStorageConfig for upstream registry %s into DB", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryStorageConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamOciRegistryStorageConfig of registry(%s) is not successful", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryStorageConfig)
	}

	res, err = tx.Exec(InsertUpstreamOciRegistryCacheConfig, regId, cacheConfig.TtlInSeconds, cacheConfig.OfflineMode)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist UpstreamOciRegistryCacheConfig for upstream registry %s into DB", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryCacheConfig)
	}
	if r, err := res.RowsAffected(); r <= 0 || err != nil {
		log.Logger().Warn().Err(err).Msgf("Persisting UpstreamOciRegistryCacheConfig is not successful: %s", upstreamReg.Name)
		return "", "", db.ClassifyError(err, InsertUpstreamOciRegistryCacheConfig)
	}

	return regId, upstreamReg.Name, nil
}

// DeleteUpstreamRegistry implements UpstreamDAO.
func (u *upstreamDaoImpl) DeleteUpstreamRegistry(regId string) (err error) {
	res, err := u.db.Exec(DeleteUpstreamOciRegistry, regId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to delete UpstreamRegistry by id : %s", regId)
		return db.ClassifyError(err, DeleteUpstreamOciRegistry)
	}
	if r, err1 := res.RowsAffected(); r <= 0 && err1 != nil {
		log.Logger().Warn().Err(err1).Msgf("Deleting UpstreamRegistry was not successful by id %s", regId)
		return db.ClassifyError(err1, DeleteUpstreamOciRegistry)
	}
	return nil
}

// GetUpstreamRegistry implements UpstreamDAO.
func (u *upstreamDaoImpl) GetUpstreamRegistry(regId string) (*types.UpstreamOCIRegEntity, error) {
	var reg types.UpstreamOCIRegEntity
	reg.Id = regId
	err := u.db.QueryRow(GetUpstreamOciRegistry, regId).
		Scan(&reg.Name, &reg.Port, &reg.Status, &reg.UpstreamUrl, &reg.CreatedAt, &reg.UpdatedAt)
	if err != nil {
		return nil, db.ClassifyError(err, GetUpstreamOciRegistry)
	}
	return &reg, nil
}

func (u *upstreamDaoImpl) GetUpstreamRegistryWithConfig(regId string) (*types.UpstreamOCIRegEntity,
	*types.UpstreamOCIRegAccessConfig, *types.UpstreamOCIRegAuthConfig,
	*types.UpstreamOCIRegCacheConfig, *types.UpstreamOCIRegStorageConfig, error) {
	var reg types.UpstreamOCIRegEntity
	var auth types.UpstreamOCIRegAuthConfig
	var cache types.UpstreamOCIRegCacheConfig
	var storage types.UpstreamOCIRegStorageConfig
	var access types.UpstreamOCIRegAccessConfig

	var credentialJsonBytes []byte

	err := u.db.QueryRow(GetUpstreamOCIRegistryWithConfig, regId).Scan(
		&reg.Name, &reg.Port, &reg.UpstreamUrl, &reg.CreatedAt, &reg.UpdatedAt,
		&auth.AuthType, &credentialJsonBytes, &auth.TokenEndpoint, &auth.Certificate, &auth.UpdatedAt,
		&access.ProxyEnabled, &access.ProxyUrl, &access.ConnectionTimeoutInSeconds, &access.ReadTimeoutInSeconds,
		&access.MaxConnections, &access.MaxRetries, &access.RetryDelayInSeconds, &access.UpdatedAt,
		&storage.StorageLimitInMbs, &storage.CleanupPolicy, &storage.CleanupThreshold, &storage.UpdatedAt,
		&cache.TtlInSeconds, &cache.OfflineMode, &cache.UpdatedAt)

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrive upstream registery with config: %s", regId)
		return nil, nil, nil, nil, nil, db.ClassifyError(err, GetUpstreamOCIRegistryWithConfig)
	}

	err = json.Unmarshal(credentialJsonBytes, &auth.CredentialJson)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when unmarshalling auth.CredentialJson")
		return nil, nil, nil, nil, nil, err
	}

	return &reg, &access, &auth, &cache, &storage, nil
}

// ListUpstreamRegistries implements UpstreamDAO.
func (u *upstreamDaoImpl) ListUpstreamRegistries() (registeries []*types.UpstreamOCIRegEntityWithAdditionalInfo, err error) {

	rows, err := u.db.Query(GetListUpstreamOciRegisteriesWithAdditionalInfo)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving upstream OCI registeries from db.")
		return nil, db.ClassifyError(err, GetListUpstreamOciRegisteriesWithAdditionalInfo)
	}
	defer rows.Close()

	var rowCount = 0
	for rows.Next() {
		var uptreamOciReg types.UpstreamOCIRegEntityWithAdditionalInfo

		err = rows.Scan(&uptreamOciReg.Id, &uptreamOciReg.Name, &uptreamOciReg.Port,
			&uptreamOciReg.Status, &uptreamOciReg.UpstreamUrl, &uptreamOciReg.CreatedAt,
			&uptreamOciReg.UpdatedAt, &uptreamOciReg.CachedImagesCount)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when reading resultset at %d. Moving to next line.", rowCount)
			continue
		}
		registeries = append(registeries, &uptreamOciReg)
		rowCount++
	}

	return registeries, nil
}

// UpdateUpstreamRegistry implements UpstreamDAO.
func (u *upstreamDaoImpl) UpdateUpstreamRegistry(regId string, upstreamReg *types.UpstreamOCIRegEntity,
	authConfig *types.UpstreamOCIRegAuthConfig,
	accessConfig *types.UpstreamOCIRegAccessConfig,
	storageConfig *types.UpstreamOCIRegStorageConfig,
	cacheConfig *types.UpstreamOCIRegCacheConfig) (err error) {
	tx, err := u.db.Begin()
	if err != nil {
		log.Logger().Error().Err(err).Msg("Unable to acquire db transaction to create upstream OCI registry")
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
		res, err := tx.Exec(UpdateUpstreamOciRegistry, upstreamReg.Name, upstreamReg.Port, upstreamReg.UpstreamUrl, regId)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamOciRegistry: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistry)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamOciRegistry was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistry)
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

		res, err := tx.Exec(UpdateUpstreamOciRegistryAuthConfig, authConfig.AuthType,
			credentailsJsonBytes, authConfig.TokenEndpoint, authConfig.Certificate, regId)

		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamOciRegistryAuthConfig: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistryAuthConfig)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamOciRegistryAuthConfig was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistryAuthConfig)
		}
	}

	if storageConfig != nil {
		res, err := tx.Exec(UpdateUpstreamOciRegistryStorageConfig, storageConfig.StorageLimitInMbs,
			storageConfig.CleanupPolicy, storageConfig.CleanupThreshold, regId)

		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamOciRegistryStorageConfig: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistryStorageConfig)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamOciRegistryStorageConfig was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistryStorageConfig)
		}
	}

	if cacheConfig != nil {
		res, err := tx.Exec(UpdateUpstreamOciRegistryCacheConfig, cacheConfig.TtlInSeconds, cacheConfig.OfflineMode, regId)

		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when updating UpstreamOciRegistryCacheConfig: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistryCacheConfig)
		}

		if r, err := res.RowsAffected(); r <= 0 || err != nil {
			log.Logger().Warn().Err(err).Msgf("Persisting updates to UpstreamOciRegistryCacheConfig was not successful: %s. Skipping updates to other tables", regId)
			return db.ClassifyError(err, UpdateUpstreamOciRegistryCacheConfig)
		}
	}
	return nil
}