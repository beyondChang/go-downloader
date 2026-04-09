package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/go-downloader/internal/app"
	"github.com/go-downloader/internal/config"
	"github.com/go-downloader/internal/core"
	"github.com/go-downloader/internal/engine/events"
	"github.com/go-downloader/internal/engine/types"
	"github.com/go-downloader/internal/processing"
	"github.com/go-downloader/internal/server/web"
	"github.com/go-downloader/internal/utils"
)

var (
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrDownloadNotFound   = errors.New("download not found")
	ErrNoDestinationPath  = errors.New("download has no destination path")
)

func RegisterHTTPRoutes(mux *http.ServeMux, port int, defaultOutputDir string, service core.DownloadService, authToken string) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			// Serve static assets
			http.FileServer(http.FS(web.Files)).ServeHTTP(w, r)
			return
		}
		// Serve index.html
		data, err := web.Files.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Failed to load web interface", http.StatusInternalServerError)
			return
		}

		// Inject authToken into JS
		content := string(data)
		content = strings.Replace(content, "// AUTH_TOKEN will be injected here", fmt.Sprintf("var AUTH_TOKEN = '%s';", authToken), 1)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(content))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
			"port":   port,
		})
	})

	mux.HandleFunc("/events", eventsHandler(service))

	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		handleDownload(w, r, defaultOutputDir, service)
	})

	mux.HandleFunc("/pause", requireMethod(http.MethodPost, withRequiredID(func(w http.ResponseWriter, _ *http.Request, id string) {
		if err := service.Pause(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "paused", "id": id})
	})))

	mux.HandleFunc("/resume", requireMethod(http.MethodPost, withRequiredID(func(w http.ResponseWriter, _ *http.Request, id string) {
		if err := service.Resume(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "resumed", "id": id})
	})))

	mux.HandleFunc("/delete", requireMethods(withRequiredID(func(w http.ResponseWriter, _ *http.Request, id string) {
		if err := service.Delete(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
	}), http.MethodDelete, http.MethodPost))

	mux.HandleFunc("/list", requireMethod(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
		statuses, err := service.List()
		if err != nil {
			http.Error(w, "Failed to list downloads: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if statuses == nil {
			statuses = []types.DownloadStatus{}
		}
		WriteJSONResponse(w, http.StatusOK, statuses)
	}))

	mux.HandleFunc("/history", requireMethod(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
		history, err := service.History()
		if err != nil {
			http.Error(w, "Failed to retrieve history: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if history == nil {
			history = []types.DownloadEntry{}
		}

		// Check file existence for each entry
		for i := range history {
			if history[i].DestPath != "" {
				_, err := os.Stat(history[i].DestPath)
				history[i].FileExists = err == nil
			}
		}

		sort.Slice(history, func(left, right int) bool {
			if history[left].CompletedAt == history[right].CompletedAt {
				return history[left].ID > history[right].ID
			}
			return history[left].CompletedAt > history[right].CompletedAt
		})
		WriteJSONResponse(w, http.StatusOK, history)
	}))

	mux.HandleFunc("/clear-history", requireMethod(http.MethodPost, func(w http.ResponseWriter, _ *http.Request) {
		if err := service.ClearHistory(); err != nil {
			http.Error(w, "Failed to clear history: "+err.Error(), http.StatusInternalServerError)
			return
		}
		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "History cleared"})
	}))

	mux.HandleFunc("/redownload", requireMethod(http.MethodPost, withRequiredID(func(w http.ResponseWriter, _ *http.Request, id string) {
		history, err := service.History()
		if err != nil {
			http.Error(w, "Failed to retrieve history: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var target *types.DownloadEntry
		for _, entry := range history {
			if entry.ID == id {
				target = &entry
				break
			}
		}

		if target == nil {
			http.Error(w, "Download not found in history", http.StatusNotFound)
			return
		}

		// Re-add download
		// Ensure directory exists
		dir := filepath.Dir(target.DestPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			http.Error(w, "Failed to create directory: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := service.Add(target.URL, dir, target.Filename, target.Mirrors, nil, false, 0, false); err != nil {
			http.Error(w, "Failed to re-download: "+err.Error(), http.StatusInternalServerError)
			return
		}

		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok", "message": "Re-download started"})
	})))

	mux.HandleFunc("/open-file", requireMethod(http.MethodPost, withRequiredID(func(w http.ResponseWriter, r *http.Request, id string) {
		if err := ensureOpenActionRequestAllowed(r); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		destPath, err := resolveDownloadDestPath(service, id)
		if err != nil {
			http.Error(w, err.Error(), statusCodeForResolveDownloadError(err))
			return
		}

		if err := utils.OpenFile(destPath); err != nil {
			http.Error(w, "Failed to open file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok", "id": id})
	})))

	mux.HandleFunc("/open-folder", requireMethod(http.MethodPost, withRequiredID(func(w http.ResponseWriter, r *http.Request, id string) {
		if err := ensureOpenActionRequestAllowed(r); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		destPath, err := resolveDownloadDestPath(service, id)
		if err != nil {
			http.Error(w, err.Error(), statusCodeForResolveDownloadError(err))
			return
		}

		if err := utils.OpenContainingFolder(destPath); err != nil {
			http.Error(w, "Failed to open folder: "+err.Error(), http.StatusInternalServerError)
			return
		}

		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok", "id": id})
	})))

	mux.HandleFunc("/update-url", requireMethod(http.MethodPut, withRequiredID(func(w http.ResponseWriter, r *http.Request, id string) {
		var req map[string]string
		if err := DecodeJSONBody(r, &req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		newURL := req["url"]
		if newURL == "" {
			http.Error(w, "Missing url parameter in body", http.StatusBadRequest)
			return
		}

		if err := service.UpdateURL(id, newURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "updated", "id": id, "url": newURL})
	})))

	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			WriteJSONResponse(w, http.StatusOK, app.GlobalSettings)
		case http.MethodPost:
			var newSettings config.Settings
			if err := DecodeJSONBody(r, &newSettings); err != nil {
				http.Error(w, "Invalid settings data", http.StatusBadRequest)
				return
			}
			// Update global settings
			*app.GlobalSettings = newSettings
			if err := config.SaveSettings(app.GlobalSettings); err != nil {
				http.Error(w, "Failed to save settings: "+err.Error(), http.StatusInternalServerError)
				return
			}
			WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func eventsHandler(service core.DownloadService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		stream, cleanup, err := service.StreamEvents(r.Context())
		if err != nil {
			http.Error(w, "Failed to subscribe to events", http.StatusInternalServerError)
			return
		}
		defer cleanup()

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		flusher.Flush()

		done := r.Context().Done()
		for {
			select {
			case <-done:
				return
			case msg, ok := <-stream:
				if !ok {
					return
				}

				frames, err := events.EncodeSSEMessages(msg)
				if err != nil {
					utils.Debug("Error encoding SSE event: %v", err)
					continue
				}
				if len(frames) == 0 {
					continue
				}

				for _, frame := range frames {
					_, _ = fmt.Fprintf(w, "event: %s\n", frame.Event)
					_, _ = fmt.Fprintf(w, "data: %s\n\n", frame.Data)
				}
				flusher.Flush()
			}
		}
	}
}

func requireMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return requireMethods(next, method)
}

func requireMethods(next http.HandlerFunc, methods ...string) http.HandlerFunc {
	allowed := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		allowed[method] = struct{}{}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[r.Method]; !ok {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func withRequiredID(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}
		next(w, r, id)
	}
}

func WriteJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		utils.Debug("Failed to encode response: %v", err)
	}
}

