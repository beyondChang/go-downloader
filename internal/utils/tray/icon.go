package tray

import (
	"go-downloader/internal/utils/icon"
)

// IconData 读取托盘图标，从嵌入资源加载并转换为 ICO 格式
func IconData() []byte {
	return icon.LoadLogoICO()
}
