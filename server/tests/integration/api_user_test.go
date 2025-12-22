package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_Success(t *testing.T) {
	tcs := []struct {
		name       string
		body       map[string]any
		statusCode int
	}{
		{
			name:       "With username, email, role and display name",
			body:       map[string]any{"username": "test001", "email": "test0001@test.com", "display_name": "test0001", "role": "Admin"},
			statusCode: http.StatusCreated,
		},
		{
			name:       "With username, email and role only",
			body:       map[string]any{"username": "test002", "email": "test0002@test.com", "role": "Admin"},
			statusCode: http.StatusCreated,
		},
		{
			name:       "With role Maintainer",
			body:       map[string]any{"username": "test003", "email": "test0003@test.com", "display_name": "test0003", "role": "Maintainer"},
			statusCode: http.StatusCreated,
		},
		{
			name:       "With role Developer",
			body:       map[string]any{"username": "test004", "email": "test0004@test.com", "display_name": "test0004", "role": "Maintainer"},
			statusCode: http.StatusCreated,
		},
		{
			name:       "With role guest",
			body:       map[string]any{"username": "test005", "email": "test0005@test.com", "display_name": "test0005", "role": "Maintainer"},
			statusCode: http.StatusCreated,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUsers),
				bytes.NewReader(body))
			require.NoError(t, err)

			req.Header.Set("Content-Type", testdata.ApplicationJson)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			helpers.AssertStatusCode(t, resp, tc.statusCode)

			recoveryUUID := resp.Header.Get(testdata.HeaderAccountSetupUUID)
			require.NotEmpty(t, recoveryUUID, "recovery uuid is expected when running tests in dev mode")

			verifyResp, err := http.Get(fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointAccountSetupInfo, recoveryUUID)))
			require.NoError(t, err)
			helpers.AssertStatusCode(t, verifyResp, http.StatusOK)

			var resMap map[string]any
			defer verifyResp.Body.Close()
			err = json.NewDecoder(verifyResp.Body).Decode(&resMap)
			require.NoError(t, err)

			// Verification of persisted details
			assert.Empty(t, resMap["error_message"], "Expected no error message in response")
			assert.Equal(t, tc.body["username"], resMap["username"])
			assert.Equal(t, tc.body["email"], resMap["email"])
			assert.Equal(t, tc.body["role"], resMap["role"])

			if displayName, ok := tc.body["display_name"]; ok {
				assert.Equal(t, displayName, resMap["display_name"])
			} else {
				// If display_name wasn't provided, it usually defaults to the username or empty
				assert.NotEmpty(t, resMap["display_name"])
			}

			assert.NotEmpty(t, resMap["user_id"], "User ID should be populated")

		})
	}

	// TODO: cleanup may be needed
}

