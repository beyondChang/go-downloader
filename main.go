// Package main 程序入口
// Go Downloader - 多线程下载管理器
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-downloader/internal/config"
	"go-downloader/internal/database"
	"go-downloader/internal/server"
	"go-downloader/internal/utils/logger"
	"go-downloader/web"

	"github.com/spf13/cobra"
)

// 版本信息，编译时通过 -ldflags 注入
var (
	Version = "dev"
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:     "go-downloader",
	Short:   "Go Downloader - 多线程下载管理器",
	Long:    "Go Downloader 是一个支持多线程下载、暂停恢复、文件分类的下载管理器。",
	Version: Version,
	Run:     run,
}

// run 主运行函数
func run(cmd *cobra.Command, args []string) {
	// 初始化配置
	config.Init()
	
	// 初始化日志
	logger.Init(config.GetDataDir())
	defer logger.Close()

	// 设置 Web 静态资源
	server.WebAssets = web.Assets

	// 创建 HTTP 服务器
	srv, asm := server.NewServer()

	// 打印启动信息
	fmt.Printf("Go Downloader v%s 正在启动...\n", Version)
	fmt.Printf("监听地址: %s\n", srv.Addr)
	fmt.Printf("数据目录: %s\n", config.GetDataDir())
	
	// 获取自动启动状态
	if asm.IsEnabled() {
		fmt.Println("开机自启: 已启用")
	} else {
		fmt.Println("开机自启: 未启用")
	}

	// 启动服务器（非阻塞）
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	fmt.Println("\n正在关闭服务器...")
	
	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器关闭失败: %v", err)
	}

	// 关闭数据库连接
	db := database.New()
	if err := db.Close(); err != nil {
		logger.Error("数据库关闭失败: %v", err)
	}

	fmt.Println("服务器已关闭")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
