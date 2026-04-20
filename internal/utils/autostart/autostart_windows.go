package autostart

import (
	"os"

	"golang.org/x/sys/windows/registry"

	"go-downloader/internal/utils/logger"
)

const (
	// runKey Windows 注册表中自启动项的路径
	runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
)

// windowsManager Windows 平台自启动管理器，通过注册表实现
type windowsManager struct {
	appName string // 注册表中的键名
}

// newManager 创建 Windows 平台自启动管理器
func newManager(appName string) Manager {
	return &windowsManager{appName: appName}
}

// Enable 在注册表中注册开机自启动，键值为当前 exe 的绝对路径
func (m *windowsManager) Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		logger.Error("获取可执行文件路径失败: %v", err)
		return err
	}

	// 打开当前用户的 Run 注册表键
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		logger.Error("打开注册表失败: %v", err)
		return err
	}
	defer k.Close()

	// 写入自启动项
	if err := k.SetStringValue(m.appName, `"`+exePath+`"`); err != nil {
		logger.Error("写入注册表失败: %v", err)
		return err
	}

	logger.Info("已注册开机自启动: %s", exePath)
	return nil
}

// Disable 从注册表中删除开机自启动项
func (m *windowsManager) Disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		logger.Error("打开注册表失败: %v", err)
		return err
	}
	defer k.Close()

	if err := k.DeleteValue(m.appName); err != nil {
		logger.Error("删除注册表项失败: %v", err)
		return err
	}

	logger.Info("已取消开机自启动")
	return nil
}

// IsEnabled 查询注册表中是否存在自启动项且路径匹配
func (m *windowsManager) IsEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue(m.appName)
	if err != nil {
		return false
	}

	// 检查注册表值是否包含当前 exe 路径
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	return val == `"`+exePath+`"` || val == exePath
}
