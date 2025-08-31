package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

const (
	CodeFileNotFound                    = 6001
	CodePermissionDenied                = 6002
	CodeConcurrentWriteAccessNotAllowed = 6003
	CodeFileIsNotDirectory              = 6004
	CodeInsufficientDiskSpace           = 6005
	CodeReadOnlyFS                      = 6006
	CodeFileAlreadyExists               = 6007
	CodeIOError                         = 6008
	CodeFsOpTimeout                     = 6009
	CodeTooLongFileName                 = 6010
	CodeTooManyOpenFiles                = 6011
	CodeFileCorrupted                   = 6012
	CodeInvalidOffset                   = 6013
	CodeUnclassifiedError               = 6099
)

var (
	ErrFileNotFound = &StorageError{
		errorCode: CodeFileNotFound,
	}
	ErrPermissionDenied = &StorageError{
		errorCode: CodePermissionDenied,
	}
	ErrConcurrentAccessDenied = &StorageError{
		errorCode: CodeConcurrentWriteAccessNotAllowed,
	}
	ErrNotDirectory = &StorageError{
		errorCode: CodeFileIsNotDirectory,
	}
	ErrInsufficientDiskSpace = &StorageError{
		errorCode: CodeInsufficientDiskSpace,
	}
	ErrReadOnlyFS = &StorageError{
		errorCode: CodeReadOnlyFS,
	}
	ErrFileAlreadyExists = &StorageError{
		errorCode: CodeFileAlreadyExists,
	}
	ErrIOError = &StorageError{
		errorCode: CodeIOError,
	}
	ErrFsOpTimeout = &StorageError{
		errorCode: CodeFsOpTimeout,
	}
	ErrTooLongFileName = &StorageError{
		errorCode: CodeTooLongFileName,
	}
	ErrTooManyOpenFiles = &StorageError{
		errorCode: CodeTooManyOpenFiles,
	}
	ErrFileCorrupted = &StorageError{
		errorCode: CodeFileCorrupted,
	}
	ErrInvalidOffset = &StorageError{
		errorCode: CodeInvalidOffset,
	}
	ErrUnclassifiedError = &StorageError{
		errorCode: CodeUnclassifiedError,
	}
)

type StorageError struct {
	location  string
	action    string
	err       error
	errorCode int
}

func (se *StorageError) Error() string {
	if se.err != nil {
		return fmt.Sprintf("storage error: [%d] occurred. operation = %s, location = %s, cause = %v",
			se.errorCode, se.action, se.location, se.err)
	}
	return fmt.Sprintf("storage error: [%d] occurred. operation = %s, location = %s",
		se.errorCode, se.action, se.location)
}

func (se *StorageError) Unwrap() error {
	return se.err
}

func (se *StorageError) Is(target error) bool {
	if sErr, ok := target.(*StorageError); ok {
		return sErr.errorCode == se.errorCode
	}
	return false
}

func (se *StorageError) Code() int {
	return se.errorCode
}

func (se *StorageError) Location() string {
	return se.location
}

func (se *StorageError) Action() string {
	return se.action
}

func ConcurrentAccessDeniedError(action, location string) *StorageError {
	return &StorageError{
		errorCode: CodeConcurrentWriteAccessNotAllowed,
		action:    action,
		location:  location,
	}
}

func NotDirectoryError(action, location string) *StorageError {
	return &StorageError{
		errorCode: CodeFileIsNotDirectory,
		action:    action,
		location:  location,
	}
}

func FileAlreadyExistsError(action, location string) *StorageError {
	return &StorageError{
		errorCode: CodeFileAlreadyExists,
		action:    action,
		location:  location,
	}
}

func ReadOnlyFSError(action, location string) *StorageError {
	return &StorageError{
		errorCode: CodeReadOnlyFS,
		action:    action,
		location:  location,
	}
}

func FileCorruptedError(action, location string) *StorageError {
	return &StorageError{
		errorCode: CodeFileCorrupted,
		action:    action,
		location:  location,
	}
}

func InvalidOffsetError(action, location string) *StorageError {
	return &StorageError{
		errorCode: CodeInvalidOffset,
		action:    action,
		location:  location,
	}
}

// ClassifyError investigates err received and returns appropriate StorageError.
// The current implementation only considers local file system.
// TODO: This can be improved to handle other storage backends.
func ClassifyError(err error, action, location string) *StorageError {
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) || errors.Is(err, fs.ErrNotExist) {
		return &StorageError{
			err:       err,
			errorCode: CodeFileNotFound,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, os.ErrPermission) || errors.Is(err, fs.ErrPermission) {
		return &StorageError{
			err:       err,
			errorCode: CodePermissionDenied,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, os.ErrExist) || errors.Is(err, fs.ErrExist) {
		return &StorageError{
			err:       err,
			errorCode: CodeFileAlreadyExists,
			action:    action,
			location:  location,
		}
	}

	// Handle syscall errors
	if errors.Is(err, syscall.ENOSPC) {
		return &StorageError{
			err:       err,
			errorCode: CodeInsufficientDiskSpace,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, syscall.EROFS) {
		return &StorageError{
			err:       err,
			errorCode: CodeReadOnlyFS,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, syscall.EEXIST) {
		return &StorageError{
			err:       err,
			errorCode: CodeFileAlreadyExists,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, syscall.EIO) {
		return &StorageError{
			err:       err,
			errorCode: CodeIOError,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, syscall.ETIMEDOUT) {
		return &StorageError{
			err:       err,
			errorCode: CodeFsOpTimeout,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, syscall.ENAMETOOLONG) {
		return &StorageError{
			err:       err,
			errorCode: CodeTooLongFileName,
			action:    action,
			location:  location,
		}
	}

	if errors.Is(err, syscall.EMFILE) || errors.Is(err, syscall.ENFILE) {
		return &StorageError{
			err:       err,
			errorCode: CodeTooManyOpenFiles,
			action:    action,
			location:  location,
		}
	}

	return &StorageError{
		err:       err,
		errorCode: CodeUnclassifiedError,
		action:    action,
		location:  location,
	}
}

func UnwrapStorageError(err error) (ok bool, errorCode int) {
	if se, ok := err.(*StorageError); ok {
		return ok, se.errorCode
	}
	return false, CodeUnclassifiedError
}