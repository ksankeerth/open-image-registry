package store

import (
	"context"
	"time"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type UserStore interface {
	Create(ctx context.Context, username, email, displayName, password, salt string) (id string, err error)

	Delete(ctx context.Context, userId string) (err error)

	Update(ctx context.Context, userId, displayName string) error

	UpdateEmail(ctx context.Context, userId, newEmail string) (err error)

	UpdateDisplayName(ctx context.Context, userId, displayName string) (err error)

	LockAccount(ctx context.Context, username string, lockedReason int) (err error)

	UnlockAccount(ctx context.Context, username string, resetFailures bool) (err error)

	RecordFailedAttempt(ctx context.Context, username string) error

	CheckAvailability(ctx context.Context, username, email string) (usernameAvail, emailAvail bool, err error)

	// identifier can be either id or username
	Get(ctx context.Context, identifier string) (*models.UserAccount, error)

	GetByUsername(ctx context.Context, username string) (*models.UserAccount, error)

	GetPasswordAndSalt(ctx context.Context, userId string) (password, salt string, err error)

	UpdatePasswordAndSalt(ctx context.Context, userId, password, salt string) (err error)

	ListUserAccounts(ctx context.Context, conditions *ListQueryConditions) (users []*models.UserAccountView, total int, err error)

	GetRole(ctx context.Context, userId string) (role string, err error)

	AssignRole(ctx context.Context, userId, role string) error

	UnAssignRole(ctx context.Context, userId string) error

	AreAccountsActive(ctx context.Context, userIds []string) (valid bool, err error)

	RecordLastAccessedTime(ctx context.Context, userId string, t time.Time) error

	DeleteAllNonAdminAccounts(ctx context.Context) error
}