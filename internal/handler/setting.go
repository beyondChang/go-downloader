// Package handler 设置相关处理器
// 提供应用程序设置的查询和更新功能
package handler

import (
	"encoding/json"
	"net/http"
)

// ListSettings 获取所有设置项
// 返回应用程序的当前配置
func (h *Handler) ListSettings(w http.ResponseWriter, r *http.Request) {
	// 获取自动启动状态
	autoStartEnabled := h.autostartManager.IsEnabled()
	
	// TODO: 从配置服务获取更多设置
	settings := map[string]interface{}{
		"auto_start":   autoStartEnabled,   // 是否开机自启
		"open_on_start": false,              // 启动时打开浏览器
		"local_only":   true,                // 仅本地访问
	}
	
	writeJSON(w, http.StatusOK, settings)
}

// UpdateSettings 更新设置项
// 接收 JSON 格式的设置更新请求
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体")
		return
	}
	
	// 处理自动启动设置
	if autoStart, ok := req["auto_start"].(bool); ok {
		var err error
		if autoStart {
			err = h.autostartManager.Enable()
		} else {
			err = h.autostartManager.Disable()
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "更新自动启动设置失败")
			return
		}
	}
	
	// TODO: 更新其他设置到配置服务
	
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
