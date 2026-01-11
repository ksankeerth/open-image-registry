package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/integration/seeder"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Run("LoginAndLogout", a.testAuthFlowSuccess)
}

func (a *AuthTestSuite) Name() string {
	return a.name
}

func (a *AuthTestSuite) APIVersion() string {
	return a.apiVersion
}

func (a *AuthTestSuite) testAuthFlowSuccess(t *testing.T) {
	// Setup: Create a real user to log in with
	username := "auth-flow-user"
	password := "SecurePass123!"
	a.seeder.ProvisionUserWithPassword(t, username, "authflow@t.com", "Admin", password)

	targetURL := a.testBaseURL + fmt.Sprintf(testdata.EndpointNamespaceByID, "non-existent-ns")
	loginURL := a.testBaseURL + testdata.EndpointLogin
	logoutURL := a.testBaseURL + testdata.EndpointLogout

	t.Run("Full Auth Lifecycle", func(t *testing.T) {
		// 1. Try to access resource without token -> Expect 401 Unauthorized
		req1, _ := http.NewRequest(http.MethodGet, targetURL, nil)
		resp1, err := http.DefaultClient.Do(req1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp1.StatusCode, "Should be unauthorized initially")
		resp1.Body.Close()

		// 2. Login to get the session cookie
		loginBody, _ := json.Marshal(map[string]string{
			"username": username,
			"password": password,
		})
		reqLogin, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(loginBody))
		reqLogin.Header.Set("Content-Type", "application/json")

		respLogin, err := http.DefaultClient.Do(reqLogin)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, respLogin.StatusCode)

		// Capture cookies from the login response
		cookies := respLogin.Cookies()
		require.NotEmpty(t, cookies, "Login should return session cookies")
		respLogin.Body.Close()

		// 3. Access same resource with the session cookie -> Expect 404 Not Found (Authorized but missing)
		req2, _ := http.NewRequest(http.MethodGet, targetURL, nil)
		for _, cookie := range cookies {
			req2.AddCookie(cookie)
		}

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp2.StatusCode, "Should be authorized (404) after login")
		resp2.Body.Close()

		// 4. Logout with the session cookie -> Expect 200 OK
		reqLogout, _ := http.NewRequest(http.MethodPost, logoutURL, nil)
		for _, cookie := range cookies {
			reqLogout.AddCookie(cookie)
		}

		respLogout, err := http.DefaultClient.Do(reqLogout)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, respLogout.StatusCode)
		respLogout.Body.Close()

		// 5. Access same resource again -> Expect 401 Unauthorized (Session revoked)
		req3, _ := http.NewRequest(http.MethodGet, targetURL, nil)
		// Even if we send the old cookies, the server-side session should be gone
		for _, cookie := range cookies {
			req3.AddCookie(cookie)
		}

		resp3, err := http.DefaultClient.Do(req3)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp3.StatusCode, "Should be unauthorized after logout")
		resp3.Body.Close()
	})
}