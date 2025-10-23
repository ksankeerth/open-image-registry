# User Management Guide - Open Image Registry

## Overview

Open Image Registry provides a secure user account lifecycle management system with multi-stage account setup, password recovery, and account locking mechanisms. This document explains how user accounts are created, managed, and secured.

---

## User Account Lifecycle

### Stage 1: Onboarding (Admin Creates Account)

An administrator creates a new user account by providing:

- **Username** - Unique identifier for the user (can be modified later by the user)
- **Email** - Unique email address for account recovery and notifications
- **Display Name** (Optional) - User's full name or preferred display name (can be set later)
- **Role** - Access level: Admin, Maintainer, Developer, or Guest

#### What Happens Behind the Scenes

1. **Validation** - System checks that the username and email don't already exist in the database
2. **Account Creation** - User account is created with:
   - `USERNAME` and `EMAIL` stored as provided
   - `DISPLAY_NAME` set to "NOT SET" if not provided (satisfies NOT NULL constraint)
   - `LOCKED` status set to `true`
   - `LOCKED_REASON` set to "New Account - Verification Required"
   - `FAILED_ATTEMPTS` set to 0
3. **Invite Sent** - Invitation email sent to the provided email address with a unique setup link
4. **Recovery Entry Created** - A password recovery record is created with:
   - Unique `RECOVERY_UUID`
   - `REASON_TYPE` = 1 (New Account Setup)
   - Link valid for setup completion

#### UI Validation Requirements

Before submitting a new user creation form, the system should:

- Check if the username is available (real-time or on blur)
- Check if the email is available (real-time or on blur)
- Show warning/error messages if either already exists
- Prevent form submission if validation fails

#### API Validation Endpoint

```
POST /api/v1/users/validate
{
  "username": "string (optional)",
  "email": "string (optional)"
}

Response:
{
  "username_available": boolean,
  "email_available": boolean
}
```

---

### Stage 2: Account Setup (User Completes Initial Setup)

The new user receives an email with a setup link and clicks it to:

1. **Enter Password** - Create a secure password (validated for strength)
2. **Set/Confirm Details**:
   - Optionally enter or modify display name
   - Confirm email address
   - (Optional) Modify username if required

#### Password Validation

The system validates passwords for:
- Minimum length (recommended: 12 characters)
- Complexity requirements (uppercase, lowercase, numbers, special characters)
- Not matching username or email
- Not containing common patterns

Validation should occur both in the UI and on the backend API before accepting the password.

#### What Happens After Setup Completion

1. **Account Unlocked** - `LOCKED` status set to `false`
2. **Account Activated** - User can now log in
3. **Recovery Record Cleared** - Password recovery entry deleted from the database
4. **Locked Reason Cleared** - `LOCKED_REASON` and `LOCKED_AT` timestamps cleared

#### Important Note

Account unlock only occurs if the `LOCKED_REASON` was "New Account - Verification Required". Users cannot unlock accounts locked for other reasons through password reset.

---

### Stage 3: Active Account Usage

Once account setup is complete, users can:

- Log in with their username/email and password
- Access resources based on their assigned role
- Change their password anytime
- Update their profile (display name, etc.)

Successful login resets `FAILED_ATTEMPTS` to 0.

---

## Password Management

### Case 1: User-Initiated Password Reset

A user can reset their password only if their account is **active and unlocked**.

**User chooses:**
- "I forgot my password" - Reason: Forgot Password
- "I want to change my password" - Reason: User Initiated Change

**What Happens:**

1. System creates a password recovery entry with the chosen reason
2. Email sent with unique password reset link
3. **Account is NOT locked** during password reset (to prevent abuse)
4. User clicks the link and enters a new password
5. Password recovery record is deleted
6. User can log in with the new password

#### API Endpoint for Password Reset Request

```
POST /api/v1/users/password-reset/request
{
  "email": "string",
  "reason": "forgot" | "change"
}

Response:
{
  "success": true,
  "message": "Password reset email sent"
}
```

#### API Endpoint for Password Validation

```
POST /api/v1/users/password-reset/validate-password
{
  "password": "string",
  "recovery_uuid": "string"
}

Response:
{
  "valid": boolean,
  "errors": {
    "strength": ["Password too weak", "Missing uppercase letter"]
  }
}
```

#### API Endpoint for Password Reset Completion

```
POST /api/v1/users/password-reset/complete
{
  "recovery_uuid": "string",
  "password": "string"
}

Response:
{
  "success": true,
  "message": "Password reset successful"
}
```

---

## Account Locking

### Automatic Locking: Failed Login Attempts

When a user fails to log in:

1. `FAILED_ATTEMPTS` counter increments by 1
2. If `FAILED_ATTEMPTS` reaches 5:
   - Account is automatically locked
   - `LOCKED` status set to `true`
   - `LOCKED_REASON` set to "Failed Login Attempts"
   - `LOCKED_AT` timestamp recorded

**Reset:** One successful login resets `FAILED_ATTEMPTS` to 0.

### Manual Locking: Admin Action

Administrators can manually lock any account:

- `LOCKED` status set to `true`
- `LOCKED_REASON` set to "Manually Locked by Admin"
- `LOCKED_AT` timestamp recorded
- Optional admin note/reason stored (if needed)

#### API Endpoint for Manual Lock

