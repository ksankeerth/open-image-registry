package seeder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/require"
)

// ProvisionUser will create a test user account with given details if no user exists with given
// username. If a user is found with given username, then it'll check email, role and locked status
// if any mismatch found, it would fail the tests. Otherwise, It'll create new user acccount with given
// values.
func (s *TestDataSeeder) ProvisionUser(t *testing.T, username, email, role string) (userID string) {
	t.Helper()

	exists, mismatch, userID := s.checkUser(t, username, role, email, false)
	if exists && mismatch {
		require.Fail(t, "user exists with different values than given")
		return ""
	}

	if exists {
		return
	}

	payload := map[string]any{
		"username":     username,
		"email":        email,
		"role":         role,
		"display_name": username, // Defaulting display name to username
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err, "failed to marshal create user body")

	createURL := fmt.Sprintf("%s%s", s.baseURL, testdata.EndpointUsers)
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", testdata.ApplicationJson)
	token := s.AdminToken(t)
	helpers.SetAuthCookie(req, token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to execute create user request")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "unexpected status code on creation")

	var respBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err, "failed to parse create user response")
	userID, ok := respBody["user_id"].(string)
	require.True(t, ok, "failed to extract user id from create user response")
	require.NotEmpty(t, userID, "failed to extract user id from create user response")

	// Extract the Recovery/Setup UUID from headers
	recoveryUUID := resp.Header.Get(testdata.HeaderAccountSetupUUID)
	require.NotEmpty(t, recoveryUUID, "account setup UUID was not found in response headers")

	// Generate a random password for completion
	randomPassword := generateRandomString(12)

	// Complete Account Setup
	setupPayload := map[string]any{
		"user_id":      userID,
		"uuid":         recoveryUUID,
		"password":     randomPassword,
		"username":     username,
		"display_name": username,
	}
	setupBody, _ := json.Marshal(setupPayload)

	setupURL := fmt.Sprintf("%s%s", s.baseURL, fmt.Sprintf(testdata.EndpointAccountSetupComplete, recoveryUUID))
	setupReq, err := http.NewRequest(http.MethodPost, setupURL, bytes.NewReader(setupBody))
	require.NoError(t, err)

	helpers.SetAuthCookie(setupReq, token)

	setupReq.Header.Set("Content-Type", testdata.ApplicationJson)

	setupResp, err := http.DefaultClient.Do(setupReq)
	require.NoError(t, err, "failed to execute setup completion")

	defer setupResp.Body.Close()

	require.Equal(t, http.StatusOK, setupResp.StatusCode, "failed to complete account setup: status")

	return userID
}

// ProvisionUserWithPassword works like ProvisionUser but allows specifying a custom password.
func (s *TestDataSeeder) ProvisionUserWithPassword(t *testing.T, username, email, role, password string) (userID string) {
	t.Helper()

	exists, mismatch, userID := s.checkUser(t, username, role, email, false)
	if exists && mismatch {
		require.Fail(t, "user exists with different values than given")
		return ""
	}

	if exists {
		return userID
	}

	// 1. Create the user (Initial state)
	payload := map[string]any{
		"username":     username,
		"email":        email,
		"role":         role,
		"display_name": username,
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err, "failed to marshal create user body")

	createURL := fmt.Sprintf("%s%s", s.baseURL, testdata.EndpointUsers)
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", testdata.ApplicationJson)
	helpers.SetAuthCookie(req, s.AdminToken(t))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to execute create user request")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "unexpected status code on user creation")

	var respBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	userID = respBody["user_id"].(string)

	// 2. Extract Setup UUID from headers
	recoveryUUID := resp.Header.Get(testdata.HeaderAccountSetupUUID)
	require.NotEmpty(t, recoveryUUID, "account setup UUID missing from response headers")

	// 3. Complete Setup with the provided password
	setupPayload := map[string]any{
		"user_id":      userID,
		"uuid":         recoveryUUID,
		"password":     password,
		"username":     username,
		"display_name": username,
	}
	setupBody, _ := json.Marshal(setupPayload)

	setupURL := fmt.Sprintf("%s%s", s.baseURL, fmt.Sprintf(testdata.EndpointAccountSetupComplete, recoveryUUID))
	setupReq, err := http.NewRequest(http.MethodPost, setupURL, bytes.NewReader(setupBody))
	require.NoError(t, err)
	setupReq.Header.Set("Content-Type", testdata.ApplicationJson)

	setupResp, err := http.DefaultClient.Do(setupReq)
	require.NoError(t, err, "failed to execute setup completion")
	defer setupResp.Body.Close()

	require.Equal(t, http.StatusOK, setupResp.StatusCode, "failed to complete account setup with provided password")

	return userID
}

