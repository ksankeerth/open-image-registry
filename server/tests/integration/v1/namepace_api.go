package v1

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/integration/seeder"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type NamespaceTestSuite struct {
	apiVersion  string
	name        string
	seeder      *seeder.TestDataSeeder
	testBaseURL string
}

func NewNamespaceTestSuite(seeder *seeder.TestDataSeeder, baseURL string) *NamespaceTestSuite {
	return &NamespaceTestSuite{
		apiVersion:  "v1",
		name:        "Namespace API",
		seeder:      seeder,
		testBaseURL: baseURL,
	}
}

func (n *NamespaceTestSuite) Name() string {
	return n.name
}

func (n *NamespaceTestSuite) APIVersion() string {
	return n.apiVersion
}

func (n *NamespaceTestSuite) Run(t *testing.T) {
	t.Run("CreateNamespace_Success", n.testCreateNamespaceSuccess)
	t.Run("CreateNamespace_Validation", n.testCreateNamespaceValidation)
	t.Run("CreateNamespace_Conflicts", n.testCreateNamespaceConflicts)
	t.Run("GetAndExists", n.testNamespaceGetAndExists)

	t.Run("UpdateNamespace", n.testUpdateNamespace)

	t.Run("DeleteNamespace", n.testDeleteNamespace)

	// State & Visibility
	t.Run("StateChange", n.testNamespaceStateChange)
	t.Run("VisibilityChange", n.testNamespaceVisibilityChange)

	// Access Control
	t.Run("UserAccess_List", n.testNamespaceUserAccessList)
	t.Run("GrantAccess", n.testNamespaceGrantAccess)
	t.Run("RevokeAccess", n.testNamespaceRevokeAccess)

	// Collections
	t.Run("ListNamespaces", n.testListNamespaces)
	t.Run("ListRepositories", n.testNamespaceListRepositories)

	// Protocol
	t.Run("Generic_HTTPMethods", n.testNamespaceHTTPMethods)
	t.Run("Generic_ContentType", n.testNamespaceContentType)
}

func (n *NamespaceTestSuite) testCreateNamespaceSuccess(t *testing.T) {
	// Prepare Data
	m1 := n.seeder.ProvisionUser(t, "maint01", "m1@test.com", "Maintainer")
	m2 := n.seeder.ProvisionUser(t, "maint02", "m2@test.com", "Maintainer")

	tcs := []struct {
		name       string
		body       map[string]any
		statusCode int
	}{
		{
			name: "Without description",
			body: map[string]any{
				"name":        "ns-no-desc",
				"maintainers": []string{m1},
				"is_public":   false,
				"purpose":     "Team",
			},
			statusCode: http.StatusCreated,
		},
		{
			name: "With description and multiple maintainers",
			body: map[string]any{
				"name":        "ns-full",
				"description": "A test namespace",
				"maintainers": []string{m1, m2},
				"is_public":   false,
				"purpose":     "Team",
			},
			statusCode: http.StatusCreated,
		},
		{
			name: "Public Namespace",
			body: map[string]any{
				"name":        "ns-public1",
				"description": "A test namespace",
				"maintainers": []string{m1, m2},
				"is_public":   true,
				"purpose":     "Team",
			},
			statusCode: http.StatusCreated,
		},
		{
			name: "Project Namespace",
			body: map[string]any{
				"name":        "ns-project1",
				"description": "A test namespace",
				"maintainers": []string{m1, m2},
				"is_public":   false,
				"purpose":     "Project",
			},
			statusCode: http.StatusCreated,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tc.body)
			req, err := http.NewRequest(http.MethodPost, n.testBaseURL+testdata.EndpointNamespaces, bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tc.statusCode, resp.StatusCode)
		})
	}
}

