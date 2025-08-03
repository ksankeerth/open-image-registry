package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	storage_errors "github.com/ksankeerth/open-image-registry/errors/storage"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
)

const (
	PropertyStoragePath = "fs.storage.path"
	DirPermissions      = 0755
	FilePermissions     = 0644
)

type localFileStorage struct {
	storageDir string
	// Use this lock only when modifying file paths and writing into file.
	// For reading file, we will not use this file lock. Instead of returning partially written
	// results would suffice. Docker clients can retry if the content is not complete.
	fileLocks *lib.KeyLock
}

func NewLFS(props map[string]string) *localFileStorage {
	storagePath, _ := props[PropertyStoragePath]
	return &localFileStorage{
		storageDir: storagePath,
		fileLocks:  lib.NewKeyLock(),
	}
}

func (lfs *localFileStorage) Init() error {
	fileInfo, err := os.Stat(lfs.storageDir)
	if err == nil {
		if !fileInfo.IsDir() {
			log.Logger().Error().Msgf("File: %s is not a directory", lfs.storageDir)
			return fmt.Errorf("file: %s is not a directory", lfs.storageDir)
		}
		log.Logger().Debug().Msgf("Storage Directory: %s already exists.", lfs.storageDir)
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		log.Logger().Debug().Msgf("Storage Directory: %s does not exist.", lfs.storageDir)
		err := os.MkdirAll(lfs.storageDir, os.FileMode(DirPermissions))
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to create storage directory: %s", lfs.storageDir)
			return err
		}
		log.Logger().Info().Msgf("Storage Directory: %s was created successfully.", lfs.storageDir)
		return nil
	} else if errors.Is(err, os.ErrPermission) {
		log.Logger().Error().Err(err).Msgf("No peremissions to access storage directory: %s", lfs.storageDir)
		return err
	} else {
		log.Logger().Error().Err(err).Msgf("Unexpected error occured when initializing storage direcctory: %s", lfs.storageDir)
		return err
	}
}

func (lfs *localFileStorage) ReadFile(location string) ([]byte, error) {
	targetPath := filepath.Join(lfs.storageDir, location)

	file, err := os.Open(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Logger().Error().Err(err).Msgf("Location: %s does not exist in LFS.", targetPath)
		} else if errors.Is(err, os.ErrPermission) {
			log.Logger().Error().Err(err).Msgf("Permission denied to access location: %s.", targetPath)
		}
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("File: %s exists but error occured when reading the file.", targetPath)
		return nil, err
	}

	return data, nil
}

func (lfs *localFileStorage) PutFile(location string, data []byte) error {
	targetPath := filepath.Join(lfs.storageDir, location)

	if !lfs.fileLocks.Lock(targetPath) {
		return storage_errors.ConcurrentAccessDeniedError("put", targetPath)
	}
	defer lfs.fileLocks.Unlock(targetPath)

	err := os.MkdirAll(filepath.Dir(targetPath), os.FileMode(DirPermissions))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to create directories: %s.", filepath.Dir(targetPath))
		return storage_errors.NotDirectoryError("mkdirall", filepath.Dir(targetPath))
	}

	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(FilePermissions))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to create file: %s", targetPath)
		return storage_errors.ClassifyError(err, "open", targetPath)
	}
	defer file.Close()

	n, err := file.Write(data)

	if err != nil || n != len(data) {
		log.Logger().Error().Err(err).Msgf("Unable to write data into file: %s. Partially written file will be removed", targetPath)
		err1 := os.Remove(targetPath)
		if err1 != nil {
			log.Logger().Error().Err(err1).Msgf("Removing partially wrritten file failed: %s", targetPath)
		}
		return storage_errors.ClassifyError(err, "write", targetPath)
	}

	return nil
}

