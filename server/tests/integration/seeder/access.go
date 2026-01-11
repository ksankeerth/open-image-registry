package seeder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/require"
)

// GrantAccess seeds an access record for a specific resource.
func (s *TestDataSeeder) GrantAccess(t *testing.T, resourceID, resourceType, userID, accessLevel string) {
	t.Helper()

	body := map[string]any{
		"user_id":       userID,
		"access_level":  accessLevel,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"granted_by":    "admin",
	}

	reqBody, err := json.Marshal(body)
	require.NoError(t, err, "failed to marshal grant access body")

	url := fmt.Sprintf("%s%s", s.baseURL, fmt.Sprintf(testdata.EndpointNamespaceUsers, resourceID))

	// 1. Create the request object manually instead of using http.Post
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	require.NoError(t, err, "failed to create grant access request")
	req.Header.Set("Content-Type", "application/json")

	// 2. Generate the admin token and attach the cookie using your helpers
	token := s.AdminToken(t)
	helpers.SetAuthCookie(req, token)

	// 3. Perform the request using the default HTTP client
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to send grant access request")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to grant access: expected 200/201 but got %d", resp.StatusCode)
	}
}