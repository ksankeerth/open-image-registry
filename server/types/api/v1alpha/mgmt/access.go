package mgmt

import "time"

type NamespaceAccess struct {
	ID          string     `json:"id"`
	Namespace   string     `json:"namespace"`
	ResourceID  string     `json:"resource_id"`
	UserID      string     `json:"user_id"`
	AccessLevel string     `json:"access_level"`
	GrantedBy   string     `json:"granted_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type RepositoryAccess struct {
	ID          string     `json:"id"`
	Namespace   string     `json:"namespace"`
	Repository  string     `json:"repository"`
	ResourceID  string     `json:"resource_id"`
	UserID      string     `json:"user_id"`
	AccessLevel string     `json:"access_level"`
	GrantedBy   string     `json:"granted_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type UserNamespaceAccessResponse struct {
	Username   string             `json:"username"`
	AccessList []*NamespaceAccess `json:"access_list"`
}

type UserRepositoryAccessResponse struct {
	Username   string              `json:"username"`
	AccessList []*RepositoryAccess `json:"access_list"`
}

type AccessGrantRequest struct {
	UserID       string `json:"user_id"`
	AccessLevel  string `json:"access_level"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	GrantedBy    string `json:"granted_by"`
}

type AccessRevokeRequest struct {
	UserID       string `json:"user_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

type ResourceAccessViewDTO struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resource_type"`
	ResourceName string    `json:"resource_name"`
	ResourceID   string    `json:"resource_id"`
	AccessLevel  string    `json:"access_level"`
	UserId       string    `json:"user_id"`
	Username     string    `json:"username"`
	GrantedUser  string    `json:"granted_user"` // granted username
	GrantedBy    string    `json:"granted_by"`   // granted_user_id
	GrantedAt    time.Time `json:"granted_at"`
}