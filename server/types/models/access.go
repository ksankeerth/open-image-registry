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
