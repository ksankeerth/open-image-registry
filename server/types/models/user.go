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
}

type PasswordRecovery struct {
	UserId     string
	RecoveryId string
	ReasonType int
	CreatedAt  time.Time
}