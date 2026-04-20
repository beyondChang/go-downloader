package app

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"go-downloader/internal/config"
	"go-downloader/internal/core"
	"go-downloader/internal/download"
	"go-downloader/internal/engine/events"
	"go-downloader/internal/engine/state"
	"go-downloader/internal/engine/types"
	"go-downloader/internal/processing"
	"go-downloader/internal/utils"
	"github.com/google/uuid"
)

// Global state variables moved from cmd package
var (
	ActiveDownloads         int32
	PendingEnqueue          int32
	GlobalPool              *download.WorkerPool
	GlobalProgressCh        chan any
	GlobalService           core.DownloadService
	GlobalLifecycleCleanup  func()
	StartupIntegrityMessage string
	GlobalSettings          *config.Settings
	GlobalLifecycle         *processing.LifecycleManager
	GlobalLifecycleMu       sync.Mutex
	GlobalEnqueueCtx        context.Context
	GlobalEnqueueCancel     context.CancelFunc
	GlobalEnqueueMu         sync.Mutex
)

func BuildPoolIsNameActive(getAll func() []types.DownloadConfig) processing.IsNameActiveFunc {
	if getAll == nil {
		return nil
	}

	return func(dir, name string) bool {
		dir = utils.EnsureAbsPath(strings.TrimSpace(dir))
		name = strings.TrimSpace(name)
		if dir == "" || name == "" {
			return false
		}

		for _, cfg := range getAll() {
			existingName := strings.TrimSpace(cfg.Filename)
			existingDir := strings.TrimSpace(cfg.OutputPath)
			if cfg.DestPath != "" {
				existingDir = filepath.Dir(cfg.DestPath)
				if existingName == "" {
					existingName = filepath.Base(cfg.DestPath)
				}
			}
			if cfg.State != nil {
				if stateName := strings.TrimSpace(cfg.State.GetFilename()); stateName != "" {
					existingName = stateName
				}
				if stateDestPath := strings.TrimSpace(cfg.State.GetDestPath()); stateDestPath != "" {
					existingDir = filepath.Dir(stateDestPath)
					if existingName == "" {
						existingName = filepath.Base(stateDestPath)
					}
				}
			}
			if existingDir == "" || existingName == "" {
				continue
			}
			if utils.EnsureAbsPath(existingDir) == dir && existingName == name {
				return true
			}
		}
		return false
	}
}

func NewLocalLifecycleManager(service core.DownloadService, getAll func() []types.DownloadConfig) *processing.LifecycleManager {
	var addFunc processing.AddDownloadFunc
	var addWithIDFunc processing.AddDownloadWithIDFunc
	if service != nil {
		addFunc = service.Add
		addWithIDFunc = service.AddWithID
	}

	return processing.NewLifecycleManager(addFunc, addWithIDFunc, BuildPoolIsNameActive(getAll))
}

func StartLifecycleEventWorker(service core.DownloadService, mgr *processing.LifecycleManager) (func(), error) {
	if service == nil || mgr == nil {
		return nil, nil
	}

	managerStream, managerCleanup, err := service.StreamEvents(context.Background())
	if err != nil {
		return nil, err
	}
	go mgr.StartEventWorker(managerStream)
	return managerCleanup, nil
}

func CurrentLifecycle() *processing.LifecycleManager {
	GlobalLifecycleMu.Lock()
	defer GlobalLifecycleMu.Unlock()
	return GlobalLifecycle
}

func ResetGlobalEnqueueContext() {
	GlobalEnqueueMu.Lock()
	defer GlobalEnqueueMu.Unlock()
	if GlobalEnqueueCancel != nil {
		GlobalEnqueueCancel()
	}
	GlobalEnqueueCtx, GlobalEnqueueCancel = context.WithCancel(context.Background())
}

func EnsureEnqueueContextLocked() {
	if GlobalEnqueueCtx == nil || GlobalEnqueueCancel == nil {
		GlobalEnqueueCtx, GlobalEnqueueCancel = context.WithCancel(context.Background())
	}
}

func CurrentEnqueueContext() context.Context {
	GlobalEnqueueMu.Lock()
	defer GlobalEnqueueMu.Unlock()
	EnsureEnqueueContextLocked()
	return GlobalEnqueueCtx
}

func CurrentEnqueueCancel() context.CancelFunc {
	GlobalEnqueueMu.Lock()
	defer GlobalEnqueueMu.Unlock()
	EnsureEnqueueContextLocked()
	return GlobalEnqueueCancel
}

