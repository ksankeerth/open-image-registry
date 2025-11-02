package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/types/query"
	"github.com/ksankeerth/open-image-registry/utils"
)

type userDaoImpl struct {
	db *sql.DB
	*TransactionManager
}

func (u *userDaoImpl) CreateUserAccount(username, email, displayName, password, salt string, txKey string) (id string, err error) {
	var userId string

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return "", db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(InsertUserAccount, username, email, displayName, password, salt).Scan(&userId)
	} else {
		err = u.db.QueryRow(InsertUserAccount, username, email, displayName, password, salt).Scan(&userId)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when creating user account for %s", username)
		return "", db_errors.ClassifyError(err, InsertUserAccount)
	}
	return userId, nil
}

func (u *userDaoImpl) DeleteUserAccount(userId string, txKey string) (deleted bool, err error) {
	var res sql.Result
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(MarkUserAccountAsDeleted, userId)
	} else {
		res, err = u.db.Exec(MarkUserAccountAsDeleted, userId)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when deleting user account: %s", userId)
		return false, db_errors.ClassifyError(err, MarkUserAccountAsDeleted)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading results of delete query: %s", userId)
		return false, db_errors.ClassifyError(err, MarkUserAccountAsDeleted)
	}
	if rows != 1 {
		return false, nil
	}
	return true, nil
}

func (u *userDaoImpl) UpdateUserDisplayName(userId, displayName string, txKey string) (updated bool, err error) {
	var res sql.Result
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UpdateUserDisplayName, displayName, userId)
	} else {
		res, err = u.db.Exec(UpdateUserDisplayName, displayName, userId)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when updating display name of user account: %s", userId)
		return false, db_errors.ClassifyError(err, UpdateUserDisplayName)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading results of update query: %s", userId)
	}
	if rows != 1 {
		return false, nil
	}
	return true, nil
}

func (u *userDaoImpl) UpdateUserEmail(userId, newEmail string, txKey string) (updated bool, err error) {
	var res sql.Result
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UpdateUserEmail, newEmail, userId)
	} else {
		res, err = u.db.Exec(UpdateUserEmail, newEmail, userId)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when updating email of user account: %s", userId)
		return false, db_errors.ClassifyError(err, UpdateUserEmail)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading results of update query: %s", userId)
	}
	if rows != 1 {
		return false, nil
	}
	return true, nil
}

func (u *userDaoImpl) LockUserAccount(username string, lockedReason int, txKey string) (locked bool, err error) {
	var res sql.Result
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(LockUserAccount, lockedReason, username)
	} else {
		res, err = u.db.Exec(LockUserAccount, lockedReason, username)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when locking user account: %s", username)
		return false, db_errors.ClassifyError(err, LockUserAccount)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when reading result of update query: %s", username)
		return false, db_errors.ClassifyError(err, LockUserAccount)
	}
	if rows != 1 {
		return false, nil
	}
	return true, nil
}

func (u *userDaoImpl) UnlockUserAccount(username string, txKey string) (unlocked bool, err error) {
	var res sql.Result
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UnlockUserAccount, username)
	} else {
		res, err = u.db.Exec(UnlockUserAccount, username)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when unlocking user account: %s", username)
		return false, db_errors.ClassifyError(err, UnlockUserAccount)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when reading result of update query: %s", username)
		return false, db_errors.ClassifyError(err, UnlockUserAccount)
	}
	if rows != 1 {
		return false, nil
	}
	return true, nil
}

func (u *userDaoImpl) RecordFailedAttempt(username string, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UpdateFailedAttempts, username)
	} else {
		res, err = u.db.Exec(UpdateFailedAttempts, username)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when updating user account: %s", username)
		return db_errors.ClassifyError(err, UpdateFailedAttempts)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurr when reading result of update query: %s", username)
		return db_errors.ClassifyError(err, UpdateFailedAttempts)
	}
	if rows != 1 {
		log.Logger().Error().Msgf("UserAccount update seems to have failed: %s", username)
		return db_errors.ErrUnclassifiedError
	}
	return nil
}

