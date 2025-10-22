package models

import "time"

type OAuthSession struct {
	SessionId      string
	UserId         string
	ScopeHash      string
	IssuedAt       time.Time
	ExpiresAt      *time.Time
	LastAccessedAt *time.Time
	UserAgent      string
	ClientIp       string
	GrantType      string
}

type ScopeRoleBinding struct {
	ScopeName string
	RoleName  string
}