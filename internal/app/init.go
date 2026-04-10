package app

import (
	"os"

	"github.com/go-downloader/internal/download"
	"github.com/go-downloader/internal/utils"
)

func Initialize() {
	if err := InitializeGlobalState(); err != nil {
		utils.Debug("Error initializing downloader: %v", err)
		utils.Notify("启动失败", "无法初始化应用状态，请检查日志。")
		os.Exit(1)
	}

	GlobalSettings = GetSettings()

	GlobalProgressCh = make(chan any, 1000)
	GlobalPool = download.NewWorkerPool(GlobalProgressCh, GlobalSettings.Network.MaxConcurrentDownloads)

	if err := EnsureGlobalLocalServiceAndLifecycle(); err != nil {
		utils.Debug("Error initializing download engine: %v", err)
		utils.Notify("启动失败", "无法启动下载引擎，请检查日志。")
		os.Exit(1)
	}

	StartupIntegrityMessage = RunStartupIntegrityCheck()
	ResumePausedDownloads()
}
