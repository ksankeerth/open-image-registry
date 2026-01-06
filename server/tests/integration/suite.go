package integration

import "testing"

type APITestSuite interface {
	Name() string
	APIVersion() string
	Run(t *testing.T)
}
