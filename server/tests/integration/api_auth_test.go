package integration

import "testing"

func TestUserLogin_Success(t *testing.T) {
	// tcs
	// 1. New sessioin
	// 2. Renew session
	// 3. Expired session
}

func TestUserLogin_Failure(t *testing.T) {
	// tcs
	// 1. WithoutBody
	// 2. With GetMethod
	// 3. WithPutMethod
	// 4. WithPatchMethod
	// 5. With Invalid username
	// 6. With Invalid password
	// 7. With Invalid Content-Type
}

func TestUserLogin_AccountLock(t *testing.T) {
	// tcs
	// 1. Login with invalid credentials multiple times
	// 2. Login to already locked account
	// 3.
}