```
POST /api/v1/users/{user_id}/lock
{
  "reason": "string (optional admin note)"
}

Response: 
- HTTP 200 - If successful
- HTTP 409 - If already locked
- HTTP 500 - If any database errors occurred
```

---

## Account Unlocking

### Unlocking Process

Users cannot self-unlock their accounts. Unlock requires:

1. **Admin Action** - Administrator must manually unlock the account using the admin panel
2. **No Verification** - Unlike password reset, no email verification is needed
3. **Account Immediately Active** - User can log in immediately after unlock

#### API Endpoint for Unlock

```
POST /api/v1/users/{user_id}/unlock

Response:
 - HTTP 200 - Success
 - HTTP 409 - New account
 - HTTP 500 - If database errors occurred
```

#### Unlock Restrictions

- Cannot unlock through the password reset process
- Only admins can unlock accounts
- Recommended: Send a notification email to the user after unlock

---

## Account Deletion

### Deletion Rules

Only administrators can delete user accounts.

**What Happens Behind the Scenes**

Instead of removing the user account record from the database, we keep the record with a flag `DELETED` marked as `true`. We also change the `USERNAME` and `EMAIL` with a `[DELETED]` prefix, and the `PASSWORD` and `SALT` will be set to `[DELETED]`.

Additionally, records with `DELETED` set to `true` cannot be fetched using API calls.

---

## Password Recovery Reasons Reference

The system tracks password recovery actions with reason codes for audit and security purposes:

| Reason Type           | Code | Trigger                        | Account Locked? | User Action                           | Outcome                     |
| --------------------- | ---- | ------------------------------ | --------------- | ------------------------------------- | --------------------------- |
| New Account Setup     | 1    | Admin creates account          | Yes             | User sets password via email link     | Account unlocked, activated |
| Forgot Password       | 2    | User requests password reset   | No              | User sets new password via email link | Password changed            |
| User Initiated Change | 3    | User changes password manually | No              | User sets new password via email link | Password changed            |

**Security Note:** Types 2 and 3 do not lock the account to prevent attackers from locking legitimate users out by requesting password resets.

---

## Account Locking Reasons Reference

The system records why an account is locked for transparency and audit purposes:

| Reason                              | Code | Trigger                     | Auto-Unlock?            | Resolution               |
| ----------------------------------- | ---- | --------------------------- | ----------------------- | ------------------------ |
| New Account - Verification Required | 1    | Admin creates account       | Yes (on password setup) | Complete initial setup   |
| Failed Login Attempts (5+)          | 2    | Login failures exceed limit | No                      | Contact admin for unlock |
| Manually Locked by Admin            | 3    | Admin manual action         | No                      | Contact admin for unlock |

**Security Note:** Only reason 1 automatically unlocks. All other locks require explicit admin action.

---

## User Roles

The system supports four role levels with increasing permissions:

- **Guest** - Read-only access to public resources
- **Developer** - Can push/pull images, manage personal namespaces
- **Maintainer** - Can manage repositories and users in their organization
- **Admin** - Full system access, can manage all users and settings

---

## Database Schema Notes

### USER_ACCOUNT Table

- `USERNAME` and `EMAIL` are UNIQUE to prevent duplicates
- `DISPLAY_NAME` cannot be NULL; set to "NOT SET" if not provided
- `LOCKED` is INTEGER (0 = unlocked, 1 = locked)
- `LOCKED_REASON` stores code (1, 2, or 3) matching recovery reason types
- `LOCKED_AT` tracks when the lock occurred for audit purposes
- `FAILED_ATTEMPTS` automatically resets to 0 on successful login

### USER_PASSWORD_RECOVERY Table

- `RECOVERY_UUID` is the primary key (unique recovery link per user)
- `USER_ID` has a UNIQUE constraint to enforce one active recovery per user
- `REASON_TYPE` uses a CHECK constraint to enforce valid codes (1, 2, 3)
- Record is automatically deleted when recovery is completed or user is deleted

### USER_ROLE_ASSIGNMENT Table

- `PRIMARY KEY (USER_ID)` enforces a single role per user (no multi-role support)
- `ON DELETE CASCADE` ensures cleanup when a user is deleted

---

## Security Considerations

1. **Password Hashing** - Always hash passwords with salt before storage (never store plain text)
2. **Rate Limiting** - Implement rate limiting on login to prevent brute force attacks
3. **Recovery Link Expiry** - Password recovery links should expire within 24-48 hours
4. **Audit Logging** - Log all account creation, locking, unlocking, and deletion actions
5. **Email Verification** - Verify email ownership during account creation and changes
6. **Failed Attempt Tracking** - Track and log all failed login attempts with timestamps and IP addresses
7. **Admin Notifications** - Alert admins when accounts are locked due to failed attempts

---

## Implementation Checklist

- [ ] Implement username/email uniqueness validation
- [ ] Implement password strength validation (UI + API)
- [ ] Create password recovery UUID generation (use cryptographically secure random)
- [ ] Implement email sending for invites and password resets
- [ ] Set up recovery link expiry (24-48 hours)
- [ ] Implement failed login attempt tracking
- [ ] Create admin unlock functionality
- [ ] Implement audit logging for all account actions
- [ ] Add rate limiting on login endpoint
- [ ] Create cleanup job to delete expired recovery records
- [ ] Add email verification step during account creation
- [ ] Document password hashing algorithm used

---