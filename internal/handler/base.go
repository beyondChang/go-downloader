// Package handler HTTP 请求处理器
// 提供所有 API 端点的处理函数
package handler

import (
	"encoding/json"
	"net/http"

	"go-downloader/internal/database"
	"go-downloader/internal/utils/autostart"
)

// Handler 结构体，持有依赖服务
// 通过依赖注入方式管理数据库连接和自动启动管理器
type Handler struct {
	db              database.Service      // 数据库服务
	autostartManager autostart.Manager    // 自动启动管理器
}

// New 创建 Handler 实例，注入依赖服务
func New(db database.Service, asm autostart.Manager) *Handler {
	return &Handler{
		db:              db,
		autostartManager: asm,
	}
}

// HelloWorld 返回 Hello World 测试响应
// 用于验证服务是否正常运行
func (h *Handler) HelloWorld(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Hello World"})
}

// Health 返回数据库健康状态信息
// 用于健康检查和监控
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.db.Health())
}

// writeJSON 辅助函数，将数据序列化为 JSON 并写入响应
// 统一处理 Content-Type 和错误响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError 辅助函数，写入错误响应
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
