package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	sqlite3 "modernc.org/sqlite"

	"github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
)

var initCalled int32

// InitDB initializes the DB and transaction manager.
// Returns an error if initialization failed.
// This method must be called once and repeated calls will panic
func InitDB() (database *sql.DB, tm *TransactionManager, err error) {

	if !atomic.CompareAndSwapInt32(&initCalled, 0, 1) {
		panic("InitDB() called multiple times - should only be called once from main.go")
	}

	serverDir, err := os.Getwd()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to get current working directory")
		return
	}

	// Register sqlite driver with hooks
	sql.Register("sqlite-hooked", &sqlite3.Driver{})

	dbPath := filepath.Join(filepath.Dir(serverDir), "registry_sqlite.db")
	database, err = sql.Open("sqlite-hooked", fmt.Sprintf("file:%s?cache=shared&_fk=1", dbPath))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to initialize database from configuration")
		return
	}

	// Verify database connection
	if err = database.Ping(); err != nil {
		database.Close()
		database = nil
		log.Logger().Error().Err(err).Msgf("Ping failed to the database")
		return
	}

	// Run schema migrations
	scriptsPath := filepath.Join(filepath.Dir(serverDir), "db-scripts/sqlite/registry.sql")
	contentBytes, err := os.ReadFile(scriptsPath)
	if err != nil {
		database.Close()
		database = nil
		log.Logger().Error().Err(err).Msgf("Error occured when reading schema file")
		return
	}

	if _, err = database.Exec(string(contentBytes)); err != nil {
		database.Close()
		database = nil
		log.Logger().Error().Err(err).Msgf("Error occured when initializing schema")
		return
	}

	tm = &TransactionManager{
		db:       database,
		keyLocks: lib.NewKeyLock(),
	}

	if database == nil || tm == nil {
		panic("not allowed to initialized database multiple times")
	}

	return database, tm, nil
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
	// TODO: instead of db.ErrTxCreationFailed, we should send another error to indicate that
	// there is a transaction with same key.
	// Caller can implement retry.
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

func NewImageRegistryDAO(database *sql.DB, tm *TransactionManager) ImageRegistryDAO {
	return &imageRegistryDaoImpl{
		db:                 database,
		TransactionManager: tm,
	}
}

func NewUpstreamDAO(database *sql.DB, tm *TransactionManager) UpstreamDAO {
	return &upstreamDaoImpl{
		db:                 database,
		TransactionManager: tm,
	}
}