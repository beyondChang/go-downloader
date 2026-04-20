//go:build !windows

// Package config 开机自启配置
// 非 Windows 平台的占位实现
package config

// SetAutoRun 设置或取消开机自启（非 Windows 平台暂不支持）
func SetAutoRun(enable bool) error {
	return nil
}

// GetAutoRun 检查是否已设置开机自启（非 Windows 平台暂不支持）
func GetAutoRun() (bool, error) {
	return false, nil
}
