package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sqlite3 "modernc.org/sqlite"

	"github.com/ksankeerth/open-image-registry/errors/db"
	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types"
)

var database *sql.DB
var tm *TransactionManager
var imgRegDao ImageRegistryDAO
var upstreamDao UpstreamDAO
var once sync.Once

func InitDB() (*sql.DB, error) {
	var err error
	once.Do(func() {
		serverDir, err1 := os.Getwd()
		if err != nil {
			log.Logger().Err(err1).Msg("Unable to detect server director in runtime using os.Getwd()")
			err = err1
			return
		}

		// TODO: Add hooks to DB calls so we can monitor query performance
		sql.Register("sqlite-hooked", &sqlite3.Driver{})

		database, err1 = sql.Open("sqlite-hooked", fmt.Sprintf("file:%s?cache=shared&_fk=1", filepath.Join(serverDir, "open_image_registry.db")))
		if err1 != nil {
			log.Logger().Err(err1).Msg("Unable to open SQL connection to Sqlite")
			err = err1
			return
		}

		scriptsPath := filepath.Join(serverDir, "db-scripts/sqlite.sql")
		contentBytes, err1 := os.ReadFile(scriptsPath)
		if err1 != nil {
			log.Logger().Error().Err(err1).Msgf("Unable to open DB scripts to create table: %s", scriptsPath)
			err = err1
			return
		}

		_, err1 = database.Exec(string(contentBytes))
		if err1 != nil {
			log.Logger().Error().Err(err1).Msg("Error occurred when executing DB scripts")
			err = err1
			return
		}

		tm = &TransactionManager{
			db:       database,
			keyLocks: lib.NewKeyLock(),
		}

		imgRegDao = &imageRegistryDaoImpl{
			db:                 database,
			TransactionManager: tm,
		}

		upstreamDao = &upstreamDaoImpl{
			db:                 database,
			TransactionManager: tm,
		}
	})

	if err != nil {
		return nil, err
	}

	return database, nil
}

func LoadUpstreamRegistryAddresses() (upstreamAddrs []*types.UpstreamRegistryAddress, err error) {
	if database == nil {
		log.Logger().Error().Msgf("Database reference is nil.")
		panic("Database reference is nil.")
	}
	rows, err := database.Query(LoadActiveUpstreamRegistryAddressess)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Logger().Log().Err(err).Msg("Unable to load upstream addresses")
		return nil, db_errors.ClassifyError(err, LoadActiveUpstreamRegistryAddressess)
	}
	defer rows.Close()

	for rows.Next() {
		var upstreamAddr types.UpstreamRegistryAddress

		err = rows.Scan(&upstreamAddr.Id, &upstreamAddr.Name, &upstreamAddr.Port, &upstreamAddr.UpstreamUrl)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to read results of upstream registry addresses")
			return nil, db_errors.ClassifyError(err, LoadActiveUpstreamRegistryAddressess)
		}
		upstreamAddrs = append(upstreamAddrs, &upstreamAddr)
	}
	return upstreamAddrs, nil
}

func GetImageRegDAO() ImageRegistryDAO {
	return imgRegDao
}

func GetUpstreamDAO() UpstreamDAO {
	return upstreamDao
}

func GetTransactionManager() *TransactionManager {
	once.Do(func() {
		tm = &TransactionManager{
			db:       database,
			keyLocks: lib.NewKeyLock(),
		}
	})
	return tm
}

type TransactionManager struct {
	transactionsMap sync.Map
	db              *sql.DB
	keyLocks        *lib.KeyLock
}

func (tm *TransactionManager) Begin(key string) error {
	tm.keyLocks.Lock(key)
	defer tm.keyLocks.Unlock(key)
	_, ok := tm.transactionsMap.Load(key)
	if !ok {
		log.Logger().Debug().Msgf("Starting new database transaction for key: %s", key)
		tx, err := tm.db.Begin()
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to create database transaction with key: %s", key)
			return db.ClassifyError(err, "BEGIN")
		}
		tm.transactionsMap.Store(key, tx)
		return nil
	}
	return db.ErrTxCreationFailed
}

func (tm *TransactionManager) Commit(key string) error {
	tm.keyLocks.Lock(key)
	defer tm.keyLocks.Unlock(key)
	v, ok := tm.transactionsMap.Load(key)
	defer func() {
		tm.transactionsMap.Delete(key)
	}()
	if !ok {
		log.Logger().Warn().Msgf("Commit is invoked on non-existent transaction: %s", key)
		return db.ErrTxAlreadyClosed
	}
	tx, _ := v.(*sql.Tx)
	if tx == nil {
		log.Logger().Warn().Msgf("Commit is invoked on non-existent transaction: %s", key)
		return db.ErrTxAlreadyClosed
	}
	err := tx.Commit()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Commit failed with errors for key: %s", key)
		return db.ClassifyError(err, "COMMIT")
	}
	return nil
}

func (tm *TransactionManager) Rollback(key string) error {
	tm.keyLocks.Lock(key)
	defer tm.keyLocks.Unlock(key)
	v, ok := tm.transactionsMap.Load(key)
	defer func() {
		tm.transactionsMap.Delete(key)
	}()
	if !ok {
		log.Logger().Warn().Msgf("Rollback is invoked on non-existent transaction: %s", key)
		return db.ErrTxAlreadyClosed
	}
	tx, _ := v.(*sql.Tx)
	if tx == nil {
		log.Logger().Warn().Msgf("Rollback is invoked on non-existent transaction: %s", key)
		return db.ErrTxAlreadyClosed
	}
	err := tx.Rollback()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Rollback failed with errors for key: %s", key)
		return db.ClassifyError(err, "ROLLBACK")
	}
	return nil
}

func (tm *TransactionManager) getTx(key string) *sql.Tx {
	tm.keyLocks.Lock(key)
	defer tm.keyLocks.Unlock(key)
	v, ok := tm.transactionsMap.Load(key)
	if !ok {
		return nil
	}
	tx, _ := v.(*sql.Tx)
	return tx
}