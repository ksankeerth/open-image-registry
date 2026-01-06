package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	sqlite3 "modernc.org/sqlite"

	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
)

type Store struct {
	db         *sql.DB
	acesss     *resourceAccessStore
	auth       *authStore
	blob       *blobMetaStore
	cache      *registryCacheStore
	manifest   *manifestStore
	namespace  *namespaceStore
	recovery   *accountRecoveryStore
	repository *repositoryStore
	tag        *imageTagStore
	user       *userStore
	upstream   *upstreamStore

	queries *queries
}

func New(config config.DatabaseConfig) (*Store, error) {

	sql.Register("sqlite-hooked", &sqlite3.Driver{})

	database, err := sql.Open("sqlite-hooked", fmt.Sprintf("file:%s?cache=shared&_fk=1", config.Path))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to connect to Sqlite database")
		return nil, err
	}

	// Verify database connection
	if err = database.Ping(); err != nil {
		database.Close()
		database = nil
		log.Logger().Error().Err(err).Msgf("Ping failed to the database")
		return nil, err
	}

	// Run schema migrations
	contentBytes, err := os.ReadFile(config.ScriptsPath)
	if err != nil {
		database.Close()
		database = nil
		log.Logger().Error().Err(err).Msgf("Error occured when reading schema file")
		return nil, err
	}

	if _, err = database.Exec(string(contentBytes)); err != nil {
		database.Close()
		database = nil
		log.Logger().Error().Err(err).Msgf("Error occured when initializing schema")
		return nil, err
	}

	return NewWithDB(database), nil
}

func NewWithDB(db *sql.DB) *Store {
	s := Store{db: db}

	s.namespace = newNamespaceStore(db)
	s.acesss = newResourceAccessStore(db)
	s.auth = newAuthStore(db)
	s.blob = newBlobMetaStore(db)
	s.cache = newRegistryCacheStore(db)
	s.manifest = newManifestStore(db)
	s.recovery = newAccountRecoveryStore(db)
	s.repository = newRepositoryStore(db)
	s.upstream = newUpstreamStore(db)
	s.user = newUserStore(db)
	s.tag = newImageStore(db)

	s.queries = newQueries(db)

	return &s
}

func (s *Store) Namespaces() store.NamespaceStore {
	return s.namespace
}
func (s *Store) Repositories() store.RepositoryStore {
	return s.repository
}
func (s *Store) Manifests() store.ManifestStore {
	return s.manifest
}
func (s *Store) Tags() store.ImageTagStore {
	return s.tag
}
func (s *Store) Blobs() store.BlobMetaStore {
	return s.blob
}
func (s *Store) Cache() store.RegistryCacheStore {
	return s.cache
}
func (s *Store) Users() store.UserStore {
	return s.user
}
func (s *Store) Access() store.ResourceAccessStore {
	return s.acesss
}
func (s *Store) AccountRecovery() store.AccountRecoveryStore {
	return s.recovery
}
func (s *Store) Auth() store.AuthStore {
	return s.auth
}

func (s *Store) Upstreams() store.UpstreamRegistyStore {
	return s.upstream
}

func (s *Store) ImageQueries() store.ImageQueries {
	return s.queries
}
func (s *Store) NamespaceQueries() store.NamespaceQueries {
	return s.queries
}

func (s *Store) UserQueries() store.UserQueries {
	return s.queries
}

func (s *Store) AccessQueries() store.AccessQueries {
	return s.queries
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) Begin(ctx context.Context) (*sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return tx, nil
}