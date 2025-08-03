package lib

import (
	"sync"

	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/utils"
)

const MaxAtomicLockKeys = 5

// KeyLock provides a way to lock using keys to avoid
// concurrent access to crticial sections.
// It tracks all the existing keys which are locked using a map.
// If a key exists in `locks“, it means a go-routine has locked that
// particular key.
type KeyLock struct {
	mu    sync.Mutex
	locks map[string]struct{}
}

func NewKeyLock() *KeyLock {
	return &KeyLock{
		locks: make(map[string]struct{}),
	}
}

// Lock attempts to get the lock if no-other go-routine hold the lock
// for given `key`. If this lock was already taken by another go-routine,
// without waiting, this will return.
func (kl *KeyLock) Lock(key string) bool {
	kl.mu.Lock()
	locked := kl.unsafeLock(key)
	kl.mu.Unlock()
	return locked
}

// unsafeLock first checks the given exists in locks. If it
// present, another go-route hold this lock. So immediately returns false
// to indicate that caller that lock was not aquired.
func (kl *KeyLock) unsafeLock(key string) bool {
	_, exists := kl.locks[key]

	if !exists {
		kl.locks[key] = struct{}{}
		return true
	}
	// key already exists. another go-routine hold this lock. So the caller has to wait and
	// acquire lock later
	return false
}

func (kl *KeyLock) LockKeysAtomically(keys ...string) bool {
	if len(keys) > MaxAtomicLockKeys {
		return false
	}
	uniqueKeys := utils.RemoveDuplicateKeys(keys)

	lastLockAcquiredIndex := -1

	kl.mu.Lock()
	defer kl.mu.Unlock()

	for i, key := range uniqueKeys {
		if _, exists := kl.locks[key]; exists {
			// some of keys are already locked by other go-routine
			// we cannot get locks for all the keys atomically. hence, return false to
			// caller. Caller can retry later.
			lastLockAcquiredIndex = i - 1
			break
		} else {
			locked := kl.unsafeLock(key)
			if !locked {
				lastLockAcquiredIndex = i - 1
				log.Logger().Warn().Msgf("Unable to lock for key: %s.Possible race conditions.", key)
				break
			}
		}
	}

	// We have to release locks aquired since we were not able to lock for all the keys.
	if lastLockAcquiredIndex >= 0 {
		for i := 0; i <= lastLockAcquiredIndex; i++ {
			kl.unsafeUnlock(uniqueKeys[i])
		}
		return false
	}

	return true
}

func (kl *KeyLock) UnlockKeysAtomically(keys ...string) bool {
	if len(keys) > MaxAtomicLockKeys {
		return false
	}
	uniqueKeys := utils.RemoveDuplicateKeys(keys)

	kl.mu.Lock()
	for _, key := range uniqueKeys {
		kl.unsafeUnlock(key)
	}
	kl.mu.Unlock()
	return true
}

// Unlock ensures that it removes the key from `locks`.
// caller can only expects that the key will be removed if it exists.
func (kl *KeyLock) Unlock(key string) {
	kl.mu.Lock()
	kl.unsafeUnlock(key)
	kl.mu.Unlock()
}

func (kl *KeyLock) unsafeUnlock(key string) {
	_, exists := kl.locks[key]
	if !exists {
		log.Logger().Warn().Msgf("Unlock was called on non-existent key: %s", key)
		return
	}
	delete(kl.locks, key)
}
