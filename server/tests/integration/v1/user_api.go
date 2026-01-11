package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/integration/seeder"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserTestSuite struct {
	apiVersion  string
	name        string
	seeder      *seeder.TestDataSeeder
	testBaseURL string
}

func NewUserTestSuite(seeder *seeder.TestDataSeeder, baseURL string) *UserTestSuite {
	return &UserTestSuite{
		apiVersion:  "v1",
		name:        "User API",
		seeder:      seeder,
		testBaseURL: baseURL,
	}
}

func (u *UserTestSuite) Name() string {
	return u.name
}

func (u *UserTestSuite) APIVersion() string {
	return u.apiVersion
}

func (u *UserTestSuite) Run(t *testing.T) {
	t.Run("CreateUser_Success", u.testCreateUserSuccess)
	t.Run("CreateUser_Validation", u.testCreateUserValidationErrors)
	t.Run("CreateUser_Conflicts", u.testCreateUserConflicts)

	t.Run("CheckAvailability_Success", u.testCheckAvailabilitySuccess)
	t.Run("CheckAvailability_Validation", u.testCheckAvailabilityValidation)

	t.Run("ValidatePassword_Success", u.testValidatePassword)

	t.Run("UpdateUser_Success", u.testUpdateUserSuccess)
	t.Run("UpdateUser_Validation", u.testUpdateUserValidationErrors)

	t.Run("AccountSetup_Flow", u.testAccountSetupFlow)
	t.Run("RoleChange_Constraints", u.testRoleChangeConstraints)
	t.Run("LockUnlock_Flow", u.testLockUnlockFlow)
	t.Run("DeleteUser_Constraints", u.testDeleteUserConstraints)

	t.Run("Generic_ContentType", u.testGenericContentType)
}

func (u *UserTestSuite) testCreateUserSuccess(t *testing.T) {
	tcs := []struct {
		name       string
		body       map[string]any
		statusCode int
	}{
		{"All fields", map[string]any{"username": "test001", "email": "test0001@test.com", "display_name": "test0001", "role": "Admin"}, http.StatusCreated},
		{"Minimal fields", map[string]any{"username": "test002", "email": "test0002@test.com", "role": "Admin"}, http.StatusCreated},
		{"Maintainer Role", map[string]any{"username": "test003", "email": "test0003@test.com", "role": "Maintainer"}, http.StatusCreated},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)

			// 1. Create User Request
			req, err := http.NewRequest(http.MethodPost, u.testBaseURL+testdata.EndpointUsers, bytes.NewReader(b))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.statusCode, resp.StatusCode)

			// Verify the account setup UUID is present
			recoveryUUID := resp.Header.Get(testdata.HeaderAccountSetupUUID)
			assert.NotEmpty(t, recoveryUUID)

			// 2. Setup Info Request (Public endpoint, but keeping pattern consistent)
			verifyURL := u.testBaseURL + fmt.Sprintf(testdata.EndpointAccountSetupInfo, recoveryUUID)
			verifyReq, err := http.NewRequest(http.MethodGet, verifyURL, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(verifyReq, u.seeder.AdminToken(t))

			verifyResp, err := http.DefaultClient.Do(verifyReq)
			require.NoError(t, err)
			defer verifyResp.Body.Close()

			assert.Equal(t, http.StatusOK, verifyResp.StatusCode)

			var resMap map[string]any
			json.NewDecoder(verifyResp.Body).Decode(&resMap)
			assert.Equal(t, tc.body["username"], resMap["username"])
		})
	}
}

func (u *UserTestSuite) testCreateUserValidationErrors(t *testing.T) {
	tcs := []struct {
		name string
		body any
	}{
		{"Invalid JSON", "{invalid"},
		{"Missing Email", map[string]any{"username": "u1", "role": "Admin"}},
		{"Invalid Email", map[string]any{"username": "u2", "email": "not-email", "role": "Admin"}},
		{"Invalid Role", map[string]any{"username": "u3", "email": "t@t.com", "role": "GodMode"}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req, err := http.NewRequest(http.MethodPost, u.testBaseURL+testdata.EndpointUsers, &buf)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func (u *UserTestSuite) testCreateUserConflicts(t *testing.T) {
	// Seed initial user
	u.seeder.ProvisionUser(t, "test009", "test0009@test.com", "Admin")

	tcs := []struct {
		name string
		body map[string]any
	}{
		{"Same Username", map[string]any{"username": "test009", "email": "diff@test.com", "role": "Admin"}},
		{"Same Email", map[string]any{"username": "diff", "email": "test0009@test.com", "role": "Admin"}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)

			req, err := http.NewRequest(http.MethodPost, u.testBaseURL+testdata.EndpointUsers, bytes.NewReader(b))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusConflict, resp.StatusCode)
		})
	}
}