func TestCreateUser_ValidationErrors(t *testing.T) {
	tcs := []struct {
		name       string
		body       any
		statusCode int
	}{
		{
			name:       "With invalid JSON payload",
			body:       "{invalid-json",
			statusCode: http.StatusBadRequest,
		},
		{
			name: "With email missing",
			body: map[string]any{
				"username":     "testuser",
				"display_name": "Test User",
				"role":         "Admin",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "With username missing",
			body: map[string]any{
				"email":        "test@example.com",
				"display_name": "Test User",
				"role":         "Admin",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "With role missing",
			body: map[string]any{
				"username":     "testuser",
				"email":        "test@example.com",
				"display_name": "Test User",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "With invalid email format",
			body: map[string]any{
				"username": "testuser",
				"email":    "not-an-email",
				"role":     "Admin",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "With invalid username (too short/special chars)",
			body: map[string]any{
				"username": "a",
				"email":    "test@example.com",
				"role":     "Admin",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "With invalid role name",
			body: map[string]any{
				"username": "testuser",
				"email":    "test@example.com",
				"role":     "SuperUserWhichDoesNotExist",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				err := json.NewEncoder(&buf).Encode(tc.body)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUsers), &buf)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestCreateUser_Conflicts(t *testing.T) {
	// Create the original user to
	originalUser := map[string]any{
		"username":     "test009",
		"email":        "test0009@test.com",
		"display_name": "test0009",
		"role":         "Admin",
	}

	body, _ := json.Marshal(originalUser)
	setupReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUsers), bytes.NewReader(body))
	setupReq.Header.Set("Content-Type", "application/json")
	setupResp, err := http.DefaultClient.Do(setupReq)
	require.NoError(t, err)
	helpers.AssertStatusCode(t, setupResp, http.StatusCreated)

	// test scenarios
	tcs := []struct {
		name       string
		body       map[string]any
		statusCode int
	}{
		{
			name: "With same username",
			body: map[string]any{
				"username":     "test009", // Conflict
				"email":        "different@test.com",
				"display_name": "diff",
				"role":         "Admin",
			},
			statusCode: http.StatusConflict,
		},
		{
			name: "With same email",
			body: map[string]any{
				"username":     "different_user",
				"email":        "test0009@test.com", // Conflict
				"display_name": "diff",
				"role":         "Admin",
			},
			statusCode: http.StatusConflict,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUsers), bytes.NewReader(b))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestCreateUser_HTTPMethods(t *testing.T) {
	tcs := []struct {
		name       string
		method     string
		statusCode int
	}{
		{
			name:       "With HTTP PUT method",
			method:     http.MethodPut,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With HTTP PATCH method",
			method:     http.MethodPatch,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, testBaseURL+testdata.EndpointUsers, nil)
			require.NoError(t, err)
			require.NotNil(t, req)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestCreateUser_ContentType(t *testing.T) {
	tcs := []struct {
		name        string
		contentType string
		body        string
		statusCode  int
	}{
		{
			name:        "With application/xml",
			contentType: "application/xml",
			body: `<?xml version="1.0" encoding="UTF-8"?>
<user>
  <username>john.doe</username>
  <email>john.doe@example.com</email>
  <role>Admin</role>
</user>`,
			statusCode: http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUsers), strings.NewReader(tc.body))
			require.NoError(t, err)
			require.NotNil(t, req)

			req.Header.Set("Content-Type", tc.contentType)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NotNil(t, resp)

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestUpdateUser_Success(t *testing.T) {
	// prepare data to test update user account
	userBody := map[string]any{
		"username":     "updatetest",
		"email":        "update@test.com",
		"role":         "Admin",
		"display_name": "Original Name",
	}
	b, _ := json.Marshal(userBody)
	resp, err := http.Post(fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUsers), "application/json",
		bytes.NewReader(b))
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var resBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&resBody)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	userId := resBody["user_id"]
	require.NotEmpty(t, userId)

	// update displayname
	newName := "Updated Display Name"
	updateBody := mgmt.UpdateUserAccountRequest{DisplayName: newName}
	ub, _ := json.Marshal(updateBody)

	userURL := fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointUserByID, userId))
	req, _ := http.NewRequest(http.MethodPut, userURL, bytes.NewReader(ub))
	updateResp, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	// verify by calling get user
	userResp, err := http.Get(userURL)
	require.NoError(t, err)
	require.NotNil(t, userResp)

	require.Equal(t, http.StatusOK, userResp.StatusCode)

	var userResBody map[string]any
	err = json.NewDecoder(userResp.Body).Decode(&userResBody)
	defer userResp.Body.Close()

	require.NoError(t, err)

	assert.Equal(t, newName, userResBody["display_name"])
	assert.Equal(t, userResBody["username"], userBody["username"])
}

func TestUpdateUser_ValidationErrors(t *testing.T) {
	tcs := []struct {
		name       string
		body       string
		statusCode int
	}{
		{
			name:       "Invalid JSON body",
			body:       `{"display_name": "missing-quote}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Missing body",
			body:       "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Large displayname chars > 255",
			body:       fmt.Sprintf(`{"display_name": "%s"}`, strings.Repeat("a", 256)),
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Use a dummy ID for validation tests
			userURL := fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointUserByID, "123"))
			req, err := http.NewRequest(http.MethodPut, userURL, strings.NewReader(tc.body))
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestUpdateUser_HTTPMethods(t *testing.T) {
	tcs := []struct {
		name       string
		method     string
		statusCode int
	}{
		{
			name:       "With POST method",
			method:     http.MethodPost,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With PATCH method",
			method:     http.MethodPatch,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			userURL := fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointUserByID, "123"))
			req, err := http.NewRequest(tc.method, userURL, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestUpdateUser_ContentType(t *testing.T) {
	tcs := []struct {
		name        string
		contentType string
		body        string
		statusCode  int
	}{
		{
			name:        "With application/xml",
			contentType: "application/xml",
			body:        `<user><name>John</name></user>`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "With text/plain",
			contentType: "text/plain",
			body:        `display_name=John`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Content-Type checks usually happen before DB lookups, so a dummy ID is fine
			userURL := fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointUserByID, "123"))

			// Note: Changed to MethodPut to match your handler's intended verb
			req, err := http.NewRequest(http.MethodPut, userURL, strings.NewReader(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.contentType)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestCheckEmailUsernameAvailablity_Success(t *testing.T) {
	// prepare data to test
	userBody := map[string]any{
		"username": "emailusernameavailability",
		"email":    "emailusernameavailability@test.com",
		"role":     "Admin",
	}
	b, _ := json.Marshal(userBody)
	resp, err := http.Post(fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUsers), "application/json",
		bytes.NewReader(b))
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	tcs := []struct {
		name              string
		email             string
		username          string
		statusCode        int
		usernameAvailable bool
		emailAvailable    bool
	}{
		{
			name:              "With available username",
			username:          "avail-username-100",
			email:             "",
			statusCode:        http.StatusOK,
			usernameAvailable: true,
		},
		{
			name:           "With available email",
			email:          "avail-email@test.com",
			username:       "",
			emailAvailable: true,
			statusCode:     http.StatusOK,
		},
		{
			name:              "With available username and email",
			username:          "avail-username-100",
			email:             "avail-email@test.com",
			statusCode:        http.StatusOK,
			usernameAvailable: true,
			emailAvailable:    true,
		},
		{
			name:              "With unavailable username and email",
			username:          "emailusernameavailability",
			email:             "emailusernameavailability@test.com",
			statusCode:        http.StatusOK,
			usernameAvailable: false,
			emailAvailable:    false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]any{
				"username": tc.username,
				"email":    tc.email,
			}

			reqBytes, err := json.Marshal(reqBody)
			require.NoError(t, err)
			require.NotNil(t, reqBytes)

			resp, err := http.Post(fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidateUser), testdata.ApplicationJson,
				bytes.NewReader(reqBytes))
			require.NoError(t, err)
			require.NotNil(t, resp)
			helpers.AssertStatusCode(t, resp, tc.statusCode)

			var respBody map[string]any
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.usernameAvailable, respBody["username_available"])
			assert.Equal(t, tc.emailAvailable, respBody["email_available"])
		})
	}
}

func TestCheckEmailUsernameAvailablity_ValidationErrors(t *testing.T) {
	tcs := []struct {
		name       string
		body       map[string]any
		statusCode int
	}{
		{
			name: "Invalid Request body",
			body: map[string]any{
				"invalid": "invalid",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBytes, _ := json.Marshal(tc.body)

			resp, err := http.Post(fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidateUser),
				testdata.ApplicationJson, bytes.NewReader(reqBytes))
			require.NoError(t, err)
			defer resp.Body.Close()
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestCheckEmailUsernameAvailablity_HTTPMethods(t *testing.T) {
	t.Skip() // TODO
	//TODO: https://github.com/ksankeerth/open-image-registry/issues/33
	// After fixing #33, We need to enable this test
	tcs := []struct {
		name       string
		method     string
		statusCode int
	}{
		{
			name:       "With HTTP PUT method",
			method:     http.MethodPut,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With HTTP PATCH method",
			method:     http.MethodPatch,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With HTTP DELETE method",
			method:     http.MethodDelete,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With HTTP GET method",
			method:     http.MethodGet,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidateUser), nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestCheckEmailUsernameAvailablity_ContentType(t *testing.T) {
	tcs := []struct {
		name        string
		contentType string
		body        string
		statusCode  int
	}{
		{
			name:        "With application/xml",
			contentType: "application/xml",
			body:        `<request><email>test@example.com</email></request>`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "With text/plain",
			contentType: "text/plain",
			body:        `username=testuser`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "With multipart/form-data",
			contentType: "multipart/form-data; boundary=something",
			body:        `--something...`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidateUser), strings.NewReader(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.contentType)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestValidatePassword_Success(t *testing.T) {
	url := fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidatePassword)
	request := map[string]any{
		"password": "Test123#wesabse",
	}

	requestBytes, err := json.Marshal(request)
	require.NoError(t, err)
	require.NotNil(t, requestBytes)

	resp, err := http.Post(url, testdata.ApplicationJson, bytes.NewReader(requestBytes))
	require.NoError(t, err)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	defer resp.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, true, respBody["is_valid"])
	assert.Empty(t, respBody["msg"])
}

func TestValidatePassword_ValidationErrors(t *testing.T) {
	tcs := []struct {
		name           string
		body           string
		expectedStatus int
		expectedValid  bool
		checkMsg       bool // whether to check for a specific error message
	}{
		{
			name:           "Invalid body (Malformed JSON)",
			body:           `{"password": "validPassword123!`, // Missing closing quote/brace
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid password format (Too short)",
			body:           `{"password": "Short1!"}`,
			expectedStatus: http.StatusOK, // The API worked, but the password failed validation
			expectedValid:  false,
			checkMsg:       true,
		},
		{
			name:           "Invalid password format (No digits)",
			body:           `{"password": "NoDigitsAllowed!"}`,
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			checkMsg:       true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidatePassword)
			resp, err := http.Post(url, testdata.ApplicationJson, strings.NewReader(tc.body))
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedStatus == http.StatusOK {
				var respBody map[string]any
				err = json.NewDecoder(resp.Body).Decode(&respBody)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedValid, respBody["is_valid"])
				if tc.checkMsg {
					assert.NotEmpty(t, respBody["msg"])
				}
			}
		})
	}
}

func TestValidatePassword_HTTPMethods(t *testing.T) {
	t.Skip() // TODO
	//TODO: https://github.com/ksankeerth/open-image-registry/issues/33
	// After fixing #33, We need to enable this test
	tcs := []struct {
		name       string
		method     string
		statusCode int
	}{
		{name: "With HTTP PUT method", method: http.MethodPut, statusCode: http.StatusMethodNotAllowed},
		{name: "With HTTP PATCH method", method: http.MethodPatch, statusCode: http.StatusMethodNotAllowed},
		{name: "With HTTP GET method", method: http.MethodGet, statusCode: http.StatusMethodNotAllowed},
		{name: "With HTTP DELETE method", method: http.MethodDelete, statusCode: http.StatusMethodNotAllowed},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidatePassword)
			req, err := http.NewRequest(tc.method, url, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestValidatePassword_ContentType(t *testing.T) {
	tcs := []struct {
		name        string
		contentType string
		body        string
		statusCode  int
	}{
		{
			name:        "With application/xml",
			contentType: "application/xml",
			body:        `<password>Secret123!</password>`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "With text/plain",
			contentType: "text/plain",
			body:        `password=Secret123!`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointValidatePassword)
			req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.contentType)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

// TODO: UpdateEmail, this is not needed at the moment
// We'll focus on this later. Update email handling, notification, and ways to
// verify the mail require careful thinking and an enterprise feature.
// sometimes, use of external IAM provide can solve this , so let's focus on this later
func TestUpdateEmail_Success(t *testing.T) {
	// 1. valid email & receive notification
	// 2. changing email of lock user account should keep the account in same state
	// 3. if the account is locked, only admin users should be able to change the email
	// 4. if the email is same as current, just pass without errors
}

func TestUpdateEmail_ValidationErrors(t *testing.T) {
	// 1. invalid email format
	// 2. invalid request or missing request body
	// 3. invalid user id => 404
	// 4. if account is locked, email change request from non-admin user should be rejected with 403
}

func TestUpdateEmail_Conflicts(t *testing.T) {
	// 1. email is used by another user
}

func TestChangeRole_Success(t *testing.T) {
	// 1. Admin will be able to change role of any users if it is a role promotion
	// 2. if the role is same, just pass without doing modification in db
	// 3. If it is a role demotion,  we have to check some conditions

}

func TestChangeRole_ValidationErrors(t *testing.T) {
	// 1. Should not allow changing role with lesser access if user holds some resource access with current level. => 403
	// 2. Role changes are not allowed to locked accounts
	// 3. missing role query param
	// 4. invalid user id => 404
}

func TestChangeRole_HTTPMethods(t *testing.T) {
	t.Skip() // TODO
	//TODO: https://github.com/ksankeerth/open-image-registry/issues/33
	// After fixing #33, We need to enable this test
	tcs := []struct {
		name       string
		method     string
		statusCode int
	}{
		{
			name:       "With HTTP PUT method",
			method:     http.MethodPut,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With HTTP POST method",
			method:     http.MethodPost,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, fmt.Sprintf("%s%s", testBaseURL, testdata.EndpointUserChangeRole), nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

// TODO: We can remove PUT /api/v1/users/{id}/display-name and keep this handled through PUT /api/v1/users/{id}
// therefore no tests cases for this resource

func TestDeleteUser_Success(t *testing.T) {
	// 1. delete a maintainer who don't have any access to any resources by admin
	// 2. delete a developer who don't have any access to any resources by admin
	// 3. delete a guest who don't have any access to any resources by admin
}

func TestDeleteUser_ValidationErrors(t *testing.T) {
	// 1. self delete is not allowed
	// 2. admin accounts cannot be deleted only allowed to locked
	// 3. deleting a maintainer who has access to a namespace
	// 4. deleting a developer who has access to a namespace and repository
	// 5. deleting a guest who has access to a repository
	// 6. deleting a guest who has access to upstream registry
	// 7. deleting a non-existent user
}

func TestPasswordChangeRequest_Success(t *testing.T) {
	// 1. request to change password of an existing user by him self
}

func TestPasswordChangeRequest_ValidationErrors(t *testing.T) {
	// 1. admin cannot change password
	// 2. invalid user id => 404
	// 3. recover id is invalid => 404
	// 4. old password is not valid => 403
	// 5. new password doesn't match with policy => 403
	// 6. invalid body => 400
	// 7. missing recovery id => 400
}

func TestPasswordChangeRequest_ContentType(t *testing.T) {
	tcs := []struct {
		name        string
		contentType string
		body        string
		statusCode  int
	}{
		{
			name:        "With application/xml",
			contentType: "application/xml",
			body:        `<request><email>test@example.com</email></request>`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointUserPassword, "dummy-uuid")), strings.NewReader(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.contentType)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestPasswordChangeRequest_HTTPMethods(t *testing.T) {
	tcs := []struct {
		name       string
		method     string
		statusCode int
	}{
		{
			name:       "With HTTP POST method",
			method:     http.MethodPost,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With HTTP PATCH method",
			method:     http.MethodPatch,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointUserPassword, "dummy-uuid")), nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}
func TestLockUser_Success(t *testing.T) {
	// 1. admin can lock guest
	// 2. admin can lock developer
	// 3. admin can lock maintainer
	// 4. admin can lock admin
	// 5. just pass if the account is already locked
}

func TestLockUser_ValidationErrors(t *testing.T) {
	// 1. maintainer cannot lock
	// 2. developer cannot lock
	// 3. guest cannot lock
	// 4. admin cannot do self-lock
	// 5. invalid user-id => 404
}

func TestUnlockUser_Success(t *testing.T) {
	// 1. Admin Unlock user accounts which were not deleted
	// 2. Admin can unlock maintainer user account
	// 3. Admin can unlock developer user account
	// 4. Admin can unlock guest user account
}

func TestUnlockUser_ValidationErrors(t *testing.T) {
	// 1. Self-unlock is not allowed
	// 2. non-admin users cannot unlock others user account
	// 3. deleted accounts cannot be unlocked
	// 4. Locked for initial verification and user invitation cannot be unlocked
}

func TestGetUserAccountSetupInfo_Success(t *testing.T) {
	// 1. Able to retrieve details using recovery Id
	// 2. no authentication required
}

func TestGetUserAccountSetupInfo_ValidationErrors(t *testing.T) {
	// 1. Not allowed to use an expired id => 403
	// 2. non-exitent ids = > 404
	// 3. if user is deleted => 404
}

func TestCompleteUserAccountSetup_Success(t *testing.T) {
	// 1. able to complete user account setup with new username
	// 2. able to complete user account setup with empty display name
	// 3. able to complete user account setup with non-empty display name
}

func TestCompleteUserAccountSetup_ValidationErrors(t *testing.T) {
	// 1. non-existent user-id => 403
	// 2. non-exitent recovery id => 403
	// 3. weak password => 403
	// 4. empty username => 400
	// 5. empty payload or missing body => 400
}

func TestCompleteUserAccountSetup_HTTPMethods(t *testing.T) {
	t.Skip() // TODO
	//TODO: https://github.com/ksankeerth/open-image-registry/issues/33
	// After fixing #33, We need to enable this tests
	tcs := []struct {
		name       string
		method     string
		statusCode int
	}{
		{
			name:       "With HTTP PATCH method",
			method:     http.MethodPatch,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "With HTTP PUT method",
			method:     http.MethodPut,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointAccountSetupComplete, "dummy-id")), nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func TestCompleteUserAccountSetup_ContentType(t *testing.T) {
	tcs := []struct {
		name        string
		contentType string
		body        string
		statusCode  int
	}{
		{
			name:        "With application/xml",
			contentType: "application/xml",
			body:        `<setup><token>123</token></setup>`,
			statusCode:  http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", testBaseURL, fmt.Sprintf(testdata.EndpointAccountSetupComplete, "dumm-uuid")), strings.NewReader(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.contentType)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}