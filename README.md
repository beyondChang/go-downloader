# go-downloader

## 运行

```bash
go run .
```

## 打包

双击 `build/build.bat` 即可一键打包，流程如下：

1. 自动检查安装 [go-winres](https://github.com/tc-hib/go-winres) 工具
2. 根据 `build/winres/winres.json` 生成资源文件（图标+版本信息+manifest）
3. 编译 Windows GUI 程序
4. 检测 UPX，存在则自动压缩

### 手动打包

```bash
# 安装 go-winres（首次）
go install github.com/tc-hib/go-winres@latest

# 生成资源文件
cd build && go-winres make && copy rsrc.syso ..\ && cd ..

# 编译
go build -ldflags="-s -w -H=windowsgui" -o go-downloader.exe .
```

### 自定义版本信息

编辑 `build/winres/winres.json` 中的 `RT_VERSION` 字段，可修改：
- FileDescription — 程序描述
- CompanyName — 公司/作者名
- LegalCopyright — 版权信息
- FileVersion / ProductVersion — 版本号
