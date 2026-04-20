package app

import (
	"fmt"
	"path/filepath"
	"sync/atomic"

	"go-downloader/internal/config"
	"go-downloader/internal/engine/state"
	"go-downloader/internal/utils"
)

func RunStartupIntegrityCheck() string {
	// Normalize downloads stuck in "downloading" status from a prior crash/kill.
	if normalized, err := state.NormalizeStaleDownloads(); err != nil {
		msg := fmt.Sprintf("Startup: failed to normalize stale downloads: %v", err)
		utils.Debug("%s", msg)
	} else if normalized > 0 {
		utils.Debug("Startup: normalized %d stale downloading entries to paused", normalized)
	}

	// Validate integrity
	if removed, err := state.ValidateIntegrity(); err != nil {
		msg := fmt.Sprintf("Startup integrity check failed: %v", err)
		return msg
	} else if removed > 0 {
		msg := fmt.Sprintf("Startup integrity check: removed %d corrupted/orphaned downloads", removed)
		return msg
	}
	utils.Debug("%s", "Startup integrity check: no issues found")
	return ""
}

func InitializeGlobalState() error {
	stateDir := config.GetStateDir()
	logsDir := config.GetLogsDir()
	stateDBPath := filepath.Join(stateDir, "downloader.db")

	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create downloader directories: %w", err)
	}

	// Config engine state
	state.Configure(stateDBPath)

	// Config logging
	utils.ConfigureDebug(logsDir)

	// Clean up old logs
	retention := GetSettings().General.LogRetentionCount
	if retention > 0 {
		utils.CleanupLogs(retention - 1)
	} else {
		utils.CleanupLogs(retention)
	}
	return nil
}

func ResumePausedDownloads() {
	settings := GetSettings()

	pausedEntries, err := state.LoadPausedDownloads()
	if err != nil {
		return
	}

	for _, entry := range pausedEntries {
		if entry.Status == "paused" && !settings.General.AutoResume {
			continue
		}
		if GlobalService == nil || entry.ID == "" {
			continue
		}
		if err := GlobalService.Resume(entry.ID); err == nil {
			atomic.AddInt32(&ActiveDownloads, 1)
		}
	}
}
