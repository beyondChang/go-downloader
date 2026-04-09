package tray

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"runtime"

	"github.com/getlantern/systray"
	"github.com/go-downloader/internal/app"
	"github.com/go-downloader/internal/assets"
	"github.com/go-downloader/internal/utils"
	"github.com/nfnt/resize"
	"github.com/sergeymakinen/go-ico"
)

var webURL string

// Run starts the system tray loop
func Run(url string) {
	webURL = url
	systray.Run(onReady, onExit)
}

func onReady() {
	// 尝试加载托盘图标
	iconData := assets.LogoData
	if runtime.GOOS == "windows" {
		if converted, err := convertToICO(assets.LogoData); err == nil {
			iconData = converted
		} else {
			utils.Debug("%s", fmt.Sprintf("Failed to convert icon to ICO: %v", err))
		}
	}
	systray.SetIcon(iconData)

	systray.SetTooltip("Downloader")

	mShow := systray.AddMenuItem("显示主界面", "打开 Downloader Web 界面")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "关闭 Downloader")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				openWeb()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	// Handle left-click on tray icon (Windows only currently supports this via systray package well)
	// Some systray forks support mLeftClick, but getlantern/systray uses mShow pattern for cross-platform.
	// For now, mShow serves as the explicit way to open the web.
}

func convertToICO(pngData []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}

	// Resize to standard tray icon size for better compatibility
	img = resize.Resize(32, 32, img, resize.Lanczos3)

	var buf bytes.Buffer
	if err := ico.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func onExit() {
	// Clean up and shutdown the application
	app.ExecuteGlobalShutdown("tray exit")
	os.Exit(0)
}

func openWeb() {
	if webURL != "" {
		if err := utils.OpenURL(webURL); err != nil {
			utils.Debug("%s", fmt.Sprintf("Failed to open web URL: %v", err))
		}
	}
}
