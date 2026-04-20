// Package config 配置管理
// 使用 Viper 库管理 YAML 配置文件
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go-downloader/internal/utils"
)

var (
	globalConfig *Config
	configOnce   sync.Once
	configMu     sync.RWMutex
)

// Config 所有用户可配置的应用设置
type Config struct {
	General     GeneralConfig     `mapstructure:"general"`
	Network     NetworkConfig     `mapstructure:"network"`
	Performance PerformanceConfig `mapstructure:"performance"`
	Categories  CategoryConfig    `mapstructure:"categories"`
}

// GeneralConfig 应用行为设置
type GeneralConfig struct {
	DefaultDownloadDir           string `mapstructure:"default_download_dir"`
	WarnOnDuplicate              bool   `mapstructure:"warn_on_duplicate"`
	DownloadCompleteNotification bool   `mapstructure:"download_complete_notification"`
	AllowRemoteOpenActions       bool   `mapstructure:"allow_remote_open_actions"`
	ExtensionPrompt              bool   `mapstructure:"extension_prompt"`
	AutoResume                   bool   `mapstructure:"auto_resume"`
	AutoRun                      bool   `mapstructure:"auto_run"`
	SkipUpdateCheck              bool   `mapstructure:"skip_update_check"`
	ClipboardMonitor             bool   `mapstructure:"clipboard_monitor"`
	Theme                        int    `mapstructure:"theme"`
	LogRetentionCount            int    `mapstructure:"log_retention_count"`
	Port                         int    `mapstructure:"port"`
	OpenOnStart                  bool   `mapstructure:"open_on_start"`
	LocalOnly                    bool   `mapstructure:"local_only"`
}

// NetworkConfig 网络连接参数
type NetworkConfig struct {
	MaxConnectionsPerHost  int           `mapstructure:"max_connections_per_host"`
	MaxConcurrentDownloads int           `mapstructure:"max_concurrent_downloads"`
	UserAgent              string        `mapstructure:"user_agent"`
	ProxyURL               string        `mapstructure:"proxy_url"`
	SequentialDownload     bool          `mapstructure:"sequential_download"`
	MinChunkSize           int64         `mapstructure:"min_chunk_size"`
	WorkerBufferSize       int           `mapstructure:"worker_buffer_size"`
	ConnectTimeout         time.Duration `mapstructure:"connect_timeout"`
	ResponseTimeout        time.Duration `mapstructure:"response_timeout"`
	RequestTimeout         time.Duration `mapstructure:"request_timeout"`
	MaxRedirects           int           `mapstructure:"max_redirects"`
}

// PerformanceConfig 性能调优参数
type PerformanceConfig struct {
	MaxTaskRetries        int           `mapstructure:"max_task_retries"`
	SlowWorkerThreshold   float64       `mapstructure:"slow_worker_threshold"`
	SlowWorkerGracePeriod time.Duration `mapstructure:"slow_worker_grace_period"`
	StallTimeout          time.Duration `mapstructure:"stall_timeout"`
	SpeedEmaAlpha         float64       `mapstructure:"speed_ema_alpha"`
}

// CategoryConfig 文件分类设置
type CategoryConfig struct {
	CategoryEnabled bool       `mapstructure:"category_enabled"`
	Categories      []Category `mapstructure:"categories"`
}



// GetConfigPath 返回配置文件路径
func GetConfigPath() string {
	configDir := GetConfigDir()
	return filepath.Join(configDir, "config.yaml")
}

// GetConfigDir 返回配置目录
func GetConfigDir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "downloader")
	}

	// Windows: %APPDATA%\downloader
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "downloader")
	}

	// Linux/macOS: ~/.config/downloader
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".config", "downloader")
}

// InitConfig 初始化配置系统
func InitConfig() error {
	var initErr error
	configOnce.Do(func() {
		initErr = loadConfig()
	})
	return initErr
}

