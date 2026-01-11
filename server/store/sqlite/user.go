package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type userStore struct {
	db *sql.DB
}

func newUserStore(db *sql.DB) *userStore {
	return &userStore{db: db}
}

func (r *userStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (u *userStore) Create(ctx context.Context, username, email, displayName, password, salt string) (id string, err error) {
	q := u.getQuerier(ctx)

	err = q.QueryRowContext(ctx, UserCreateAccountQuery,
		username, email, displayName, password, salt,
	).Scan(&id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create user")
		return "", dberrors.ClassifyError(err, UserCreateAccountQuery)
	}
	return id, nil
}

func (u *userStore) Delete(ctx context.Context, userId string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserDeleteAccountQuery, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete user")
		return dberrors.ClassifyError(err, UserDeleteAccountQuery)
	}
	return nil
}

func (u *userStore) Update(ctx context.Context, userId, displayName string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserUpdateAccountQuery, displayName, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update user")
		return dberrors.ClassifyError(err, UserUpdateAccountQuery)
	}
	return nil
}

func (u *userStore) UpdateEmail(ctx context.Context, userId, newEmail string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserUpdateEmailAccountQuery, newEmail, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update user email")
		return dberrors.ClassifyError(err, UserUpdateEmailAccountQuery)
	}
	return nil
}

func (u *userStore) UpdateDisplayName(ctx context.Context, userId, displayName string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserUpdateDisplayNameAccountQuery, displayName, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to update user display name")
		return dberrors.ClassifyError(err, UserUpdateDisplayNameAccountQuery)
	}
	return nil
}

func (u *userStore) LockAccount(ctx context.Context, username string, lockedReason int) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserLockUserAccountByUsernameQuery, lockedReason, username)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to lock user account")
		return dberrors.ClassifyError(err, UserLockUserAccountByUsernameQuery)
	}
	return nil
}

func (u *userStore) UnlockAccount(ctx context.Context, username string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserUnlockUserAccountByUsernameQuery, username)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to unlock user account")
		return dberrors.ClassifyError(err, UserUnlockUserAccountByUsernameQuery)
	}
	return nil
}

func (u *userStore) RecordFailedAttempt(ctx context.Context, username string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserIncrementFailedAttemptQuery, username)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to increment failed attempts")
		return dberrors.ClassifyError(err, UserIncrementFailedAttemptQuery)
	}
	return nil
}

func (u *userStore) CheckAvailability(ctx context.Context, username, email string) (usernameAvail, emailAvail bool, err error) {
	q := u.getQuerier(ctx)

	rows, err := q.QueryContext(ctx, UserValidateUsernameAndEmailQuery, username, email)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to check username and email availability")
		return false, false, dberrors.ClassifyError(err, UserValidateUsernameAndEmailQuery)
	}
	defer rows.Close()

	var usernames []string
	var emails []string

	for rows.Next() {
		var un, em string
		err = rows.Scan(&un, &em)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to check username and email availability")
			return false, false, dberrors.ClassifyError(err, UserValidateUsernameAndEmailQuery)
		}
		if un != "" {
			usernames = append(usernames, un)
		}

		if em != "" {
			emails = append(emails, em)
		}
	}

	return username != "" && !slices.Contains(usernames, username), email != "" && !slices.Contains(emails, email), nil
}

func (u *userStore) Get(ctx context.Context, identifier string) (*models.UserAccount, error) {
	q := u.getQuerier(ctx)

	var m models.UserAccount
	var createdAt, updatedAt string
	var lockedAt sql.NullString
	var locked int

	row := q.QueryRowContext(ctx, UserGetUserAccountQuery, identifier, identifier)

	err := row.Scan(
		&m.Id, &m.Username, &m.Email, &m.DisplayName,
		&locked, &m.LockedReason, &m.FailedAttempts,
		&createdAt, &updatedAt, &lockedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve user")
		return nil, dberrors.ClassifyError(err, UserGetUserAccountQuery)
	}

	m.Locked = locked != 0

	if m.Locked {
		if !lockedAt.Valid {
			log.Logger().Warn().Msgf("user account(%s) is in locked state but locked timestamp is not available", identifier)
		} else {
			m.LockedAt, err = utils.ParseSqliteTimestamp(lockedAt.String)
			if err != nil {
				log.Logger().Error().Err(err).Msg("failed to retrieve user")
				return nil, dberrors.ClassifyError(err, UserGetUserAccountQuery)
			}
		}
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve user")
		return nil, dberrors.ClassifyError(err, UserGetUserAccountQuery)
	}
	m.CreatedAt = *createdTime

	m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve user")
		return nil, dberrors.ClassifyError(err, UserGetUserAccountQuery)
	}

	return &m, nil
}