func (u *UserTestSuite) testUpdateUserSuccess(t *testing.T) {
	// Prepare data
	username := "updatetest"
	userID := u.seeder.ProvisionUser(t, username, "update@test.com", "Admin")

	// Perform Update
	newName := "Updated Display Name"
	updateBody, _ := json.Marshal(map[string]string{"display_name": newName})

	userURL := u.testBaseURL + fmt.Sprintf(testdata.EndpointUserByID, userID)
	req, _ := http.NewRequest(http.MethodPut, userURL, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify
	getReq, _ := http.NewRequest(http.MethodGet, userURL, nil)
	helpers.SetAuthCookie(getReq, u.seeder.AdminToken(t))

	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var res map[string]any
	json.NewDecoder(getResp.Body).Decode(&res)
	assert.Equal(t, newName, res["display_name"])
}

func (u *UserTestSuite) testUpdateUserValidationErrors(t *testing.T) {
	userURL := u.testBaseURL + fmt.Sprintf(testdata.EndpointUserByID, "123")
	tcs := []struct {
		name string
		body string
	}{
		{"Malformed JSON", `{"name": "broken}`},
		{"Name too long", fmt.Sprintf(`{"display_name": "%s"}`, strings.Repeat("a", 256))},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPut, userURL, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func (u *UserTestSuite) testCheckAvailabilitySuccess(t *testing.T) {
	// Seed a user to check against
	username := "availability_check"
	email := "avail@test.com"
	u.seeder.ProvisionUser(t, username, email, "Admin")

	tcs := []struct {
		name     string
		username string
		email    string
		uAvail   bool
		eAvail   bool
	}{
		{"Both Available", "newuser", "new@test.com", true, true},
		{"Username Taken", username, "new@test.com", false, true},
		{"Email Taken", "newuser", email, true, false},
		{"Both Taken", username, email, false, false},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(map[string]string{
				"username": tc.username,
				"email":    tc.email,
			})

			req, _ := http.NewRequest(http.MethodPost, u.testBaseURL+testdata.EndpointValidateUser, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			var res map[string]any
			json.NewDecoder(resp.Body).Decode(&res)
			assert.Equal(t, tc.uAvail, res["username_available"])
			assert.Equal(t, tc.eAvail, res["email_available"])
		})
	}
}

func (u *UserTestSuite) testCheckAvailabilityValidation(t *testing.T) {
	tcs := []struct {
		name       string
		body       map[string]any
		statusCode int
	}{
		{name: "Invalid body", body: map[string]any{"invalid_key": "data"}, statusCode: http.StatusBadRequest},
		{name: "Invalid username", body: map[string]any{"username": "&testusername", "email": ""}, statusCode: http.StatusBadRequest},
		{name: "Invalid email", body: map[string]any{"username": "", "email": "not-an-email"}, statusCode: http.StatusBadRequest},
		{name: "Invalid username and email", body: map[string]any{"username": "*testusername", "email": "not-valid-email"}, statusCode: http.StatusBadRequest},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, u.testBaseURL+testdata.EndpointValidateUser, bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			helpers.AssertStatusCode(t, resp, tc.statusCode)
		})
	}
}

func (u *UserTestSuite) testValidatePassword(t *testing.T) {
	tcs := []struct {
		name  string
		pass  string
		valid bool
	}{
		{"Valid Password", "ValidPassword123!", true},
		{"Too Short", "sh1!", false},
		{"No Digits", "NoDigitsHere!", false},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(map[string]string{"password": tc.pass})

			req, err := http.NewRequest(http.MethodPost, u.testBaseURL+testdata.EndpointValidatePassword, bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(req, u.seeder.AdminToken(t))

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			helpers.AssertStatusCode(t, resp, http.StatusOK)

			var res map[string]any
			json.NewDecoder(resp.Body).Decode(&res)
			assert.Equal(t, tc.valid, res["is_valid"])
			if !tc.valid {
				assert.NotEmpty(t, res["msg"])
			}
		})
	}
}