// loadConfig 加载配置文件
func loadConfig() error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// 设置默认值
	setDefaults(v)

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，创建默认配置
			utils.Debug("配置文件不存在，创建默认配置: %s", configPath)
			if err := v.SafeWriteConfig(); err != nil {
				utils.Debug("警告：无法写入默认配置: %v", err)
			}
		} else {
			utils.Debug("警告：读取配置文件失败: %v", err)
		}
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	configMu.Lock()
	globalConfig = &cfg
	configMu.Unlock()

	utils.Debug("配置加载完成: %s", configPath)
	return nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// General 默认值
	v.SetDefault("general.default_download_dir", getDefaultDownloadDir())
	v.SetDefault("general.warn_on_duplicate", true)
	v.SetDefault("general.download_complete_notification", true)
	v.SetDefault("general.allow_remote_open_actions", false)
	v.SetDefault("general.extension_prompt", false)
	v.SetDefault("general.auto_resume", false)
	v.SetDefault("general.auto_run", false)
	v.SetDefault("general.skip_update_check", false)
	v.SetDefault("general.clipboard_monitor", true)
	v.SetDefault("general.theme", 0)
	v.SetDefault("general.log_retention_count", 5)
	v.SetDefault("general.port", 8080)
	v.SetDefault("general.open_on_start", true)
	v.SetDefault("general.local_only", true)

	// Network 默认值
	v.SetDefault("network.max_connections_per_host", 32)
	v.SetDefault("network.max_concurrent_downloads", 3)
	v.SetDefault("network.user_agent", "")
	v.SetDefault("network.proxy_url", "")
	v.SetDefault("network.sequential_download", false)
	v.SetDefault("network.min_chunk_size", 2097152)       // 2 MB
	v.SetDefault("network.worker_buffer_size", 524288)    // 512 KB
	v.SetDefault("network.connect_timeout", "30s")
	v.SetDefault("network.response_timeout", "60s")
	v.SetDefault("network.request_timeout", "5m")
	v.SetDefault("network.max_redirects", 10)

	// Performance 默认值
	v.SetDefault("performance.max_task_retries", 3)
	v.SetDefault("performance.slow_worker_threshold", 0.3)
	v.SetDefault("performance.slow_worker_grace_period", "5s")
	v.SetDefault("performance.stall_timeout", "3s")
	v.SetDefault("performance.speed_ema_alpha", 0.3)

	// Categories 默认值
	v.SetDefault("categories.category_enabled", false)
	v.SetDefault("categories.categories", getDefaultCategories())
}

// getDefaultDownloadDir 获取默认下载目录
func getDefaultDownloadDir() string {
	// Windows: 检查 D:\system\downloads 或使用用户下载目录
	customPath := filepath.Join("D:", "system", "downloads")
	if _, err := os.Stat(customPath); err == nil {
		return customPath
	}

	if dlDir := os.Getenv("USERPROFILE"); dlDir != "" {
		return filepath.Join(dlDir, "Downloads")
	}

	// Linux/macOS: ~/Downloads
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Downloads")
}

