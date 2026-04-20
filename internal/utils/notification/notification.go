package notification

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/go-toast/toast"

	"go-downloader/internal/config"
	"go-downloader/internal/utils/icon"
	"go-downloader/internal/utils/logger"
)

var (
	iconPath string
	iconOnce sync.Once
)

// ensureIcon 将图标写入用户缓存目录，首次调用时执行
func ensureIcon() string {
	iconOnce.Do(func() {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			cacheDir = os.TempDir()
		}
		appCache := filepath.Join(cacheDir, "go-downloader")
		if err := os.MkdirAll(appCache, 0755); err != nil {
			return
		}

		ext := ".png"
		data := icon.LoadLogoPNG(64)
		if runtime.GOOS == "windows" {
			ext = ".ico"
			data = icon.LoadLogoICO()
		}

		path := filepath.Join(appCache, "logo"+ext)
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			if err := os.WriteFile(path, data, 0644); err != nil {
				return
			}
		}
		iconPath = path
	})
	return iconPath
}

// Send 发送系统通知
func Send(title, message string) {
	notification := toast.Notification{
		AppID:  "go-downloader",
		Title:  title,
		Message: message,
		Icon:   ensureIcon(),
	}

	if err := notification.Push(); err != nil {
		logger.Error("通知发送失败: %v", err)
	}
}

// SendWithURL 发送系统通知，点击通知可打开指定 URL
func SendWithURL(title, message, url string) {
	notification := toast.Notification{
		AppID:               "go-downloader",
		Title:               title,
		Message:             message,
		Icon:                ensureIcon(),
		ActivationType:      "protocol",
		ActivationArguments: url,
	}

	if err := notification.Push(); err != nil {
		logger.Error("通知发送失败: %v", err)
	}
}

// StartURL 返回服务启动后的访问地址
func StartURL() string {
	return fmt.Sprintf("http://localhost:%d", config.GetPort())
}
