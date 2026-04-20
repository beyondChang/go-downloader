// Package web 前端静态资源
// 使用 go:embed 将前端文件嵌入到二进制文件中
package web

import "embed"

// Assets 嵌入的前端静态资源
// 包含 css、js、assets 目录和 index.html 主页面
//
//go:embed all:css all:js all:assets *.html
var Assets embed.FS

// LogoData 应用程序图标数据，用于系统托盘
//
//go:embed assets/logo.png
var LogoData []byte
