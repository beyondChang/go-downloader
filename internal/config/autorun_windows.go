//go:build windows

// Package config 开机自启配置
// Windows 平台通过注册表实现开机自启
package config

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// autoRunKeyName 注册表键名
const autoRunKeyName = "go-downloader"

// getExePath 获取可执行文件路径
func getExePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exePath)
}

// SetAutoRun 设置或取消开机自启
// 通过 Windows 注册表的 Run 键实现
func SetAutoRun(enable bool) error {
	exePath, err := getExePath()
	if err != nil {
		return err
	}

	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	if enable {
		return key.SetStringValue(autoRunKeyName, exePath)
	}
	return key.DeleteValue(autoRunKeyName)
}

// GetAutoRun 检查是否已设置开机自启
func GetAutoRun() (bool, error) {
	exePath, err := getExePath()
	if err != nil {
		return false, err
	}

	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer key.Close()

	val, _, err := key.GetStringValue(autoRunKeyName)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}

	return val == exePath, nil
}
