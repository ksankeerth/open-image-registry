# OpenImageRegistry API Documentation

**Base URL:** `/api/v1`

**Version:** v1alpha

## Overview

OpenImageRegistry is a Docker registry alternative with built-in WebUI, horizontal scalability, and support for proxying/caching images from well-known registries.

## Technical Details

### Account Locking

**Automatic Locking Triggers:**
- Exceeding maximum failed login attempts
- New account pending verification
- Administrative lock

**Lock Reasons:**
- `failed_login_attempts` - Too many failed logins
- `new_account_verification_required` - Account setup pending
- `admin_locked` - Manually locked by administrator

**Unlock Requirements:**
- Failed login locks: Administrator must unlock
- New account locks: Complete account setup
- Admin locks: Administrator must unlock

### Session Management Details

**Session Properties:**
- Session ID: UUID v4
- Grant Type: `password`
- Scope Hash: Calculated from authorized scopes
- User Agent: Captured from request header
- Client IP: Extracted from X-Forwarded-For header
- Expiry: 900 seconds (15 minutes) from issue time

**Session Lifecycle:**
1. User authenticates with credentials
2. System validates username and password
3. New session created with unique ID
4. Session cookie set in response
5. Session expires after inactivity timeout
6. New login invalidates previous session

### Email Notifications

**Account Setup Email:**
- Sent when new user account is created
- Contains account setup verification link
- Link expires after use
- In development mode, UUID returned in response header


### Content-Type Headers

**Required for JSON Endpoints:**
- Request: `Content-Type: application/json`
- Response: `Content-Type: application/json`

---

## Authentication

### Login

Authenticates a user and creates a session.

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (Success - 200 OK):**
```json
{
  "success": true,
  "errorMessage": "",
  "sessionId": "string",
  "authorizedScopes": ["string"],
  "expiresAt": "2024-01-01T00:00:00Z",
  "user": {
    "userId": "string",
    "username": "string",
    "role": "string"
  }
}
```

**Response (Failed - 403 Forbidden):**
```json
{
  "success": false,
  "errorMessage": "Invalid username or password!",
  "sessionId": "",
  "authorizedScopes": [],
  "expiresAt": "2024-01-01T00:00:00Z"
}
```

**Cookies Set:**
- Session cookie with 900 seconds expiry on successful login

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `403 Forbidden` - Invalid credentials or locked account
- `500 Internal Server Error` - Server error

**Notes:**
- Account locks after multiple failed login attempts (max attempts configured in system)
- Session expires after 900 seconds (15 minutes) of inactivity
- Client IP is extracted from `X-Forwarded-For` header for audit logging

---

## User Management

### List Users

Retrieves a paginated list of user accounts.

**Endpoint:** `GET /api/v1/users`