func (u *UserTestSuite) testAccountSetupFlow(t *testing.T) {
	// Prepare data
	u1, r1 := u.seeder.CreateUser(t, "accountsetup01", "accountsetup01@t.com", "Guest")
	require.NotEmpty(t, r1)
	u.seeder.SetAccountRecoveryReason(t, r1, constants.ReasonPasswordRecoveryForgotPassowrd)

	u2, r2 := u.seeder.CreateUser(t, "accountsetup02", "accountsetup02@t.com", "Maintainer")
	require.NotEmpty(t, r2)

	u3, r3 := u.seeder.CreateUser(t, "accountsetup03", "accountsetup03@t.com", "Admin")
	require.NotEmpty(t, r3)

	displayName := "newly updated displayname"

	tcs := []struct {
		name           string
		infoStatus     int
		completeStatus int
		error_msg      string
		displayName    string
		username       string
		password       string
		recoveryID     string
		userID         string
	}{
		{
			name:       "Invalid Recovery ID or Link",
			recoveryID: "d04f0f6b60f2a3fe",
			userID:     "",
			infoStatus: http.StatusNotFound,
			error_msg:  "This account setup link is no longer valid. It may have already been used..",
		},
		{
			name:       "Invalid Account Recovery Reason",
			recoveryID: r1,
			userID:     u1,
			username:   "accountsetup01",
			infoStatus: http.StatusNotFound,
			error_msg:  "This link is not valid. Please check it again ...",
		},
		{
			name:           "Account setup with invalid password",
			recoveryID:     r2,
			userID:         u2,
			username:       "accountsetup02",
			password:       "short",
			infoStatus:     http.StatusOK,
			completeStatus: http.StatusBadRequest,
		},
		{
			name:           "Account setup with valid password and display name",
			recoveryID:     r3,
			userID:         u3,
			username:       "accountsetup03",
			infoStatus:     http.StatusOK,
			completeStatus: http.StatusOK,
			displayName:    displayName,
			password:       "Valid13%stssrw",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// 1. GET Account Info
			infoURL := u.testBaseURL + fmt.Sprintf(testdata.EndpointAccountSetupInfo, tc.recoveryID)
			infoReq, err := http.NewRequest(http.MethodGet, infoURL, nil)
			require.NoError(t, err)
			helpers.SetAuthCookie(infoReq, u.seeder.AdminToken(t))

			infoResp, err := http.DefaultClient.Do(infoReq)
			require.NoError(t, err)
			defer infoResp.Body.Close()

			require.Equal(t, tc.infoStatus, infoResp.StatusCode, "Account Setup Info status mismatch")

			if tc.infoStatus != http.StatusOK {
				var errRes map[string]any
				err := json.NewDecoder(infoResp.Body).Decode(&errRes)
				require.NoError(t, err)
				if tc.error_msg != "" {
					require.Equal(t, tc.error_msg, errRes["error_message"].(string))
				}
				return
			}

			// 2. POST Complete Setup
			setupBody, _ := json.Marshal(map[string]string{
				"user_id":      tc.userID,
				"uuid":         tc.recoveryID,
				"username":     tc.username,
				"password":     tc.password,
				"display_name": tc.displayName,
			})

			completeURL := u.testBaseURL + fmt.Sprintf(testdata.EndpointAccountSetupComplete, tc.recoveryID)
			setupReq, err := http.NewRequest(http.MethodPost, completeURL, bytes.NewReader(setupBody))
			require.NoError(t, err)
			setupReq.Header.Set("Content-Type", "application/json")
			helpers.SetAuthCookie(setupReq, u.seeder.AdminToken(t))

			setupResp, err := http.DefaultClient.Do(setupReq)
			require.NoError(t, err)
			defer setupResp.Body.Close()

			require.Equal(t, tc.completeStatus, setupResp.StatusCode, "Account Setup Complete: Status mismatch")
		})
	}
}

// TODO
func (u *UserTestSuite) testRoleChangeConstraints(t *testing.T) {
	// 1. Seed a user
	// 2. Attempt role changes via testdata.EndpointUserChangeRole
}

// TODO
func (u *UserTestSuite) testLockUnlockFlow(t *testing.T) {
	// Verify admin can lock/unlock and self-lock is forbidden
}

// TODO
func (u *UserTestSuite) testDeleteUserConstraints(t *testing.T) {
	// Test deletion logic and resource-access blockers
}

func (u *UserTestSuite) testGenericContentType(t *testing.T) {
	endpoints := []string{
		u.testBaseURL + testdata.EndpointUsers,
		u.testBaseURL + testdata.EndpointValidateUser,
		u.testBaseURL + testdata.EndpointValidatePassword,
	}

	for _, ep := range endpoints {
		req, _ := http.NewRequest(http.MethodPost, ep, strings.NewReader("<xml></xml>"))
		req.Header.Set("Content-Type", "application/xml")
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	}
}