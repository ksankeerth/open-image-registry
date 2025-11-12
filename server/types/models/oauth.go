package models

import "time"

type OAuthSession struct {
	ID             string
	UserID         string
	ScopeHash      string
	IssuedAt       time.Time
	ExpiresAt      *time.Time
	LastAccessedAt *time.Time
	UserAgent      string
	ClientIP       string
	GrantType      string
}

type ScopeRoleBinding struct {
	ScopeName string
	RoleName  string
}
