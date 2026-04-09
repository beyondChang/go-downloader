# CLI 使用指南

Downloader 提供了一个强大的命令行界面（CLI），用于自动化和脚本编写。有关配置选项，请参阅 [SETTINGS.md](SETTINGS.md)。

## 命令表

| 命令 | 功能说明 | 关键参数 (Flags) | 备注 |
| :--- | :--- | :--- | :--- |
| `downloader [url]...` | 启动带 Web 界面的 Downloader。可排队可选的 URL。 | `--batch, -b`<br>`--port, -p`<br>`--output, -o`<br>`--no-resume`<br>`--exit-when-done` | `-o` 默认为当前工作目录 (CWD)。如果设置了 `--host`，则进入远程模式。 |
| `downloader server [url]...` | 启动无界面服务器（Headless）。可排队可选的 URL。 | `--batch, -b`<br>`--port, -p`<br>`--output, -o`<br>`--exit-when-done`<br>`--no-resume`<br>`--token` | `-o` 默认为当前工作目录。主要的无界面模式命令。 |
| `downloader connect [host:port]` | 将浏览器连接到正在运行的服务器。未指定目标时自动检测本地服务器。 | `--insecure-http` | 远程使用的便捷别名。 |
| `downloader add <url>...` | 通过 CLI/API 添加下载排队。 | `--batch, -b`<br>`--output, -o` | `-o` 默认为当前工作目录。别名：`get`。 |
| `downloader ls [id]` | 列出下载任务，或显示单个下载详情。 | `--json`<br>`--watch` | 别名：`l`。 |
| `downloader pause <id>` | 通过 ID 或前缀暂停下载。 | `--all` | |
| `downloader resume <id>` | 通过 ID 或前缀恢复已暂停的下载。 | `--all` | |
| `downloader refresh <id> <url>` | 更新已暂停或出错下载的源 URL。 | 无 | 使用新链接重新连接。 |
| `downloader rm <id>` | 通过 ID 或前缀移除下载任务。 | `--clean` | 别名：`kill`。 |
| `downloader token` | 打印当前的 API 认证 Token。 | 无 | 适用于远程客户端。 |

## 服务器子命令（兼容性）

| 命令 | 功能说明 |
| :--- | :--- |
| `downloader server start [url]...` | `downloader server [url]...` 的旧版等效命令。 |
| `downloader server stop` | 通过 PID 文件停止正在运行的服务器进程。 |
| `downloader server status` | 从 PID/端口状态打印运行/未运行状态。 |

## 全局参数 (Global Flags)

这些是持久化参数，可用于所有命令。

| 参数 | 说明 |
| :--- | :--- |
| `--host <host:port>` | CLI 操作的目标服务器。 |
| `--token <token>` | 用于 API 请求的 Bearer Token。 |

## 环境变量

| 变量名 | 说明 |
| :--- | :--- |
| `SURGE_HOST` | 未提供 `--host` 时的默认主机。 |
| `SURGE_TOKEN` | 未提供 `--token` 时的默认 Token。 |