func (lfs *localFileStorage) ListFiles(location string) ([]string, error) {
	var files []string

	targetPath := filepath.Join(lfs.storageDir, location)

	f, err := os.Open(targetPath)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to open directory: %s", targetPath)
		return files, storage_errors.ClassifyError(err, "open", targetPath)
	}
	defer f.Close()

	fileInfos, err := f.ReadDir(-1)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to read the directory: %s", targetPath)
	}

	for _, fi := range fileInfos {
		if !fi.IsDir() {
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

func (lfs *localFileStorage) RenameFile(oldLocation, newLocation string) error {
	if oldLocation == newLocation {
		return nil
	}
	oldLocFullPath := filepath.Join(lfs.storageDir, oldLocation)
	newLocFullPath := filepath.Join(lfs.storageDir, newLocation)

	locked := lfs.fileLocks.LockKeysAtomically(oldLocFullPath, newLocFullPath)
	if !locked {
		return storage_errors.ConcurrentAccessDeniedError("rename", newLocFullPath)
	}
	defer lfs.fileLocks.UnlockKeysAtomically(oldLocFullPath, newLocFullPath)

	if err := os.MkdirAll(filepath.Dir(newLocFullPath), os.FileMode(DirPermissions)); err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to create required directories to rename file: %s to %s.", oldLocFullPath, newLocFullPath)
		return storage_errors.ClassifyError(err, "mkdirall", filepath.Dir(newLocFullPath))
	}

	if err := os.Rename(oldLocFullPath, newLocFullPath); err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to rename file: %s to %s.", oldLocFullPath, newLocFullPath)
		return storage_errors.ClassifyError(err, "rename", newLocFullPath)
	}
	return nil
}

func (lfs *localFileStorage) DeleteFile(location string) error {
	targetPath := filepath.Join(lfs.storageDir, location)

	locked := lfs.fileLocks.Lock(targetPath)
	if !locked {
		return storage_errors.ConcurrentAccessDeniedError("delete", targetPath)
	}
	defer lfs.fileLocks.Unlock(targetPath)

	err := os.Remove(targetPath)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when removing file: %s", targetPath)
		return storage_errors.ClassifyError(err, "delete", targetPath)
	}

	return nil
}

func (lfs *localFileStorage) Size(location string) (int64, error) {
	targetPath := filepath.Join(lfs.storageDir, location)

	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when checking file: %s", targetPath)
		return -1, storage_errors.ClassifyError(err, "stat", targetPath)
	}
	return fileInfo.Size(), nil
}

func (lfs *localFileStorage) PutFileChunk(location string, chunk []byte, offset int64) error {
	targetPath := filepath.Join(lfs.storageDir, location)

	if offset < 0 {
		log.Logger().Warn().Msgf("Offset is negative value for chunk file write")
		return storage_errors.InvalidOffsetError("put_chunk", targetPath)
	}

	locked := lfs.fileLocks.Lock(targetPath)
	if !locked {
		// We'll let the caller to make decision whether to retry or remove the partial updates
		return storage_errors.ConcurrentAccessDeniedError("put_chunk", targetPath)
	}
	defer lfs.fileLocks.Unlock(targetPath)

	createFile := false
	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) && offset == 0 {
			createFile = true
		} else {
			createFile = false
			log.Logger().Error().Err(err).Msgf("Error when checking file: %s", targetPath)
			return storage_errors.ClassifyError(err, "stat", targetPath)
		}
	}

	if createFile {
		if err = os.MkdirAll(filepath.Dir(targetPath), DirPermissions); err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to create directories: %s", filepath.Dir(targetPath))
			return storage_errors.ClassifyError(err, "mkdirall", filepath.Dir(targetPath))
		}
	} else if fileInfo.Size() != offset {
		// file seems to be corrupted, therefore we'll remove it. So next
		// retry can pass.
		log.Logger().Warn().Msgf("File: %s seems to be corrupted. Threfore It will be removed", targetPath)
		err1 := os.Remove(targetPath)
		if err1 != nil {
			log.Logger().Error().Err(err).Msgf("Removing file: %s failed due to errors.", targetPath)
			return storage_errors.ClassifyError(err1, "remove", targetPath)
		}
		return storage_errors.FileCorruptedError("put_chunk", targetPath)
	}

	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, FilePermissions)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when opening file: %s", targetPath)
		return storage_errors.ClassifyError(err, "open", targetPath)
	}
	defer file.Close()

	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when seeking file: %s, offset: %d", targetPath, offset)
		return storage_errors.ClassifyError(err, "seek", targetPath)
	}

	n, err := file.Write(chunk)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when writing chunk to file: %s, offset: %d", targetPath, offset)
		return storage_errors.ClassifyError(err, "write", targetPath)
	}

	if n != len(chunk) {
		log.Logger().Warn().Msgf("File: %s write is incomplete. Threfore, the file will be removed rather having corrupted file.", targetPath)
		err = os.Remove(targetPath)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to remove file: %s", targetPath)
			return storage_errors.ClassifyError(err, "delete", targetPath)
		}
		return storage_errors.FileCorruptedError("write", targetPath)
	}

	return nil
}
