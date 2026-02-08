package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/integration/seeder"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/assert"
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
	t.Run("GetRepository", r.testGetRepository)
	t.Run("HeadRepository", r.testHeadRepository)
	t.Run("UpdateRepository", r.testUpdateRepository)
	t.Run("DeleteRepository", r.testDeleteRepository)
	t.Run("ChangeState", r.testChangeState)
	t.Run("ChangeVisibility", r.testChangeVisibility)
	t.Run("GrantAccess", r.testGrantAccess)
	t.Run("RevokeAccess", r.testRevokeAccess)
	t.Run("ListUserAccess", r.testListUserAccess)
}

func (r *RepositorySuite) Name() string {
	return r.name
}

func (r *RepositorySuite) APIVersion() string {
	return r.apiVersion
}

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

func (r *RepositorySuite) testGetRepository(t *testing.T) {
	m1 := r.seeder.ProvisionUser(t, "getrepo-maintainer1", "getrepo-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	n1 := r.seeder.CreateNamespace(t, "get-repo-testns-1", "", "Team", false, m1)
	require.NotEmpty(t, n1)

	r1 := r.seeder.CreateRepository(t, "get-repo1", "", "admin", n1, false)
	require.NotEmpty(t, r1)

	tcs := []struct {
		name       string
		statusCode int
		repoId     string
	}{
		{
			name:       "Non existent repository id",
			statusCode: http.StatusNotFound,
			repoId:     "non-existenet-id",
		},
		{
			name:       "Valid repository id",
			statusCode: http.StatusOK,
			repoId:     r1,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryByID, tc.repoId)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
			if tc.statusCode == http.StatusOK {
				var resBody struct {
					ID string `json:"id"`
				}
				err = json.NewDecoder(resp.Body).Decode(&resBody)
				require.NoError(t, err)

				assert.Equal(t, tc.repoId, resBody.ID)
			}
		})
	}
}

func (r *RepositorySuite) testHeadRepository(t *testing.T) {
	m1 := r.seeder.ProvisionUser(t, "headrepo-maintainer1", "head-repo-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	n1 := r.seeder.CreateNamespace(t, "head-repo-testns-1", "", "Team", false, m1)
	require.NotEmpty(t, n1)

	r1 := r.seeder.CreateRepository(t, "head-repo1", "", "admin", n1, false)
	require.NotEmpty(t, r1)

	tcs := []struct {
		name       string
		statusCode int
		repoId     string
	}{
		{
			name:       "Non existent repository id",
			statusCode: http.StatusNotFound,
			repoId:     "non-existenet-id",
		},
		{
			name:       "Valid repository id",
			statusCode: http.StatusOK,
			repoId:     r1,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryByID, tc.repoId)

			req, err := http.NewRequest(http.MethodHead, url, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)

		})
	}
}

