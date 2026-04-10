# Project go-downloder

### 构建指南
在终端执行以下命令进行构建：
```shell
go build -ldflags "-H=windowsgui -s -w" -o go-downloader.exe
```

**在 Trae/VS Code 中直接执行：**
1. 按下 `Ctrl+Shift+B` 即可直接触发默认构建任务。
2. 或者点击菜单栏 `终端` -> `运行任务...` -> `Build Go Downloader`。

**使用说明：**
- **启动**：双击 `go-downloader.exe`。程序会静默启动并最小化到系统托盘。
- **管理**：在托盘图标上右键点击，选择“显示主界面”即可在浏览器中打开管理后台。