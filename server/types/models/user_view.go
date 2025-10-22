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