//go:build !windows

package autostart

import "go-downloader/internal/utils/logger"

// otherManager 非 Windows 平台的空实现
type otherManager struct {
	appName string
}

// newManager 创建空实现的自启动管理器（非 Windows 平台暂不支持）
func newManager(appName string) Manager {
	return &otherManager{appName: appName}
}

// Enable 非 Windows 平台暂不支持自启动注册
func (m *otherManager) Enable() error {
	logger.Info("当前平台不支持开机自启动")
	return nil
}

// Disable 非 Windows 平台暂不支持取消自启动
func (m *otherManager) Disable() error {
	return nil
}

// IsEnabled 非 Windows 平台始终返回 false
func (m *otherManager) IsEnabled() bool {
	return false
}
