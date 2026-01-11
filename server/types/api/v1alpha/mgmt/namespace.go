package mgmt

import "time"

type CreateNamespaceRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsPublic    bool     `json:"is_public"`
	Purpose     string   `json:"purpose"`
	Maintainers []string `json:"maintainers"`
}

type CreateNamespaceResponse struct {
	Id string `json:"id"`
}

type NamespaceResponse struct {
	ID          string     `json:"id"`
	RegistryID  string     `json:"registry_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Purpose     string     `json:"purpose"`
	IsPublic    bool       `json:"is_public"`
	State       string     `json:"state"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type UpdateNamespaceRequest struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
}

type NamespaceViewDTO struct {
	RegistryID  string     `json:"registry_id"`
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	IsPublic    bool       `json:"is_public"`
	Purpose     string     `json:"purpose"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	Developers  []string   `json:"developers"`
	Maintainers []string   `json:"maintainers"`
}
