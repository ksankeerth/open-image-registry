package store

import "context"

type Store interface {
	TxBeginner

	// Stores
	Namespaces() NamespaceStore
	Repositories() RepositoryStore
	Manifests() ManifestStore
	Tags() ImageTagStore
	Blobs() BlobMetaStore
	Cache() RegistryCacheStore
	Users() UserStore
	Access() ResourceAccessStore
	AccountRecovery() AccountRecoveryStore
	Auth() AuthStore
	Upstreams() UpstreamRegistyStore

	// Queries
	ImageQueries() ImageQueries
	NamespaceQueries() NamespaceQueries

	// Lifecycle
	Close() error
	Ping(ctx context.Context) error
}