func (u *userDaoImpl) GetUserAccount(username string, txKey string) (*models.UserAccount, error) {
	var userAcccount models.UserAccount
	var err error
	var locked int
	var createdAt string
	var updatedAt string

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return nil, db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetUserAccount, username).Scan(&userAcccount.Id, &userAcccount.Username,
			&userAcccount.Email, &userAcccount.DisplayName, &locked, &userAcccount.FailedAttempts, &createdAt, &updatedAt)
	} else {
		err = u.db.QueryRow(GetUserAccount, username).Scan(&userAcccount.Id, &userAcccount.Username,
			&userAcccount.Email, &userAcccount.DisplayName, &locked, &userAcccount.FailedAttempts, &createdAt, &updatedAt)
	}
	if errors.Is(err, sql.ErrNoRows) {
		log.Logger().Error().Msgf("No user account found for username: %s", username)
		return nil, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user account: %s", username)
		return nil, db_errors.ClassifyError(err, GetUserAccount)
	}
	if locked != 0 {
		userAcccount.Locked = true
	}
	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", createdAt)
	}
	if createdTime != nil {
		userAcccount.CreatedAt = *createdTime
	}

	userAcccount.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", updatedAt)
	}
	return &userAcccount, nil
}

func (u *userDaoImpl) ValidateUsernameAndEmail(username, email string, txKey string) (usernameAvail,
	emailAvail bool, err error) {
	var un, mail string
	var rows *sql.Rows
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, false, db_errors.ErrTxAlreadyClosed
		}
		rows, err = tx.Query(ValidateUsernameAndEmail, username, email)
	} else {
		rows, err = u.db.Query(ValidateUsernameAndEmail, username, email)
	}

	usernameAvail = true
	emailAvail = true

	for rows.Next() {
		err = rows.Scan(&un, &mail)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Error occurred when reading result rows")
			return false, false, db_errors.ClassifyError(err, ValidateUsernameAndEmail)
		}
		if un == username {
			usernameAvail = false
		}
		if mail == email {
			emailAvail = false
		}
	}

	return
}

func (u *userDaoImpl) GetUserAccountById(userId string, txKey string) (*models.UserAccount, error) {
	var userAcccount models.UserAccount
	var err error
	var locked int
	var createdAt string
	var updatedAt sql.NullString
	var lockedAt sql.NullString

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return nil, db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetUserAccountById, userId).Scan(&userAcccount.Id, &userAcccount.Username, &userAcccount.Email,
			&userAcccount.DisplayName, &locked, &userAcccount.LockedReason, &userAcccount.FailedAttempts, &createdAt, &updatedAt, &lockedAt)
	} else {
		err = u.db.QueryRow(GetUserAccountById, userId).Scan(&userAcccount.Id, &userAcccount.Username, &userAcccount.Email,
			&userAcccount.DisplayName, &locked, &userAcccount.LockedReason, &userAcccount.FailedAttempts, &createdAt, &updatedAt, &lockedAt)
	}
	if errors.Is(err, sql.ErrNoRows) {
		log.Logger().Error().Msgf("No user account found for userId: %s", userId)
		return nil, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving user account: %s", userId)
		return nil, db_errors.ClassifyError(err, GetUserAccountById)
	}
	if locked != 0 {
		userAcccount.Locked = true
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", createdAt)
	}
	if createdTime != nil {
		userAcccount.CreatedAt = *createdTime
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", createdAt)
	}

	if updatedAt.Valid {
		userAcccount.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", updatedAt.String)
		}
	}

	if lockedAt.Valid {
		userAcccount.LockedAt, err = utils.ParseSqliteTimestamp(lockedAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", lockedAt.String)
		}
	}

	return &userAcccount, nil
}

func (u *userDaoImpl) GetUsernameById(userId string, txKey string) (string, error) {
	var err error
	var username string
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return "", db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetUsernameById, userId).Scan(&username)
	} else {
		err = u.db.QueryRow(GetUsernameById, userId).Scan(&username)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving username by userId: %s", userId)
		return "", db_errors.ClassifyError(err, GetUsernameById)
	}

	return username, nil
}

func (u *userDaoImpl) GetUserPasswordAndSaltById(userId string, txKey string) (password, salt string,
	err error) {
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return "", "", db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetUserPasswordAndSalt, userId).Scan(&password, &salt)
	} else {
		err = u.db.QueryRow(GetUserPasswordAndSalt, userId).Scan(&password, &salt)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving password from database for user: %s", userId)
		return "", "", db_errors.ClassifyError(err, GetUserPasswordAndSalt)
	}
	return
}

func (u *userDaoImpl) UpdateUserPasswordAndSalt(userId, password, salt string,
	txKey string) (updated bool, err error) {
	var res sql.Result

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UpdatePassword, password, salt, userId)
	} else {
		res, err = u.db.Exec(UpdatePassword, password, salt, userId)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when updating password for user: %s", userId)
		return false, db_errors.ClassifyError(err, UpdatePassword)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, db_errors.ClassifyError(err, UpdatePassword)
	}

	if rows == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

