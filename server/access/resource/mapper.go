package resource

import "net/http"

// MapToHTTP provides a standardized HTTP status and message.
// The isBulk parameter adjusts the message to be more accurate for multiple users.
func (g AccessOpFailure) MapToHTTP(isBulk bool) (int, string) {
	switch g {
	case Success:
		if isBulk {
			return http.StatusOK, "Access changes processed for all users successfully"
		}
		return http.StatusOK, "Access change processed successfully"

	case InitiatorNotFound:
		return http.StatusNotFound, "The person initiating this request was not found"

	case InitiatorNotAdmin:
		return http.StatusForbidden, "Only administrators are authorized to perform this operation"

	case GranteeNotFound:
		if isBulk {
			return http.StatusNotFound, "One or more target users were not found"
		}
		return http.StatusNotFound, "The target user was not found"

	case GranteeIsAdmin:
		if isBulk {
			return http.StatusForbidden, "One or more target users are administrators; their access levels cannot be modified"
		}
		return http.StatusForbidden, "The target user is an administrator; their access level cannot be modified"

	case ResourceNotFound:
		return http.StatusNotFound, "The requested resource does not exist"

	case ResourceDisabled:
		return http.StatusForbidden, "This resource is currently disabled and permissions cannot be changed"

	case ExceedsRole:
		if isBulk {
			return http.StatusForbidden, "One or more users do not have the base role required for this access level"
		}
		return http.StatusForbidden, "The user does not have the base role required for this access level"

	case Conflict:
		if isBulk {
			return http.StatusConflict, "One or more users already have a different access level that conflicts with this request"
		}
		return http.StatusConflict, "The user already has a different access level that conflicts with this request"

	case HasSameAccessAlready:
		if isBulk {
			return http.StatusOK, "All users already have this exact access level; no changes made"
		}
		return http.StatusOK, "The user already has this exact access level; no changes made"

	case NotAllowedAccessLevel:
		return http.StatusForbidden, "The requested access level is not allowed for this resource type"

	case RedundantAccess:
		return http.StatusConflict, "User already has the requested access via inheritance"

	case AccessNotFound:
		if isBulk {
			return http.StatusNotFound, "One or more access records to be revoked were not found"
		}
		return http.StatusNotFound, "The specified access record does not exist"

	case CannotRevokeSelf:
		return http.StatusForbidden, "Security Policy: You cannot revoke your own administrative permissions"

	case UnexpectedError:
		return http.StatusInternalServerError, "An unexpected system error occurred"

	default:
		return http.StatusInternalServerError, "Unknown operation error"
	}
}