**Query Parameters:**
- `page` (integer, optional) - Page number (default: 1)
- `limit` (integer, optional) - Items per page (default: 10)
- `sortField` (string, optional) - Field to sort by (allowed fields: defined in system)
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- Must be from the system's allowed user filter fields list
- `locked` field accepts single boolean value only

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "users": [
    {
      "userId": "string",
      "username": "string",
      "email": "string",
      "displayName": "string",
      "role": "string",
      "locked": false,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters or filter fields
- `500 Internal Server Error` - Database error

---

### Create User

Creates a new user account and sends account setup email.

**Endpoint:** `POST /api/v1/users`

**Request Body:**
```json
{
  "username": "string",
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `username`: Must be valid username format
- `email`: Must be valid email format
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`
- `displayName`: Optional

**Response (201 Created):**
```json
{
  "username": "string",
  "userId": "string"
}
```

**Response Headers (Development Mode):**
- `Account-Setup-Id`: UUID for account setup (only in development mode)

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `409 Conflict` - Username or email already exists
- `500 Internal Server Error` - Server error

**Notes:**
- New accounts are locked by default until account setup is completed
- Account setup email is sent with verification link
- In development mode with mock email enabled, setup UUID is returned in response header

---

### Update User

Updates user account information.

**Endpoint:** `PUT /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `email`: Required, must be valid email format
- `displayName`: Optional, max 255 characters
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `500 Internal Server Error` - Database error

---

### Delete User

Deletes a user account.

**Endpoint:** `DELETE /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Database error

---

### Update User Email

Updates a user's email address.

**Endpoint:** `PUT /api/v1/users/{id}/email`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "email": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `email`: Must be valid email format

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Update User Display Name

Updates a user's display name.

**Endpoint:** `PUT /api/v1/users/{id}/display-name`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "displayName": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `displayName`: Required (cannot be empty)

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Change Password

Changes a user's password using recovery/setup process.

**Endpoint:** `PUT /api/v1/users/{id}/password`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "recoveryId": "string",
  "oldPassword": "string",
  "password": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- Password must meet system password requirements

**Response (200 OK):**
```json
{
  "invalidId": false,
  "expired": false,
  "oldPasswordDiff": false,
  "invalidUserAccount": false,
  "changed": true
}
```

**Response Fields:**
- `invalidId`: Recovery ID is invalid or not found
- `expired`: Recovery link has expired
- `oldPasswordDiff`: Old password doesn't match
- `invalidUserAccount`: User account not found
- `changed`: Password successfully changed

**Error Responses:**
- `400 Bad Request` - Invalid request body or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Lock User Account

Locks a user account, preventing login.

**Endpoint:** `PUT /api/v1/users/{id}/lock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Account is already locked
- `500 Internal Server Error` - Database error

---

### Unlock User Account

Unlocks a user account.

**Endpoint:** `PUT /api/v1/users/{id}/unlock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Cannot unlock new accounts pending verification
- `500 Internal Server Error` - Database error

**Notes:**
- New accounts locked for verification cannot be unlocked manually
- Must complete account setup process instead

---

### Validate Username/Email

Checks availability of username and/or email.

**Endpoint:** `POST /api/v1/users/validate`

**Request Body:**
```json
{
  "username": "string",
  "email": "string"
}
```

**Validation Rules:**
- At least one of `username` or `email` must be provided

**Response (200 OK):**
```json
{
  "usernameAvailable": true,
  "emailAvailable": true
}
```

**Error Responses:**
- `400 Bad Request` - Both username and email are empty
- `500 Internal Server Error` - Database error

---

### Validate Password

Validates a password against system requirements.

**Endpoint:** `POST /api/v1/users/validate-password`

**Request Body:**
```json
{
  "password": "string"
}
```

**Response (200 OK):**
```json
{
  "isValid": true,
  "msg": "Password is valid"
}
```

**Error Responses:**
- `400 Bad Request` - Unable to parse request

---

### Get Account Setup Info

Retrieves information for account setup/verification.

**Endpoint:** `GET /api/v1/users/account-setup/{uuid}`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID from email

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "username": "string",
  "email": "string",
  "role": "string",
  "displayName": "string"
}
```

**Error Responses:**
- `404 Not Found` - Setup link is invalid or already used
- `500 Internal Server Error` - Database error

---

### Complete Account Setup

Completes the account setup process for new users.

**Endpoint:** `POST /api/v1/users/account-setup/{uuid}/complete`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID

**Request Body:**
```json
{
  "uuid": "string",
  "userId": "string",
  "username": "string",
  "displayName": "string",
  "password": "string"
}
```

**Validation Rules:**
- `uuid` in body must match `uuid` in path
- `username`: Must be valid username format
- `password`: Must meet system password requirements
- `userId`: Required

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `500 Internal Server Error` - Database error

**Notes:**
- Unlocks the account upon successful completion
- Removes the account recovery record
- User can login after completing this step

---

### Get Current User

Retrieves the currently authenticated user's information.

**Endpoint:** `GET /api/v1/users/me`

**Status:** Not yet implemented

---

### Update Current User

Updates the currently authenticated user's information.

**Endpoint:** `PUT /api/v1/users/me`

**Status:** Not yet implemented

---

## Namespace Management

Namespaces are used to organize repositories. They can be associated with teams or projects.

### List Namespaces

Retrieves a paginated list of namespaces.

**Endpoint:** `GET /api/v1/access/namespaces`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- Other namespace-specific fields as defined in system

**Response (200 OK):**
```json
{
  "total": 50,
  "page": 1,
  "limit": 10,
  "namespaces": [
    {
      "id": "string",
      "registryId": "string",
      "name": "string",
      "purpose": "project",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Namespace

Creates a new namespace with assigned maintainers.

**Endpoint:** `POST /api/v1/access/namespaces`

**Request Body:**
```json
{
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "maintainers": ["userId1", "userId2"]
}
```

**Validation Rules:**
- `name`: Must be valid namespace format
- `purpose`: Required. Must be `project` or `team`
- `maintainers`: Required. At least one maintainer must be specified
- All maintainers must have active accounts with 'Maintainer' role

**Response (201 Created):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid maintainer roles
- `409 Conflict` - Namespace name already exists
- `500 Internal Server Error` - Server error

**Notes:**
- Maintainers automatically receive maintainer-level access to the namespace
- Namespace names must be unique within the registry

---

### Check Namespace Exists

Checks if a namespace exists by identifier (ID or name).

**Endpoint:** `HEAD /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response:**
- `200 OK` - Namespace exists
- `404 Not Found` - Namespace does not exist
- `500 Internal Server Error` - Server error

---

### Get Namespace

Retrieves detailed information about a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### Update Namespace

Updates namespace information.

**Endpoint:** `PUT /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "id": "string",
  "description": "string",
  "purpose": "project"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Purpose defaults to existing value if not provided

---

### Delete Namespace

Deletes a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Namespace State

Changes the state of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/state`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Changing to non-active state affects associated repositories

---

### Change Namespace Visibility

Changes the visibility (public/private) of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/visibility`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility in disabled state
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace is in disabled state
- Changing to private affects associated repositories

---

### List Namespace Users

Lists users with access to a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 25,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "namespace",
      "accessLevel": "maintainer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

### Grant Namespace Access

Grants a user access to a namespace.

**Endpoint:** `POST /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `namespace`
- `accessLevel`: Must be `guest`, `developer`, or `maintainer`
- `identifier` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or namespace not found
- `500 Internal Server Error` - Server error

---

### Revoke Namespace Access

Revokes a user's access to a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}/users/{userID}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### List Namespace Repositories

Lists all repositories within a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/repositories`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 30,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

## Repository Management

Repositories store container images within namespaces.

### List Repositories

Retrieves a paginated list of repositories.

**Endpoint:** `GET /api/v1/access/repositories`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- `tags` - Range filter with format `>value` or `<value` (max 2 values)
- Other repository-specific fields

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Repository

Creates a new repository within a namespace.

**Endpoint:** `POST /api/v1/access/repositories`

**Request Body:**
```json
{
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "createdBy": "string"
}
```

**Validation Rules:**
- `name`: Must be valid repository format
- `namespaceId`: Required, must be valid namespace
- `createdBy`: Required, username of creator

**Response (200 OK):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid namespace
- `409 Conflict` - Repository with same identifier already exists
- `500 Internal Server Error` - Server error

---

### Check Repository Exists

Checks if a repository exists by ID.

**Endpoint:** `HEAD /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response:**
- `200 OK` - Repository exists
- `404 Not Found` - Repository does not exist
- `500 Internal Server Error` - Server error

---

### Get Repository

Retrieves detailed information about a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "tagCount": 5,
  "createdBy": "string",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Update Repository

Updates repository information.

**Endpoint:** `PUT /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "description": "string"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Delete Repository

Deletes a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Repository State

Changes the state of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/state`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Cannot change state when namespace is disabled
- Cannot change to `active` when namespace is deprecated

---

### Change Repository Visibility

Changes the visibility (public/private) of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/visibility`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility when disabled
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace or repository is disabled

---

### List Repository Users

Lists users with access to a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 15,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "repository",
      "accessLevel": "developer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

**Notes:**
- Returns users with access to both the repository and parent namespace

---

### Grant Repository Access

Grants a user access to a repository.

**Endpoint:** `POST /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `repository`
- `accessLevel`: Must be `guest` or `developer`
- `id` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or repository not found
- `500 Internal Server Error` - Server error

---

### Revoke Repository Access

Revokes a user's access to a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}/users/{userID}`

**Path Parameters:**
- `id` (string, required) - Repository ID
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

## Common Response Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters or body
- `403 Forbidden` - Authentication failed or insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate name)
- `500 Internal Server Error` - Server error

---

## Authentication & Session Management

### Session Details
- Sessions expire after 900 seconds (15 minutes)
- Session cookie is set on successful login
- Failed login attempts are tracked per user
- Accounts lock after exceeding max failed login attempts
- Locked accounts must be unlocked by administrator

### Security Features
- Password hashing with salt
- Failed login attempt tracking
- Account locking mechanisms
- Session management with expiry
- Client IP tracking for audit

---

## Access Control

### Resource Types
- `namespace` - Organizational container for repositories
- `repository` - Container image storage

### Access Levels

**Namespace Access Levels:**
- `maintainer` - Full control over namespace and repositories
- `developer` - Can push/pull images, create repositories
- `guest` - Read-only access

**Repository Access Levels:**
- `developer` - Can push/pull images
- `guest` - Read-only access (pull only)

### User Roles

The system supports the following user roles:

- `admin` - Full administrative access to the system
- `maintainer` - Can maintain namespaces and repositories
- `developer` - Can work with repositories
- `guest` - Read-only access

## Validation Rules

### Username Validation

**Format:** `^[a-zA-Z0-9._-]{3,32}# OpenImageRegistry API Documentation

**Base URL:** `/api/v1`

**Version:** v1alpha

## Overview

OpenImageRegistry is a Docker registry alternative with built-in WebUI, horizontal scalability, and support for proxying/caching images from well-known registries.

---

## Authentication

### Login

Authenticates a user and creates a session.

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (Success - 200 OK):**
```json
{
  "success": true,
  "errorMessage": "",
  "sessionId": "string",
  "authorizedScopes": ["string"],
  "expiresAt": "2024-01-01T00:00:00Z",
  "user": {
    "userId": "string",
    "username": "string",
    "role": "string"
  }
}
```

**Response (Failed - 403 Forbidden):**
```json
{
  "success": false,
  "errorMessage": "Invalid username or password!",
  "sessionId": "",
  "authorizedScopes": [],
  "expiresAt": "2024-01-01T00:00:00Z"
}
```

**Cookies Set:**
- Session cookie with 900 seconds expiry on successful login

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `403 Forbidden` - Invalid credentials or locked account
- `500 Internal Server Error` - Server error

**Notes:**
- Account locks after multiple failed login attempts (max attempts configured in system)
- Session expires after 900 seconds (15 minutes) of inactivity
- Client IP is extracted from `X-Forwarded-For` header for audit logging

---

## User Management

### List Users

Retrieves a paginated list of user accounts.

**Endpoint:** `GET /api/v1/users`

**Query Parameters:**
- `page` (integer, optional) - Page number (default: 1)
- `limit` (integer, optional) - Items per page (default: 10)
- `sortField` (string, optional) - Field to sort by (allowed fields: defined in system)
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- Must be from the system's allowed user filter fields list
- `locked` field accepts single boolean value only

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "users": [
    {
      "userId": "string",
      "username": "string",
      "email": "string",
      "displayName": "string",
      "role": "string",
      "locked": false,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters or filter fields
- `500 Internal Server Error` - Database error

---

### Create User

Creates a new user account and sends account setup email.

**Endpoint:** `POST /api/v1/users`

**Request Body:**
```json
{
  "username": "string",
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `username`: Must be valid username format
- `email`: Must be valid email format
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`
- `displayName`: Optional

**Response (201 Created):**
```json
{
  "username": "string",
  "userId": "string"
}
```

**Response Headers (Development Mode):**
- `Account-Setup-Id`: UUID for account setup (only in development mode)

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `409 Conflict` - Username or email already exists
- `500 Internal Server Error` - Server error

**Notes:**
- New accounts are locked by default until account setup is completed
- Account setup email is sent with verification link
- In development mode with mock email enabled, setup UUID is returned in response header

---

### Update User

Updates user account information.

**Endpoint:** `PUT /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `email`: Required, must be valid email format
- `displayName`: Optional, max 255 characters
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `500 Internal Server Error` - Database error

---

### Delete User

Deletes a user account.

**Endpoint:** `DELETE /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Database error

---

### Update User Email

Updates a user's email address.

**Endpoint:** `PUT /api/v1/users/{id}/email`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "email": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `email`: Must be valid email format

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Update User Display Name

Updates a user's display name.

**Endpoint:** `PUT /api/v1/users/{id}/display-name`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "displayName": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `displayName`: Required (cannot be empty)

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Change Password

Changes a user's password using recovery/setup process.

**Endpoint:** `PUT /api/v1/users/{id}/password`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "recoveryId": "string",
  "oldPassword": "string",
  "password": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- Password must meet system password requirements

**Response (200 OK):**
```json
{
  "invalidId": false,
  "expired": false,
  "oldPasswordDiff": false,
  "invalidUserAccount": false,
  "changed": true
}
```

**Response Fields:**
- `invalidId`: Recovery ID is invalid or not found
- `expired`: Recovery link has expired
- `oldPasswordDiff`: Old password doesn't match
- `invalidUserAccount`: User account not found
- `changed`: Password successfully changed

**Error Responses:**
- `400 Bad Request` - Invalid request body or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Lock User Account

Locks a user account, preventing login.

**Endpoint:** `PUT /api/v1/users/{id}/lock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Account is already locked
- `500 Internal Server Error` - Database error

---

### Unlock User Account

Unlocks a user account.

**Endpoint:** `PUT /api/v1/users/{id}/unlock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Cannot unlock new accounts pending verification
- `500 Internal Server Error` - Database error

**Notes:**
- New accounts locked for verification cannot be unlocked manually
- Must complete account setup process instead

---

### Validate Username/Email

Checks availability of username and/or email.

**Endpoint:** `POST /api/v1/users/validate`

**Request Body:**
```json
{
  "username": "string",
  "email": "string"
}
```

**Validation Rules:**
- At least one of `username` or `email` must be provided

**Response (200 OK):**
```json
{
  "usernameAvailable": true,
  "emailAvailable": true
}
```

**Error Responses:**
- `400 Bad Request` - Both username and email are empty
- `500 Internal Server Error` - Database error

---

### Validate Password

Validates a password against system requirements.

**Endpoint:** `POST /api/v1/users/validate-password`

**Request Body:**
```json
{
  "password": "string"
}
```

**Response (200 OK):**
```json
{
  "isValid": true,
  "msg": "Password is valid"
}
```

**Error Responses:**
- `400 Bad Request` - Unable to parse request

---

### Get Account Setup Info

Retrieves information for account setup/verification.

**Endpoint:** `GET /api/v1/users/account-setup/{uuid}`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID from email

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "username": "string",
  "email": "string",
  "role": "string",
  "displayName": "string"
}
```

**Error Responses:**
- `404 Not Found` - Setup link is invalid or already used
- `500 Internal Server Error` - Database error

---

### Complete Account Setup

Completes the account setup process for new users.

**Endpoint:** `POST /api/v1/users/account-setup/{uuid}/complete`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID

**Request Body:**
```json
{
  "uuid": "string",
  "userId": "string",
  "username": "string",
  "displayName": "string",
  "password": "string"
}
```

**Validation Rules:**
- `uuid` in body must match `uuid` in path
- `username`: Must be valid username format
- `password`: Must meet system password requirements
- `userId`: Required

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `500 Internal Server Error` - Database error

**Notes:**
- Unlocks the account upon successful completion
- Removes the account recovery record
- User can login after completing this step

---

### Get Current User

Retrieves the currently authenticated user's information.

**Endpoint:** `GET /api/v1/users/me`

**Status:** Not yet implemented

---

### Update Current User

Updates the currently authenticated user's information.

**Endpoint:** `PUT /api/v1/users/me`

**Status:** Not yet implemented

---

## Namespace Management

Namespaces are used to organize repositories. They can be associated with teams or projects.

### List Namespaces

Retrieves a paginated list of namespaces.

**Endpoint:** `GET /api/v1/access/namespaces`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- Other namespace-specific fields as defined in system

**Response (200 OK):**
```json
{
  "total": 50,
  "page": 1,
  "limit": 10,
  "namespaces": [
    {
      "id": "string",
      "registryId": "string",
      "name": "string",
      "purpose": "project",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Namespace

Creates a new namespace with assigned maintainers.

**Endpoint:** `POST /api/v1/access/namespaces`

**Request Body:**
```json
{
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "maintainers": ["userId1", "userId2"]
}
```

**Validation Rules:**
- `name`: Must be valid namespace format
- `purpose`: Required. Must be `project` or `team`
- `maintainers`: Required. At least one maintainer must be specified
- All maintainers must have active accounts with 'Maintainer' role

**Response (201 Created):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid maintainer roles
- `409 Conflict` - Namespace name already exists
- `500 Internal Server Error` - Server error

**Notes:**
- Maintainers automatically receive maintainer-level access to the namespace
- Namespace names must be unique within the registry

---

### Check Namespace Exists

Checks if a namespace exists by identifier (ID or name).

**Endpoint:** `HEAD /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response:**
- `200 OK` - Namespace exists
- `404 Not Found` - Namespace does not exist
- `500 Internal Server Error` - Server error

---

### Get Namespace

Retrieves detailed information about a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### Update Namespace

Updates namespace information.

**Endpoint:** `PUT /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "id": "string",
  "description": "string",
  "purpose": "project"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Purpose defaults to existing value if not provided

---

### Delete Namespace

Deletes a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Namespace State

Changes the state of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/state`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Changing to non-active state affects associated repositories

---

### Change Namespace Visibility

Changes the visibility (public/private) of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/visibility`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility in disabled state
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace is in disabled state
- Changing to private affects associated repositories

---

### List Namespace Users

Lists users with access to a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 25,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "namespace",
      "accessLevel": "maintainer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

### Grant Namespace Access

Grants a user access to a namespace.

**Endpoint:** `POST /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `namespace`
- `accessLevel`: Must be `guest`, `developer`, or `maintainer`
- `identifier` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or namespace not found
- `500 Internal Server Error` - Server error

---

### Revoke Namespace Access

Revokes a user's access to a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}/users/{userID}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### List Namespace Repositories

Lists all repositories within a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/repositories`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 30,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

## Repository Management

Repositories store container images within namespaces.

### List Repositories

Retrieves a paginated list of repositories.

**Endpoint:** `GET /api/v1/access/repositories`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- `tags` - Range filter with format `>value` or `<value` (max 2 values)
- Other repository-specific fields

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Repository

Creates a new repository within a namespace.

**Endpoint:** `POST /api/v1/access/repositories`

**Request Body:**
```json
{
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "createdBy": "string"
}
```

**Validation Rules:**
- `name`: Must be valid repository format
- `namespaceId`: Required, must be valid namespace
- `createdBy`: Required, username of creator

**Response (200 OK):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid namespace
- `409 Conflict` - Repository with same identifier already exists
- `500 Internal Server Error` - Server error

---

### Check Repository Exists

Checks if a repository exists by ID.

**Endpoint:** `HEAD /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response:**
- `200 OK` - Repository exists
- `404 Not Found` - Repository does not exist
- `500 Internal Server Error` - Server error

---

### Get Repository

Retrieves detailed information about a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "tagCount": 5,
  "createdBy": "string",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Update Repository

Updates repository information.

**Endpoint:** `PUT /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "description": "string"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Delete Repository

Deletes a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Repository State

Changes the state of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/state`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Cannot change state when namespace is disabled
- Cannot change to `active` when namespace is deprecated

---

### Change Repository Visibility

Changes the visibility (public/private) of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/visibility`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility when disabled
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace or repository is disabled

---

### List Repository Users

Lists users with access to a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 15,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "repository",
      "accessLevel": "developer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

**Notes:**
- Returns users with access to both the repository and parent namespace

---

### Grant Repository Access

Grants a user access to a repository.

**Endpoint:** `POST /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `repository`
- `accessLevel`: Must be `guest` or `developer`
- `id` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or repository not found
- `500 Internal Server Error` - Server error

---

### Revoke Repository Access

Revokes a user's access to a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}/users/{userID}`

**Path Parameters:**
- `id` (string, required) - Repository ID
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

## Common Response Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters or body
- `403 Forbidden` - Authentication failed or insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate name)
- `500 Internal Server Error` - Server error

---

## Authentication & Session Management

### Session Details
- Sessions expire after 900 seconds (15 minutes)
- Session cookie is set on successful login
- Failed login attempts are tracked per user
- Accounts lock after exceeding max failed login attempts
- Locked accounts must be unlocked by administrator

### Security Features
- Password hashing with salt
- Failed login attempt tracking
- Account locking mechanisms
- Session management with expiry
- Client IP tracking for audit

---

## Access Control

### Resource Types
- `namespace` - Organizational container for repositories
- `repository` - Container image storage

### Access Levels

**Namespace Access Levels:**
- `maintainer` - Full control over namespace and repositories
- `developer` - Can push/pull images, create repositories
- `guest` - Read-only access

**Repository Access Levels:**
- `developer` - Can push/pull images
- `guest` - Read-only access (pull only)

### User Roles

The system supports the following user roles:

- `admin` - Full administrative access to the system
- `maintainer` - Can maintain namespaces and repositories
- `developer` - Can work with repositories
- `guest` - Read-only access



**Rules:**
- Length: 3-32 characters
- Allowed characters: alphanumeric, dot (.), underscore (_), hyphen (-)
- Must start and end with alphanumeric character

**Examples:**
- Valid: `john_doe`, `user123`, `test-user.dev`
- Invalid: `ab` (too short), `user@domain` (invalid character), `_user` (starts with underscore)

### Email Validation

**Format:** `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}# OpenImageRegistry API Documentation

**Base URL:** `/api/v1`

**Version:** v1alpha

## Overview

OpenImageRegistry is a Docker registry alternative with built-in WebUI, horizontal scalability, and support for proxying/caching images from well-known registries.

---

## Authentication

### Login

Authenticates a user and creates a session.

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (Success - 200 OK):**
```json
{
  "success": true,
  "errorMessage": "",
  "sessionId": "string",
  "authorizedScopes": ["string"],
  "expiresAt": "2024-01-01T00:00:00Z",
  "user": {
    "userId": "string",
    "username": "string",
    "role": "string"
  }
}
```

**Response (Failed - 403 Forbidden):**
```json
{
  "success": false,
  "errorMessage": "Invalid username or password!",
  "sessionId": "",
  "authorizedScopes": [],
  "expiresAt": "2024-01-01T00:00:00Z"
}
```

**Cookies Set:**
- Session cookie with 900 seconds expiry on successful login

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `403 Forbidden` - Invalid credentials or locked account
- `500 Internal Server Error` - Server error

**Notes:**
- Account locks after multiple failed login attempts (max attempts configured in system)
- Session expires after 900 seconds (15 minutes) of inactivity
- Client IP is extracted from `X-Forwarded-For` header for audit logging

---

## User Management

### List Users

Retrieves a paginated list of user accounts.

**Endpoint:** `GET /api/v1/users`

**Query Parameters:**
- `page` (integer, optional) - Page number (default: 1)
- `limit` (integer, optional) - Items per page (default: 10)
- `sortField` (string, optional) - Field to sort by (allowed fields: defined in system)
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- Must be from the system's allowed user filter fields list
- `locked` field accepts single boolean value only

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "users": [
    {
      "userId": "string",
      "username": "string",
      "email": "string",
      "displayName": "string",
      "role": "string",
      "locked": false,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters or filter fields
- `500 Internal Server Error` - Database error

---

### Create User

Creates a new user account and sends account setup email.

**Endpoint:** `POST /api/v1/users`

**Request Body:**
```json
{
  "username": "string",
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `username`: Must be valid username format
- `email`: Must be valid email format
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`
- `displayName`: Optional

**Response (201 Created):**
```json
{
  "username": "string",
  "userId": "string"
}
```

**Response Headers (Development Mode):**
- `Account-Setup-Id`: UUID for account setup (only in development mode)

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `409 Conflict` - Username or email already exists
- `500 Internal Server Error` - Server error

**Notes:**
- New accounts are locked by default until account setup is completed
- Account setup email is sent with verification link
- In development mode with mock email enabled, setup UUID is returned in response header

---

### Update User

Updates user account information.

**Endpoint:** `PUT /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `email`: Required, must be valid email format
- `displayName`: Optional, max 255 characters
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `500 Internal Server Error` - Database error

---

### Delete User

Deletes a user account.

**Endpoint:** `DELETE /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Database error

---

### Update User Email

Updates a user's email address.

**Endpoint:** `PUT /api/v1/users/{id}/email`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "email": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `email`: Must be valid email format

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Update User Display Name

Updates a user's display name.

**Endpoint:** `PUT /api/v1/users/{id}/display-name`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "displayName": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `displayName`: Required (cannot be empty)

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Change Password

Changes a user's password using recovery/setup process.

**Endpoint:** `PUT /api/v1/users/{id}/password`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "recoveryId": "string",
  "oldPassword": "string",
  "password": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- Password must meet system password requirements

**Response (200 OK):**
```json
{
  "invalidId": false,
  "expired": false,
  "oldPasswordDiff": false,
  "invalidUserAccount": false,
  "changed": true
}
```

**Response Fields:**
- `invalidId`: Recovery ID is invalid or not found
- `expired`: Recovery link has expired
- `oldPasswordDiff`: Old password doesn't match
- `invalidUserAccount`: User account not found
- `changed`: Password successfully changed

**Error Responses:**
- `400 Bad Request` - Invalid request body or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Lock User Account

Locks a user account, preventing login.

**Endpoint:** `PUT /api/v1/users/{id}/lock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Account is already locked
- `500 Internal Server Error` - Database error

---

### Unlock User Account

Unlocks a user account.

**Endpoint:** `PUT /api/v1/users/{id}/unlock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Cannot unlock new accounts pending verification
- `500 Internal Server Error` - Database error

**Notes:**
- New accounts locked for verification cannot be unlocked manually
- Must complete account setup process instead

---

### Validate Username/Email

Checks availability of username and/or email.

**Endpoint:** `POST /api/v1/users/validate`

**Request Body:**
```json
{
  "username": "string",
  "email": "string"
}
```

**Validation Rules:**
- At least one of `username` or `email` must be provided

**Response (200 OK):**
```json
{
  "usernameAvailable": true,
  "emailAvailable": true
}
```

**Error Responses:**
- `400 Bad Request` - Both username and email are empty
- `500 Internal Server Error` - Database error

---

### Validate Password

Validates a password against system requirements.

**Endpoint:** `POST /api/v1/users/validate-password`

**Request Body:**
```json
{
  "password": "string"
}
```

**Response (200 OK):**
```json
{
  "isValid": true,
  "msg": "Password is valid"
}
```

**Error Responses:**
- `400 Bad Request` - Unable to parse request

---

### Get Account Setup Info

Retrieves information for account setup/verification.

**Endpoint:** `GET /api/v1/users/account-setup/{uuid}`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID from email

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "username": "string",
  "email": "string",
  "role": "string",
  "displayName": "string"
}
```

**Error Responses:**
- `404 Not Found` - Setup link is invalid or already used
- `500 Internal Server Error` - Database error

---

### Complete Account Setup

Completes the account setup process for new users.

**Endpoint:** `POST /api/v1/users/account-setup/{uuid}/complete`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID

**Request Body:**
```json
{
  "uuid": "string",
  "userId": "string",
  "username": "string",
  "displayName": "string",
  "password": "string"
}
```

**Validation Rules:**
- `uuid` in body must match `uuid` in path
- `username`: Must be valid username format
- `password`: Must meet system password requirements
- `userId`: Required

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `500 Internal Server Error` - Database error

**Notes:**
- Unlocks the account upon successful completion
- Removes the account recovery record
- User can login after completing this step

---

### Get Current User

Retrieves the currently authenticated user's information.

**Endpoint:** `GET /api/v1/users/me`

**Status:** Not yet implemented

---

### Update Current User

Updates the currently authenticated user's information.

**Endpoint:** `PUT /api/v1/users/me`

**Status:** Not yet implemented

---

## Namespace Management

Namespaces are used to organize repositories. They can be associated with teams or projects.

### List Namespaces

Retrieves a paginated list of namespaces.

**Endpoint:** `GET /api/v1/access/namespaces`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- Other namespace-specific fields as defined in system

**Response (200 OK):**
```json
{
  "total": 50,
  "page": 1,
  "limit": 10,
  "namespaces": [
    {
      "id": "string",
      "registryId": "string",
      "name": "string",
      "purpose": "project",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Namespace

Creates a new namespace with assigned maintainers.

**Endpoint:** `POST /api/v1/access/namespaces`

**Request Body:**
```json
{
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "maintainers": ["userId1", "userId2"]
}
```

**Validation Rules:**
- `name`: Must be valid namespace format
- `purpose`: Required. Must be `project` or `team`
- `maintainers`: Required. At least one maintainer must be specified
- All maintainers must have active accounts with 'Maintainer' role

**Response (201 Created):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid maintainer roles
- `409 Conflict` - Namespace name already exists
- `500 Internal Server Error` - Server error

**Notes:**
- Maintainers automatically receive maintainer-level access to the namespace
- Namespace names must be unique within the registry

---

### Check Namespace Exists

Checks if a namespace exists by identifier (ID or name).

**Endpoint:** `HEAD /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response:**
- `200 OK` - Namespace exists
- `404 Not Found` - Namespace does not exist
- `500 Internal Server Error` - Server error

---

### Get Namespace

Retrieves detailed information about a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### Update Namespace

Updates namespace information.

**Endpoint:** `PUT /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "id": "string",
  "description": "string",
  "purpose": "project"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Purpose defaults to existing value if not provided

---

### Delete Namespace

Deletes a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Namespace State

Changes the state of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/state`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Changing to non-active state affects associated repositories

---

### Change Namespace Visibility

Changes the visibility (public/private) of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/visibility`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility in disabled state
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace is in disabled state
- Changing to private affects associated repositories

---

### List Namespace Users

Lists users with access to a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 25,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "namespace",
      "accessLevel": "maintainer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

### Grant Namespace Access

Grants a user access to a namespace.

**Endpoint:** `POST /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `namespace`
- `accessLevel`: Must be `guest`, `developer`, or `maintainer`
- `identifier` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or namespace not found
- `500 Internal Server Error` - Server error

---

### Revoke Namespace Access

Revokes a user's access to a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}/users/{userID}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### List Namespace Repositories

Lists all repositories within a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/repositories`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 30,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

## Repository Management

Repositories store container images within namespaces.

### List Repositories

Retrieves a paginated list of repositories.

**Endpoint:** `GET /api/v1/access/repositories`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- `tags` - Range filter with format `>value` or `<value` (max 2 values)
- Other repository-specific fields

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Repository

Creates a new repository within a namespace.

**Endpoint:** `POST /api/v1/access/repositories`

**Request Body:**
```json
{
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "createdBy": "string"
}
```

**Validation Rules:**
- `name`: Must be valid repository format
- `namespaceId`: Required, must be valid namespace
- `createdBy`: Required, username of creator

**Response (200 OK):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid namespace
- `409 Conflict` - Repository with same identifier already exists
- `500 Internal Server Error` - Server error

---

### Check Repository Exists

Checks if a repository exists by ID.

**Endpoint:** `HEAD /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response:**
- `200 OK` - Repository exists
- `404 Not Found` - Repository does not exist
- `500 Internal Server Error` - Server error

---

### Get Repository

Retrieves detailed information about a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "tagCount": 5,
  "createdBy": "string",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Update Repository

Updates repository information.

**Endpoint:** `PUT /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "description": "string"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Delete Repository

Deletes a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Repository State

Changes the state of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/state`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Cannot change state when namespace is disabled
- Cannot change to `active` when namespace is deprecated

---

### Change Repository Visibility

Changes the visibility (public/private) of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/visibility`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility when disabled
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace or repository is disabled

---

### List Repository Users

Lists users with access to a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 15,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "repository",
      "accessLevel": "developer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

**Notes:**
- Returns users with access to both the repository and parent namespace

---

### Grant Repository Access

Grants a user access to a repository.

**Endpoint:** `POST /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `repository`
- `accessLevel`: Must be `guest` or `developer`
- `id` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or repository not found
- `500 Internal Server Error` - Server error

---

### Revoke Repository Access

Revokes a user's access to a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}/users/{userID}`

**Path Parameters:**
- `id` (string, required) - Repository ID
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

## Common Response Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters or body
- `403 Forbidden` - Authentication failed or insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate name)
- `500 Internal Server Error` - Server error

---

## Authentication & Session Management

### Session Details
- Sessions expire after 900 seconds (15 minutes)
- Session cookie is set on successful login
- Failed login attempts are tracked per user
- Accounts lock after exceeding max failed login attempts
- Locked accounts must be unlocked by administrator

### Security Features
- Password hashing with salt
- Failed login attempt tracking
- Account locking mechanisms
- Session management with expiry
- Client IP tracking for audit

---

## Access Control

### Resource Types
- `namespace` - Organizational container for repositories
- `repository` - Container image storage

### Access Levels

**Namespace Access Levels:**
- `maintainer` - Full control over namespace and repositories
- `developer` - Can push/pull images, create repositories
- `guest` - Read-only access

**Repository Access Levels:**
- `developer` - Can push/pull images
- `guest` - Read-only access (pull only)

### User Roles

The system supports the following user roles:

- `admin` - Full administrative access to the system
- `maintainer` - Can maintain namespaces and repositories
- `developer` - Can work with repositories
- `guest` - Read-only access



**Rules:**
- Must contain @ symbol
- Valid local part (before @): alphanumeric, dot, underscore, percent, plus, hyphen
- Valid domain part: alphanumeric, dot, hyphen
- TLD must be at least 2 characters

**Examples:**
- Valid: `user@example.com`, `john.doe+test@company.co.uk`
- Invalid: `user@`, `@example.com`, `user@domain`

### Password Validation

**Requirements:**
- Minimum length: 12 characters
- Maximum length: 64 characters
- Must contain at least one uppercase letter (A-Z)
- Must contain at least one lowercase letter (a-z)
- Must contain at least one digit (0-9)
- Must contain at least one symbol from: `!@#$%^&*`

**Error Messages:**
- "Password must be at least 12 characters long"
- "Password cannot exceed 64 characters"
- "Password must contain at least one uppercase letter"
- "Password must contain at least one lowercase letter"
- "Password must contain at least one number"
- "Password must contain at least one symbol (!@#$%^&*)"

**Examples:**
- Valid: `MyP@ssw0rd123`, `Secure#Pass2024!`
- Invalid: `password` (no uppercase, no digit, no symbol, too short), `PASSWORD123` (no lowercase, no symbol)

### Namespace Name Validation

**Format:** `^[a-zA-Z0-9_-]+# OpenImageRegistry API Documentation

**Base URL:** `/api/v1`

**Version:** v1alpha

## Overview

OpenImageRegistry is a Docker registry alternative with built-in WebUI, horizontal scalability, and support for proxying/caching images from well-known registries.

---

## Authentication

### Login

Authenticates a user and creates a session.

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (Success - 200 OK):**
```json
{
  "success": true,
  "errorMessage": "",
  "sessionId": "string",
  "authorizedScopes": ["string"],
  "expiresAt": "2024-01-01T00:00:00Z",
  "user": {
    "userId": "string",
    "username": "string",
    "role": "string"
  }
}
```

**Response (Failed - 403 Forbidden):**
```json
{
  "success": false,
  "errorMessage": "Invalid username or password!",
  "sessionId": "",
  "authorizedScopes": [],
  "expiresAt": "2024-01-01T00:00:00Z"
}
```

**Cookies Set:**
- Session cookie with 900 seconds expiry on successful login

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `403 Forbidden` - Invalid credentials or locked account
- `500 Internal Server Error` - Server error

**Notes:**
- Account locks after multiple failed login attempts (max attempts configured in system)
- Session expires after 900 seconds (15 minutes) of inactivity
- Client IP is extracted from `X-Forwarded-For` header for audit logging

---

## User Management

### List Users

Retrieves a paginated list of user accounts.

**Endpoint:** `GET /api/v1/users`

**Query Parameters:**
- `page` (integer, optional) - Page number (default: 1)
- `limit` (integer, optional) - Items per page (default: 10)
- `sortField` (string, optional) - Field to sort by (allowed fields: defined in system)
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- Must be from the system's allowed user filter fields list
- `locked` field accepts single boolean value only

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "users": [
    {
      "userId": "string",
      "username": "string",
      "email": "string",
      "displayName": "string",
      "role": "string",
      "locked": false,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters or filter fields
- `500 Internal Server Error` - Database error

---

### Create User

Creates a new user account and sends account setup email.

**Endpoint:** `POST /api/v1/users`

**Request Body:**
```json
{
  "username": "string",
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `username`: Must be valid username format
- `email`: Must be valid email format
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`
- `displayName`: Optional

**Response (201 Created):**
```json
{
  "username": "string",
  "userId": "string"
}
```

**Response Headers (Development Mode):**
- `Account-Setup-Id`: UUID for account setup (only in development mode)

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `409 Conflict` - Username or email already exists
- `500 Internal Server Error` - Server error

**Notes:**
- New accounts are locked by default until account setup is completed
- Account setup email is sent with verification link
- In development mode with mock email enabled, setup UUID is returned in response header

---

### Update User

Updates user account information.

**Endpoint:** `PUT /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `email`: Required, must be valid email format
- `displayName`: Optional, max 255 characters
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `500 Internal Server Error` - Database error

---

### Delete User

Deletes a user account.

**Endpoint:** `DELETE /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Database error

---

### Update User Email

Updates a user's email address.

**Endpoint:** `PUT /api/v1/users/{id}/email`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "email": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `email`: Must be valid email format

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Update User Display Name

Updates a user's display name.

**Endpoint:** `PUT /api/v1/users/{id}/display-name`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "displayName": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `displayName`: Required (cannot be empty)

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Change Password

Changes a user's password using recovery/setup process.

**Endpoint:** `PUT /api/v1/users/{id}/password`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "recoveryId": "string",
  "oldPassword": "string",
  "password": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- Password must meet system password requirements

**Response (200 OK):**
```json
{
  "invalidId": false,
  "expired": false,
  "oldPasswordDiff": false,
  "invalidUserAccount": false,
  "changed": true
}
```

**Response Fields:**
- `invalidId`: Recovery ID is invalid or not found
- `expired`: Recovery link has expired
- `oldPasswordDiff`: Old password doesn't match
- `invalidUserAccount`: User account not found
- `changed`: Password successfully changed

**Error Responses:**
- `400 Bad Request` - Invalid request body or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Lock User Account

Locks a user account, preventing login.

**Endpoint:** `PUT /api/v1/users/{id}/lock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Account is already locked
- `500 Internal Server Error` - Database error

---

### Unlock User Account

Unlocks a user account.

**Endpoint:** `PUT /api/v1/users/{id}/unlock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Cannot unlock new accounts pending verification
- `500 Internal Server Error` - Database error

**Notes:**
- New accounts locked for verification cannot be unlocked manually
- Must complete account setup process instead

---

### Validate Username/Email

Checks availability of username and/or email.

**Endpoint:** `POST /api/v1/users/validate`

**Request Body:**
```json
{
  "username": "string",
  "email": "string"
}
```

**Validation Rules:**
- At least one of `username` or `email` must be provided

**Response (200 OK):**
```json
{
  "usernameAvailable": true,
  "emailAvailable": true
}
```

**Error Responses:**
- `400 Bad Request` - Both username and email are empty
- `500 Internal Server Error` - Database error

---

### Validate Password

Validates a password against system requirements.

**Endpoint:** `POST /api/v1/users/validate-password`

**Request Body:**
```json
{
  "password": "string"
}
```

**Response (200 OK):**
```json
{
  "isValid": true,
  "msg": "Password is valid"
}
```

**Error Responses:**
- `400 Bad Request` - Unable to parse request

---

### Get Account Setup Info

Retrieves information for account setup/verification.

**Endpoint:** `GET /api/v1/users/account-setup/{uuid}`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID from email

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "username": "string",
  "email": "string",
  "role": "string",
  "displayName": "string"
}
```

**Error Responses:**
- `404 Not Found` - Setup link is invalid or already used
- `500 Internal Server Error` - Database error

---

### Complete Account Setup

Completes the account setup process for new users.

**Endpoint:** `POST /api/v1/users/account-setup/{uuid}/complete`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID

**Request Body:**
```json
{
  "uuid": "string",
  "userId": "string",
  "username": "string",
  "displayName": "string",
  "password": "string"
}
```

**Validation Rules:**
- `uuid` in body must match `uuid` in path
- `username`: Must be valid username format
- `password`: Must meet system password requirements
- `userId`: Required

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `500 Internal Server Error` - Database error

**Notes:**
- Unlocks the account upon successful completion
- Removes the account recovery record
- User can login after completing this step

---

### Get Current User

Retrieves the currently authenticated user's information.

**Endpoint:** `GET /api/v1/users/me`

**Status:** Not yet implemented

---

### Update Current User

Updates the currently authenticated user's information.

**Endpoint:** `PUT /api/v1/users/me`

**Status:** Not yet implemented

---

## Namespace Management

Namespaces are used to organize repositories. They can be associated with teams or projects.

### List Namespaces

Retrieves a paginated list of namespaces.

**Endpoint:** `GET /api/v1/access/namespaces`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- Other namespace-specific fields as defined in system

**Response (200 OK):**
```json
{
  "total": 50,
  "page": 1,
  "limit": 10,
  "namespaces": [
    {
      "id": "string",
      "registryId": "string",
      "name": "string",
      "purpose": "project",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Namespace

Creates a new namespace with assigned maintainers.

**Endpoint:** `POST /api/v1/access/namespaces`

**Request Body:**
```json
{
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "maintainers": ["userId1", "userId2"]
}
```

**Validation Rules:**
- `name`: Must be valid namespace format
- `purpose`: Required. Must be `project` or `team`
- `maintainers`: Required. At least one maintainer must be specified
- All maintainers must have active accounts with 'Maintainer' role

**Response (201 Created):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid maintainer roles
- `409 Conflict` - Namespace name already exists
- `500 Internal Server Error` - Server error

**Notes:**
- Maintainers automatically receive maintainer-level access to the namespace
- Namespace names must be unique within the registry

---

### Check Namespace Exists

Checks if a namespace exists by identifier (ID or name).

**Endpoint:** `HEAD /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response:**
- `200 OK` - Namespace exists
- `404 Not Found` - Namespace does not exist
- `500 Internal Server Error` - Server error

---

### Get Namespace

Retrieves detailed information about a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### Update Namespace

Updates namespace information.

**Endpoint:** `PUT /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "id": "string",
  "description": "string",
  "purpose": "project"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Purpose defaults to existing value if not provided

---

### Delete Namespace

Deletes a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Namespace State

Changes the state of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/state`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Changing to non-active state affects associated repositories

---

### Change Namespace Visibility

Changes the visibility (public/private) of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/visibility`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility in disabled state
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace is in disabled state
- Changing to private affects associated repositories

---

### List Namespace Users

Lists users with access to a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 25,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "namespace",
      "accessLevel": "maintainer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

### Grant Namespace Access

Grants a user access to a namespace.

**Endpoint:** `POST /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `namespace`
- `accessLevel`: Must be `guest`, `developer`, or `maintainer`
- `identifier` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or namespace not found
- `500 Internal Server Error` - Server error

---

### Revoke Namespace Access

Revokes a user's access to a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}/users/{userID}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### List Namespace Repositories

Lists all repositories within a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/repositories`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 30,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

## Repository Management

Repositories store container images within namespaces.

### List Repositories

Retrieves a paginated list of repositories.

**Endpoint:** `GET /api/v1/access/repositories`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- `tags` - Range filter with format `>value` or `<value` (max 2 values)
- Other repository-specific fields

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Repository

Creates a new repository within a namespace.

**Endpoint:** `POST /api/v1/access/repositories`

**Request Body:**
```json
{
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "createdBy": "string"
}
```

**Validation Rules:**
- `name`: Must be valid repository format
- `namespaceId`: Required, must be valid namespace
- `createdBy`: Required, username of creator

**Response (200 OK):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid namespace
- `409 Conflict` - Repository with same identifier already exists
- `500 Internal Server Error` - Server error

---

### Check Repository Exists

Checks if a repository exists by ID.

**Endpoint:** `HEAD /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response:**
- `200 OK` - Repository exists
- `404 Not Found` - Repository does not exist
- `500 Internal Server Error` - Server error

---

### Get Repository

Retrieves detailed information about a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "tagCount": 5,
  "createdBy": "string",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Update Repository

Updates repository information.

**Endpoint:** `PUT /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "description": "string"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Delete Repository

Deletes a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Repository State

Changes the state of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/state`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Cannot change state when namespace is disabled
- Cannot change to `active` when namespace is deprecated

---

### Change Repository Visibility

Changes the visibility (public/private) of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/visibility`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility when disabled
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace or repository is disabled

---

### List Repository Users

Lists users with access to a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 15,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "repository",
      "accessLevel": "developer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

**Notes:**
- Returns users with access to both the repository and parent namespace

---

### Grant Repository Access

Grants a user access to a repository.

**Endpoint:** `POST /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `repository`
- `accessLevel`: Must be `guest` or `developer`
- `id` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or repository not found
- `500 Internal Server Error` - Server error

---

### Revoke Repository Access

Revokes a user's access to a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}/users/{userID}`

**Path Parameters:**
- `id` (string, required) - Repository ID
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

## Common Response Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters or body
- `403 Forbidden` - Authentication failed or insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate name)
- `500 Internal Server Error` - Server error

---

## Authentication & Session Management

### Session Details
- Sessions expire after 900 seconds (15 minutes)
- Session cookie is set on successful login
- Failed login attempts are tracked per user
- Accounts lock after exceeding max failed login attempts
- Locked accounts must be unlocked by administrator

### Security Features
- Password hashing with salt
- Failed login attempt tracking
- Account locking mechanisms
- Session management with expiry
- Client IP tracking for audit

---

## Access Control

### Resource Types
- `namespace` - Organizational container for repositories
- `repository` - Container image storage

### Access Levels

**Namespace Access Levels:**
- `maintainer` - Full control over namespace and repositories
- `developer` - Can push/pull images, create repositories
- `guest` - Read-only access

**Repository Access Levels:**
- `developer` - Can push/pull images
- `guest` - Read-only access (pull only)

### User Roles

The system supports the following user roles:

- `admin` - Full administrative access to the system
- `maintainer` - Can maintain namespaces and repositories
- `developer` - Can work with repositories
- `guest` - Read-only access



**Rules:**
- Allowed characters: alphanumeric, underscore (_), hyphen (-)
- No length limits specified (controlled by database)
- Must be unique within the registry

**Examples:**
- Valid: `my-project`, `team_alpha`, `namespace123`
- Invalid: `my.namespace` (dot not allowed), `namespace with spaces`, `namespace@test`

### Repository Name Validation

**Format:** `^[a-zA-Z0-9_-]+# OpenImageRegistry API Documentation

**Base URL:** `/api/v1`

**Version:** v1alpha

## Overview

OpenImageRegistry is a Docker registry alternative with built-in WebUI, horizontal scalability, and support for proxying/caching images from well-known registries.

---

## Authentication

### Login

Authenticates a user and creates a session.

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (Success - 200 OK):**
```json
{
  "success": true,
  "errorMessage": "",
  "sessionId": "string",
  "authorizedScopes": ["string"],
  "expiresAt": "2024-01-01T00:00:00Z",
  "user": {
    "userId": "string",
    "username": "string",
    "role": "string"
  }
}
```

**Response (Failed - 403 Forbidden):**
```json
{
  "success": false,
  "errorMessage": "Invalid username or password!",
  "sessionId": "",
  "authorizedScopes": [],
  "expiresAt": "2024-01-01T00:00:00Z"
}
```

**Cookies Set:**
- Session cookie with 900 seconds expiry on successful login

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `403 Forbidden` - Invalid credentials or locked account
- `500 Internal Server Error` - Server error

**Notes:**
- Account locks after multiple failed login attempts (max attempts configured in system)
- Session expires after 900 seconds (15 minutes) of inactivity
- Client IP is extracted from `X-Forwarded-For` header for audit logging

---

## User Management

### List Users

Retrieves a paginated list of user accounts.

**Endpoint:** `GET /api/v1/users`

**Query Parameters:**
- `page` (integer, optional) - Page number (default: 1)
- `limit` (integer, optional) - Items per page (default: 10)
- `sortField` (string, optional) - Field to sort by (allowed fields: defined in system)
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- Must be from the system's allowed user filter fields list
- `locked` field accepts single boolean value only

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "users": [
    {
      "userId": "string",
      "username": "string",
      "email": "string",
      "displayName": "string",
      "role": "string",
      "locked": false,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters or filter fields
- `500 Internal Server Error` - Database error

---

### Create User

Creates a new user account and sends account setup email.

**Endpoint:** `POST /api/v1/users`

**Request Body:**
```json
{
  "username": "string",
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `username`: Must be valid username format
- `email`: Must be valid email format
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`
- `displayName`: Optional

**Response (201 Created):**
```json
{
  "username": "string",
  "userId": "string"
}
```

**Response Headers (Development Mode):**
- `Account-Setup-Id`: UUID for account setup (only in development mode)

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `409 Conflict` - Username or email already exists
- `500 Internal Server Error` - Server error

**Notes:**
- New accounts are locked by default until account setup is completed
- Account setup email is sent with verification link
- In development mode with mock email enabled, setup UUID is returned in response header

---

### Update User

Updates user account information.

**Endpoint:** `PUT /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "email": "string",
  "displayName": "string",
  "role": "string"
}
```

**Validation Rules:**
- `email`: Required, must be valid email format
- `displayName`: Optional, max 255 characters
- `role`: Required. Must be one of: `admin`, `maintainer`, `developer`, `guest`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `500 Internal Server Error` - Database error

---

### Delete User

Deletes a user account.

**Endpoint:** `DELETE /api/v1/users/{id}`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Database error

---

### Update User Email

Updates a user's email address.

**Endpoint:** `PUT /api/v1/users/{id}/email`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "email": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `email`: Must be valid email format

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Update User Display Name

Updates a user's display name.

**Endpoint:** `PUT /api/v1/users/{id}/display-name`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "displayName": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- `displayName`: Required (cannot be empty)

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid payload or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Change Password

Changes a user's password using recovery/setup process.

**Endpoint:** `PUT /api/v1/users/{id}/password`

**Path Parameters:**
- `id` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "recoveryId": "string",
  "oldPassword": "string",
  "password": "string"
}
```

**Validation Rules:**
- `userId` in body must match `id` in path
- Password must meet system password requirements

**Response (200 OK):**
```json
{
  "invalidId": false,
  "expired": false,
  "oldPasswordDiff": false,
  "invalidUserAccount": false,
  "changed": true
}
```

**Response Fields:**
- `invalidId`: Recovery ID is invalid or not found
- `expired`: Recovery link has expired
- `oldPasswordDiff`: Old password doesn't match
- `invalidUserAccount`: User account not found
- `changed`: Password successfully changed

**Error Responses:**
- `400 Bad Request` - Invalid request body or mismatched user IDs
- `500 Internal Server Error` - Database error

---

### Lock User Account

Locks a user account, preventing login.

**Endpoint:** `PUT /api/v1/users/{id}/lock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Account is already locked
- `500 Internal Server Error` - Database error

---

### Unlock User Account

Unlocks a user account.

**Endpoint:** `PUT /api/v1/users/{id}/unlock`

**Path Parameters:**
- `id` (string, required) - User ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `409 Conflict` - Cannot unlock new accounts pending verification
- `500 Internal Server Error` - Database error

**Notes:**
- New accounts locked for verification cannot be unlocked manually
- Must complete account setup process instead

---

### Validate Username/Email

Checks availability of username and/or email.

**Endpoint:** `POST /api/v1/users/validate`

**Request Body:**
```json
{
  "username": "string",
  "email": "string"
}
```

**Validation Rules:**
- At least one of `username` or `email` must be provided

**Response (200 OK):**
```json
{
  "usernameAvailable": true,
  "emailAvailable": true
}
```

**Error Responses:**
- `400 Bad Request` - Both username and email are empty
- `500 Internal Server Error` - Database error

---

### Validate Password

Validates a password against system requirements.

**Endpoint:** `POST /api/v1/users/validate-password`

**Request Body:**
```json
{
  "password": "string"
}
```

**Response (200 OK):**
```json
{
  "isValid": true,
  "msg": "Password is valid"
}
```

**Error Responses:**
- `400 Bad Request` - Unable to parse request

---

### Get Account Setup Info

Retrieves information for account setup/verification.

**Endpoint:** `GET /api/v1/users/account-setup/{uuid}`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID from email

**Response (200 OK):**
```json
{
  "id": "string",
  "userId": "string",
  "username": "string",
  "email": "string",
  "role": "string",
  "displayName": "string"
}
```

**Error Responses:**
- `404 Not Found` - Setup link is invalid or already used
- `500 Internal Server Error` - Database error

---

### Complete Account Setup

Completes the account setup process for new users.

**Endpoint:** `POST /api/v1/users/account-setup/{uuid}/complete`

**Path Parameters:**
- `uuid` (string, required) - Account setup UUID

**Request Body:**
```json
{
  "uuid": "string",
  "userId": "string",
  "username": "string",
  "displayName": "string",
  "password": "string"
}
```

**Validation Rules:**
- `uuid` in body must match `uuid` in path
- `username`: Must be valid username format
- `password`: Must meet system password requirements
- `userId`: Required

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body or validation error
- `500 Internal Server Error` - Database error

**Notes:**
- Unlocks the account upon successful completion
- Removes the account recovery record
- User can login after completing this step

---

### Get Current User

Retrieves the currently authenticated user's information.

**Endpoint:** `GET /api/v1/users/me`

**Status:** Not yet implemented

---

### Update Current User

Updates the currently authenticated user's information.

**Endpoint:** `PUT /api/v1/users/me`

**Status:** Not yet implemented

---

## Namespace Management

Namespaces are used to organize repositories. They can be associated with teams or projects.

### List Namespaces

Retrieves a paginated list of namespaces.

**Endpoint:** `GET /api/v1/access/namespaces`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order: `asc` or `desc`
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- Other namespace-specific fields as defined in system

**Response (200 OK):**
```json
{
  "total": 50,
  "page": 1,
  "limit": 10,
  "namespaces": [
    {
      "id": "string",
      "registryId": "string",
      "name": "string",
      "purpose": "project",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Namespace

Creates a new namespace with assigned maintainers.

**Endpoint:** `POST /api/v1/access/namespaces`

**Request Body:**
```json
{
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "maintainers": ["userId1", "userId2"]
}
```

**Validation Rules:**
- `name`: Must be valid namespace format
- `purpose`: Required. Must be `project` or `team`
- `maintainers`: Required. At least one maintainer must be specified
- All maintainers must have active accounts with 'Maintainer' role

**Response (201 Created):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid maintainer roles
- `409 Conflict` - Namespace name already exists
- `500 Internal Server Error` - Server error

**Notes:**
- Maintainers automatically receive maintainer-level access to the namespace
- Namespace names must be unique within the registry

---

### Check Namespace Exists

Checks if a namespace exists by identifier (ID or name).

**Endpoint:** `HEAD /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response:**
- `200 OK` - Namespace exists
- `404 Not Found` - Namespace does not exist
- `500 Internal Server Error` - Server error

---

### Get Namespace

Retrieves detailed information about a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "name": "string",
  "purpose": "project",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### Update Namespace

Updates namespace information.

**Endpoint:** `PUT /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "id": "string",
  "description": "string",
  "purpose": "project"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Purpose defaults to existing value if not provided

---

### Delete Namespace

Deletes a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Namespace State

Changes the state of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/state`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Changing to non-active state affects associated repositories

---

### Change Namespace Visibility

Changes the visibility (public/private) of a namespace.

**Endpoint:** `PATCH /api/v1/access/namespaces/{identifier}/visibility`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility in disabled state
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace is in disabled state
- Changing to private affects associated repositories

---

### List Namespace Users

Lists users with access to a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 25,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "namespace",
      "accessLevel": "maintainer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

### Grant Namespace Access

Grants a user access to a namespace.

**Endpoint:** `POST /api/v1/access/namespaces/{identifier}/users`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `namespace`
- `accessLevel`: Must be `guest`, `developer`, or `maintainer`
- `identifier` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or namespace not found
- `500 Internal Server Error` - Server error

---

### Revoke Namespace Access

Revokes a user's access to a namespace.

**Endpoint:** `DELETE /api/v1/access/namespaces/{identifier}/users/{userID}`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "namespace"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Namespace not found
- `500 Internal Server Error` - Server error

---

### List Namespace Repositories

Lists all repositories within a namespace.

**Endpoint:** `GET /api/v1/access/namespaces/{identifier}/repositories`

**Path Parameters:**
- `identifier` (string, required) - Namespace ID or name

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 30,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

---

## Repository Management

Repositories store container images within namespaces.

### List Repositories

Retrieves a paginated list of repositories.

**Endpoint:** `GET /api/v1/access/repositories`

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order
- `filters` (object, optional) - Filter criteria

**Allowed Filter Fields:**
- `is_public` - Boolean value (single value only)
- `tags` - Range filter with format `>value` or `<value` (max 2 values)
- Other repository-specific fields

**Response (200 OK):**
```json
{
  "total": 100,
  "page": 1,
  "limit": 10,
  "repositories": [
    {
      "id": "string",
      "registryId": "string",
      "namespaceId": "string",
      "name": "string",
      "description": "string",
      "isPublic": true,
      "state": "active",
      "tagCount": 5,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

---

### Create Repository

Creates a new repository within a namespace.

**Endpoint:** `POST /api/v1/access/repositories`

**Request Body:**
```json
{
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "createdBy": "string"
}
```

**Validation Rules:**
- `name`: Must be valid repository format
- `namespaceId`: Required, must be valid namespace
- `createdBy`: Required, username of creator

**Response (200 OK):**
```json
{
  "id": "string"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request or invalid namespace
- `409 Conflict` - Repository with same identifier already exists
- `500 Internal Server Error` - Server error

---

### Check Repository Exists

Checks if a repository exists by ID.

**Endpoint:** `HEAD /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response:**
- `200 OK` - Repository exists
- `404 Not Found` - Repository does not exist
- `500 Internal Server Error` - Server error

---

### Get Repository

Retrieves detailed information about a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
```json
{
  "id": "string",
  "registryId": "string",
  "namespaceId": "string",
  "name": "string",
  "description": "string",
  "isPublic": true,
  "state": "active",
  "tagCount": 5,
  "createdBy": "string",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Update Repository

Updates repository information.

**Endpoint:** `PUT /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "description": "string"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request body
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

### Delete Repository

Deletes a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Response (200 OK):**
Empty response body

**Error Responses:**
- `500 Internal Server Error` - Server error

---

### Change Repository State

Changes the state of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/state`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `state` (string, required) - New state: `active`, `deprecated`, or `disabled`

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Missing or invalid state parameter
- `403 Forbidden` - State transition not allowed
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**State Transition Rules:**
- Cannot change from `active` to `disabled` directly
- Cannot change state when namespace is disabled
- Cannot change to `active` when namespace is deprecated

---

### Change Repository Visibility

Changes the visibility (public/private) of a repository.

**Endpoint:** `PATCH /api/v1/access/repositories/{id}/visibility`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `public` (boolean, required) - Set to `true` for public, `false` for private

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid query parameter
- `403 Forbidden` - Cannot change visibility when disabled
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

**Notes:**
- Cannot change visibility when namespace or repository is disabled

---

### List Repository Users

Lists users with access to a repository.

**Endpoint:** `GET /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Query Parameters:**
- `page` (integer, optional) - Page number
- `limit` (integer, optional) - Items per page
- `sortField` (string, optional) - Field to sort by
- `sortOrder` (string, optional) - Sort order

**Response (200 OK):**
```json
{
  "total": 15,
  "page": 1,
  "limit": 10,
  "accesses": [
    {
      "userId": "string",
      "username": "string",
      "resourceId": "string",
      "resourceType": "repository",
      "accessLevel": "developer",
      "grantedBy": "string",
      "grantedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Server error

**Notes:**
- Returns users with access to both the repository and parent namespace

---

### Grant Repository Access

Grants a user access to a repository.

**Endpoint:** `POST /api/v1/access/repositories/{id}/users`

**Path Parameters:**
- `id` (string, required) - Repository ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository",
  "accessLevel": "developer",
  "grantedBy": "string"
}
```

**Validation Rules:**
- `resourceType`: Must be `repository`
- `accessLevel`: Must be `guest` or `developer`
- `id` in URL must match `resourceId` in body

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - Invalid request or URL/body mismatch
- `403 Forbidden` - Cannot override existing access level
- `404 Not Found` - User, granted-by user, or repository not found
- `500 Internal Server Error` - Server error

---

### Revoke Repository Access

Revokes a user's access to a repository.

**Endpoint:** `DELETE /api/v1/access/repositories/{id}/users/{userID}`

**Path Parameters:**
- `id` (string, required) - Repository ID
- `userID` (string, required) - User ID

**Request Body:**
```json
{
  "userId": "string",
  "resourceId": "string",
  "resourceType": "repository"
}
```

**Response (200 OK):**
Empty response body

**Error Responses:**
- `400 Bad Request` - URL identifier doesn't match request
- `404 Not Found` - Repository not found
- `500 Internal Server Error` - Server error

---

## Common Response Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters or body
- `403 Forbidden` - Authentication failed or insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate name)
- `500 Internal Server Error` - Server error

---

## Authentication & Session Management

### Session Details
- Sessions expire after 900 seconds (15 minutes)
- Session cookie is set on successful login
- Failed login attempts are tracked per user
- Accounts lock after exceeding max failed login attempts
- Locked accounts must be unlocked by administrator

### Security Features
- Password hashing with salt
- Failed login attempt tracking
- Account locking mechanisms
- Session management with expiry
- Client IP tracking for audit

---

## Access Control

### Resource Types
- `namespace` - Organizational container for repositories
- `repository` - Container image storage

### Access Levels

**Namespace Access Levels:**
- `maintainer` - Full control over namespace and repositories
- `developer` - Can push/pull images, create repositories
- `guest` - Read-only access

**Repository Access Levels:**
- `developer` - Can push/pull images
- `guest` - Read-only access (pull only)

### User Roles

The system supports the following user roles:

- `admin` - Full administrative access to the system
- `maintainer` - Can maintain namespaces and repositories
- `developer` - Can work with repositories
- `guest` - Read-only access



**Rules:**
- Same format as namespace names
- Allowed characters: alphanumeric, underscore (_), hyphen (-)
- Must be unique within the namespace
- Combined identifier format: `namespace/repository`

**Examples:**
- Valid: `api-service`, `web_app`, `backend-v2`
- Invalid: `my.repo` (dot not allowed), `repo with spaces`, `repo:latest` (colon not allowed)

### Display Name Validation

**Rules:**
- Maximum length: 255 characters
- Can contain any characters
- Optional field (defaults to "Not Set" if empty)

### Image Digest Format

**Format:** `sha256:[0-9a-f]{64}`

**Rules:**
- Must start with `sha256:` prefix
- Followed by 64 hexadecimal characters
- Used for content-addressable image layer identification

**Example:**
- `sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890`

### Resource States

All namespaces and repositories have a state:

- `active` - Resource is active and fully operational
- `deprecated` - Resource is marked as deprecated but still accessible
- `disabled` - Resource is disabled and not accessible

**Important State Transition Rules:**
- Cannot transition directly from `active` to `disabled` (must go through `deprecated`)
- Repository state changes are constrained by parent namespace state
- Cannot activate a repository when namespace is deprecated or disabled
- Cannot change namespace state to disabled when repositories are active
- State changes to non-active values may affect child resources

### Namespace Purpose Types

- `project` - Namespace for a specific project
- `team` - Namespace for team-wide resources

### Timestamp Format

All timestamps in the API follow **ISO 8601** format with timezone information:

**Format:** `YYYY-MM-DDTHH:MM:SS.sssZ` or `YYYY-MM-DDTHH:MM:SS.sss±HH:MM`

**Examples:**
- UTC: `2024-01-15T10:30:45.123Z`
- With timezone: `2024-01-15T10:30:45.123+05:30`

**Fields with timestamps:**
- `createdAt` - Resource creation time
- `updatedAt` - Last modification time
- `expiresAt` - Session/token expiration time
- `grantedAt` - Access grant time
- `last_loggedin_at` - User's last login time

---

## Query Parameters & Filtering

### List Endpoints

All list endpoints support the following query parameters:

**Pagination:**
- `page` (integer) - Page number (default: 1)
- `limit` (integer) - Items per page (default: 10, max: 100)

**Sorting:**
- `sortField` (string) - Field name to sort by (must be from allowed fields)
- `sortOrder` (string) - `asc` or `desc` (default: `asc`)

**Filtering:**
- `filters` (array) - Array of filter objects with field, operator, and values

**Search:**
- `searchTerm` (string) - Full-text search term (searches across relevant text fields)

### Filter Specifications by Resource

**User List Filters:**

Allowed filter fields:
- `role` - Filter by user role (`admin`, `maintainer`, `developer`, `guest`)
- `locked` - Filter by lock status (boolean, single value only)

Allowed sort fields:
- `username` - Sort by username
- `email` - Sort by email address
- `role` - Sort by role
- `display_name` - Sort by display name
- `last_loggedin_at` - Sort by last login timestamp

**Namespace List Filters:**

Allowed filter fields:
- `purpose` - Filter by purpose (`project`, `team`)
- `state` - Filter by state (`active`, `deprecated`, `disabled`)
- `is_public` - Filter by visibility (boolean, single value only)

Allowed sort fields:
- `name` - Sort by namespace name
- `created_at` - Sort by creation timestamp

**Repository List Filters:**

Allowed filter fields:
- `state` - Filter by state (`active`, `deprecated`, `disabled`)
- `is_public` - Filter by visibility (boolean, single value only)
- `tags` - Filter by tag count (range filter: `>10`, `<100`)
- `namespace_id` - Filter by parent namespace ID

Allowed sort fields:
- `name` - Sort by repository name
- `tags` - Sort by tag count
- `created_at` - Sort by creation timestamp

**Resource Access List Filters:**

Allowed filter fields:
- `access_level` - Filter by access level
- `resource_type` - Filter by type (`namespace`, `repository`)
- `resource_id` - Filter by specific resource ID

Allowed sort fields:
- `user` - Sort by username
- `granted_user` - Sort by user who granted access
- `granted_at` - Sort by grant timestamp

### Filter Operators

- `equal` - Exact match
- `contains` - String contains (case-insensitive)
- `in` - Value in list
- `range` - Numeric range (used with `>` or `<` prefix)
- `or` - Logical OR between filters

### Special Filter Rules

**Boolean Filters:**
- Fields like `is_public` and `locked` accept only single boolean values
- Example: `{"field": "is_public", "operator": "equal", "values": [true]}`

**Range Filters:**
- Tag count filter accepts values with `>` or `<` prefix
- Format: `>10` (greater than 10) or `<100` (less than 100)
- Maximum of 2 values: one lower bound and one upper bound
- Cannot have two lower bounds or two upper bounds
- Example: `{"field": "tags", "operator": "range", "values": [">5", "<50"]}`

### Filter Examples

**Filter repositories with more than 10 tags:**
```
GET /api/v1/access/repositories?filters=[{"field":"tags","operator":"range","values":[">10"]}]
```

**Filter public namespaces sorted by creation date:**
```
GET /api/v1/access/namespaces?filters=[{"field":"is_public","operator":"equal","values":[true]}]&sortField=created_at&sortOrder=desc
```

**Filter locked users:**
```
GET /api/v1/users?filters=[{"field":"locked","operator":"equal","values":[true]}]
```

---

## Error Handling

### Error Response Format

```json
{
  "error": "string",
  "message": "string",
  "statusCode": 400
}
```

### Common Error Scenarios

**400 Bad Request:**
- Invalid request body format
- Missing required fields
- Validation errors
- URL/body parameter mismatches

**403 Forbidden:**
- Invalid credentials
- Account locked
- Insufficient permissions
- State transition not allowed

**404 Not Found:**
- Resource does not exist
- User not found
- Invalid recovery link

**409 Conflict:**
- Duplicate resource name
- Account already locked
- Cannot override existing access

**500 Internal Server Error:**
- Database errors
- Transaction failures
- Internal processing errors

---

## Glossary

**Registry** - The top-level container for all namespaces and repositories. In most deployments, this is a single hosted registry instance.

**Namespace** - An organizational unit that groups related repositories. Can represent a project or team.

**Repository** - A collection of container images with the same name but different tags or versions.

**Digest** - A SHA256 hash of content, used for content-addressable storage and integrity verification.

**Manifest** - Metadata describing a container image, including layers and configuration.

**Blob** - A binary large object representing an image layer or configuration.

**Tag** - A human-readable label pointing to a specific image manifest (e.g., `latest`, `v1.0.0`).

**Access Level** - The permission level granted to a user for a resource (guest, developer, maintainer).

**Resource State** - The operational status of a namespace or repository (active, deprecated, disabled).

**Session** - An authenticated user session with a limited lifetime (900 seconds).

**Scope** - The set of permissions associated with an authentication session.

**Grant** - The act of giving a user access to a resource at a specific level.

**Maintainer** - A user role with full control over namespaces and repositories.

**Upstream** - An external registry that can be proxied and cached.

**Identifier** - Can be either a UUID or a human-readable name, used to reference resources.

**Recovery Link** - A time-limited URL used for account setup or password reset.

