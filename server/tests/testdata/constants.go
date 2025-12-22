package testdata

// Test User Credentials
const (
	DefaultAdmin         = "admin"
	DefaultAdminPassword = "Admin@123456"
	DefaultAdminEmail    = "admin@test.com"

	TestAdmin1         = "john_doe"
	TestAdmin1Password = "Admin123@we"
	TestAdmin1Email    = "john_doe@test.com"

	TestAdmin2         = "sarah_admin"
	TestAdmin2Password = "SecureAdmin#2023"
	TestAdmin2Email    = "sarah.admin@test.com"

	TestAdmin3         = "mike_superadmin"
	TestAdmin3Password = "SuperAdmin$987Pass"
	TestAdmin3Email    = "mike.super@test.com"

	TestMaintainer1         = "alex_paul"
	TestMaintainer1Password = "ads#A23sweQwerty"
	TestMaintainer1Email    = "alex_paul@test.com"

	TestMaintainer2         = "emma_wilson"
	TestMaintainer2Password = "Maintain@Emma456"
	TestMaintainer2Email    = "emma.wilson@test.com"

	TestMaintainer3         = "david_maintain"
	TestMaintainer3Password = "David#Secure789"
	TestMaintainer3Email    = "david.maintain@test.com"

	TestDeveloper1         = "bob_developer"
	TestDeveloper1Password = "DevBob@123Pass"
	TestDeveloper1Email    = "bob.dev@test.com"

	TestDeveloper2         = "alice_coder"
	TestDeveloper2Password = "AliceCodes#2024"
	TestDeveloper2Email    = "alice.coder@test.com"

	TestDeveloper3         = "charlie_dev"
	TestDeveloper3Password = "Charlie$Dev789"
	TestDeveloper3Email    = "charlie.dev@test.com"

	TestDeveloper4         = "diana_programmer"
	TestDeveloper4Password = "Diana@Prog456XYZ"
	TestDeveloper4Email    = "diana.prog@test.com"

	TestDeveloper5         = "eric_engineer"
	TestDeveloper5Password = "Eric#Engineer123"
	TestDeveloper5Email    = "eric.eng@test.com"

	TestGuest1         = "guest_user1"
	TestGuest1Password = "Guest@User123456"
	TestGuest1Email    = "guest1@test.com"

	TestGuest2         = "guest_user2"
	TestGuest2Password = "SecureGuest#789"
	TestGuest2Email    = "guest2@test.com"

	TestGuest3         = "guest_user3"
	TestGuest3Password = "Guest$User2024XY"
	TestGuest3Email    = "guest3@test.com"

	TestGuest4         = "viewer_guest"
	TestGuest4Password = "Viewer@Guest456"
	TestGuest4Email    = "viewer.guest@test.com"

	TestGuest5         = "readonly_user"
	TestGuest5Password = "ReadOnly#Pass789"
	TestGuest5Email    = "readonly@test.com"
)

// Headers
const (
	HeaderAccountSetupUUID = "Account-Setup-Id"
)

// Test Invalid User details
const (
	EmptyEmail1   = ""
	InvalidEmail1 = "john_doe@"

	EmptyUsername1   = ""
	InvalidUsername1 = "John Doe"

	EmptyRole1             = ""
	InvalidAdminRole1      = "admin"
	InvalidMaintainerRole1 = "maintainer"
	InvalidDeveloperRole1  = "developer"
	InvalidGuestRole1      = "guest"
)

// Content Types
const (
	ApplicationJson = "application/json"
	ApplicationXml  = "application/xml"
)

// Endpoints
const (
	// Authentication
	EndpointLogin = "/api/v1/auth/login"

	// User Management (Base)
	EndpointUsers       = "/api/v1/users"
	EndpointCurrentUser = "/api/v1/users/me"

	// User Management ID Specific
	EndpointUserByID        = "/api/v1/users/%s"
	EndpointUserEmail       = "/api/v1/users/%s/email"
	EndpointUserDisplayName = "/api/v1/users/%s/display-name"
	EndpointUserPassword    = "/api/v1/users/%s/password"
	EndpointUserLock        = "/api/v1/users/%s/lock"
	EndpointUserUnlock      = "/api/v1/users/%s/unlock"
	EndpointUserChangeRole  = "/api/v1/users/%s/role"

	// Account Setup/Validation
	EndpointValidateUser         = "/api/v1/users/validate"
	EndpointValidatePassword     = "/api/v1/users/validate-password"
	EndpointAccountSetupInfo     = "/api/v1/users/account-setup/%s"
	EndpointAccountSetupComplete = "/api/v1/users/account-setup/%s/complete"

	// Access Management
	EndpointAccessBase   = "/api/v1/access"
	EndpointNamespaces   = "/api/v1/access/namespaces"
	EndpointRepositories = "/api/v1/access/repositories"
	EndpointUpstreams    = "/api/v1/access/upstreams"

	// Namespace Access ID Specific
	EndpointNamespaceByID       = "/api/v1/access/namespaces/%s"
	EndpointNamespaceState      = "/api/v1/access/namespaces/%s/state"
	EndpointNamespaceVisibility = "/api/v1/access/namespaces/%s/visibility"
	EndpointNamespaceUsers      = "/api/v1/access/namespaces/%s/users"
	EndpointNamespaceUserRevoke = "/api/v1/access/namespaces/%s/users/%s" // Identifier, UserID

	// Repository Access ID Specific
	EndpointRepositoryByID       = "/api/v1/access/repositories/%s"
	EndpointRepositoryState      = "/api/v1/access/repositories/%s/state"
	EndpointRepositoryVisibility = "/api/v1/access/repositories/%s/visibility"
	EndpointRepositoryUsers      = "/api/v1/access/repositories/%s/users"
	EndpointRepositoryUserRevoke = "/api/v1/access/repositories/%s/users/%s"

	EndpointHealthCheck = "/api/v1/health"
)