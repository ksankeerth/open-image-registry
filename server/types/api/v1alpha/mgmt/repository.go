package mgmt

import "time"

type RepositoryViewDTO struct {
	RegistryID  string     `json:"registry_id"`
	NamespaceID string     `json:"namespace_id"`
	ID          string     `json:"id"`
	Namespace   string     `json:"namespace"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	IsPublic    bool       `json:"is_public"`
	CreatedBy   string     `json:"created_by"`
	TagsCount   int        `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type CreateRepositoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	NamespaceId string `json:"namespace_id"`
	CreatedBy   string `json:"created_by"`
}

type CreateRepositoryResponse struct {
	ID string `json:"id"`
}

type RepositoryResponse struct {
	ID          string     `json:"id"`
	RegistryID  string     `json:"registry_id"`
	NamespaceID string     `json:"namespace_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IsPublic    bool       `json:"is_public"`
	State       string     `json:"state"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type UpdateRepositoryRequest struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}