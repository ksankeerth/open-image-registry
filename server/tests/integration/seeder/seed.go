package seeder

import (
	"testing"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/stretchr/testify/require"
)

type TestDataSeeder struct {
	baseURL     string
	store       store.Store
	jwtProvider lib.JWTProvider
}

func NewTestDataSeeder(baseURL string, store store.Store, jwtProvider lib.JWTProvider) *TestDataSeeder {
	return &TestDataSeeder{
		baseURL:     baseURL,
		store:       store,
		jwtProvider: jwtProvider,
	}
}

func (s *TestDataSeeder) AdminToken(t *testing.T) string {
	t.Helper()

	token, err := s.jwtProvider.Sign(map[string]any{
		constants.ClaimRole:    "Admin",
		constants.ClaimSubject: "admin",
	})
	require.NoError(t, err)

	return token
}