func resolveDownloadDestPath(service core.DownloadService, id string) (string, error) {
	if service == nil {
		return "", ErrServiceUnavailable
	}

	status, err := service.GetStatus(id)
	if err == nil && status != nil {
		if destPath := filepath.Clean(status.DestPath); destPath != "" && destPath != "." {
			return destPath, nil
		}
	}

	history, err := service.History()
	if err != nil {
		return "", fmt.Errorf("failed to read history: %w", err)
	}

	for _, entry := range history {
		if entry.ID != id {
			continue
		}
		destPath := filepath.Clean(entry.DestPath)
		if destPath == "" || destPath == "." {
			return "", fmt.Errorf("%w: %s", ErrNoDestinationPath, id)
		}
		return destPath, nil
	}

	return "", fmt.Errorf("%w: %s", ErrDownloadNotFound, id)
}

func statusCodeForResolveDownloadError(err error) int {
	switch {
	case errors.Is(err, ErrDownloadNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrServiceUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func ensureOpenActionRequestAllowed(r *http.Request) error {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
		xri := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if xff == "" && xri == "" {
			return nil
		}
	}

	settings := app.GetSettings()
	if settings != nil && settings.General.AllowRemoteOpenActions {
		return nil
	}

	return fmt.Errorf("open actions are only allowed from local host")
}

func DecodeJSONBody(r *http.Request, dst interface{}) error {
	defer func() {
		_ = r.Body.Close()
	}()
	return json.NewDecoder(r.Body).Decode(dst)
}

func handleDownload(w http.ResponseWriter, r *http.Request, defaultOutputDir string, service core.DownloadService) {
	if handleDownloadStatusRequest(w, r, service) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if service == nil {
		http.Error(w, "Service unavailable", http.StatusInternalServerError)
		return
	}

	resolved, err := resolveDownloadRequest(r, defaultOutputDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if maybeRequireDownloadApproval(w, service, resolved) {
		return
	}

	newID, err := app.EnqueueDownloadRequest(r.Context(), service, resolved)
	if err != nil {
		app.RecordPreflightDownloadError(resolved.URLForAdd, resolved.OutPath, err)
		app.PublishSystemLog(fmt.Sprintf("Error adding %s: %v", resolved.URLForAdd, err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	atomic.AddInt32(&app.ActiveDownloads, 1)
	WriteJSONResponse(w, http.StatusOK, map[string]string{
		"status":  "queued",
		"message": "Download queued successfully",
		"id":      newID,
	})
}

func handleDownloadStatusRequest(w http.ResponseWriter, r *http.Request, service core.DownloadService) bool {
	if r.Method != http.MethodGet {
		return false
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return true
	}

	if service == nil {
		http.Error(w, "Service unavailable", http.StatusInternalServerError)
		return true
	}

	status, err := service.GetStatus(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return true
	}

	WriteJSONResponse(w, http.StatusOK, status)
	return true
}

func decodeAndValidateDownloadRequest(r *http.Request) (app.DownloadRequest, error) {
	var req app.DownloadRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		return req, fmt.Errorf("invalid json: %w", err)
	}
	if req.URL == "" {
		return req, fmt.Errorf("url is required")
	}
	if strings.Contains(req.Filename, "..") {
		return req, fmt.Errorf("invalid filename")
	}
	if strings.Contains(req.Filename, "/") || strings.Contains(req.Filename, "\\") {
		return req, fmt.Errorf("invalid filename")
	}
	if strings.Contains(req.Path, "..") {
		return req, fmt.Errorf("invalid path")
	}
	if req.RelativeToDefaultDir && req.Path != "" {
		if filepath.IsAbs(req.Path) {
			return req, fmt.Errorf("invalid path")
		}
		cleanPath := filepath.Clean(req.Path)
		if cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
			return req, fmt.Errorf("invalid path")
		}
		req.Path = cleanPath
	}
	return req, nil
}

func resolveDownloadRequest(r *http.Request, defaultOutputDir string) (*app.ResolvedDownloadRequest, error) {
	settings := app.GetSettings()
	req, err := decodeAndValidateDownloadRequest(r)
	if err != nil {
		return nil, err
	}

	utils.Debug("Received download request: URL=%s, Path=%s", req.URL, req.Path)

	outPath := utils.EnsureAbsPath(app.ResolveOutputDir(req.Path, req.RelativeToDefaultDir, defaultOutputDir, settings))
	urlForAdd, mirrorsForAdd := normalizeDownloadTargets(req.URL, req.Mirrors)
	isDuplicate, isActive := resolveDuplicateState(urlForAdd, settings)

	utils.Debug("Download request: URL=%s, SkipApproval=%v, isDuplicate=%v, isActive=%v", urlForAdd, req.SkipApproval, isDuplicate, isActive)

	return &app.ResolvedDownloadRequest{
		Request:       req,
		Settings:      settings,
		OutPath:       outPath,
		URLForAdd:     urlForAdd,
		MirrorsForAdd: mirrorsForAdd,
		IsDuplicate:   isDuplicate,
		IsActive:      isActive,
	}, nil
}

func normalizeDownloadTargets(url string, mirrors []string) (string, []string) {
	if len(mirrors) == 0 && strings.Contains(url, ",") {
		return app.ParseURLArg(url)
	}
	return url, mirrors
}

func resolveDuplicateState(urlForAdd string, settings *config.Settings) (bool, bool) {
	activeDownloadsFunc := func() map[string]*types.DownloadConfig {
		active := make(map[string]*types.DownloadConfig)
		for _, cfg := range app.GlobalPool.GetAll() {
			c := cfg
			active[c.ID] = &c
		}
		return active
	}

	dupResult := processing.CheckForDuplicate(urlForAdd, settings, activeDownloadsFunc)
	if dupResult == nil {
		return false, false
	}
	return dupResult.Exists, dupResult.IsActive
}

func maybeRequireDownloadApproval(w http.ResponseWriter, service core.DownloadService, resolved *app.ResolvedDownloadRequest) bool {
	req := resolved.Request

	if req.SkipApproval {
		utils.Debug("Extension request: skipping all prompts, proceeding with download")
		return false
	}

	shouldPrompt := resolved.Settings.General.ExtensionPrompt || (resolved.Settings.General.WarnOnDuplicate && resolved.IsDuplicate)
	if !shouldPrompt {
		return false
	}

	utils.Debug("Auto-approving download: %s (Duplicate: %v)", req.URL, resolved.IsDuplicate)
	return false
}
