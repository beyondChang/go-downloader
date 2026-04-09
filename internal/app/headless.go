package app

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-downloader/internal/engine/events"
	"github.com/go-downloader/internal/utils"
)

// StartHeadlessConsumer starts a goroutine to consume progress messages and log to stdout
func StartHeadlessConsumer() {
	go func() {
		if GlobalService == nil {
			return
		}
		stream, cleanup, err := GlobalService.StreamEvents(context.Background())
		if err != nil {
			utils.Debug("Failed to start event stream: %v", err)
			return
		}
		defer cleanup()

		for msg := range stream {
			switch m := msg.(type) {
			case events.DownloadStartedMsg:
				fmt.Printf("Started: %s [%s]\n", m.Filename, TruncateID(m.DownloadID))
			case events.DownloadCompleteMsg:
				atomic.AddInt32(&ActiveDownloads, -1)
				fmt.Printf("Completed: %s [%s] (in %s)\n", m.Filename, TruncateID(m.DownloadID), m.Elapsed)
			case events.DownloadErrorMsg:
				atomic.AddInt32(&ActiveDownloads, -1)
				fmt.Printf("Error: %s [%s]: %v\n", m.Filename, TruncateID(m.DownloadID), m.Err)
			case events.DownloadQueuedMsg:
				fmt.Printf("Queued: %s [%s]\n", m.Filename, TruncateID(m.DownloadID))
			case events.DownloadPausedMsg:
				fmt.Printf("Paused: %s [%s]\n", m.Filename, TruncateID(m.DownloadID))
			case events.DownloadResumedMsg:
				fmt.Printf("Resumed: %s [%s]\n", m.Filename, TruncateID(m.DownloadID))
			case events.DownloadRemovedMsg:
				fmt.Printf("Removed: %s [%s]\n", m.Filename, TruncateID(m.DownloadID))
			}
		}
	}()
}

// TruncateID shortens a UUID to its first 8 characters for display
func TruncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
