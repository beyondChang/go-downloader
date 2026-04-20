package app

import (
	"fmt"
	"sync"

	"go-downloader/internal/utils"
)

var (
	globalShutdownOnce sync.Once
	globalShutdownErr  error
	globalShutdownFn   = DefaultGlobalShutdown
)

func DefaultGlobalShutdown() error {
	CancelGlobalEnqueue()

	// Shutdown the service FIRST so that PauseAll() can emit DownloadPausedMsg
	// events while the lifecycle event worker is still alive to persist them.
	// If we close the lifecycle stream before shutdown, pause state is lost
	// and downloads vanish from the list on terminal close.
	var err error
	if GlobalService != nil {
		err = GlobalService.Shutdown()
	} else if GlobalPool != nil {
		GlobalPool.GracefulShutdown()
	}

	if cleanup := TakeLifecycleCleanup(); cleanup != nil {
		cleanup()
	}

	return err
}

func ExecuteGlobalShutdown(reason string) error {
	globalShutdownOnce.Do(func() {
		utils.Debug("Executing graceful shutdown (%s)", reason)
		globalShutdownErr = globalShutdownFn()
		if globalShutdownErr != nil {
			globalShutdownErr = fmt.Errorf("graceful shutdown failed: %w", globalShutdownErr)
		}
	})
	return globalShutdownErr
}

func ResetGlobalShutdownCoordinatorForTest(fn func() error) {
	globalShutdownOnce = sync.Once{}
	globalShutdownErr = nil
	ResetGlobalEnqueueContext()
	_ = TakeLifecycleCleanup()
	if fn != nil {
		globalShutdownFn = fn
		return
	}
	globalShutdownFn = DefaultGlobalShutdown
}
