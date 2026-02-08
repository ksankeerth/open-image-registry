package seeder

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/require"
)

func (s *TestDataSeeder) CreateRepository(t *testing.T, name, description, createdBy, nsId string,
	isPublic bool) (id string) {
	t.Helper()

	body := map[string]any{
		"name":         name,
		"description":  description,
		"is_public":    isPublic,
		"namespace_id": nsId,
		"created_by":   createdBy,
	}

	reqBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, s.baseURL+testdata.EndpointRepositories, bytes.NewReader(reqBytes))
	require.NoError(t, err)

	helpers.SetAuthCookie(req, s.AdminToken(t))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody struct{ Id string }
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	require.NotEmpty(t, respBody.Id)

	return respBody.Id
}

func (s *TestDataSeeder) SetRepositoryDeprecated(t *testing.T, id string) {
	t.Helper()

	err := s.store.Repositories().SetState(context.Background(), id, "Deprecated")
	require.NoError(t, err)
}

func (s *TestDataSeeder) SetRepositoryDisabled(t *testing.T, id string) {
	t.Helper()

	err := s.store.Repositories().SetState(context.Background(), id, "Disabled")
	require.NoError(t, err)
}
