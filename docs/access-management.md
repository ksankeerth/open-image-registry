# Access Management Guide - Open Image Registry

> **Note:** This is a draft specification document that serves as both a guide for implementation and a reference for refining the backend to match this spec. The specification may evolve if changes prove to be more intuitive and practical during development.

## Table of Contents
- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Roles & Permissions](#roles--permissions)
- [Access Levels Explained](#access-levels-explained)
- [Resource Hierarchy](#resource-hierarchy)
- [Access Control Rules](#access-control-rules)
- [Common Operations](#common-operations)
- [Database Schema](#database-schema)
- [Implementation Examples](#implementation-examples)
- [Best Practices](#best-practices)

---

## Overview

OpenImageRegistry is a horizontally scalable Docker registry with built-in WebUI that supports:
- **Hosting** Docker images in namespaces and repositories
- **Proxying** and caching images from upstream registries
- **Multi-tenant** namespace-based access control

### Key Features
- Fine-grained access control at namespace and repository levels
- Support for team-based or project-based logical grouping
- Hierarchical permission inheritance with override rules
- Tag stability management (stable vs unstable tags)
- Public/private namespace visibility control
- Repository state management

---

## Core Concepts

### 1. Resource Hierarchy

```
Namespace (public/private, active/deprecated/disabled)
  ├── Repository 1 (active/deprecated/disabled)
  │     ├── Tag 1 (stable/unstable)
  │     ├── Tag 2
  │     └── Tag 3
  ├── Repository 2
  │     └── ...
  └── Repository N

Upstream Registry (separate hierarchy)
  └── Proxy + Cache configuration
```

### 2. Access Model

**Global Role + Resource-Level Access**

- Each user has **ONE global role** (Admin, Maintainer, Developer, Guest, Machine)
- **Roles** determine what opportunities or capabilities a user has
- **Access levels** on resources (namespace/repository) determine what a specific user can actually do
- Users can have **different access levels** across namespaces/repositories
- Example: A user with **Maintainer role** can be:
  - **Maintainer** of Namespace A (full control)
  - **Developer** on Namespace B (limited access)
  - **Guest** on Namespace C (read-only)

---

## Roles & Permissions

### Understanding Roles vs Access Levels

**Role (Global):**
- Defines the maximum capabilities a user can have
- Think of it as a "license" or "qualification"
- Does NOT automatically grant access to any specific resource

**Access Level (Resource-specific):**
- Determines actual permissions on a namespace or repository
- Must be explicitly granted
- Cannot exceed what the user's global role allows

**Example:**
```
User: alice
Global Role: Developer (has the capability to be a developer)
Access Levels:
  - platform-eng namespace: developer access (can push/pull)
  - ml-models namespace: guest access (can only pull)
  - data-pipelines namespace: no access (cannot access at all)
```

### Admin Role
**System-wide administrator**

| Permission | Description |
|------------|-------------|
| Create/Delete Namespaces | Full namespace lifecycle control |
| Manage All Resources | Access to all namespaces without explicit assignment |
| Self-Assign as Maintainer | Can assign themselves to any namespace |
| Grant Any Access | Can grant guest/developer/maintainer access |
| Change Lifecycle States | active → deprecated → disabled → active |
| Delete Stable Tags | Can remove any tag |

### Maintainer Role
**Namespace-level administrator**

| Permission | Description |
|------------|-------------|
| Create Repositories | Within assigned namespaces |
| Delete Repositories | Within assigned namespaces |
| Configure Auto-Create | Control if developers can auto-create repos on push |
| Grant Access | Give developer/guest access within their namespaces |
| Manage Lifecycle | Change namespace states (active/deprecated/disabled) |
| Mark Tags as Stable | Protect tags from developer deletion |
| Delete Stable Tags | Can remove stable tags |
| Repository Access Control | Grant specific access to repositories |

**Prerequisites:**
- User must have **Maintainer** or **Admin** global role to be assigned as maintainer of a namespace

**Important Note:**
- ❌ Users with maintainer access at namespace level **CANNOT** be granted any access at repository level (redundant - they already have full control)

### Developer Role
**Basic push/pull access**

| Permission | Description |
|------------|-------------|
| Push Images | Upload new images and create tags |
| Pull Images | Download images |
| Delete Unstable Tags | Remove tags not marked as stable |
| Auto-Create Repository | Only if enabled by maintainer (disabled by default) |

**Prerequisites:**
- User must have **Developer** or **Maintainer** global role to be granted developer access to a namespace

**Limitations:**
- ❌ Cannot delete stable tags
- ❌ Cannot create repositories (unless auto-create is enabled)
- ❌ Cannot manage access
- ❌ Cannot change lifecycle states

### Guest Role
**Read-only access**

| Permission | Description |
|------------|-------------|
| Pull Images | Download images from authorized namespaces/repositories |

**Current Limitations:**
- ⚠️ No UI access (dedicated guest portal to be developed)
- Can be granted access at **namespace** or **repository** level
- Granted by Admin or Maintainer

### Machine Role
**Automation/CI-CD accounts**

⚠️ **Future Implementation (TBD)**
- Authentication mechanism (token/API key)
- Authorization model
- Who can create machine accounts
- Access level assignment

---

## Access Levels Explained

### Namespace-Level Access

When a user is granted access at the namespace level, this access is **inherited by all repositories** under that namespace by default.

**Available Access Levels:**
- **Maintainer**: Full control over the namespace and all repositories
- **Developer**: Can push/pull to all repositories in the namespace
- **Guest**: Can pull from all repositories in the namespace (read-only)

**Inheritance Example:**
```
User: alice
Namespace: platform-eng (developer access)
Result:
  ✅ Can push/pull to platform-eng/api-gateway
  ✅ Can push/pull to platform-eng/frontend
  ✅ Can push/pull to platform-eng/database
  (All repositories inherit developer access)
```

### Repository-Level Access

Repository-level access allows **fine-grained control** for specific repositories within a namespace.

**Available Access Levels:**
- **Developer**: Can push/pull to this specific repository
- **Guest**: Can pull from this specific repository (read-only)
- ❌ **No Maintainer access** at repository level

**Common Use Case:**
```
Scenario: Team with restricted project access
  - All team members: guest access at namespace (view-only for everything)
  - Project team members: developer access at repository (push/pull for their project)

Example:
  User: bob
  Namespace: platform-eng (guest access)
  Repository: platform-eng/mobile-app (developer access)
  
  Result:
    ✅ Bob can pull from all repositories in platform-eng
    ✅ Bob can push/pull to platform-eng/mobile-app specifically
    ❌ Bob cannot push to other repositories in platform-eng
```

### Access Override Rules

1. **Namespace developer/maintainer access CANNOT receive repository-level access**
   - It's redundant since they already have access to all repositories
   - System will block these attempts

2. **Namespace guest access CAN be elevated at repository level**
   - Guest → Developer for specific repositories
   - Allows selective promotion for specific projects

**Example:**
```
❌ REDUNDANT (Blocked):
  User has: namespace developer access
  Attempt: Grant repository developer access
  Result: DENIED - already has access via namespace

✅ ALLOWED (Upgrade):
  User has: namespace guest access
  Attempt: Grant repository developer access
  Result: ALLOWED - elevates access for this specific repository
```

---

## Resource Hierarchy

### Namespace
**Logical grouping for teams, projects, or tenants**

**Properties:**
- **Name**: Must follow specific format/naming rules
- **Description**: Human-readable description
- **Visibility**: public | private
- **Maintainer(s)**: One or more maintainers
- **State**: active | deprecated | disabled

**Visibility:**
- **Public**: Visible to all users (but access still requires explicit grants)
- **Private**: Only visible to users with explicit access

**Lifecycle States:**

```
┌─────────┐
│ ACTIVE  │ ←──────────────┐
└────┬────┘                 │
     │                      │
     ↓                      │
┌────────────┐              │
│ DEPRECATED │              │
│ (pull only)│              │
└─────┬──────┘              │
      │                     │
      ↓                     │
┌──────────┐                │
│ DISABLED │ ───────────────┘
└──────────┘
```

- **Active**: Full read/write access
- **Deprecated**: Pull-only (no push/delete allowed)
- **Disabled**: No access (except admin)
  - When namespace is disabled, repository states cannot be changed
  - Repository state is effectively the same as namespace state

**Namespace Creation:**
- Only **Admin** can create namespaces
- Only **Admin** can delete namespaces
- Admin can self-assign as maintainer (but has full access without assignment)

**Admin's Common Tasks:**
```
✓ Create namespace "team-a"
✓ Set visibility (public/private)
✓ Assign maintainers to "team-a"
✓ View maintainers of "team-a"
✓ Remove maintainer from "team-a"
✓ Change lifecycle state
✓ Delete namespace
```

### Repository
**Container for related Docker images**

**Properties:**
- **Name**: Repository identifier
- **Namespace**: Parent namespace
- **State**: active | deprecated | disabled
- **Access Control**: Inherited from namespace + specific grants

**Repository States:**
- **Active**: Full read/write access (if user has permission)
- **Deprecated**: Pull-only (no push/delete allowed)
- **Disabled**: No access

**State Dependencies:**
- If namespace is **deprecated or disabled**, repository state is considered the same as namespace state
- If namespace is **disabled**, repository state cannot be changed
- Repository state can be managed independently only when namespace is **active**

**Repository Creation:**
- **Maintainers** can create repositories
- **Developers** can auto-create repositories **only if** maintainer enabled it (disabled by default)

**Access Control at Repository Level:**

Namespace-level access is inherited when creating a repository, but it's possible to grant additional access at the repository level for fine-grained control.

**Common Pattern:**
```
Example: Team-based access with project restrictions
  
  Namespace: engineering (all team members have guest access)
  
  Repositories:
    - frontend (team A: developer access)
    - backend (team B: developer access)
    - mobile (team C: developer access)
  
  Result:
    ✅ All team members can view/pull all repositories
    ✅ Each team can push only to their assigned repository
```

**Important:** Repository-level access can only be **Guest** or **Developer**. No maintainer access at repository level.

### Tag
**Specific version of a Docker image**

**Properties:**
- **Tag Name**: Version identifier (e.g., v1.0.0, latest, dev)
- **Stable Flag**: Boolean (marked by maintainer/admin)

**Tag Stability:**
- **Unstable (default)**: Can be deleted by developers
- **Stable (marked)**: Protected from developer deletion
  - Only **Admin** or **Maintainer** can mark as stable
  - Only **Admin** or **Maintainer** can delete stable tags
  - Prevents accidental deletion of production images

---

## Access Control Rules

### Rule 1: Role Prerequisites

Before granting resource-level access, verify global role:

| Resource Access | Required Global Role |
|-----------------|---------------------|
| Maintainer of Namespace | Admin OR Maintainer |
| Developer of Namespace | Developer OR Maintainer OR Admin |
| Guest of Namespace | Any role |

**Explanation:**
- A user's global role determines what access levels they're **eligible** to receive
- Access must still be explicitly granted by admin or maintainer

### Rule 2: Access Hierarchy & Inheritance

**Namespace-Level Access Inheritance:**

When a user has **developer access at namespace level**:
- ✅ Automatically has developer access to **all repositories** under that namespace
- ❌ **Cannot** be granted developer or guest access at repository level (redundant and blocked by system)
- The namespace-level access takes precedence

**Repository-Level Override:**

When a user has **guest (read-only) access at namespace level**:
- ✅ **Can** be granted developer access to **selected repositories** under that namespace
- This allows fine-grained promotion: namespace guest → repository developer for specific repos

**Access Level Restrictions:**

```
Repository Level Access Limitations:
├─→ Only GUEST or DEVELOPER access allowed
├─→ No MAINTAINER access at repository level
└─→ Users with namespace-level developer/maintainer access CANNOT receive repository-level access (redundant)
```

### Rule 3: Namespace Maintainer Blocking Rule

**Critical Rule:** Users with **maintainer access at namespace level** cannot be granted any access at repository level.

**Enforcement:**
- System must **block** attempts to grant repository-level access to namespace maintainers
- This applies to attempts by both admins and other maintainers
- Rationale: Maintainers already have full control at namespace level

**Example Scenarios:**

❌ **BLOCKED:**
```
User: alice
Namespace: platform-eng (alice is maintainer)
Attempt: Grant alice developer access to platform-eng/nginx-proxy
Result: DENIED - alice is already namespace maintainer
```

✅ **ALLOWED:**
```
User: bob
Namespace: platform-eng (bob has guest access)
Attempt: Grant bob developer access to platform-eng/nginx-proxy
Result: ALLOWED - bob can have elevated access to specific repository
```

### Rule 4: Permission Inheritance

**Namespace-level access applies to all repositories by default.**

**Exception:** Repository-level grants can elevate guest access to developer access for specific repositories.

### Rule 5: State-Based Access Control

**Namespace State Impact:**
- **Active**: All access levels work as normal
- **Deprecated**: Only pull operations allowed (developers cannot push)
- **Disabled**: No access except for admins

**Repository State Impact:**
- **Active**: All access levels work as normal (if namespace is active)
- **Deprecated**: Only pull operations allowed
- **Disabled**: No access
- **If namespace is deprecated/disabled**: Repository state is considered the same as namespace state

### Rule 6: Admin Override

**Admin users:**
- Have implicit full access to all namespaces (without explicit assignment)
- Can assign themselves as maintainer to any namespace
- Can bypass most restrictions (except tag stability for safety)

---

## Common Operations

### Access Grant Validation

Before granting access, the system performs these checks:

```
Grant Access Request
    │
    ├─→ Target: Repository Level?
    │       │
    │       ├─→ User has namespace maintainer access?
    │       │       │
    │       │       └─→ Yes → ❌ DENY (cannot grant repository access to namespace maintainer)
    │       │
    │       ├─→ User has namespace developer access?
    │       │       │
    │       │       └─→ Yes → ❌ DENY (already has access via namespace)
    │       │
    │       └─→ Access level is maintainer?
    │               │
    │               └─→ Yes → ❌ DENY (no maintainer access at repository level)
    │
    └─→ Proceed with access grant
```

### Push Operation Workflow

When a developer pushes an image: `docker push registry.example.com/namespace/repository:tag`

**Access Checks (in order):**

1. **Check namespace state**
   - If deprecated or disabled → DENY push

2. **Check repository state**
   - If deprecated or disabled → DENY push

3. **Check namespace-level developer access**
   - If user has developer access at namespace: ✅ Access granted (inherited to all repositories)

4. **Check repository-level access** (if no namespace developer access)
   - User has explicit repository developer access?

5. **Tag stability check**
   - If tag exists and is stable, only maintainers can override

**Decision Tree:**

```
Push Request
    │
    ├─→ Namespace state is deprecated/disabled? → ❌ DENY
    │
    ├─→ Repository state is deprecated/disabled? → ❌ DENY
    │
    ├─→ User is namespace maintainer? → ✅ ALLOW (full control)
    │
    ├─→ User has namespace developer access? → ✅ ALLOW (inherited to all repos)
    │
    ├─→ User has repository developer access? → Continue
    │       │
    │       └─→ No → ❌ DENY (no access)
    │
    └─→ Tag is stable?
            │
            ├─→ Yes → User is maintainer?
            │           │
            │           ├─→ No → ❌ DENY (stable tag protection)
            │           └─→ Yes → ✅ ALLOW
            │
            └─→ No → ✅ ALLOW
```

### Pull Operation Workflow

When a user pulls an image: `docker pull registry.example.com/namespace/repository:tag`

**Access Checks:**

1. Check namespace state (disabled blocks even pulls, except for admin)
2. Check repository state (disabled blocks even pulls)
3. Check repository-specific access (guest or developer)
4. Fall back to namespace access (any level: guest, developer, or maintainer)

**Decision Tree:**

```
Pull Request
    │
    ├─→ User is admin? → ✅ ALLOW (bypass all restrictions)
    │
    ├─→ Namespace state is disabled? → ❌ DENY
    │
    ├─→ Repository state is disabled? → ❌ DENY
    │
    ├─→ User is namespace maintainer? → ✅ ALLOW
    │
    ├─→ User has namespace developer access? → ✅ ALLOW (inherited)
    │
    ├─→ Has repository-level access (guest or developer)? → ✅ ALLOW
    │
    └─→ Has namespace guest access? → ✅ ALLOW
            │
            └─→ No access → ❌ DENY
```

### Access Management Examples

**Scenario 1: Basic Namespace Access**
```
User: alice (global role: Developer)
Action: Grant developer access to namespace "platform-eng"
Result: ✅ Alice can push/pull to ALL repositories in platform-eng
        ❌ Alice CANNOT be granted access at repository level (redundant)
```

**Scenario 2: Guest with Repository Promotion**
```
User: bob (global role: Developer)
Step 1: Grant guest access to namespace "platform-eng"
        → Bob can pull from all repos in platform-eng
Step 2: Grant developer access to repository "platform-eng/critical-service"
        → ✅ ALLOWED - Bob now has developer access to this specific repo
        → Bob remains guest for all other repos in platform-eng
```

**Scenario 3: Maintainer Access Blocking**
```
User: carol (global role: Maintainer)
Step 1: Assign carol as maintainer of namespace "data-eng"
        → Carol has full control over data-eng and all its repositories
Step 2: Attempt to grant carol developer access to "data-eng/etl-pipeline"
        → ❌ BLOCKED - Carol is already namespace maintainer (redundant)
```

**Scenario 4: Team-Based Project Access**
```
Namespace: "engineering" (visibility: private)
Team members: alice, bob, carol, dave

Access Setup:
  - All team members: guest access at namespace
  - alice, bob: developer access to "engineering/frontend"
  - carol, dave: developer access to "engineering/backend"

Result:
  ✅ All can view and pull all repositories
  ✅ alice, bob can push to frontend only
  ✅ carol, dave can push to backend only
  ❌ No one can push to repositories they don't have developer access to
```

**Scenario 5: State-Based Access Control**
```
Namespace: "legacy-apps" (state: deprecated)
User: alice (developer access at namespace)

Action: Push to "legacy-apps/old-api"
Result: ❌ DENIED - namespace is deprecated (pull-only)

Action: Pull from "legacy-apps/old-api"
Result: ✅ ALLOWED - can still pull
```

### List Namespaces

**As Admin:**
- See ALL namespaces (public and private)

**As Maintainer/Developer/Guest:**
- See only authorized namespaces
- Public namespaces appear in browse/search but access still requires explicit grants

---

## Best Practices

### 1. Access Granting Strategy

- **Start broad, refine as needed:**
  - Grant guest access at namespace level for general visibility
  - Elevate to developer access at repository level for specific responsibilities
  
- **Avoid redundant access grants:**
  - Don't grant repository-level access to users who already have namespace developer access
  - System will block these attempts

### 2. Role vs Access Level

- **Assign appropriate global roles:**
  - Global role determines maximum capability
  - Doesn't automatically grant access to resources
  
- **Grant explicit access levels:**
  - Even admins can be granted specific access levels for clarity
  - Access levels determine actual permissions on resources

### 3. Maintainer Management

- **Namespace maintainers have full control:**
  - No need to grant them repository-level access
  - System enforces this by blocking such attempts

### 4. Repository Access Control

- **Use repository-level access for fine-grained control:**
  - When teams need different access to different projects
  - When namespace has many repositories with varying sensitivity

### 5. State Management

- **Use namespace states appropriately:**
  - **Active**: For current, actively developed projects
  - **Deprecated**: For legacy code that should be preserved but not modified
  - **Disabled**: For archived projects or during maintenance

- **Repository states follow namespace:**
  - When namespace is deprecated/disabled, repositories inherit this restriction
  - Manage repository states independently only when needed


---
