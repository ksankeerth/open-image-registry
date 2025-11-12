package models

import "time"

type NamespaceModel struct {
	RegistryId  string
	Id          string
	Name        string
	Description string
	State       string
	Purpose     string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type RepositoryModel struct {
	ID         string
	RegistryID  string
	NamespaceID string
	Name        string
	Description string
	IsPublic    bool
	State       string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type ImageBlobMetaModel struct {
	NamespaceID  string
	RegistryID   string
	RepositoryID string
	Digest       string
	Size         int
	Location     string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type ImageManifestModel struct {
	ID           string
	Digest       string
	Size         int
	MediaType    string
	Content      string
	NamespaceID  string
	RegistryID   string
	RepositoryID string
	UniqueDigest string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type ImageTagModel struct {
	Id           string
	NamespaceId  string
	RegistryId   string
	RepositoryId string
	Tag          string
	IsStable     bool
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type RegistryCacheModel struct {
	NamespaceID  string
	RegistryID   string
	RepositoryID string
	Identifier   string
	Digest       string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type ResourceAccessModel struct {
	Id           string
	ResourceType string
	UserId       string
	AccessLevel  string
	GrantedBy    string
	GrantedAt    time.Time
}
