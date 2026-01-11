package models

import "time"

type UserAccount struct {
	Id             string
	Username       string
	Email          string
	DisplayName    string
	Locked         bool
	LockedReason   int
	FailedAttempts int
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	LockedAt       *time.Time
	LastAccessedAt *time.Time
}

type AccountRecovery struct {
	UserID     string
	RecoveryID string
	ReasonType int
	CreatedAt  time.Time
}