func CancelGlobalEnqueue() {
	GlobalEnqueueMu.Lock()
	cancel := GlobalEnqueueCancel
	GlobalEnqueueMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func TakeLifecycleCleanup() func() {
	GlobalLifecycleMu.Lock()
	defer GlobalLifecycleMu.Unlock()
	cleanup := GlobalLifecycleCleanup
	GlobalLifecycleCleanup = nil
	return cleanup
}

func CurrentPoolConfigs() []types.DownloadConfig {
	if GlobalPool == nil {
		return nil
	}
	return GlobalPool.GetAll()
}

func LifecycleForLocalService(service core.DownloadService) (*processing.LifecycleManager, error) {
	lifecycle := CurrentLifecycle()
	if service == nil || GlobalService == nil || service != GlobalService {
		return lifecycle, nil
	}
	return EnsureLocalLifecycle(GlobalService, CurrentPoolConfigs)
}

func EnsureGlobalLocalServiceAndLifecycle() error {
	if GlobalService == nil {
		localService := core.NewLocalDownloadServiceWithInput(GlobalPool, GlobalProgressCh)
		GlobalService = localService

		lifecycle, err := EnsureLocalLifecycle(localService, CurrentPoolConfigs)
		if err != nil {
			return err
		}

		lifecycle.SetEngineHooks(processing.EngineHooks{
			Pause:               GlobalPool.Pause,
			ExtractPausedConfig: GlobalPool.ExtractPausedConfig,
			GetStatus:           GlobalPool.GetStatus,
			AddConfig:           GlobalPool.Add,
			Cancel:              GlobalPool.Cancel,
			UpdateURL:           GlobalPool.UpdateURL,
			PublishEvent:        localService.Publish,
		})

		localService.SetLifecycleHooks(core.LifecycleHooks{
			Pause:       lifecycle.Pause,
			Resume:      lifecycle.Resume,
			ResumeBatch: lifecycle.ResumeBatch,
			Cancel:      lifecycle.Cancel,
			UpdateURL:   lifecycle.UpdateURL,
		})
	} else {
		_, err := EnsureLocalLifecycle(GlobalService, CurrentPoolConfigs)
		return err
	}
	return nil
}

func PublishSystemLog(message string) {
	if GlobalService != nil {
		_ = GlobalService.Publish(events.SystemLogMsg{Message: message})
		return
	}
	utils.Debug("System Log: %s", message)
}

func RecordPreflightDownloadError(url, outPath string, err error) {
	if err == nil || strings.TrimSpace(url) == "" {
		return
	}

	filename := strings.TrimSpace(processing.InferFilenameFromURL(url))
	destPath := ""
	if filename != "" && strings.TrimSpace(outPath) != "" {
		destPath = filepath.Join(outPath, filename)
	}

	entry := types.DownloadEntry{
		ID:       uuid.New().String(),
		URL:      url,
		URLHash:  state.URLHash(url),
		DestPath: destPath,
		Filename: filename,
		Status:   "error",
	}
	if addErr := state.AddToMasterList(entry); addErr != nil {
		utils.Debug("Failed to persist preflight download error for %s: %v", url, addErr)
	}
	if GlobalService != nil {
		_ = GlobalService.Publish(events.DownloadErrorMsg{
			DownloadID: entry.ID,
			Filename:   filename,
			DestPath:   destPath,
			Err:        err,
		})
	}
}

func EnsureLocalLifecycle(service core.DownloadService, getAll func() []types.DownloadConfig) (*processing.LifecycleManager, error) {
	GlobalLifecycleMu.Lock()
	defer GlobalLifecycleMu.Unlock()

	if GlobalLifecycle == nil {
		GlobalLifecycle = NewLocalLifecycleManager(service, getAll)
	}
	if GlobalLifecycleCleanup == nil {
		cleanup, err := StartLifecycleEventWorker(service, GlobalLifecycle)
		if err != nil {
			return nil, err
		}
		GlobalLifecycleCleanup = cleanup
	}
	return GlobalLifecycle, nil
}

func IsExplicitOutputPath(outPath, defaultDir string) bool {
	return utils.EnsureAbsPath(strings.TrimSpace(outPath)) != utils.EnsureAbsPath(strings.TrimSpace(defaultDir))
}

func GetSettings() *config.Settings {
	if GlobalSettings != nil {
		return GlobalSettings
	}
	settings, err := config.LoadSettings()
	if err != nil {
		return config.DefaultSettings()
	}
	return settings
}

// ParseURLArg parses a command line argument that might contain comma-separated mirrors
// Returns the primary URL and a list of all mirrors (including the primary)
func ParseURLArg(arg string) (string, []string) {
	parts := strings.Split(arg, ",")
	var urls []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			urls = append(urls, trimmed)
		}
	}
	if len(urls) == 0 {
		return "", nil
	}
	return urls[0], urls
}
