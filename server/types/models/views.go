package models

import "time"

type UserAccountView struct {
	Id                        string
	Username                  string
	Email                     string
	DisplayName               string
	Locked                    bool
	LockedReason              int
	LockedAt                  *time.Time
	FailedAttempts            int
	CreatedAt                 time.Time
	UpdatedAt                 *time.Time
	Role                      string
	PasswordRecoveryId        string
	PasswordRecoveryReason    int
	PasswordRecoveryCreatedAt *time.Time
	LastLoggedInAt            *time.Time
}

type ResourceAccessView struct {
	ID           string
	ResourceType string
	ResourceName string
	ResourceID   string
	AccessLevel  string
	UserId       string
	Username     string
	GrantedUser  string // granted username
	GrantedBy    string // granted_user_id
	GrantedAt    time.Time
}

type UpstreamAddressView struct {
	ID          string
	Name        string
	Port        int
	UpstreamUrl string
}
