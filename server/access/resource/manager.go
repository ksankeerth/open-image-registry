package resource

import "github.com/ksankeerth/open-image-registry/store"

type Manager struct {
	store store.Store
}

func NewManager(store store.Store) *Manager {
	return &Manager{
		store: store,
	}
}