package app

import (
	"fmt"
	"path/filepath"

	"github.com/go-downloader/internal/config"
	"github.com/go-downloader/internal/utils"

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
	utils.Debug("AcquireLock() invoked")
	// Ensure config dir exists
	if err := config.EnsureDirs(); err != nil {
		utils.Debug("EnsureDirs() failed: %v", err)
		return false, fmt.Errorf("failed to ensure config dirs: %w", err)
	}

	runtimeDir := config.GetRuntimeDir()
	utils.Debug("Runtime directory: %s", runtimeDir)
	lockPath := filepath.Join(runtimeDir, "downloader.lock")
	utils.Debug("Lock file path: %s", lockPath)
	fileLock := flock.New(lockPath)

	utils.Debug("Attempting to acquire file lock")
	locked, err := fileLock.TryLock()
	if err != nil {
		utils.Debug("TryLock() encountered an error: %v", err)
		return false, fmt.Errorf("failed to try lock: %w", err)
	}

	if locked {
		utils.Debug("File lock acquired")
		// We are the master
		instanceLock = &InstanceLock{
			flock: fileLock,
			path:  lockPath,
		}
		return true, nil
	}

	utils.Debug("File lock is already held by another instance")
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