func (u *userStore) GetByUsername(ctx context.Context, username string) (*models.UserAccount, error) {
	q := u.getQuerier(ctx)

	var m models.UserAccount
	var createdAt, updatedAt string
	var locked int
	var lockedAt sql.NullString

	row := q.QueryRowContext(ctx, UserGetUserAccountByUsernameQuery, username)

	err := row.Scan(
		&m.Id, &m.Username, &m.Email, &m.DisplayName,
		&locked, &m.LockedReason, &m.FailedAttempts,
		&createdAt, &updatedAt, &lockedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve user")
		return nil, dberrors.ClassifyError(err, UserGetUserAccountByUsernameQuery)
	}

	m.Locked = locked != 0

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve user")
		return nil, dberrors.ClassifyError(err, UserGetUserAccountByUsernameQuery)
	}
	m.CreatedAt = *createdTime

	m.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve user")
		return nil, dberrors.ClassifyError(err, UserGetUserAccountByUsernameQuery)
	}

	m.LockedAt, err = utils.ParseSqliteTimestamp(lockedAt.String)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve user")
		return nil, dberrors.ClassifyError(err, UserGetUserAccountByUsernameQuery)
	}

	return &m, nil
}

func (u *userStore) GetPasswordAndSalt(ctx context.Context, userId string) (password, salt string, err error) {
	q := u.getQuerier(ctx)

	err = q.QueryRowContext(ctx, UserGetPasswordAndSaltQuery, userId).Scan(&password, &salt)
	if err != nil {
		return "", "", dberrors.ClassifyError(err, UserGetPasswordAndSaltQuery)
	}

	return
}

func (u *userStore) UpdatePasswordAndSalt(ctx context.Context, userId, password, salt string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserUpdatePasswordQuery, password, salt, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrieve user details")
		return dberrors.ClassifyError(err, UserUpdatePasswordQuery)
	}

	return nil
}

func (u *userStore) ListUserAccounts(ctx context.Context, conditions *store.ListQueryConditions) (users []*models.UserAccountView, total int, err error) {
	qb := store.NewQueryBuilder(store.DBTypeSqlite).
		WithSearchFields("USERNAME", "EMAIL", "DISPLAY_NAME").
		WithFieldTransformation("role", "ROLE_NAME").
		WithFieldTransformation("last_loggedin_at", "LAST_LOGGEDIN_AT").
		WithBooleanField("locked").
		WithAllowedSortFields("username", "email", "last_loggedin_at", "role_name").
		WithAllowedFilterFields("role", "locked")

	listQuery, countQuery, args, err := qb.Build(
		UserListBaseQuery,
		UserCountBaseQuery,
		conditions,
	)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to build user list query")
		return nil, 0, fmt.Errorf("build query: %w", err)
	}

	// Execute count query first (without limit/offset)
	countArgs := args[:len(args)-2]

	q := u.getQuerier(ctx)

	err = q.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to count total user accounts")
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// Execute list query
	rows, err := q.QueryContext(ctx, listQuery, args...)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to retrieve user accounts")
		return nil, 0, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	// Scan results
	users = make([]*models.UserAccountView, 0)
	for rows.Next() {
		user, err := u.scanUserAccountView(rows)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Failed to scan user account")
			continue // Skip bad rows instead of failing entire query
		}
		users = append(users, user)
	}

	// Check for row iteration errors
	if err = rows.Err(); err != nil {
		log.Logger().Error().Err(err).Msg("Error during row iteration")
		return nil, 0, fmt.Errorf("iterate rows: %w", err)
	}

	return users, total, nil
}

