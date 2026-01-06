package seeder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/require"
)

// GrantAccess seeds an access record for a specific resource.
// It uses the EndpointNamespaceUsers or a generic access endpoint depending on your API design.
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

	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err, "failed to send grant access request")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to grant access: expected 200/201 but got %d", resp.StatusCode)
	}
}
