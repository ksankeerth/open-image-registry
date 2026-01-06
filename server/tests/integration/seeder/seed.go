package seeder

import "github.com/ksankeerth/open-image-registry/store"

type TestDataSeeder struct {
	baseURL string
	store   store.Store
}

func NewTestDataSeeder(baseURL string, store store.Store) *TestDataSeeder {
	return &TestDataSeeder{
		baseURL: baseURL,
		store:   store,
	}
}