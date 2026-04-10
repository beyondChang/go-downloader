package app

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/go-downloader/internal/config"
	"github.com/go-downloader/internal/core"
	"github.com/go-downloader/internal/processing"
	"github.com/go-downloader/internal/utils"
)

// DownloadRequest represents a download request
type DownloadRequest struct {
	URL                  string            `json:"url"`
	Filename             string            `json:"filename,omitempty"`
	Path                 string            `json:"path,omitempty"`
	RelativeToDefaultDir bool              `json:"relative_to_default_dir,omitempty"`
	Mirrors              []string          `json:"mirrors,omitempty"`
	SkipApproval         bool              `json:"skip_approval,omitempty"`
	Headers              map[string]string `json:"headers,omitempty"`
	IsExplicitCategory   bool              `json:"is_explicit_category,omitempty"`
}

type ResolvedDownloadRequest struct {
	Request       DownloadRequest
	Settings      *config.Settings
	OutPath       string
	URLForAdd     string
	MirrorsForAdd []string
	IsDuplicate   bool
	IsActive      bool
}

func EnqueueDownloadRequest(ctx context.Context, service core.DownloadService, resolved *ResolvedDownloadRequest) (string, error) {
	lifecycle, err := LifecycleForLocalService(service)
	if err != nil {
		return "", fmt.Errorf("failed to initialize lifecycle manager: %w", err)
	}

	req := resolved.Request
	if lifecycle != nil {
		return lifecycle.Enqueue(ctx, &processing.DownloadRequest{
			URL:                resolved.URLForAdd,
			Filename:           req.Filename,
			Path:               resolved.OutPath,
			Mirrors:            resolved.MirrorsForAdd,
			Headers:            req.Headers,
			IsExplicitCategory: req.IsExplicitCategory,
			SkipApproval:       req.SkipApproval,
		})
	}

	return service.Add(resolved.URLForAdd, resolved.OutPath, req.Filename, resolved.MirrorsForAdd, req.Headers, req.IsExplicitCategory, 0, false)
}

// ProcessDownloads handles the logic of adding downloads to the local pool
func ProcessDownloads(urls []string, outputDir string) int {
	successCount := 0

	if GlobalService == nil {
		utils.Debug("Error: GlobalService not initialized")
		return 0
	}

	settings := GetSettings()

	lifecycle, err := LifecycleForLocalService(GlobalService)
	if err != nil {
		utils.Debug("Error: unable to initialize lifecycle manager: %v", err)
		return 0
	}

	for _, arg := range urls {
		if arg == "" {
			continue
		}

		url, mirrors := ParseURLArg(arg)
		if url == "" {
			continue
		}

		outPath := ResolveOutputDir(outputDir, false, "", settings)
		outPath = utils.EnsureAbsPath(outPath)

		isExplicit := IsExplicitOutputPath(outPath, settings.General.DefaultDownloadDir)
		if lifecycle == nil {
			err := fmt.Errorf("lifecycle manager unavailable")
			RecordPreflightDownloadError(url, outPath, err)
			PublishSystemLog(fmt.Sprintf("Error adding %s: %v", url, err))
			continue
		}

		_, err := lifecycle.Enqueue(CurrentEnqueueContext(), &processing.DownloadRequest{
			URL:                url,
			Path:               outPath,
			Mirrors:            mirrors,
			IsExplicitCategory: isExplicit,
		})
		if err != nil {
			RecordPreflightDownloadError(url, outPath, err)
			PublishSystemLog(fmt.Sprintf("Error adding %s: %v", url, err))
			continue
		}
		atomic.AddInt32(&ActiveDownloads, 1)
		successCount++
	}
	return successCount
}

func ResolveOutputDir(reqPath string, relativeToDefaultDir bool, defaultOutputDir string, settings *config.Settings) string {
	outPath := reqPath

	if relativeToDefaultDir && reqPath != "" {
		baseDir := settings.General.DefaultDownloadDir
		if baseDir == "" {
			baseDir = defaultOutputDir
		}
		if baseDir == "" {
			baseDir = "."
		}
		outPath = filepath.Join(baseDir, reqPath)
	} else if outPath == "" {
		if defaultOutputDir != "" {
			outPath = defaultOutputDir
		} else if settings.General.DefaultDownloadDir != "" {
			outPath = settings.General.DefaultDownloadDir
		} else {
			outPath = "."
		}
	}

	return outPath
}
