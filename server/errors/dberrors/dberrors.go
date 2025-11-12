package dberrors

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	sqlite3 "modernc.org/sqlite"
)

const (
	CodeUniqueConstraintViolation = 5001
	CodeNotFound                  = 5002
	CodeConnectionFailed          = 5003
	CodeTxCreateFailed            = 5004
	CodeTxAlreadyClosed           = 5005
	CodeUnclassifiedError         = 5049
)

var (
	ErrUniqueConstraint = &DatabaseError{
		errCode: CodeUniqueConstraintViolation,
	}
	ErrNotFound = &DatabaseError{
		errCode: CodeNotFound,
	}
	ErrConnectionFailed = &DatabaseError{
		errCode: CodeConnectionFailed,
	}
	ErrTxCreationFailed = &DatabaseError{
		errCode: CodeTxCreateFailed,
	}
	ErrTxAlreadyClosed = &DatabaseError{
		errCode: CodeTxAlreadyClosed,
	}
	ErrUnclassifiedError = &DatabaseError{
		errCode: CodeUnclassifiedError,
	}
)

type DatabaseError struct {
	query   string
	err     error
	errCode int
}

func (de *DatabaseError) Error() string {
	msg := fmt.Sprintf("db error: [%d]", de.errCode)
	if de.query != "" {
		msg += fmt.Sprintf(" (query: %q)", de.query)
	}
	return msg
}

func (de *DatabaseError) Unwrap() error {
	return de.err
}

func (de *DatabaseError) Is(target error) bool {
	if dbErr, ok := target.(*DatabaseError); ok {
		return dbErr.errCode == de.errCode
	}
	return false
}

const (
	SqliteUniqueConstraintViolation int = 2067
)

// ClassifyError invesitgates err received and return appropriate DatabaseError
// The current implementation only considers sqlite db.
func ClassifyError(err error, query string) *DatabaseError {
	// If the err is nil, proably other conditions were not met to be successful
	if err == nil {
		return &DatabaseError{
			errCode: CodeUnclassifiedError,
			query:   query,
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return &DatabaseError{
			err:     err,
			errCode: CodeNotFound,
			query:   query,
		}
	}

	if errors.Is(err, sql.ErrConnDone) {
		return &DatabaseError{
			err:     err,
			errCode: CodeConnectionFailed,
			query:   query,
		}
	}

	var sqliteErr *sqlite3.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case SqliteUniqueConstraintViolation:
			return &DatabaseError{
				err:     err,
				errCode: CodeUniqueConstraintViolation,
				query:   query,
			}
		}
	}

	return &DatabaseError{
		err:     err,
		errCode: CodeUnclassifiedError,
		query:   query,
	}
}

func UnwrapDBError(err error) (ok bool, errorCode int) {
	if de, ok := err.(*DatabaseError); ok {
		return ok, de.errCode
	}
	return false, CodeUnclassifiedError
}

func IsNotFound(err error) bool {
	dbErr, ok := err.(*DatabaseError)
	if ok && dbErr.errCode == CodeNotFound {
		return true
	}
	return false
}

func IsUniqueConstraint(err error) (yes bool, columnName string) {
	dbErr, ok := err.(*DatabaseError)
	if ok && dbErr.errCode == CodeUniqueConstraintViolation {

		return true, extractColumnNameFromSqliteUniqueConstraintError(dbErr.err.Error())
	}
	return false, ""
}

var sqliteUniqueConstraintRegex = regexp.MustCompile(`UNIQUE constraint failed: \w+\.(\w+)`)

func extractColumnNameFromSqliteUniqueConstraintError(errMsg string) string {
	matches := sqliteUniqueConstraintRegex.FindStringSubmatch(errMsg)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}
