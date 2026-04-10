package main

import (
	"fmt"
	"os"

	"github.com/go-downloader/cmd"
	"github.com/go-downloader/internal/utils"
)

func main() {
	utils.Debug("Application starting...")

	// 无论通过什么方式启动，优先尝试直接启动 GUI 流程以提高双击响应成功率
	// 如果带有命令行参数（且非空），则按正常 Cobra 流程走，否则直接进 GUI
	if len(os.Args) <= 1 {
		utils.Debug("No arguments provided, starting GUI directly")
		if err := cmd.RunGUI(); err != nil {
			utils.Debug("Fatal error in GUI mode: %v", err)
			utils.Notify("程序启动失败", fmt.Sprintf("错误原因: %v", err))
		}
		return
	}

	if err := cmd.Execute(); err != nil {
		utils.Debug("Fatal error during execution: %v", err)
		utils.Notify("程序启动失败", fmt.Sprintf("错误原因: %v", err))
	}
}
