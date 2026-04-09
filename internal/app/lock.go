package app

import (
	"fmt"
	"path/filepath"

	"github.com/go-downloader/internal/config"

	"github.com/gofrs/flock"
)

// InstanceLock wraps the file locking mechanism
type InstanceLock struct {
	flock *flock.Flock
	path  string
}

// Global lock instance
var instanceLock *InstanceLock

// AcquireLock attempts to acquire the single instance lock.
// Returns true if the lock was acquired (this is the master instance).
// Returns false if the lock is already held (another instance is running).
// Returns an error if the locking process failed unexpectedly.
func AcquireLock() (bool, error) {
	// Ensure config dir exists
	if err := config.EnsureDirs(); err != nil {
		return false, fmt.Errorf("failed to ensure config dirs: %w", err)
	}

	lockPath := filepath.Join(config.GetRuntimeDir(), "downloader.lock")
	fileLock := flock.New(lockPath)

	locked, err := fileLock.TryLock()
	if err != nil {
		return false, fmt.Errorf("failed to try lock: %w", err)
	}

	if locked {
		// We are the master
		instanceLock = &InstanceLock{
			flock: fileLock,
			path:  lockPath,
		}
		return true, nil
	}

	// Another instance holds the lock
	return false, fmt.Errorf("downloader is already running")
}

// ReleaseLock releases the lock if it is held by this instance.
func ReleaseLock() error {
	if instanceLock != nil && instanceLock.flock != nil {
		return instanceLock.flock.Unlock()
	}
	return nil
}