// getDefaultCategories 获取默认分类列表
func getDefaultCategories() []map[string]interface{} {
	homeDir, _ := os.UserHomeDir()
	docDir := filepath.Join("D:", "system", "documents")
	
	// 检查文档目录是否存在，不存在则使用 Downloads
	if _, err := os.Stat(docDir); os.IsNotExist(err) {
		docDir = getDefaultDownloadDir()
	}

	return []map[string]interface{}{
		{
			"name":        "视频",
			"description": "MP4、MKV、AVI 等视频文件",
			"pattern":     "(?i)\\.(mp4|mkv|avi|mov|wmv|flv|webm|m4v|mpg|mpeg|3gp)$",
			"path":        filepath.Join(homeDir, "Videos"),
		},
		{
			"name":        "音乐",
			"description": "MP3、FLAC 等音频文件",
			"pattern":     "(?i)\\.(mp3|flac|wav|aac|ogg|wma|m4a|opus)$",
			"path":        filepath.Join(homeDir, "Music"),
		},
		{
			"name":        "压缩包",
			"description": "ZIP、RAR 等压缩文件",
			"pattern":     "(?i)\\.(zip|rar|7z|tar|gz|bz2|xz|zst|tgz)$",
			"path":        getDefaultDownloadDir(),
		},
		{
			"name":        "文档",
			"description": "PDF、Word、Excel 等文档文件",
			"pattern":     "(?i)\\.(pdf|doc|docx|xls|xlsx|ppt|pptx|odt|ods|txt|rtf|csv|epub)$",
			"path":        docDir,
		},
		{
			"name":        "程序",
			"description": "可执行文件和安装包",
			"pattern":     "(?i)\\.(exe|msi|deb|rpm|appimage|dmg|pkg|flatpak|snap|sh|run|bin)$",
			"path":        getDefaultDownloadDir(),
		},
		{
			"name":        "图片",
			"description": "JPEG、PNG 等图片文件",
			"pattern":     "(?i)\\.(jpg|jpeg|png|gif|bmp|svg|webp|ico|tiff|psd)$",
			"path":        filepath.Join(homeDir, "Pictures"),
		},
	}
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		// 如果未初始化，返回默认配置
		return defaultConfig()
	}
	return globalConfig
}

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		General: GeneralConfig{
			DefaultDownloadDir:           getDefaultDownloadDir(),
			WarnOnDuplicate:              true,
			DownloadCompleteNotification: true,
			AllowRemoteOpenActions:       false,
			ExtensionPrompt:              false,
			AutoResume:                   false,
			AutoRun:                      false,
			SkipUpdateCheck:              false,
			ClipboardMonitor:             true,
			Theme:                        0,
			LogRetentionCount:            5,
			Port:                         8080,
			OpenOnStart:                  true,
			LocalOnly:                    true,
		},
		Network: NetworkConfig{
			MaxConnectionsPerHost:  32,
			MaxConcurrentDownloads: 3,
			UserAgent:              "",
			ProxyURL:               "",
			SequentialDownload:     false,
			MinChunkSize:           2 * 1024 * 1024, // 2 MB
			WorkerBufferSize:       512 * 1024,      // 512 KB
			ConnectTimeout:         30 * time.Second,
			ResponseTimeout:        60 * time.Second,
			RequestTimeout:         5 * time.Minute,
			MaxRedirects:           10,
		},
		Performance: PerformanceConfig{
			MaxTaskRetries:        3,
			SlowWorkerThreshold:   0.3,
			SlowWorkerGracePeriod: 5 * time.Second,
			StallTimeout:          3 * time.Second,
			SpeedEmaAlpha:         0.3,
		},
		Categories: CategoryConfig{
			CategoryEnabled: false,
			Categories: []Category{
				{Name: "视频", Description: "MP4、MKV、AVI 等视频文件", Pattern: "(?i)\\.(mp4|mkv|avi|mov|wmv|flv|webm|m4v|mpg|mpeg|3gp)$", Path: filepath.Join(homeDir, "Videos")},
				{Name: "音乐", Description: "MP3、FLAC 等音频文件", Pattern: "(?i)\\.(mp3|flac|wav|aac|ogg|wma|m4a|opus)$", Path: filepath.Join(homeDir, "Music")},
				{Name: "压缩包", Description: "ZIP、RAR 等压缩文件", Pattern: "(?i)\\.(zip|rar|7z|tar|gz|bz2|xz|zst|tgz)$", Path: getDefaultDownloadDir()},
				{Name: "文档", Description: "PDF、Word、Excel 等文档文件", Pattern: "(?i)\\.(pdf|doc|docx|xls|xlsx|ppt|pptx|odt|ods|txt|rtf|csv|epub)$", Path: filepath.Join("D:", "system", "documents")},
				{Name: "程序", Description: "可执行文件和安装包", Pattern: "(?i)\\.(exe|msi|deb|rpm|appimage|dmg|pkg|flatpak|snap|sh|run|bin)$", Path: getDefaultDownloadDir()},
				{Name: "图片", Description: "JPEG、PNG 等图片文件", Pattern: "(?i)\\.(jpg|jpeg|png|gif|bmp|svg|webp|ico|tiff|psd)$", Path: filepath.Join(homeDir, "Pictures")},
			},
		},
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(cfg *Config) error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// 设置配置值
	v.Set("general", cfg.General)
	v.Set("network", cfg.Network)
	v.Set("performance", cfg.Performance)
	v.Set("categories", cfg.Categories)

	// 写入配置文件
	if err := v.WriteConfigAs(configPath); err != nil {
		return err
	}

	// 更新全局配置
	configMu.Lock()
	globalConfig = cfg
	configMu.Unlock()

	utils.Debug("配置已保存: %s", configPath)
	return nil
}

// ReloadConfig 重新加载配置文件
func ReloadConfig() error {
	configMu.Lock()
	defer configMu.Unlock()

	// 重置 once 以允许重新加载
	configOnce = sync.Once{}
	return loadConfig()
}