func (u *userDaoImpl) CountUserAccounts(txKey string) (total int, err error) {
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return -1, db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(CountTotalUsers).Scan(&total)
	} else {
		err = u.db.QueryRow(CountTotalUsers).Scan(&total)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when counting toal user accounts")
		return -1, db_errors.ClassifyError(err, CountTotalUsers)
	}
	return
}

func (u *userDaoImpl) ListUserAccounts(conditions *query.ListModelsConditions, txKey string) (users []*models.UserAccountView, total int, err error) {
	listQuery, countQuery, args := u.buildListUserQuery(*conditions)

	var rows *sql.Rows
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return nil, -1, db_errors.ErrTxAlreadyClosed
		}
		rows, err = tx.Query(listQuery, args...)
	} else {
		rows, err = u.db.Query(listQuery, args...)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when retrieving user accounts")
		return nil, -1, db_errors.ClassifyError(err, listQuery)
	}
	defer rows.Close()

	users = make([]*models.UserAccountView, 0)
	for rows.Next() {
		var userAccount models.UserAccountView
		var createdAt, updatedAt, lockedAt, pwRecoveryAt, lastLoggedInAt, recoveryId sql.NullString
		var locked, lockedReason, pwRecoveryReason sql.NullInt16

		err = rows.Scan(&userAccount.Id, &userAccount.Username, &userAccount.Email, &userAccount.DisplayName, &locked,
			&lockedReason, &lockedAt, &userAccount.FailedAttempts, &createdAt, &updatedAt,
			&userAccount.Role, &recoveryId, &pwRecoveryReason,
			&pwRecoveryAt, &lastLoggedInAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Error occurred when reading user accounts")
			return nil, -1, db_errors.ClassifyError(err, listQuery)
		}

		// Parse CreatedAt (required)
		createdTime, err := utils.ParseSqliteTimestamp(createdAt.String)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", createdAt)
			continue
		}
		if createdTime != nil {
			userAccount.CreatedAt = *createdTime
		}

		// Parse UpdatedAt (optional)
		if updatedAt.Valid {
			userAccount.UpdatedAt, err = utils.ParseSqliteTimestamp(updatedAt.String)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", updatedAt.String)
			}
		}

		// Parse LockedAt (optional)
		if lockedAt.Valid {
			userAccount.LockedAt, err = utils.ParseSqliteTimestamp(lockedAt.String)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", lockedAt.String)
			}
		}

		// Parse PasswordRecoveryCreatedAt (optional)
		if pwRecoveryAt.Valid {
			userAccount.PasswordRecoveryCreatedAt, err = utils.ParseSqliteTimestamp(pwRecoveryAt.String)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", pwRecoveryAt.String)
			}
		}

		// Parse LastLoggedInAt (optional)
		if lastLoggedInAt.Valid {
			userAccount.LastLoggedInAt, err = utils.ParseSqliteTimestamp(lastLoggedInAt.String)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Error occurred when parsing sqlite timestamp: %s", lastLoggedInAt)
			}
		}

		// Convert int to bool for Locked
		if locked.Valid && locked.Int16 != 0 {
			userAccount.Locked = true
		}
		if lockedReason.Valid && lockedReason.Int16 != 0 {
			userAccount.LockedReason = int(lockedReason.Int16)
		}
		if pwRecoveryReason.Valid && pwRecoveryReason.Int16 != 0 {
			userAccount.PasswordRecoveryReason = int(pwRecoveryReason.Int16)
		}
		if recoveryId.Valid {
			userAccount.PasswordRecoveryId = recoveryId.String
		}

		users = append(users, &userAccount)
	}

	// Check for row iteration errors
	if err = rows.Err(); err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred during row iteration")
		return nil, -1, db_errors.ClassifyError(err, listQuery)
	}

	// Execute count query
	var countTotal int
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return nil, -1, db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(countQuery, args...).Scan(&countTotal)
	} else {
		err = u.db.QueryRow(countQuery, args...).Scan(&countTotal)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when counting total user accounts")
		return nil, -1, db_errors.ClassifyError(err, countQuery)
	}

	return users, countTotal, nil
}

