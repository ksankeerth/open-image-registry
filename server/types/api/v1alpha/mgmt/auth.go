package mgmt

import "time"

type UserProfileInfo struct {
	UserId       string              `json:"user_id"`
	Username     string              `json:"username"`
	Role         string              `json:"role"`
	Namespaces   []*NamespaceAccess  `json:"namespaces"`
	Repositories []*RepositoryAccess `json:"repositories"`
}

type AuthLoginResponse struct {
	Success          bool            `json:"success"`
	ErrorMessage     string          `json:"error_message"`
	SessionId        string          `json:"session_id"`
	AuthorizedScopes []string        `json:"authorized_scopes"`
	ExpiresAt        time.Time       `json:"expires_at"`
	User             UserProfileInfo `json:"user"`
}

type AuthLoginRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Scopes   []string `json:"scopes"`
}