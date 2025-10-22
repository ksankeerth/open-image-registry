package mgmt

import "time"

type NamespaceAccess struct {
	ID          string    `json:"id"`
	Namespace   string    `json:"namespace"`
	ResourceID  string    `json:"resource_id"`
	UserID      string    `json:"user_id"`
	AccessLevel string    `json:"access_level"`
	GrantedBy   string    `json:"granted_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type RepositoryAccess struct {
	ID          string    `json:"id"`
	Namespace   string    `json:"namespace"`
	Repository  string    `json:"repository"`
	ResourceID  string    `json:"resource_id"`
	UserID      string    `json:"user_id"`
	AccessLevel string    `json:"access_level"`
	GrantedBy   string    `json:"granted_by"`
	CreatedAt   time.Time `json:"created_at"`
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