func (r *RepositorySuite) testUpdateRepository(t *testing.T) {
	m1 := r.seeder.ProvisionUser(t, "updaterepo-maintainer1", "updaterepo-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	n1 := r.seeder.CreateNamespace(t, "updaterepo-testns", "", "Team", false, m1)
	require.NotEmpty(t, n1)

	r1 := r.seeder.CreateRepository(t, "updaterepo-test1", "", "admin", n1, false)
	require.NotEmpty(t, r1)

	r2 := r.seeder.CreateRepository(t, "updaterepo-test2", "test2", "admin", n1, false)

	tcs := []struct {
		name       string
		repoId     string
		body       map[string]any
		statusCode int
	}{
		{
			name:       "Overwrite existing description",
			repoId:     r1,
			statusCode: http.StatusOK,
			body: map[string]any{
				"description": "Description is changed",
			},
		},
		{
			name:       "Set empty description",
			repoId:     r2,
			statusCode: http.StatusOK,
			body: map[string]any{
				"description": "",
			},
		},
		{
			name:       "Update non-existent repo",
			repoId:     "non-existent-id",
			statusCode: http.StatusNotFound,
			body: map[string]any{
				"description": "new description",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryByID, tc.repoId)

			reqBytes, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBytes))
			require.NoError(t, err)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (r *RepositorySuite) testDeleteRepository(t *testing.T) {
	m1 := r.seeder.ProvisionUser(t, "delete-repo-m1", "delete-repo-m1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	n1 := r.seeder.CreateNamespace(t, "delete-repo-testns1", "", "Team", false, m1)
	require.NotEmpty(t, n1)

	r1 := r.seeder.CreateRepository(t, "delete-repo1", "", "admin", n1, false)
	require.NotEmpty(t, r1)

	tcs := []struct {
		name       string
		repoId     string
		statusCode int
	}{
		{
			name:       "Delete valid repository",
			statusCode: http.StatusOK,
			repoId:     r1,
		},
		{
			name:       "Delete non-existent repository",
			repoId:     "non-existent-id",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryByID, tc.repoId)

			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)
			require.NotNil(t, req)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}

}

func (r *RepositorySuite) testChangeState(t *testing.T) {
	// Prepare data
	m1 := r.seeder.ProvisionUser(t, "changestate-maintainer1", "changestate-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	n1 := r.seeder.CreateNamespace(t, "changestaterepo-testns1", "", "Team", false, m1)
	require.NotEmpty(t, n1)

	r1 := r.seeder.CreateRepository(t, "changestate-repo1", "", "admin", n1, false)
	require.NotEmpty(t, r1)

	r2 := r.seeder.CreateRepository(t, "changestate-repo2", "", "admin", n1, false)
	require.NotEmpty(t, r2)
	r.seeder.SetRepositoryDeprecated(t, r2)

	r3 := r.seeder.CreateRepository(t, "changestate-repo3", "", "admin", n1, false)
	require.NotEmpty(t, r3)
	r.seeder.SetRepositoryDisabled(t, r3)

	r4 := r.seeder.CreateRepository(t, "changestate-repo4", "", "admin", n1, false)
	require.NotEmpty(t, r4)
	r.seeder.SetRepositoryDeprecated(t, r4)

	r5 := r.seeder.CreateRepository(t, "changestate-repo5", "", "admin", n1, false)
	require.NotEmpty(t, r5)
	r.seeder.SetRepositoryDisabled(t, r5)

	r6 := r.seeder.CreateRepository(t, "changestate-repo6", "", "admin", n1, false)
	require.NotEmpty(t, r6)

	r7 := r.seeder.CreateRepository(t, "changestate-repo7", "", "admin", n1, false)
	require.NotEmpty(t, r7)

	tcs := []struct {
		name       string
		newState   string
		repoId     string
		statusCode int
	}{
		{
			name:       "Non-existent repository",
			newState:   "Deprecated",
			repoId:     "non-existent-id",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Invalid new state",
			repoId:     r1,
			newState:   "invalid-state",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Active to Deprecated",
			repoId:     r1,
			newState:   "Deprecated",
			statusCode: http.StatusOK,
		},
		{
			name:       "Deprecated to Disabled",
			repoId:     r2,
			newState:   "Disabled",
			statusCode: http.StatusOK,
		},
		{
			name:       "Deprecated to Active",
			repoId:     r4,
			newState:   "Active",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Disabled to Active",
			repoId:     r5,
			newState:   "Active",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "No-op: Same state",
			repoId:     r6,
			newState:   "Active",
			statusCode: http.StatusOK,
		},
		{
			name:       "Active to Disabled",
			repoId:     r7,
			newState:   "Disabled",
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryState, tc.repoId) + "?state=" + tc.newState
			req, err := http.NewRequest(http.MethodPatch, url, nil)
			require.NoError(t, err)
			require.NotNil(t, req)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (r *RepositorySuite) testChangeVisibility(t *testing.T) {
	// Prepare data
	m := r.seeder.ProvisionUser(t, "repo-vischange-maintainer1", "repo-vischange-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m)

	n1 := r.seeder.CreateNamespace(t, "repo-vis-change-engineeringns1", "engineering team", "Team", false, m)
	require.NotEmpty(t, n1)
	r1 := r.seeder.CreateRepository(t, "repo-vis-change-test1", "", "admin", n1, false)
	require.NotEmpty(t, r1)
	r3 := r.seeder.CreateRepository(t, "repo-vis-change-test3", "", "admin", n1, false)
	require.NotEmpty(t, r3)

	n2 := r.seeder.CreateNamespace(t, "repo-vis-change-engineeringns2", "engineering team", "Team", true, m)
	require.NotEmpty(t, n2)
	r2 := r.seeder.CreateRepository(t, "repo-vis-change-test2", "", "admin", n2, false)
	require.NotEmpty(t, r2)

	n3 := r.seeder.CreateNamespace(t, "repo-vis-change-engineeringns3", "engineering team", "Team", true, m)
	require.NotEmpty(t, n3)
	r4 := r.seeder.CreateRepository(t, "repo-vis-change-test4", "", "admin", n3, true)
	require.NotEmpty(t, r4)
	r.seeder.SetNamespaceDisabled(t, n3)

	n6 := r.seeder.CreateNamespace(t, "repo-vis-change-engineeringns6", "engineering team", "Team", true, m)
	require.NotEmpty(t, n6)
	r5 := r.seeder.CreateRepository(t, "repo-vis-change-test5", "", "admin", n6, true)
	require.NotEmpty(t, r5)
	r.seeder.SetRepositoryDisabled(t, r5)

	r6 := r.seeder.CreateRepository(t, "repo-vis-change-test6", "", "admin", n6, true)
	require.NotEmpty(t, r6)

	r7 := r.seeder.CreateRepository(t, "repo-vis-change-test7", "", "admin", n6, false)
	require.NotEmpty(t, r7)

	tcs := []struct {
		name       string
		repoId     string
		isPublic   string
		statusCode int
	}{
		{
			name:       "Private namespace cannot have public repository",
			repoId:     r1,
			isPublic:   "true",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Can change visibility from private to public if namespace is public",
			repoId:     r2,
			isPublic:   "true",
			statusCode: http.StatusOK,
		},
		{
			name:       "Can change visibility from public to private if namespace is public",
			repoId:     r3,
			isPublic:   "false",
			statusCode: http.StatusOK,
		},
		{
			name:       "Cannot change visibility if Namespace is disabled",
			repoId:     r4,
			isPublic:   "false",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Cannot change visiblity if Repository is disabled",
			repoId:     r5,
			isPublic:   "false",
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name:       "Cannot change visibility to a non-existent repository",
			repoId:     "non-existent-id",
			isPublic:   "true",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "No-op: Change private to private",
			repoId:     r7,
			isPublic:   "false",
			statusCode: http.StatusOK,
		},
		{
			name:       "No-op: Change public to public",
			repoId:     r6,
			isPublic:   "true",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryVisibility, tc.repoId) + "?public=" + tc.isPublic

			req, err := http.NewRequest(http.MethodPatch, url, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (r *RepositorySuite) testListUserAccess(t *testing.T) {
	// Prepare data
	m1 := r.seeder.ProvisionUser(t, "repo-useraccess-maintainer1", "repo-useraccess-maintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	ns := r.seeder.CreateNamespace(t, "repo-useraccess-engineering", "engineering team", "Team", false, m1)
	require.NotEmpty(t, ns)

	repo := r.seeder.CreateRepository(t, "test-repo", "", "admin", ns, false)
	require.NotEmpty(t, repo)

	m2 := r.seeder.ProvisionUser(t, "repo-useraccess-maintainer2", "repo-useraccess-maintainer2@t.com", "Maintainer")
	require.NotEmpty(t, m2)
	r.seeder.GrantAccess(t, repo, "Repository", m2, "Developer")

	m3 := r.seeder.ProvisionUser(t, "repo-useraccess-maintainer3", "repo-useraccess-maintainer3@t.com", "Maintainer")
	require.NotEmpty(t, m3)
	r.seeder.GrantAccess(t, repo, "Repository", m3, "Developer")

	d1 := r.seeder.ProvisionUser(t, "repo-useraccess-developer1", "repo-useraccess-developer1@t.com", "Developer")
	require.NotEmpty(t, d1)
	r.seeder.GrantAccess(t, repo, "Repository", d1, "Developer")

	d2 := r.seeder.ProvisionUser(t, "repo-useraccess-developer2", "repo-useraccess-developer2@t.com", "Developer")
	require.NotEmpty(t, d2)
	r.seeder.GrantAccess(t, repo, "Repository", d2, "Guest")

	d3 := r.seeder.ProvisionUser(t, "repo-useraccess-developer3", "repo-useraccess-developer3@t.com", "Developer")
	require.NotEmpty(t, d3)
	r.seeder.GrantAccess(t, repo, "Repository", d3, "Developer")

	d4 := r.seeder.ProvisionUser(t, "repo-useraccess-developer4", "repo-useraccess-developer4@t.com", "Developer")
	require.NotEmpty(t, d4)
	r.seeder.GrantAccess(t, repo, "Repository", d4, "Developer")

	g1 := r.seeder.ProvisionUser(t, "repo-useraccess-guest1", "repo-useraccess-guest1@t.com", "Guest")
	require.NotEmpty(t, g1)
	r.seeder.GrantAccess(t, repo, "Repository", g1, "Guest")

	g2 := r.seeder.ProvisionUser(t, "repo-useraccess-guest2", "repo-useraccess-guest2@t.com", "Guest")
	require.NotEmpty(t, g2)
	r.seeder.GrantAccess(t, repo, "Repository", g2, "Guest")

	g3 := r.seeder.ProvisionUser(t, "repo-useraccess-guest3", "repo-useraccess-guest3@t.com", "Guest")
	require.NotEmpty(t, g3)
	r.seeder.GrantAccess(t, repo, "Repository", g3, "Guest")

	g4 := r.seeder.ProvisionUser(t, "repo-useraccess-guest4", "repo-useraccess-guest4@t.com", "Guest")
	require.NotEmpty(t, g4)
	r.seeder.GrantAccess(t, repo, "Repository", g4, "Guest")

	// Grant access at namespace level to test inheritance
	nsUser := r.seeder.ProvisionUser(t, "repo-useraccess-ns-user", "repo-useraccess-ns-user@t.com", "Developer")
	require.NotEmpty(t, nsUser)
	r.seeder.GrantAccess(t, ns, "Namespace", nsUser, "Developer")

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
			name:             "List all access for repository (Default) - includes namespace access",
			queryParams:      map[string]string{},
			statusCode:       http.StatusOK,
			total:            12, // 11 direct repo access + 1 namespace access
			countCurrentPage: 12,
			expectedUserIds:  []any{m1, m2, m3, d1, d2, d3, d4, g1, g2, g3, g4, nsUser},
		},
		// This needs to be further consideration.
		// {
		// 	name: "Filter by Access Level 'Maintainer'",
		// 	queryParams: map[string]string{
		// 		"access_level": "Maintainer",
		// 	},
		// 	statusCode:       http.StatusOK,
		// 	total:            2,
		// 	countCurrentPage: 2,
		// 	expectedUserIds:  []any{m1, m2},
		// },
		// {
		// 	name: "Filter by Access Level 'Developer'",
		// 	queryParams: map[string]string{
		// 		"access_level": "Developer",
		// 	},
		// 	statusCode:       http.StatusOK,
		// 	total:            5, // 4 repo-level + 1 namespace-level
		// 	countCurrentPage: 5,
		// 	expectedUserIds:  []any{m3, d1, d3, d4, nsUser},
		// },
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
			name: "Search users by username (Search term 'ns-user')",
			queryParams: map[string]string{
				"search": "ns-user",
			},
			statusCode:       http.StatusOK,
			total:            1,
			countCurrentPage: 1,
			expectedUserIds:  []any{nsUser},
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
			reqURL := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryUsers, repo)
			if len(tc.queryParams) > 0 {
				qParams := url.Values{}
				for k, v := range tc.queryParams {
					qParams.Add(k, v)
				}
				reqURL += "?" + qParams.Encode()
			}

			req, err := http.NewRequest(http.MethodGet, reqURL, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

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
			accessList, ok := resBody["entities"].([]any)
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

func (r *RepositorySuite) testGrantAccess(t *testing.T) {
	// Prepare data
	m1 := r.seeder.ProvisionUser(t, "repo-grantmaintainer1", "repo-grantmaintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)
	m2 := r.seeder.ProvisionUser(t, "repo-grantmaintainer2", "repo-grantmaintainer2@t.com", "Maintainer")
	require.NotEmpty(t, m2)
	d1 := r.seeder.ProvisionUser(t, "repo-grantdev1", "repo-grantdev1@t.com", "Developer")
	require.NotEmpty(t, d1)
	d2 := r.seeder.ProvisionUser(t, "repo-grantdev2", "repo-grantdev2@t.com", "Developer")
	require.NotEmpty(t, d2)
	d3 := r.seeder.ProvisionUser(t, "repo-grantdev3", "repo-grantdev3@t.com", "Developer")
	require.NotEmpty(t, d3)
	d4 := r.seeder.ProvisionUser(t, "repo-grantdev4", "repo-grantdev4@t.com", "Developer")
	require.NotEmpty(t, d4)

	g1 := r.seeder.ProvisionUser(t, "repo-grantguest1", "repo-grantguest1@t.com", "Guest")
	require.NotEmpty(t, g1)
	g2 := r.seeder.ProvisionUser(t, "repo-grantguest2", "repo-grantguest2@t.com", "Guest")
	require.NotEmpty(t, g2)
	g3 := r.seeder.ProvisionUser(t, "repo-grantguest3", "repo-grantguest3@t.com", "Guest")
	require.NotEmpty(t, g3)

	a1 := r.seeder.ProvisionUser(t, "repo-grantadmin1", "repo-grantadmin1@t.com", "Admin")
	require.NotEmpty(t, a1)
	a2 := r.seeder.ProvisionUser(t, "repo-grantadmin2", "repo-grantadmin2@t.com", "Admin")
	require.NotEmpty(t, a2)
	a3 := r.seeder.ProvisionUser(t, "repo-grantadmin3", "repo-grantadmin3@t.com", "Admin")
	require.NotEmpty(t, a3)

	n1 := r.seeder.CreateNamespace(t, "repo-grantns1", "", "Team", false, m1)
	require.NotEmpty(t, n1)
	r1 := r.seeder.CreateRepository(t, "repo-grant-test1", "", "admin", n1, false)
	require.NotEmpty(t, r1)

	n2 := r.seeder.CreateNamespace(t, "repo-grantns2", "", "Team", false, m2)
	require.NotEmpty(t, n2)
	r2 := r.seeder.CreateRepository(t, "repo-grant-test2", "", "admin", n2, false)
	require.NotEmpty(t, r2)
	r.seeder.SetNamespaceDisabled(t, n2)

	n3 := r.seeder.CreateNamespace(t, "repo-grantns3", "", "Team", false, m2)
	require.NotEmpty(t, n2)
	r3 := r.seeder.CreateRepository(t, "repo-grant-test3", "", "admin", n3, false)
	require.NotEmpty(t, r3)
	r.seeder.SetRepositoryDisabled(t, r3)

	tcs := []struct {
		name       string
		body       map[string]any
		repoId     string
		statusCode int
	}{
		{
			name:       "Grant access with invalid body",
			body:       map[string]any{"test": "test"},
			repoId:     "nsgrant-fail",
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Grant access with invalid resource type",
			body: map[string]any{
				"user_id":       d1,
				"access_level":  "Developer",
				"resource_type": "repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Grant access to non-exitent repository",
			body: map[string]any{
				"user_id":       d2,
				"access_level":  "Developer",
				"resource_type": "Repository",
				"resource_id":   "non-existing-id",
				"granted_by":    "admin",
			},
			repoId:     "non-existing-id",
			statusCode: http.StatusNotFound,
		},
		{
			name: "Grant access to non-existing user",
			body: map[string]any{
				"user_id":       "non-existing-user",
				"access_level":  "Developer",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusNotFound,
		},
		{
			name: "Guest with access level 'Guest'",
			body: map[string]any{
				"user_id":       g1,
				"access_level":  "Guest",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusOK,
		},
		{
			name: "Guest with access level 'Developer'",
			body: map[string]any{
				"user_id":       g2,
				"access_level":  "Developer",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusForbidden,
		},
		{
			name: "Guest with access level 'Maintainer'",
			body: map[string]any{
				"user_id":       g3,
				"access_level":  "Maintainer",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Developer with access level 'Guest'",
			body: map[string]any{
				"user_id":       d1,
				"access_level":  "Guest",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusOK,
		},
		{
			name: "Developer with access level 'Developer'",
			body: map[string]any{
				"user_id":       d2,
				"access_level":  "Developer",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusOK,
		},
		{
			name: "Developer with access level 'Maintainer'",
			body: map[string]any{
				"user_id":       d3,
				"access_level":  "Maintainer",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Access level 'Maintainer' cannot given to repository",
			body: map[string]any{
				"user_id":       m2,
				"access_level":  "Maintainer",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "Granting new access to disabled repository is not allowed",
			body: map[string]any{
				"user_id":       d4,
				"access_level":  "Developer",
				"resource_type": "Repository",
				"resource_id":   r3,
				"granted_by":    "admin",
			},
			repoId:     r3,
			statusCode: http.StatusForbidden,
		},
		{
			name: "Granting new access is not allowed if namespace is  disabled",
			body: map[string]any{
				"user_id":       d4,
				"access_level":  "Developer",
				"resource_type": "Repository",
				"resource_id":   r2,
				"granted_by":    "admin",
			},
			repoId:     r2,
			statusCode: http.StatusForbidden,
		},
		{
			name: "Redudant access is not allowed",
			body: map[string]any{
				"user_id":       m1,
				"access_level":  "Developer",
				"resource_type": "Repository",
				"resource_id":   r1,
				"granted_by":    "admin",
			},
			repoId:     r1,
			statusCode: http.StatusConflict,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryUsers, tc.repoId)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (r *RepositorySuite) testRevokeAccess(t *testing.T) {
	// Prepare data
	m1 := r.seeder.ProvisionUser(t, "revokerepomaintainer1", "revokerepomaintainer1@t.com", "Maintainer")
	require.NotEmpty(t, m1)

	n1 := r.seeder.CreateNamespace(t, "reporevokens1", "reporevokens1", "Team", false, m1)
	require.NotEmpty(t, n1)

	r1 := r.seeder.CreateRepository(t, "reporevotest1", "", "admin", n1, false)
	require.NotEmpty(t, r1)

	d1 := r.seeder.ProvisionUser(t, "repo-grantdev1", "repo-grantdev1@t.com", "Developer")
	require.NotEmpty(t, d1)
	d2 := r.seeder.ProvisionUser(t, "repo-grantdev2", "repo-grantdev2@t.com", "Developer")
	require.NotEmpty(t, d2)

	r.seeder.GrantAccess(t, r1, "Repository", d1, "Developer")
	r.seeder.GrantAccess(t, r1, "Repository", d2, "Guest")

	tcs := []struct {
		name       string
		resourceId string
		userId     string
		statusCode int
	}{
		{
			name:       "Successfully revoke access",
			userId:     d1,
			resourceId: r1,
			statusCode: http.StatusOK,
		},
		{
			name:       "Revoke access from non-existent user",
			userId:     "non-existent-user",
			resourceId: r1,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Revoke access from invalid repository id",
			userId:     d2,
			resourceId: "non-existent-id",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			url := r.testBaseURL + fmt.Sprintf(testdata.EndpointRepositoryUsers, tc.resourceId) + "/" + tc.userId
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(req, r.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}
