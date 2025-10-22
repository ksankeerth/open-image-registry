package models

import "time"

type ResourceAccess struct {
	Id           string
	ResourceType string
	ResourceId   string
	UserId       string
	AccessLevel  string
	GrantedBy    string
	CreatedAt    time.Time
	UpdatedAt    *time.Time // optional
}

type NamespaceAccess struct {
	ID          string
	Namespace   string
	ResourceID  string
	UserID      string
	AccessLevel string
	GrantedBy   string
	CreatedAt   time.Time
	UpdatedAt   *time.Time // optional
}

type RepositoryAccess struct {
	ID          string
	Namespace   string
	Repository  string
	ResourceID  string
	UserID      string
	AccessLevel string
	GrantedBy   string
	CreatedAt   time.Time
	UpdatedAt   *time.Time // optional
}