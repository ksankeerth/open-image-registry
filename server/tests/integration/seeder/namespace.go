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

func (s *TestDataSeeder) DeleteAllNamespaces(t *testing.T) {
	t.Helper()

	err := s.store.Namespaces().DeleteAll(context.Background())
	require.NoError(t, err)
}

func (s *TestDataSeeder) CreateNamespace(t *testing.T, name, description, purpose string,
	isPublic bool, maintainers ...string) (id string) {
	t.Helper()

	for _, maintainer := range maintainers {
		exists, misMatch, _ := s.checkUser(t, maintainer, "Maintainer", "", false)
		require.Truef(t, exists, "maintainer(%s) must be availble", maintainer)
		require.Falsef(t, misMatch, "maintainer(%s) must have 'Maintainer' role", maintainer)
	}

	body := map[string]any{
		"name":        name,
		"description": description,
		"maintainers": maintainers,
		"is_public":   isPublic,
		"purpose":     purpose,
	}

	reqBody, _ := json.Marshal(body)

	// 1. Create the request manually to allow adding the cookie
	url := s.baseURL + testdata.EndpointNamespaces
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// 2. Attach the admin authentication cookie
	token := s.AdminToken(t)
	helpers.SetAuthCookie(req, token)

	// 3. Execute the request
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var resBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&resBody)
	require.NoError(t, err)

	return resBody["id"].(string)
}

func (s *TestDataSeeder) SetNamespaceDeprecated(t *testing.T, id string) {
	t.Helper()

	err := s.store.Namespaces().SetStateByID(context.Background(), id, "Deprecated")
	require.NoError(t, err)
}

func (s *TestDataSeeder) SetNamespaceDisabled(t *testing.T, id string) {
	t.Helper()

	err := s.store.Namespaces().SetStateByID(context.Background(), id, "Disabled")
	require.NoError(t, err)
}