func (n *NamespaceTestSuite) testCreateNamespaceValidation(t *testing.T) {
	// Prepare data
	d1 := n.seeder.ProvisionUser(t, "nsvalidationdev", "nsvalidation1@t.com", "Developer")

	tcs := []struct {
		name       string
		body       any
		statusCode int
		errMsg     string
	}{
		{
			name:       "Invalid request body",
			body:       struct{ name string }{name: "invalid body"},
			statusCode: http.StatusBadRequest,
			errMsg:     "",
		},
		{
			name: "With Invalid Name",
			body: map[string]any{
				"name":        "invalid-&name",
				"description": "test-ns",
				"is_public":   false,
				"purpose":     "Team",
				"maintainers": []string{"uuid"},
			},
			statusCode: http.StatusBadRequest,
			errMsg:     "Invalid Namespace name",
		},
		{
			name: "With Invalid Purpose",
			body: map[string]any{
				"name":        "ns-validation2",
				"description": "test-ns",
				"is_public":   false,
				"purpose":     "Development",
				"maintainers": []string{"uuid"},
			},
			statusCode: http.StatusBadRequest,
			errMsg:     "Namespace purpose not provided",
		},
		{
			name: "Without maintainers",
			body: map[string]any{
				"name":        "nsvalidation3",
				"description": "test-ns",
				"is_public":   false,
				"purpose":     "Team",
				"maintainers": []string{},
			},
			statusCode: http.StatusBadRequest,
			errMsg:     "Namespace should have atleast one maintainer",
		},
		{
			name: "With non-existing maintainers",
			body: map[string]any{
				"name":        "nsvalidation4",
				"description": "test-ns",
				"is_public":   false,
				"purpose":     "Team",
				"maintainers": []string{"uuid"},
			},
			statusCode: http.StatusNotFound,
			errMsg:     "One or more target users were not found",
		},
		{
			name: "With a developer as maintainer",
			body: map[string]any{
				"name":        "nsvalidation5",
				"description": "test-ns",
				"is_public":   false,
				"purpose":     "Team",
				"maintainers": []string{d1},
			},
			statusCode: http.StatusForbidden,
			errMsg:     "One or more users do not have the base role required for this access level",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tc.body)
			req, err := http.NewRequest(http.MethodPost, n.testBaseURL+testdata.EndpointNamespaces, bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tc.statusCode, resp.StatusCode)

			if tc.errMsg == "" {
				return
			}

			var respBody map[string]any
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err, "failed to parse namespace create response")

			errMsg, ok := respBody["error_message"].(string)
			require.True(t, ok, "failed to parse namespace create response")
			require.NotEmpty(t, errMsg, "failed to parse namespace create response")

			assert.Equal(t, tc.errMsg, errMsg)
		})
	}
}