// CreateUser just creates a user account. It doesn't setup password and complete account setup.
func (s *TestDataSeeder) CreateUser(t *testing.T, username, email, role string) (userID, recoveryID string) {
	t.Helper()

	exists, mismatch, userID := s.checkUser(t, username, role, email, false)
	if exists && mismatch {
		require.Fail(t, "user exists with different values than given")
		return "", ""
	}

	if exists {
		return
	}

	payload := map[string]any{
		"username":     username,
		"email":        email,
		"role":         role,
		"display_name": username, // Defaulting display name to username
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err, "failed to marshal create user body")

	createURL := fmt.Sprintf("%s%s", s.baseURL, testdata.EndpointUsers)
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewReader(body))
	require.NoError(t, err)

	token := s.AdminToken(t)
	helpers.SetAuthCookie(req, token)

	req.Header.Set("Content-Type", testdata.ApplicationJson)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to execute create user request")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "unexpected status code on creation")

	var respBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err, "failed to parse create user response")
	userID, ok := respBody["user_id"].(string)
	require.True(t, ok, "failed to extract user id from create user response")
	require.NotEmpty(t, userID, "failed to extract user id from create user response")

	// Extract the Recovery/Setup UUID from headers
	recoveryUUID := resp.Header.Get(testdata.HeaderAccountSetupUUID)
	require.NotEmpty(t, recoveryUUID, "account setup UUID was not found in response headers")

	return userID, recoveryUUID
}

func (s *TestDataSeeder) SetAccountRecoveryReason(t *testing.T, recoveryID string, reason uint) {
	err := s.store.AccountRecovery().UpdateReason(context.Background(), recoveryID, reason)
	require.NoError(t, err)
}

func (s *TestDataSeeder) DeleteAllNonAdminUsers(t *testing.T) {
	t.Helper()

	err := s.store.Users().DeleteAllNonAdminAccounts(context.Background())
	require.NoError(t, err)
}

func (s *TestDataSeeder) checkUser(t *testing.T, identifier, role, email string,
	locked bool) (exists bool, mismatch bool, userID string) {
	t.Helper()

	url := s.baseURL + fmt.Sprintf(testdata.EndpointUserByID, identifier)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err, "failed to create get user request")

	helpers.SetAuthCookie(req, s.AdminToken(t))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "failed to get user account")
	require.NotNil(t, resp)
	defer resp.Body.Close()

	checkFurther := false

	switch resp.StatusCode {
	case http.StatusOK:
		checkFurther = true
	case http.StatusNotFound:
		return false, false, ""
	default:
		require.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)
	}

	if !checkFurther {
		return false, false, ""
	}

	var resBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&resBody)
	require.NoError(t, err, "failed to parse get user account response")

	if resBody["role"].(string) != role {
		return true, true, ""
	}

	// email is empty, we wouldn't validate email
	if email != "" && resBody["email"].(string) != email {
		return true, true, ""
	}

	if resBody["locked"].(bool) != locked {
		return true, true, ""
	}

	userID = resBody["id"].(string)
	return true, false, userID
}

func generateRandomString(n int) string {
	if n < 12 {
		n = 12
	}

	// Prefix has upper and lower case letters, number and symbol to pass validation
	prefix := "Aa1!"

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n-len(prefix))
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}

	return prefix + string(b)
}