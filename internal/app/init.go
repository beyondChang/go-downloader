package app

import (
	"fmt"
	"os"

	"github.com/go-downloader/internal/download"
)

func Initialize() {
	if err := InitializeGlobalState(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing downloader: %v\n", err)
		os.Exit(1)
	}

	GlobalSettings = GetSettings()

	GlobalProgressCh = make(chan any, 1000)
	GlobalPool = download.NewWorkerPool(GlobalProgressCh, GlobalSettings.Network.MaxConcurrentDownloads)

	if err := EnsureGlobalLocalServiceAndLifecycle(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing download engine: %v\n", err)
		os.Exit(1)
	}

	StartHeadlessConsumer()

	StartupIntegrityMessage = RunStartupIntegrityCheck()
	ResumePausedDownloads()
}