func (n *NamespaceTestSuite) testCreateNamespaceConflicts(t *testing.T) {
	// Prepare data
	userID := n.seeder.ProvisionUser(t, "nsconflictuser1", "nsconflictsuser1@t.com", "Maintainer")
	require.NotEmpty(t, userID, "maintainer must be created befor tests")
	n.seeder.CreateNamespace(t, "nsconflicts1", "", "Team", false, userID)
	n.seeder.CreateNamespace(t, "nsconflicts2", "", "Project", false, userID)

	tcs := []struct {
		name          string
		namespaceName string
		statusCode    int
	}{
		{
			name:          "With existing team namespace name",
			namespaceName: "nsconflicts1",
			statusCode:    http.StatusConflict,
		},
		{
			name:          "With existing project namespace name",
			namespaceName: "nsconflicts2",
			statusCode:    http.StatusConflict,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{
				"name":        tc.namespaceName,
				"description": "",
				"maintainers": []string{userID},
				"is_public":   false,
				"purpose":     "Team",
			}

			reqBody, _ := json.Marshal(body)
			req, err := http.NewRequest(http.MethodPost, n.testBaseURL+testdata.EndpointNamespaces, bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (n *NamespaceTestSuite) testNamespaceGrantAccess(t *testing.T) {
	// Prepare data
	m1 := n.seeder.ProvisionUser(t, "nsgrantmaintainer1", "nsgrantmaintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)
	m2 := n.seeder.ProvisionUser(t, "nsgrantmaintainer2", "nsgrantmaintainer2@t.com", "Maintainer")
	require.NotEmpty(t, m2)
	d1 := n.seeder.ProvisionUser(t, "nsgrantdev1", "nsgrantdev1@t.com", "Developer")
	require.NotEmpty(t, d1)
	d2 := n.seeder.ProvisionUser(t, "nsgrantdev2", "nsgrantdev2@t.com", "Developer")
	require.NotEmpty(t, d2)
	d3 := n.seeder.ProvisionUser(t, "nsgrantdev3", "nsgrantdev3@t.com", "Developer")
	require.NotEmpty(t, d3)
	d4 := n.seeder.ProvisionUser(t, "nsgrantdev4", "nsgrantdev4@t.com", "Developer")
	require.NotEmpty(t, d4)

	g1 := n.seeder.ProvisionUser(t, "nsgrantguest1", "nsgrantguest1@t.com", "Guest")
	require.NotEmpty(t, g1)
	g2 := n.seeder.ProvisionUser(t, "nsgrantguest2", "nsgrantguest2@t.com", "Guest")
	require.NotEmpty(t, g2)
	g3 := n.seeder.ProvisionUser(t, "nsgrantguest3", "nsgrantguest3@t.com", "Guest")
	require.NotEmpty(t, g3)

	a1 := n.seeder.ProvisionUser(t, "nsgrantadmin1", "nsgrantadmin1@t.com", "Admin")
	require.NotEmpty(t, a1)
	a2 := n.seeder.ProvisionUser(t, "nsgrantadmin2", "nsgrantadmin2@t.com", "Admin")
	require.NotEmpty(t, a2)
	a3 := n.seeder.ProvisionUser(t, "nsgrantadmin3", "nsgrantadmin3@t.com", "Admin")
	require.NotEmpty(t, a3)

	nsName := "access-test-ns"
	nsId := n.seeder.CreateNamespace(t, nsName, "", "Team", false, m1)
	require.NotEmpty(t, nsId)

	tcs := []struct {
		name       string
		body       map[string]any
		identifier string
		statusCode int
	}{
		{
			name:       "Invalid body",
			body:       map[string]any{"test": "test"},
			identifier: "nsgrant-fail",
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Invalid Resource Type",
			body: map[string]any{
				"user_id":       m2,
				"access_level":  "Maintainer",
				"resource_type": "Namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Invalid Resource ID",
			body: map[string]any{
				"user_id":       m2,
				"access_level":  "Maintainer",
				"resource_type": "namespace",
				"resource_id":   "non-existing-id",
				"granted_by":    "admin",
			},
			identifier: "non-existing-id",
			statusCode: http.StatusNotFound,
		},
		{
			name: "Non-existing user",
			body: map[string]any{
				"user_id":       "non-existing-user",
				"access_level":  "Maintainer",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "Guest with access level 'Guest'",
			body: map[string]any{
				"user_id":       g1,
				"access_level":  "Guest",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "Guest with access level 'Developer'",
			body: map[string]any{
				"user_id":       g2,
				"access_level":  "Developer",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusForbidden,
		},
		{
			name: "Guest with access level 'Maintainer'",
			body: map[string]any{
				"user_id":       g3,
				"access_level":  "Maintainer",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusForbidden,
		},
		{
			name: "Developer with access level 'Guest'",
			body: map[string]any{
				"user_id":       d1,
				"access_level":  "Guest",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "Developer with access level 'Developer'",
			body: map[string]any{
				"user_id":       d2,
				"access_level":  "Developer",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "Developer with access level 'Maintainer'",
			body: map[string]any{
				"user_id":       d3,
				"access_level":  "Maintainer",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusForbidden,
		},
		{
			name: "Admin with access level 'Guest'",
			body: map[string]any{
				"user_id":       a1,
				"access_level":  "Guest",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusForbidden,
		},
		{
			name: "Admin with access level 'Developer'",
			body: map[string]any{
				"user_id":       a2,
				"access_level":  "Developer",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusForbidden,
		},
		{
			name: "Admin with access level 'Maintainer'",
			body: map[string]any{
				"user_id":       a3,
				"access_level":  "Maintainer",
				"resource_type": "namespace",
				"resource_id":   nsId,
				"granted_by":    "admin",
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			var url string
			if tc.identifier == "" {
				url = n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceUsers, nsId)
			} else {
				url = n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceUsers, tc.identifier)
			}
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}

	t.Run("Attempt to overwrite existing access", func(t *testing.T) {
		reqBody1, err := json.Marshal(map[string]any{
			"user_id":       d4,
			"access_level":  "Guest",
			"resource_type": "namespace",
			"resource_id":   nsId,
			"granted_by":    "admin"})
		require.NoError(t, err)

		req1, err := http.NewRequest(http.MethodPost, n.testBaseURL+fmt.Sprintf(testdata.EndpointNamespaceUsers, nsId),
			bytes.NewReader(reqBody1))
		require.NoError(t, err)
		req1.Header.Set("Content-Type", "application/json")
		helpers.SetAuthCookie(req1, n.seeder.AdminToken(t))

		resp1, err := http.DefaultClient.Do(req1)
		require.NoError(t, err)
		defer resp1.Body.Close()
		require.Equal(t, http.StatusOK, resp1.StatusCode)

		reqBody2, err := json.Marshal(map[string]any{
			"user_id":       d4,
			"access_level":  "Developer",
			"resource_type": "namespace",
			"resource_id":   nsId,
			"granted_by":    "admin",
		})
		require.NoError(t, err)

		req2, err := http.NewRequest(http.MethodPost, n.testBaseURL+fmt.Sprintf(testdata.EndpointNamespaceUsers, nsId),
			bytes.NewReader(reqBody2))
		require.NoError(t, err)
		req2.Header.Set("Content-Type", "application/json")
		helpers.SetAuthCookie(req2, n.seeder.AdminToken(t))

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		helpers.AssertStatusCode(t, resp2, http.StatusConflict)
	})
}

func (n *NamespaceTestSuite) testNamespaceRevokeAccess(t *testing.T) {
	// Prepare data
	m1 := n.seeder.ProvisionUser(t, "nsrevokemaintainer1", "nsrevokemaintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	nsId := n.seeder.CreateNamespace(t, "nsrevoke1", "nsrevoke1", "Team", false, m1)
	require.NotEmpty(t, nsId)

	d1 := n.seeder.ProvisionUser(t, "nsrevokedev1", "nsrevokedev1@t.com", "Developer")
	d2 := n.seeder.ProvisionUser(t, "nsrevokedev2", "nsrevokedev2@t.com", "Developer")

	n.seeder.GrantAccess(t, nsId, "namespace", d1, "Developer")

	tcs := []struct {
		name       string
		body       map[string]any
		resourceId string
		userId     string
		statusCode int
	}{
		{
			name: "Revoke user access",
			body: map[string]any{
				"user_id":       d1,
				"resource_type": "namespace",
				"resource_id":   nsId,
			},
			resourceId: nsId,
			userId:     d1,
			statusCode: http.StatusOK,
		},
		{
			name: "Non existing user",
			body: map[string]any{
				"user_id":       "non-existing-uuid",
				"resource_type": "namespace",
				"resource_id":   nsId,
			},
			userId:     "non-existing-uuid",
			resourceId: nsId,
			statusCode: http.StatusNotFound,
		},
		{
			name: "Existing user without resource access",
			body: map[string]any{
				"user_id":       d2,
				"resource_type": "namespace",
				"resource_id":   nsId,
			},
			userId:     d2,
			resourceId: nsId,
			statusCode: http.StatusNotFound,
		},
		{
			name: "Invalid body",
			body: map[string]any{
				"test": "invalid body",
			},
			resourceId: nsId,
			userId:     "invalid-id",
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Invalid resource type",
			body: map[string]any{
				"user_id":       d2,
				"resource_type": "Namespace",
				"resource_id":   nsId,
			},
			resourceId: nsId,
			userId:     d2,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Invalid resource_id",
			body: map[string]any{
				"user_id":       d2,
				"resource_type": "namespace",
				"resource_id":   "invalid-namespaceid",
			},
			userId:     d2,
			resourceId: "invalid-namespaceid",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceUsers, tc.resourceId) + "/" + tc.userId
			req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(reqBody))
			require.NoError(t, err)
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (n *NamespaceTestSuite) testNamespaceGetAndExists(t *testing.T) {
	// Prepare data
	m := n.seeder.ProvisionUser(t, "nsget-maintainer1", "nsget-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m)

	nsId := n.seeder.CreateNamespace(t, "nsget-1", "", "Team", false, m)
	require.NotEmpty(t, nsId)

	tcs := []struct {
		name       string
		identifier string
		statusCode int
	}{
		{
			name:       "Check by name",
			identifier: "nsget-1",
			statusCode: http.StatusOK,
		},
		{
			name:       "Check by id",
			identifier: nsId,
			statusCode: http.StatusOK,
		},
		{
			name:       "Check by non existing identifier",
			identifier: "non-existent-id",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceByID, tc.identifier)

			// Test GET
			reqGet, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(reqGet, n.seeder.AdminToken(t))

			resp1, err := http.DefaultClient.Do(reqGet)
			require.NoError(t, err)
			defer resp1.Body.Close()
			helpers.AssertStatusCode(t, resp1, tc.statusCode)

			// Test HEAD
			reqHead, err := http.NewRequest(http.MethodHead, url, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(reqHead, n.seeder.AdminToken(t))

			resp2, err := http.DefaultClient.Do(reqHead)
			require.NoError(t, err)
			defer resp2.Body.Close()
			helpers.AssertStatusCode(t, resp2, tc.statusCode)
		})
	}
}

func (n *NamespaceTestSuite) testUpdateNamespace(t *testing.T) {
	// Prepare data
	m := n.seeder.ProvisionUser(t, "nsupdate-maintainer1", "nsupdate-maintainer1@test.com", "Maintainer")
	require.NotEmpty(t, m)
	nsName := "nsupdate-1"
	nsId := n.seeder.CreateNamespace(t, nsName, "", "Team", false, m)
	require.NotEmpty(t, nsId)

	tcs := []struct {
		name       string
		identifier string
		body       map[string]any
		statusCode int
	}{
		{
			name:       "Update description by name",
			identifier: nsName,
			body: map[string]any{
				"id":          nsName,
				"description": "description changed by name",
				"purpose":     "Team",
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "Update description by id",
			identifier: nsId,
			body: map[string]any{
				"id":          nsName,
				"description": "description changed by id",
				"purpose":     "Team",
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "Update purpose",
			identifier: nsId,
			body: map[string]any{
				"id":      nsName,
				"purpose": "Project",
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "Update purpose with invalid-value",
			identifier: nsId,
			body: map[string]any{
				"id":      nsName,
				"purpose": "inavalid-purpose",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Update non-existing namespace",
			identifier: "non-existing-ns",
			body: map[string]any{
				"id":          "non-existing-ns",
				"description": "description1",
				"purpose":     "Team",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "Invalid body",
			body: map[string]any{
				"test": "invalid-body",
			},
			identifier: "invalid-body-test",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.body)
			require.NoError(t, err)
			require.NotEmpty(t, reqBody)

			url := n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceByID, tc.identifier)

			req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(reqBody))
			require.NoError(t, err)
			require.NotNil(t, req)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (n *NamespaceTestSuite) testDeleteNamespace(t *testing.T) {
	// Prepare data
	m := n.seeder.ProvisionUser(t, "nsdelete-maintainer1", "nsdelete-maintainer1@test.com", "Maintainer")
	require.NotEmpty(t, m)

	nsId := n.seeder.CreateNamespace(t, "nsDelete1", "", "Team", false, m)
	require.NotEmpty(t, nsId)

	nsName := "nsdelete-2"
	nsId2 := n.seeder.CreateNamespace(t, nsName, "", "Team", false, m)
	require.NotEmpty(t, nsId2)

	tcs := []struct {
		identifier string
		name       string
		statusCode int
	}{
		{
			identifier: nsId,
			name:       "Delete namespace by id",
			statusCode: http.StatusOK,
		},
		{
			identifier: nsName,
			name:       "Delete namespace by name",
			statusCode: http.StatusOK,
		},
		{
			identifier: "non existing namespace",
			name:       "Delete non-existing namespace",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceByID, tc.identifier)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)
			require.NotNil(t, req)
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (n *NamespaceTestSuite) testNamespaceStateChange(t *testing.T) {
	// Prepare data
	m := n.seeder.ProvisionUser(t, "nsstatechange-maintainer1", "nsstatechange-maintainer1@test.com",
		"Maintainer")
	require.NotEmpty(t, m)

	n1 := n.seeder.CreateNamespace(t, "nsstatechange1", "", "Team", false, m)
	require.NotEmpty(t, n1)

	n2 := n.seeder.CreateNamespace(t, "nsstatechange2", "", "Team", false, m)
	require.NotEmpty(t, n2)
	n.seeder.SetNamespaceDeprecated(t, n2)

	n3 := n.seeder.CreateNamespace(t, "nsstatechange3", "", "Team", false, m)
	require.NotEmpty(t, n3)

	n4 := n.seeder.CreateNamespace(t, "nsstatechange4", "", "Team", false, m)
	require.NotEmpty(t, n4)

	tcs := []struct {
		name       string
		identifier string
		newState   string
		statusCode int
		errMsg     string
	}{
		{
			name:       "State change from 'Active' to 'Deprecated'",
			identifier: n1,
			newState:   "Deprecated",
			statusCode: http.StatusOK,
		},
		{
			name:       "State change from 'Deprecated' to 'Disabled'",
			identifier: n2,
			newState:   "Disabled",
			statusCode: http.StatusOK,
		},
		{
			name:       "State change request with same state as current",
			newState:   "Active",
			identifier: n3,
			statusCode: http.StatusOK,
		},
		{
			name:       "State change to non-existent namespace",
			identifier: "non-existent-namespace",
			newState:   "Deprecated",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Invalid value for query param 'state'",
			newState:   "invalid-state",
			identifier: n3,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "State change from 'Active' to 'Disabled'",
			identifier: n4,
			newState:   "Disabled",
			statusCode: http.StatusForbidden,
			errMsg:     "Not allowed to change namespace state from 'Active' to 'Disabled'",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceState, tc.identifier) + "?state=" + tc.newState
			req, err := http.NewRequest(http.MethodPatch, url, nil)
			require.NoError(t, err)
			require.NotNil(t, req)
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)

			if tc.errMsg == "" {
				return
			}

			var resBody map[string]any
			err = json.NewDecoder(resp.Body).Decode(&resBody)
			require.NoError(t, err)

			errMsg := resBody["error_message"].(string)
			assert.Equal(t, tc.errMsg, errMsg)
		})
	}
}

func (n *NamespaceTestSuite) testNamespaceVisibilityChange(t *testing.T) {
	// Prepare data
	m := n.seeder.ProvisionUser(t, "nsvisibilitychange-maintainer1", "nsvisibilitychange-maintainer1@t.com",
		"Maintainer")
	require.NotEmpty(t, m)

	n1 := n.seeder.CreateNamespace(t, "nsvisibilitychange1", "", "Team", false, m)
	require.NotEmpty(t, n1)

	n2 := n.seeder.CreateNamespace(t, "nsvisibilitychange2", "", "Team", false, m)
	require.NotEmpty(t, n2)
	n.seeder.SetNamespaceDeprecated(t, n2)

	n3 := n.seeder.CreateNamespace(t, "nsvisibilitychange3", "", "Team", false, m)
	require.NotEmpty(t, n3)

	n4 := n.seeder.CreateNamespace(t, "nsvisibilitychange4", "", "Team", false, m)
	require.NotEmpty(t, n4)
	n.seeder.SetNamespaceDisabled(t, n4)

	tcs := []struct {
		name       string
		identifier string
		isPublic   string // query param value
		statusCode int
		errMsg     string
	}{
		{
			name:       "Change private namespace to public",
			identifier: n1,
			isPublic:   "true",
			statusCode: http.StatusOK,
		},
		{
			name:       "Change public namespace back to private",
			identifier: n1,
			isPublic:   "false",
			statusCode: http.StatusOK,
		},
		{
			name:       "No-op: Change private to private",
			identifier: n2,
			isPublic:   "false",
			statusCode: http.StatusOK,
		},
		{
			name:       "Invalid boolean query param",
			identifier: n3,
			isPublic:   "not-a-bool",
			statusCode: http.StatusBadRequest,
			errMsg:     "Invalid query param",
		},
		{
			name:       "Non-existent namespace",
			identifier: "ghost-namespace",
			isPublic:   "true",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Forbidden: Change visibility of disabled namespace",
			identifier: n4,
			isPublic:   "true",
			statusCode: http.StatusForbidden,
			errMsg:     "Not allowed to change visibility of a namespace when it is in disabled state",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceVisibility, tc.identifier) + "?public=" + tc.isPublic

			req, err := http.NewRequest(http.MethodPatch, url, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)

			if tc.errMsg != "" {
				var resBody map[string]any
				err = json.NewDecoder(resp.Body).Decode(&resBody)
				require.NoError(t, err)
				assert.Equal(t, tc.errMsg, resBody["error_message"])
			}
		})
	}
}

func (n *NamespaceTestSuite) testListNamespaces(t *testing.T) {
	// Clean all the namespaces created by other tests to able to predict the order
	n.seeder.DeleteAllNamespaces(t)

	// Prepare data
	m := n.seeder.ProvisionUser(t, "listns-maintainer1", "listns-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m)
	n1 := n.seeder.CreateNamespace(t, "listns-engineering", "engineering team", "Team", false, m)
	require.NotEmpty(t, n1)
	n2 := n.seeder.CreateNamespace(t, "listns-payment", "Payment Service repository", "Project", false, m)
	require.NotEmpty(t, n2)
	n3 := n.seeder.CreateNamespace(t, "listns-marketing", "Marketing department assets", "Team", true, m)
	require.NotEmpty(t, n3)
	n4 := n.seeder.CreateNamespace(t, "listns-security-scans", "Automated security vulnerability reports", "Project", false, m)
	require.NotEmpty(t, n4)
	n5 := n.seeder.CreateNamespace(t, "listns-frontend-core", "Core UI component library", "Team", true, m)
	require.NotEmpty(t, n5)
	n6 := n.seeder.CreateNamespace(t, "listns-data-science", "Jupyter notebooks and ML models", "Project", false, m)
	require.NotEmpty(t, n6)
	n7 := n.seeder.CreateNamespace(t, "listns-devops-tools", "Infrastructure as code and CI/CD scripts", "Team", false, m)
	require.NotEmpty(t, n7)
	n8 := n.seeder.CreateNamespace(t, "listns-hr-portal", "Human resources internal application", "Team", false, m)
	require.NotEmpty(t, n8)
	n9 := n.seeder.CreateNamespace(t, "listns-mobile-ios", "iOS mobile application builds", "Project", true, m)
	require.NotEmpty(t, n9)
	n10 := n.seeder.CreateNamespace(t, "listns-mobile-android", "Android mobile application builds", "Project", true, m)
	require.NotEmpty(t, n10)
	n11 := n.seeder.CreateNamespace(t, "listns-legal-docs", "Legal and compliance documentation", "Team", false, m)
	require.NotEmpty(t, n11)
	n12 := n.seeder.CreateNamespace(t, "listns-api-gateway", "Centralized API Gateway configurations", "Team", false, m)
	require.NotEmpty(t, n12)
	n13 := n.seeder.CreateNamespace(t, "listns-analytics-v2", "Customer behavior tracking service", "Project", true, m)
	require.NotEmpty(t, n13)
	n14 := n.seeder.CreateNamespace(t, "listns-customer-support", "Support ticket management system", "Team", true, m)
	require.NotEmpty(t, n14)
	n.seeder.SetNamespaceDeprecated(t, n14)
	n.seeder.SetNamespaceDisabled(t, n13)

	tcs := []struct {
		name             string
		queryParams      map[string]string
		statusCode       int
		total            int
		countCurrentPage int
		firstId          string
		lastId           string
		expectedIds      []any
		descSorting      bool
	}{
		{
			name:             "Without any filters",
			queryParams:      map[string]string{},
			statusCode:       http.StatusOK,
			total:            -1,
			countCurrentPage: -1,
		},
		{
			name: "With search term 'mobile' exists in name and description",
			queryParams: map[string]string{
				"search": "mobile",
			},
			total:            2,
			countCurrentPage: 2,
			expectedIds:      []any{n9, n10},
		},
		{
			name: "With search term 'application' exists in description",
			queryParams: map[string]string{
				"search": "application",
			},
			total:            3,
			countCurrentPage: 3,
			expectedIds:      []any{n8, n9, n10},
		},
		{
			name: "With Search term and pagination",
			queryParams: map[string]string{
				"search": "application",
				"page":   "2",
				"limit":  "1",
			},
			total:            3,
			countCurrentPage: 1,
			expectedIds:      []any{n9},
		},
		{
			name: "With filter state = 'Deprecated'",
			queryParams: map[string]string{
				"state": "Deprecated",
			},
			total:            1,
			countCurrentPage: 1,
			expectedIds:      []any{n14},
		},
		{
			name: "With filter state = 'Disabled'",
			queryParams: map[string]string{
				"state": "Disabled",
			},
			total:            1,
			countCurrentPage: 1,
			expectedIds:      []any{n13},
		},
		{
			name: "With filter purpose='Project'",
			queryParams: map[string]string{
				"purpose": "Project",
			},
			total:            6,
			countCurrentPage: 6,
			expectedIds:      []any{n2, n4, n6, n9, n10, n13},
		},
		{
			name: "With filter purpose='Project' and sorted by 'name' in descending order",
			queryParams: map[string]string{
				"purpose": "Project",
				"sort_by": "name",
				"order":   "desc",
			},
			total:            6,
			countCurrentPage: 6,
			descSorting:      true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqURL := n.testBaseURL + testdata.EndpointNamespaces

			queryParams := url.Values{}
			for key, value := range tc.queryParams {
				queryParams.Add(key, value)
			}
			reqURL += "?" + queryParams.Encode()

			req, err := http.NewRequest(http.MethodGet, reqURL, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			var resBody map[string]any
			err = json.NewDecoder(resp.Body).Decode(&resBody)
			require.NoError(t, err)

			total := resBody["total"]
			if tc.total != -1 {
				require.Equal(t, float64(tc.total), total)
			}
			namespaces, ok := resBody["namespaces"].([]any)
			require.True(t, ok)

			var resultIds []string
			var names []string
			for _, ns := range namespaces {
				ns := ns.(map[string]any)
				id, ok := ns["id"].(string)
				require.True(t, ok)
				require.NotEmpty(t, id)

				name, ok := ns["name"].(string)
				require.True(t, ok)
				require.NotEmpty(t, name)

				resultIds = append(resultIds, id)
				names = append(names, name)
			}

			if tc.countCurrentPage != -1 {
				require.Equal(t, tc.countCurrentPage, len(namespaces))
			}

			if tc.total != -1 {
				if tc.firstId != "" && len(resultIds) > 0 {
					assert.Equal(t, tc.firstId, resultIds[0])
				}

				if tc.lastId != "" && len(resultIds) > 0 {
					assert.Equal(t, tc.lastId, resultIds[len(resultIds)-1])
				}
			}

			if len(tc.expectedIds) != 0 && tc.total != -1 {
				assert.ElementsMatch(t, tc.expectedIds, resultIds)
			}

			if tc.descSorting {
				isDesc := slices.IsSortedFunc(names, func(a, b string) int {
					return cmp.Compare(b, a)
				})
				assert.True(t, isDesc, "expected names sorted in descending order")
			}
		})
	}
}

func (n *NamespaceTestSuite) testNamespaceUserAccessList(t *testing.T) {
	// Prepare data
	m1 := n.seeder.ProvisionUser(t, "useraccess-maintainer1", "useraccess-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)
	n1 := n.seeder.CreateNamespace(t, "useraccess-engineering", "engineering team", "Team", false, m1)
	require.NotEmpty(t, n1)

	m2 := n.seeder.ProvisionUser(t, "useraccess-maintainer2", "useraccess-maintainer2@t.com", "Maintainer")
	require.NotEmpty(t, m2)
	n.seeder.GrantAccess(t, n1, "namespace", m2, "Maintainer")

	m3 := n.seeder.ProvisionUser(t, "useraccess-maintainer3", "useraccess-maintainer3@t.com", "Maintainer")
	require.NotEmpty(t, m3)
	n.seeder.GrantAccess(t, n1, "namespace", m3, "Developer")

	d1 := n.seeder.ProvisionUser(t, "useraccess-developer1", "useraccess-developer1@t.com", "Developer")
	require.NotEmpty(t, d1)
	n.seeder.GrantAccess(t, n1, "namespace", d1, "Developer")

	d2 := n.seeder.ProvisionUser(t, "useraccess-developer2", "useraccess-developer2@t.com", "Developer")
	require.NotEmpty(t, d2)
	n.seeder.GrantAccess(t, n1, "namespace", d2, "Guest")

	d3 := n.seeder.ProvisionUser(t, "useraccess-developer3", "useraccess-developer3@t.com", "Developer")
	require.NotEmpty(t, d3)
	n.seeder.GrantAccess(t, n1, "namespace", d3, "Developer")

	d4 := n.seeder.ProvisionUser(t, "useraccess-developer4", "useraccess-developer4@t.com", "Developer")
	require.NotEmpty(t, d4)
	n.seeder.GrantAccess(t, n1, "namespace", d4, "Developer")

	g1 := n.seeder.ProvisionUser(t, "useraccess-guest1", "useraccess-guest1@t.com", "Guest")
	require.NotEmpty(t, g1)
	n.seeder.GrantAccess(t, n1, "namespace", g1, "Guest")

	g2 := n.seeder.ProvisionUser(t, "useraccess-guest2", "useraccess-guest2@t.com", "Guest")
	require.NotEmpty(t, g2)
	n.seeder.GrantAccess(t, n1, "namespace", g2, "Guest")

	g3 := n.seeder.ProvisionUser(t, "useraccess-guest3", "useraccess-guest3@t.com", "Guest")
	require.NotEmpty(t, g3)
	n.seeder.GrantAccess(t, n1, "namespace", g3, "Guest")

	g4 := n.seeder.ProvisionUser(t, "useraccess-guest4", "useraccess-guest4@t.com", "Guest")
	require.NotEmpty(t, g4)
	n.seeder.GrantAccess(t, n1, "namespace", g4, "Guest")

	tcs := []struct {
		name             string
		queryParams      map[string]string
		statusCode       int
		total            int
		countCurrentPage int
		expectedUserIds  []any
		expectedErr      string
	}{
		{
			name:             "List all access for namespace (Default)",
			queryParams:      map[string]string{},
			statusCode:       http.StatusOK,
			total:            11,
			countCurrentPage: 11,
			expectedUserIds:  []any{m1, m2, m3, d1, d2, d3, d4, g1, g2, g3, g4},
		},
		{
			name: "Filter by Access Level 'Maintainer'",
			queryParams: map[string]string{
				"access_level": "Maintainer",
			},
			statusCode:       http.StatusOK,
			total:            2,
			countCurrentPage: 2,
			expectedUserIds:  []any{m1, m2},
		},
		{
			name: "Filter by Access Level 'Developer'",
			queryParams: map[string]string{
				"access_level": "Developer",
			},
			statusCode:       http.StatusOK,
			total:            4,
			countCurrentPage: 4,
			expectedUserIds:  []any{m3, d1, d3, d4},
		},
		{
			name: "Filter by Access Level 'Guest' with Pagination",
			queryParams: map[string]string{
				"access_level": "Guest",
				"page":         "1",
				"limit":        "3",
			},
			statusCode:       http.StatusOK,
			total:            5,
			countCurrentPage: 3,
		},
		{
			name: "Search users by username (Search term 'guest')",
			queryParams: map[string]string{
				"search": "guest",
			},
			statusCode:       http.StatusOK,
			total:            4,
			countCurrentPage: 4,
			expectedUserIds:  []any{g1, g2, g3, g4},
		},
		{
			name: "Invalid Sort Field validation",
			queryParams: map[string]string{
				"sort_by": "invalid_field",
			},
			statusCode:  http.StatusBadRequest,
			expectedErr: "Not allowed sort field: invalid_field",
		},
		{
			name: "Invalid Filter Field validation",
			queryParams: map[string]string{
				"unknown_filter": "some_val",
			},
			statusCode:  http.StatusBadRequest,
			expectedErr: "Not allowed filter field: unknown_filter",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqURL := n.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceUsers, n1)
			if len(tc.queryParams) > 0 {
				qParams := url.Values{}
				for k, v := range tc.queryParams {
					qParams.Add(k, v)
				}
				reqURL += "?" + qParams.Encode()
			}

			req, err := http.NewRequest(http.MethodGet, reqURL, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, n.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)

			var resBody map[string]any
			err = json.NewDecoder(resp.Body).Decode(&resBody)
			require.NoError(t, err)

			if tc.statusCode != http.StatusOK {
				if tc.expectedErr != "" {
					assert.Contains(t, resBody["error_message"].(string), tc.expectedErr)
				}
				return
			}

			require.Equal(t, float64(tc.total), resBody["total"])
			accessList, ok := resBody["accesses"].([]any)
			require.True(t, ok)
			assert.Equal(t, tc.countCurrentPage, len(accessList))

			if len(tc.expectedUserIds) > 0 {
				var actualUserIds []string
				for _, entry := range accessList {
					m := entry.(map[string]any)
					actualUserIds = append(actualUserIds, m["user_id"].(string))
				}
				assert.ElementsMatch(t, tc.expectedUserIds, actualUserIds)
			}
		})
	}
}

func (n *NamespaceTestSuite) testNamespaceListRepositories(t *testing.T) {}
func (n *NamespaceTestSuite) testNamespaceHTTPMethods(t *testing.T)      {}
func (n *NamespaceTestSuite) testNamespaceContentType(t *testing.T)      {}