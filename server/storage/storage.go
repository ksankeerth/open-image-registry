package storage

import (
	"path/filepath"
	"sync"

	"github.com/ksankeerth/open-image-registry/config"
)

type BlobStorage interface {
	Init() error

	ReadFile(location string) ([]byte, error)

	PutFile(location string, data []byte) error

	ListFiles(location string) ([]string, error)

	RenameFile(oldLocation string, newLocation string) error

	PutFileChunk(location string, chunk []byte, offset int64) error

	DeleteFile(location string) error

	Size(location string) (int64, error)
}

var storage BlobStorage

var once sync.Once

// Init function initializes storage backend and Storage global variable.
// This function guarantees that the only one time Storage will be initialized.
// The current implementation initializes Local File system as storage backend.
// TODO later, this function should take config as argument, based on the configuration
// it should initialize appropriate storage backend.
// if error is returned, the caller should fix the issues and re-start the server.
func Init(config *config.StorageConfig) (err error) {
	once.Do(func() {
		if storage == nil {
			storage = NewLFS(map[string]string{
				"fs.storage.path": filepath.Join(config.Path, "storage", "lfs"),
			})
			err = storage.Init()
			if err != nil {
				storage = nil
			}
		}
	})
	return err
}

func ReadFile(location string) ([]byte, error) {
	return storage.ReadFile(location)
}

func PutFile(location string, data []byte) error {
	return storage.PutFile(location, data)
}

func ListFiles(location string) ([]string, error) {
	return storage.ListFiles(location)
}

func RenameFile(oldLocation string, newLocation string) error {
	return storage.RenameFile(oldLocation, newLocation)
}

func PutFileChunk(location string, chunk []byte, offset int64) error {
	return storage.PutFileChunk(location, chunk, offset)
}

func DeleteFile(location string) error {
	return storage.DeleteFile(location)
}

func Size(location string) (int64, error) {
	return storage.Size(location)
}