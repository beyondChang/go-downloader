// Package handler 下载相关处理器
// 提供下载任务的 CRUD 和控制操作
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ListDownloads 获取所有下载任务列表
// 返回活跃下载和已完成下载的状态
func (h *Handler) ListDownloads(w http.ResponseWriter, r *http.Request) {
	// TODO: 从下载服务获取列表
	// statuses, err := h.downloadService.List()
	
	// 临时返回空列表
	writeJSON(w, http.StatusOK, []interface{}{})
}

// CreateDownload 创建新的下载任务
// 从请求体解析 URL 和配置，启动下载
func (h *Handler) CreateDownload(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var req struct {
		URL       string `json:"url"`
		OutputDir string `json:"output_dir,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体")
		return
	}
	
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "URL 不能为空")
		return
	}
	
	// TODO: 调用下载服务创建任务
	// id, err := h.downloadService.Add(req.URL, req.OutputDir)
	
	// 临时返回
	writeJSON(w, http.StatusCreated, map[string]string{
		"id":     "temp-id",
		"status": "pending",
		"url":    req.URL,
	})
}

// GetDownload 获取单个下载任务的状态
// 返回下载进度、速度、状态等信息
func (h *Handler) GetDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少下载 ID")
		return
	}
	
	// TODO: 从下载服务获取状态
	// status, err := h.downloadService.GetStatus(id)
	
	// 临时返回
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       id,
		"status":   "downloading",
		"progress": 0,
	})
}

// DeleteDownload 删除下载任务
// 如果任务正在进行，会先暂停再删除
func (h *Handler) DeleteDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少下载 ID")
		return
	}
	
	// TODO: 调用下载服务删除任务
	// err := h.downloadService.Delete(id)
	
	writeJSON(w, http.StatusOK, map[string]string{
		"id":     id,
		"status": "deleted",
	})
}

// PauseDownload 暂停下载任务
func (h *Handler) PauseDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少下载 ID")
		return
	}
	
	// TODO: 调用下载服务暂停任务
	// err := h.downloadService.Pause(id)
	
	writeJSON(w, http.StatusOK, map[string]string{
		"id":     id,
		"status": "paused",
	})
}

// ResumeDownload 恢复暂停的下载任务
func (h *Handler) ResumeDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少下载 ID")
		return
	}
	
	// TODO: 调用下载服务恢复任务
	// err := h.downloadService.Resume(id)
	
	writeJSON(w, http.StatusOK, map[string]string{
		"id":     id,
		"status": "downloading",
	})
}

// Redownload 重新下载已完成的任务
func (h *Handler) Redownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少下载 ID")
		return
	}
	
	// TODO: 调用下载服务重新下载
	// err := h.downloadService.Redownload(id)
	
	writeJSON(w, http.StatusOK, map[string]string{
		"id":     id,
		"status": "downloading",
	})
}

// GetHistory 获取历史下载记录
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	// TODO: 从下载服务获取历史记录
	// history, err := h.downloadService.History()
	
	writeJSON(w, http.StatusOK, []interface{}{})
}

// ClearHistory 清空历史下载记录
func (h *Handler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	// TODO: 调用下载服务清空历史
	// err := h.downloadService.ClearHistory()
	
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// OpenFile 打开下载的文件
func (h *Handler) OpenFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少下载 ID")
		return
	}
	
	// TODO: 打开文件
	// err := h.fileService.OpenFile(id)
	
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// OpenFolder 打开文件所在目录
func (h *Handler) OpenFolder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少下载 ID")
		return
	}
	
	// TODO: 打开目录
	// err := h.fileService.OpenFolder(id)
	
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
