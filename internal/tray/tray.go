package tray

import (
	"fmt"
	_ "image/png"
	"os"
	"runtime"

	"github.com/getlantern/systray"
	"github.com/go-downloader/internal/app"
	"github.com/go-downloader/internal/assets"
	"github.com/go-downloader/internal/utils"
)

var webURL string

// Run starts the system tray loop
func Run(url string) {
	utils.Debug("tray.Run() started with URL: %s", url)
	webURL = url
	utils.Debug("Calling systray.Run()")
	systray.Run(onReady, onExit)
	utils.Debug("systray.Run() returned")
}

func onReady() {
	utils.Debug("tray.onReady() invoked")
	defer func() {
		if r := recover(); r != nil {
			utils.Debug("Panic in tray onReady: %v", r)
		}
	}()

	// 尝试加载托盘图标
	utils.Debug("Loading tray icon")
	iconData := assets.LogoData
	if runtime.GOOS == "windows" {
		utils.Debug("Converting icon to ICO for Windows")
		if converted, err := utils.ConvertToICO(assets.LogoData); err == nil {
			utils.Debug("Icon conversion successful")
			iconData = converted
		} else {
			utils.Debug("Failed to convert icon to ICO: %v", err)
		}
	}
	utils.Debug("Setting tray icon")
	systray.SetIcon(iconData)

	systray.SetTooltip("Downloader")

	utils.Debug("Adding tray menu items")
	mShow := systray.AddMenuItem("显示主界面", "打开 Downloader Web 界面")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "关闭 Downloader")

	go func() {
		utils.Debug("Starting tray menu event loop")
		for {
			select {
			case <-mShow.ClickedCh:
				utils.Debug("Tray menu: 'Show' clicked")
				openWeb()
			case <-mQuit.ClickedCh:
				utils.Debug("Tray menu: 'Quit' clicked")
				systray.Quit()
			}
		}
	}()
	utils.Debug("tray.onReady() completed")
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