func (u *userDaoImpl) buildListUserQuery(cond query.ListModelsConditions) (listQuery, countQuery string, args []any) {
	var sb strings.Builder
	sb.WriteString(ListUsersBaseQuery)

	searchFields := []string{"USERNAME", "EMAIL", "DISPLAY_NAME"}

	args = []any{}
	whereClauses := []string{}

	// Build filter conditions
	for _, f := range cond.Filters {
		switch f.Field {
		case "role":
			// Multiple roles: IN clause
			placeholders := make([]string, len(f.Values))
			for i, v := range f.Values {
				args = append(args, v)
				placeholders[i] = "?"
			}
			whereClauses = append(whereClauses, fmt.Sprintf("ROLE_NAME IN (%s)", strings.Join(placeholders, ",")))

		case "locked":
			// Single boolean value
			if len(f.Values) > 0 {
				whereClauses = append(whereClauses, "LOCKED = ?")
				args = append(args, f.Values[0])
			}
		}
	}

	// Build search condition
	if cond.SearchTerm != "" {
		searchClauses := make([]string, len(searchFields))
		for i, field := range searchFields {
			searchClauses[i] = fmt.Sprintf("%s LIKE ?", field)
			args = append(args, "%"+cond.SearchTerm+"%")
		}
		whereClauses = append(whereClauses, "("+strings.Join(searchClauses, " OR ")+")")
	}

	// Build WHERE clause (shared between list and count queries)
	var whereClause string
	if len(whereClauses) > 0 {
		whereClause = " AND " + strings.Join(whereClauses, " AND ")
	}

	// Build list query with ORDER BY, LIMIT, OFFSET
	sb.WriteString(whereClause)
	if cond.Sort.Field != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(cond.Sort.Field)
		sb.WriteString(" ")
		sb.WriteString(string(cond.Sort.Order))
	}

	offset := (cond.Pagination.Page - 1) * cond.Pagination.Limit
	sb.WriteString(" LIMIT ? OFFSET ?")
	args = append(args, cond.Pagination.Limit, offset)

	listQuery = sb.String()

	// Build count query
	countQuery = CountUsersBaseQuery + whereClause

	return listQuery, countQuery, args
}

func (u *userDaoImpl) UpdateUserAccount(userId, email, displayName string, txKey string) error {

	var err error
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return db_errors.ErrTxAlreadyClosed
		}
		_, err = tx.Exec(UpdateUserAccount, email, displayName, userId)
	} else {
		_, err = u.db.Exec(UpdateUserAccount, email, displayName, userId)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when updating user account: %s", userId)
		return db_errors.ClassifyError(err, UpdateUserAccount)
	}

	return nil
}

func (u *userDaoImpl) PersistPasswordRecovery(userId, recoveryUuid string, reasonType int, txKey string) error {
	var res sql.Result
	var err error
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(PersistPasswordRecoveryRecord, recoveryUuid, userId, reasonType)
	} else {
		res, err = u.db.Exec(PersistPasswordRecoveryRecord, recoveryUuid, userId, reasonType)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when persisting password recovery record")
	}
	row, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error ocurred when reading insert query result")
		return db_errors.ClassifyError(err, PersistPasswordRecoveryRecord)
	}
	if row != 1 {
		log.Logger().Error().Msgf("No record was inserted into db but without errors")
		return db_errors.ErrUnclassifiedError
	}
	return nil
}

func (u *userDaoImpl) RetrivePasswordRecovery(uuid string, txKey string) (*models.PasswordRecovery, error) {
	var pwRecovery models.PasswordRecovery
	var err error
	var createdAt string

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return nil, db_errors.ErrTxAlreadyClosed
		}

		err = tx.QueryRow(GetPasswordRecoveryById, uuid).Scan(&pwRecovery.UserId,
			&pwRecovery.ReasonType, &createdAt)

	} else {
		err = u.db.QueryRow(GetPasswordRecoveryById, uuid).Scan(&pwRecovery.UserId,
			&pwRecovery.ReasonType, &createdAt)
	}

	if errors.Is(err, sql.ErrNoRows) {
		log.Logger().Info().Msgf("No password recovery records found for id: %s", uuid)
		return nil, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving password recovery record for id: %s", uuid)
		return nil, db_errors.ClassifyError(err, GetPasswordRecoveryById)
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error parsing sqlite timestamp: %s", createdAt)
	}
	if createdTime != nil {
		pwRecovery.CreatedAt = *createdTime
	}

	pwRecovery.RecoveryId = uuid

	return &pwRecovery, nil
}

