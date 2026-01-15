package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/integration/seeder"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/require"
)

type RepositorySuite struct {
	name        string
	apiVersion  string
	seeder      *seeder.TestDataSeeder
	testBaseURL string
}

func NewRepositorySuite(seeder *seeder.TestDataSeeder, baseURL string) *RepositorySuite {
	return &RepositorySuite{
		name:        "RepositoryAPI",
		apiVersion:  "v1",
		seeder:      seeder,
		testBaseURL: baseURL,
	}
}

func (r *RepositorySuite) Run(t *testing.T) {
	t.Run("CreateRepository", r.testCreateRepository)

}

func (r *RepositorySuite) Name() string {
	return r.name
}

func (r *RepositorySuite) APIVersion() string {
	return r.apiVersion
}

// test plan
// seeder create repository
// 1. bad body -> 400
// 2. invalid name -> 400
// 3. conflict -> 409
// 4. ns doesn't exist -> 404
// 5. creator not found -> 404
func (r *RepositorySuite) testCreateRepository(t *testing.T) {
	// Prepare data
	m1 := r.seeder.ProvisionUser(t, "test-create-repo-u1", "test-create-repo-u1@t.com", constants.RoleMaintainer)
	nsId := r.seeder.CreateNamespace(t, "test-create-repo-ns", "create repo test", constants.NamespacePurposeTeam, false, m1)
	d1 := r.seeder.ProvisionUser(t, "test-create-repo-u2", "test-create-repo-u2@t.com", constants.RoleDeveloper)
	r.seeder.GrantAccess(t, nsId, constants.ResourceTypeNamespace, d1, constants.AccessLevelDeveloper)

	n2 := r.seeder.CreateNamespace(t, "test-create-repo-ns2", "create repo test2", constants.NamespacePurposeTeam, true, m1)

	n3 := r.seeder.CreateNamespace(t, "test-create-repo-ns3", "create repo test3", constants.NamespacePurposeTeam, true, m1)
	r.seeder.SetNamespaceDeprecated(t, n3)

	n4 := r.seeder.CreateNamespace(t, "test-create-repo-ns4", "create repo test4", constants.NamespacePurposeTeam, true, m1)
	r.seeder.SetNamespaceDisabled(t, n4)

	repoName := "test-create-repo6"

	r.seeder.CreateRepository(t, repoName, "", d1, nsId, false)

	tcs := []struct {
		tcName     string
		body       any
		statusCode int
	}{
		{
			tcName:     "Non JSON body",
			body:       "invalid body",
			statusCode: http.StatusBadRequest,
		},
		{
			tcName: "Invalid repository name",
			body: map[string]any{
				"name":         "6&test-invalid-repo-name",
				"description":  "",
				"is_public":    false,
				"namespace_id": nsId,
				"created_by":   d1},
			statusCode: http.StatusBadRequest,
		},
		{
			tcName: "Invalid namespace id",
			body: map[string]any{
				"name":         "test-create-repo1",
				"description":  "",
				"is_public":    false,
				"namespace_id": "invalid-ns-id",
				"created_by":   d1,
			},
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			tcName: "Non existing creator",
			body: map[string]any{
				"name":         "test-create-repo1",
				"description":  "",
				"is_public":    false,
				"namespace_id": nsId,
				"created_by":   "non-existing-user",
			},
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			tcName: "Can maintainer create repository",
			body: map[string]any{
				"name":         "test-create-repo1",
				"description":  "",
				"is_public":    false,
				"namespace_id": nsId,
				"created_by":   m1,
			},
			statusCode: http.StatusCreated,
		},
		{
			tcName: "Can developer create repository",
			body: map[string]any{
				"name":         "test-create-repo4",
				"description":  "",
				"is_public":    false,
				"namespace_id": nsId,
				"created_by":   d1,
			},
			statusCode: http.StatusCreated,
		},
		{
			tcName: "Cannot create public repository under a private namespace",
			body: map[string]any{
				"name":         "test-create-repo1",
				"description":  "",
				"is_public":    true,
				"namespace_id": nsId,
				"created_by":   d1,
			},
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			tcName: "Can create public repository",
			body: map[string]any{
				"name":         "test-create-repo3",
				"description":  "",
				"is_public":    true,
				"namespace_id": n2,
				"created_by":   d1,
			},
			statusCode: http.StatusCreated,
		},
		{
			tcName: "Cannot create repository under disabled namespace",
			body: map[string]any{
				"name":         "test-create-repo4",
				"description":  "",
				"is_public":    true,
				"namespace_id": n4,
				"created_by":   d1,
			},
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			tcName: "Cannot create repository under deprecated namespace",
			body: map[string]any{
				"name":         "test-create-repo5",
				"description":  "",
				"is_public":    false,
				"namespace_id": n3,
				"created_by":   d1,
			},
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			tcName: "Cannot create another repository with same name",
			body: map[string]any{
				"name":         repoName,
				"description":  "",
				"is_public":    false,
				"namespace_id": nsId,
				"created_by":   d1,
			},
			statusCode: http.StatusConflict,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tcName, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, r.testBaseURL+testdata.EndpointRepositories,
				bytes.NewReader(reqBody))
			require.NoError(t, err)

			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}