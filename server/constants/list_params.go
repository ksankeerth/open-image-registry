package constants

var (
	AllowedUserFilterFields = []string{
		"role",
		"locked",
	}
	AllowedUserSortFields = []string{
		"username",
		"email",
		"role",
		"display_name",
		"last_loggedin_at",
	}
)

var (
	AllowedNamespaceFilterFields = []string{"purpose", "state", "is_public"}
	AllowedNamespaceSortFields   = []string{"name", "created_at"}
)

var (
	AllowedRepositoryFilterFields = []string{"state", "is_public", "tags", "namespace_id"}
	AllowedRepositorySortFields   = []string{"name", "tags", "created_at"}
)

var (
	AllowedResourceAccessFilterFields = []string{"access_level", "user_id", "resource_type", "resource_id"}
	AllowedResourceAccessSortFields   = []string{"user", "granted_user", "granted_at"}
)

const FilterFieldNamespaceID = "namespace_id"
const FilterFieldResourceType = "resource_type"
const FilterFieldResourceID = "resource_id"
const FilterFieldIsPublic = "is_public"
const FilterFieldTagCount = "tags"
const FilterFieldRepositoryID = "repository_id"
const FilterFieldAccessLevel = "access_level"
const FilterFieldUserID = "user_id"