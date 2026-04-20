// Package config 配置管理 - 兼容层
// 提供 Settings 类型别名以兼容旧代码
package config

import (
	"time"
)

// Settings 兼容旧代码的类型别名
type Settings = Config

// GeneralSettings 兼容旧代码的类型别名
type GeneralSettings = GeneralConfig

// NetworkSettings 兼容旧代码的类型别名
type NetworkSettings = NetworkConfig

// PerformanceSettings 兼容旧代码的类型别名
type PerformanceSettings = PerformanceConfig

// CategorySettings 兼容旧代码的类型别名
type CategorySettings = CategoryConfig

// 主题常量
const (
	ThemeAdaptive = 0 // 跟随系统
	ThemeLight    = 1 // 浅色
	ThemeDark     = 2 // 深色
)

// LoadSettings 从磁盘加载设置（兼容旧代码）
func LoadSettings() (*Settings, error) {
	if err := InitConfig(); err != nil {
		return DefaultSettings(), err
	}
	return GetConfig(), nil
}

// SaveSettings 保存设置到磁盘（兼容旧代码）
func SaveSettings(s *Settings) error {
	return SaveConfig(s)
}

// DefaultSettings 返回默认设置（兼容旧代码）
func DefaultSettings() *Settings {
	return defaultConfig()
}

// GetSettingsPath 返回设置文件路径（兼容旧代码）
func GetSettingsPath() string {
	return GetConfigPath()
}

// RuntimeConfig 下载引擎运行时配置
// 用于将用户设置传递给下载引擎
type RuntimeConfig struct {
	MaxConnectionsPerHost int
	UserAgent             string
	ProxyURL              string
	SequentialDownload    bool
	MinChunkSize          int64
	WorkerBufferSize      int
	MaxTaskRetries        int
	SlowWorkerThreshold   float64
	SlowWorkerGracePeriod time.Duration
	StallTimeout          time.Duration
	SpeedEmaAlpha         float64
	ConnectTimeout        time.Duration
	ResponseTimeout       time.Duration
	RequestTimeout        time.Duration
	MaxRedirects          int
}

// ToRuntimeConfig 从用户 Settings 创建 RuntimeConfig
func (s *Settings) ToRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		MaxConnectionsPerHost: s.Network.MaxConnectionsPerHost,
		UserAgent:             s.Network.UserAgent,
		ProxyURL:              s.Network.ProxyURL,
		SequentialDownload:    s.Network.SequentialDownload,
		MinChunkSize:          s.Network.MinChunkSize,
		WorkerBufferSize:      s.Network.WorkerBufferSize,
		MaxTaskRetries:        s.Performance.MaxTaskRetries,
		SlowWorkerThreshold:   s.Performance.SlowWorkerThreshold,
		SlowWorkerGracePeriod: s.Performance.SlowWorkerGracePeriod,
		StallTimeout:          s.Performance.StallTimeout,
		SpeedEmaAlpha:         s.Performance.SpeedEmaAlpha,
		ConnectTimeout:        s.Network.ConnectTimeout,
		ResponseTimeout:       s.Network.ResponseTimeout,
		RequestTimeout:        s.Network.RequestTimeout,
		MaxRedirects:          s.Network.MaxRedirects,
	}
}
