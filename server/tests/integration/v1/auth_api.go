package v1

import (
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/integration/seeder"
)

type AuthTestSuite struct {
	name        string
	apiVersion  string
	seeder      *seeder.TestDataSeeder
	testBaseURL string
}

func NewAuthTestSuite(seeder *seeder.TestDataSeeder, baseURL string) *AuthTestSuite {
	return &AuthTestSuite{
		name:        "AuthAPI",
		apiVersion:  "v1",
		seeder:      seeder,
		testBaseURL: baseURL,
	}
}

func (a *AuthTestSuite) Run(t *testing.T) {

}

func (a *AuthTestSuite) Name() string {
	return a.name
}

func (a *AuthTestSuite) APIVersion() string {
	return a.apiVersion
}