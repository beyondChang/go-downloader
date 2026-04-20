// Package config 路径配置
// 提供各平台的标准路径配置（XDG 规范）
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
)

// getXDGBaseDir 获取 XDG 基础目录
func getXDGBaseDir(envKey, fallback string) string {
	if dir := strings.TrimSpace(os.Getenv(envKey)); dir != "" {
		if filepath.IsAbs(dir) {
			return dir
		}
	}
	return fallback
}

// GetDownloaderDir 返回配置文件目录
// Linux: $XDG_CONFIG_HOME/downloader 或 ~/.config/downloader
// macOS: ~/Library/Application Support/downloader
// Windows: %APPDATA%/downloader
func GetDownloaderDir() string {
	if runtime.GOOS == "windows" {
		// Windows 保留旧位置以兼容现有安装
		if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
			if filepath.IsAbs(appData) {
				return filepath.Join(appData, "downloader")
			}
		}
	}
	return filepath.Join(getXDGBaseDir("XDG_CONFIG_HOME", xdg.ConfigHome), "downloader")
}

// GetStateDir 返回状态文件目录
func GetStateDir() string {
	// Windows 保持状态与配置同位置以保持向后兼容
	if runtime.GOOS == "windows" {
		return GetDownloaderDir()
	}
	return filepath.Join(getXDGBaseDir("XDG_STATE_HOME", xdg.StateHome), "downloader")
}

// GetDownloadsDir 返回下载目录
func GetDownloadsDir() string {
	// 优先使用 XDG 用户目录
	if dir := strings.TrimSpace(xdg.UserDirs.Download); dir != "" {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	// 回退到 ~/Downloads
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		fallback := filepath.Join(home, "Downloads")
		if info, err := os.Stat(fallback); err == nil && info.IsDir() {
			return fallback
		}
	}

	// 最终回退：空表示当前目录
	return ""
}

// GetRuntimeDir 返回运行时目录
func GetRuntimeDir() string {
	runtimeEnv := strings.TrimSpace(os.Getenv("XDG_RUNTIME_DIR"))
	if runtimeEnv != "" && !filepath.IsAbs(runtimeEnv) {
		runtimeEnv = ""
	}

	runtimeBase := runtimeEnv
	if runtimeBase == "" {
		runtimeBase = strings.TrimSpace(xdg.RuntimeDir)
		if runtimeBase != "" && !filepath.IsAbs(runtimeBase) {
			runtimeBase = ""
		}
	}

	// 无头 Linux 会话中，XDG_RUNTIME_DIR 通常未设置
	if runtime.GOOS == "linux" && runtimeEnv == "" {
		runtimeBase = ""
	}

	if runtimeBase == "" {
		return filepath.Join(GetStateDir(), "runtime")
	}

	return filepath.Join(runtimeBase, "downloader")
}

// GetDocumentsDir 返回文档目录
func GetDocumentsDir() string {
	return xdg.UserDirs.Documents
}

// GetMusicDir 返回音乐目录
func GetMusicDir() string {
	return xdg.UserDirs.Music
}

// GetVideosDir 返回视频目录
func GetVideosDir() string {
	return xdg.UserDirs.Videos
}

// GetPicturesDir 返回图片目录
func GetPicturesDir() string {
	return xdg.UserDirs.Pictures
}

// GetLogsDir 返回日志目录
func GetLogsDir() string {
	return filepath.Join(GetStateDir(), "logs")
}

// GetDataDir 返回数据目录（兼容旧代码）
func GetDataDir() string {
	return GetDownloaderDir()
}

// GetPort 从设置获取服务端口
// 如果设置未加载，返回默认端口 8080
func GetPort() int {
	settings, err := LoadSettings()
	if err != nil {
		return 8080
	}
	return settings.General.Port
}

// EnsureDirs 创建所有必需的目录
func EnsureDirs() error {
	dirs := []string{GetDownloaderDir(), GetStateDir(), GetRuntimeDir(), GetLogsDir()}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	// Linux 运行时目录可能需要更严格的权限
	if runtime.GOOS == "linux" && os.Getenv("XDG_RUNTIME_DIR") != "" {
		_ = os.Chmod(GetRuntimeDir(), 0700)
	}

	return nil
}

// Init 初始化配置系统
// 创建必需的目录，确保配置文件存在
func Init() error {
	return EnsureDirs()
}
