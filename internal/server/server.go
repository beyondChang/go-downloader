// Package server 提供 HTTP 服务器功能
// 包含服务器创建、路由注册和中间件配置
package server

import (
	"fmt"
	"net/http"
	"time"

	"go-downloader/internal/config"
	"go-downloader/internal/database"
	"go-downloader/internal/handler"
	"go-downloader/internal/utils/autostart"
)

// Server HTTP 服务器结构体
// 包含端口、处理器和自动启动管理器
type Server struct {
	port             int                       // 监听端口
	handler          *handler.Handler          // 请求处理器
	autostartManager autostart.Manager         // 自动启动管理器
	settings         *config.Settings          // 应用设置
}

// NewServer 创建并配置 HTTP 服务器
// 返回配置好的 http.Server 和自动启动管理器
func NewServer() (*http.Server, autostart.Manager) {
	// 加载设置
	settings, err := config.LoadSettings()
	if err != nil {
		settings = config.DefaultSettings()
	}
	
	// 获取配置的端口
	port := settings.General.Port
	
	// 创建数据库实例
	db := database.New()
	
	// 根据 local_only 配置决定绑定地址
	addr := fmt.Sprintf(":%d", port)
	if settings.General.LocalOnly {
		addr = fmt.Sprintf("127.0.0.1:%d", port)
	}
	
	// 创建自动启动管理器
	asm := autostart.New("go-downloader")

	// 创建 Server 实例
	s := &Server{
		port:             port,
		handler:          handler.New(db, asm),
		autostartManager: asm,
		settings:         settings,
	}

	// 创建并配置 http.Server
	server := &http.Server{
		Addr:         addr,
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,      // 空闲连接超时
		ReadTimeout:  10 * time.Second, // 读取超时
		WriteTimeout: 30 * time.Second, // 写入超时
	}

	return server, asm
}