func (u *userDaoImpl) RetrivePasswordRecoveryByUserId(userId string, txKey string) (*models.PasswordRecovery, error) {
	var pwRecovery models.PasswordRecovery
	var err error
	var createdAt string

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return nil, db_errors.ErrTxAlreadyClosed
		}

		err = tx.QueryRow(GetPasswordRecoveryByUserId, userId).Scan(&pwRecovery.UserId, &pwRecovery.RecoveryId,
			&pwRecovery.ReasonType, &createdAt)

	} else {
		err = u.db.QueryRow(GetPasswordRecoveryByUserId, userId).Scan(&pwRecovery.UserId, &pwRecovery.RecoveryId,
			&pwRecovery.ReasonType, &createdAt)
	}

	if errors.Is(err, sql.ErrNoRows) {
		log.Logger().Info().Msgf("No password recovery records found for user: %s", userId)
		return nil, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving password recovery record of user: %s", userId)
		return nil, db_errors.ClassifyError(err, GetPasswordRecoveryByUserId)
	}

	createdTime, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error parsing sqlite timestamp: %s", createdAt)
	}
	if createdTime != nil {
		pwRecovery.CreatedAt = *createdTime
	}

	return &pwRecovery, nil
}

func (u *userDaoImpl) DeletePasswordRecovery(userId string, txKey string) (deleted bool, err error) {
	var res sql.Result
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(DeletePasswordRecoveryByUserId, userId)
	} else {
		res, err = u.db.Exec(DeletePasswordRecoveryByUserId, userId)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when deleting password recovery reference by userId: %s", userId)
		return false, db_errors.ClassifyError(err, DeletePasswordRecoveryByUserId)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when reading delete query result")
		return false, db_errors.ClassifyError(err, DeletePasswordRecoveryByUserId)
	}

	if rows <= 0 {
		return false, nil
	}

	return true, nil
}

func (u *userDaoImpl) PersistUserRole(roleName, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(PersistUserRole, roleName)
	} else {
		res, err = u.db.Exec(PersistUserRole, roleName)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when persisting role(%s)", roleName)
		return db_errors.ClassifyError(err, PersistUserRole)
	}

	_, err = res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when reading results from insert query")
		return db_errors.ClassifyError(err, PersistUserRole)
	}

	return nil
}

func (u *userDaoImpl) DeleteUserRole(roleName string, txKey string) (deleted bool, err error) {
	var res sql.Result
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(DeleteUserRole, roleName)
	} else {
		res, err = u.db.Exec(DeleteUserRole, roleName)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when deleting role: %s", roleName)
		return false, db_errors.ClassifyError(err, DeleteUserRole)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error ocurred when reading results of delete query")
		return false, db_errors.ClassifyError(err, DeleteUserRole)
	}

	if rows != 1 {
		return false, nil
	}
	return true, nil
}

func (u *userDaoImpl) AssignUserRole(roleName, userId string, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(AssignRoleToUser, userId, roleName)
	} else {
		res, err = u.db.Exec(AssignRoleToUser, userId, roleName)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when assigning role(%s) to user(%s)", roleName, userId)
		return db_errors.ClassifyError(err, AssignRoleToUser)
	}

	_, err = res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when reading results of insert query")
		return db_errors.ClassifyError(err, AssignRoleToUser)
	}

	return nil
}

func (u *userDaoImpl) RemoveUserRoleAssignment(userId string, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return db_errors.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UnassignRole, userId)
	} else {
		res, err = u.db.Exec(UnassignRole, userId)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when removing role assignment from user(%s)", userId)
		return db_errors.ClassifyError(err, UnassignRole)
	}

	_, err = res.RowsAffected()
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when reading results of delete query")
		return db_errors.ClassifyError(err, UnassignRole)
	}
	return nil
}

func (u *userDaoImpl) GetUserRole(userId string, txKey string) (roleName string, err error) {

	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return "", db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetUserRole, userId).Scan(&roleName)
	} else {
		err = u.db.QueryRow(GetUserRole, userId).Scan(&roleName)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when retriving role of user: %s", userId)
		return "", db_errors.ClassifyError(err, GetUserRole)
	}

	return
}

func (u *userDaoImpl) IsUserAssignedToRole(userId, roleName string, txKey string) (bool, error) {
	var assgined int
	var err error
	if txKey != "" {
		tx := u.getTx(txKey)
		if tx == nil {
			log.Logger().Error().Msgf("Transaction for %s already closed.", txKey)
			return false, db_errors.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(HasRole, userId, roleName).Scan(&assgined)
	} else {
		err = u.db.QueryRow(HasRole, userId, roleName).Scan(&assgined)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when checking role assignment of user: %s", userId)
		return false, db_errors.ClassifyError(err, HasRole)
	}

	if assgined == 1 {
		return true, nil
	} else {
		return false, nil
	}
}
