// Package server 路由注册
// 使用 chi 路由器注册所有 HTTP 路由和中间件
package server

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// WebAssets 前端静态资源，需在 RegisterRoutes 之前设置
// 由 main.go 或初始化代码注入嵌入的文件系统
var WebAssets fs.FS

// RegisterRoutes 注册所有路由
// 配置 CORS、日志中间件，并注册 API 和静态文件路由
func (s *Server) RegisterRoutes() http.Handler {
	// 创建 chi 路由器
	r := chi.NewRouter()
	
	// 使用日志中间件
	r.Use(middleware.Logger)

	// 配置 CORS 中间件
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300, // 预检请求缓存时间
	}))

	// 获取处理器引用
	h := s.handler

	// 注册 API 路由组
	r.Route("/api", func(r chi.Router) {
		// 基础接口
		r.Get("/", h.HelloWorld)           // Hello World 测试
		r.Get("/health", h.Health)         // 健康检查
		
		// 下载管理接口
		r.Get("/downloads", h.ListDownloads)           // 获取下载列表
		r.Post("/downloads", h.CreateDownload)         // 创建下载任务
		r.Get("/downloads/{id}", h.GetDownload)        // 获取单个下载状态
		r.Delete("/downloads/{id}", h.DeleteDownload)  // 删除下载
		r.Post("/downloads/{id}/pause", h.PauseDownload)   // 暂停下载
		r.Post("/downloads/{id}/resume", h.ResumeDownload) // 恢复下载
		r.Post("/downloads/{id}/redownload", h.Redownload) // 重新下载
		r.Get("/downloads/history", h.GetHistory)          // 获取历史记录
		r.Delete("/downloads/history", h.ClearHistory)     // 清空历史记录
		
		// 文件操作接口
		r.Post("/files/{id}/open", h.OpenFile)     // 打开文件
		r.Post("/files/{id}/folder", h.OpenFolder) // 打开目录
		
		// 设置接口
		r.Get("/settings", h.ListSettings)   // 获取设置
		r.Put("/settings", h.UpdateSettings) // 更新设置
		
		// 实时事件推送
		r.Get("/events", h.Events) // SSE 事件流
	})

	// 静态文件服务
	// 所有未匹配的路由都由前端处理（支持 SPA）
	fileServer := http.FileServer(http.FS(WebAssets))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})

	return r
}