func (s *userStore) scanUserAccountView(rows *sql.Rows) (*models.UserAccountView, error) {
	var user models.UserAccountView
	var createdAt, updatedAt, lockedAt, pwRecoveryAt, lastLoggedInAt, recoveryID sql.NullString
	var locked, lockedReason, pwRecoveryReason sql.NullInt16

	err := rows.Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.DisplayName,
		&locked,
		&lockedReason,
		&lockedAt,
		&user.FailedAttempts,
		&createdAt,
		&updatedAt,
		&user.Role,
		&recoveryID,
		&pwRecoveryReason,
		&pwRecoveryAt,
		&lastLoggedInAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan row: %w", err)
	}

	// Parse CreatedAt (required)
	if createdAt.Valid {
		createdTime, err := utils.ParseSqliteTimestamp(createdAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to parse created_at: %s", createdAt.String)
			return nil, err
		} else if createdTime != nil {
			user.CreatedAt = *createdTime
		}
	}

	// Parse UpdatedAt (optional)
	if updatedAt.Valid {
		updatedTime, err := utils.ParseSqliteTimestamp(updatedAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to parse updated_at: %s", updatedAt.String)
			return nil, err
		} else {
			user.UpdatedAt = updatedTime
		}
	}

	// Parse LockedAt (optional)
	if lockedAt.Valid {
		lockedTime, err := utils.ParseSqliteTimestamp(lockedAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to parse locked_at: %s", lockedAt.String)
			return nil, err
		} else {
			user.LockedAt = lockedTime
		}
	}

	// Parse PasswordRecoveryCreatedAt (optional)
	if pwRecoveryAt.Valid {
		pwRecoveryTime, err := utils.ParseSqliteTimestamp(pwRecoveryAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to parse password_recovery.created_at: %s", pwRecoveryAt.String)
			return nil, err
		} else {
			user.PasswordRecoveryCreatedAt = pwRecoveryTime
		}
	}

	// Parse LastLoggedInAt (optional)
	if lastLoggedInAt.Valid {
		lastLoginTime, err := utils.ParseSqliteTimestamp(lastLoggedInAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Failed to parse last_logged_in_at: %s", lastLoggedInAt.String)
			return nil, err
		} else {
			user.LastLoggedInAt = lastLoginTime
		}
	}

	// Convert SQLite integers to booleans
	if locked.Valid {
		user.Locked = locked.Int16 != 0
	}

	if lockedReason.Valid {
		user.LockedReason = int(lockedReason.Int16)
	}

	if pwRecoveryReason.Valid {
		user.PasswordRecoveryReason = int(pwRecoveryReason.Int16)
	}

	if recoveryID.Valid {
		user.PasswordRecoveryId = recoveryID.String
	}

	return &user, nil
}

func (u *userStore) GetRole(ctx context.Context, userId string) (role string, err error) {
	q := u.getQuerier(ctx)

	err = q.QueryRowContext(ctx, UserGetRoleQuery, userId).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// As per the application architecture it is impossible to have a user without a role
			log.Logger().Warn().Msgf("Unexpected behaviour a user(%s) without a role was found", userId)
			return "", nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve role of user")
		return "", dberrors.ClassifyError(err, UserGetRoleQuery)
	}

	return role, nil
}

func (u *userStore) AssignRole(ctx context.Context, userId, role string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserAssignRoleQuery, userId, role)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to assign role to user")
		return dberrors.ClassifyError(err, UserAssignRoleQuery)
	}

	return nil
}

func (u *userStore) UnAssignRole(ctx context.Context, userId string) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserUnassignRoleQuery, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to unassign role to user")
		return dberrors.ClassifyError(err, UserAssignRoleQuery)
	}

	return nil
}

func (u *userStore) AreAccountsActive(ctx context.Context, userIds []string) (valid bool, err error) {
	q := u.getQuerier(ctx)

	placeHolders := strings.Join(slices.Repeat([]string{"?"}, len(userIds)), ",")
	query := fmt.Sprintf(UserCountActiveAccountByIdsQuery, placeHolders)

	var validAccounts int
	args := make([]any, len(userIds))
	for i, v := range userIds {
		args[i] = v
	}
	err = q.QueryRowContext(ctx, query, args...).Scan(&validAccounts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Warn().Msg("none of user accounts are valid")
			return false, nil
		}
		log.Logger().Error().Err(err).Msg("failed to validate user accounts")
		return false, dberrors.ClassifyError(err, query)
	}

	return validAccounts == len(userIds), nil
}

func (u *userStore) RecordLastAccessedTime(ctx context.Context, userId string, t time.Time) error {
	q := u.getQuerier(ctx)

	_, err := q.ExecContext(ctx, UserRecordLastAccessedTime, t, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to record user last accessed time")
		return dberrors.ClassifyError(err, UserRecordLastAccessedTime)
	}

	return nil
}