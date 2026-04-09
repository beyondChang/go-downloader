# 设置与配置

本文档涵盖了 Downloader 中可用的所有配置选项。有关 CLI 命令和标志，请参阅 [USAGE.md](USAGE.md)。

## 配置文件

您可以从位于应用程序数据目录中的 `settings.json` 文件访问这些设置：

- **Windows:** `%APPDATA%\downloader\settings.json`
- **macOS:** `~/Library/Application Support/downloader/settings.json`
- **Linux:** `~/.config/downloader/settings.json`

`settings.json` 文件需要一个划分为 `general`、`network`、`performance` 和 `categories` 部分的嵌套结构。例如：

```json
{
  "general": {
    "default_download_dir": "/path/to/downloads",
    "theme": 2
  },
  "network": {
    "max_connections_per_host": 16
  },
  "performance": {
    "max_task_retries": 5
  },
  "categories": {
    "category_enabled": true
  }
}
```

*注意：您不需要指定所有键。Downloader 会自动推断缺少的键并使用其内部默认值。*

## 目录结构

Downloader 遵循操作系统约定来存储其文件。下面是它使用的所有目录以及在每个平台上可以找到它们的位置的明细。

| 目录   | 用途                           | Linux                        | macOS                                       | Windows                 |
| :---------- | :-------------------------------- | :--------------------------- | :------------------------------------------ | :---------------------- |
| **Config (配置)**  | `settings.json`                   | `~/.config/downloader/`           | `~/Library/Application Support/downloader/`      | `%APPDATA%\downloader\`      |
| **State (状态)**   | 数据库 (`downloader.db`)、auth token | `~/.local/state/downloader/`      | `~/Library/Application Support/downloader/`      | `%APPDATA%\downloader\`      |
| **Logs (日志)**    | 带时间戳的 `.log` 文件          | `~/.local/state/downloader/logs/` | `~/Library/Application Support/downloader/logs/` | `%APPDATA%\downloader\logs\` |
| **Runtime (运行时)** | PID 文件、端口文件、锁文件         | `$XDG_RUNTIME_DIR/downloader/`¹   | `$TMPDIR/downloader-runtime/`                    | `%TEMP%\downloader\`         |

> ¹ 当 `$XDG_RUNTIME_DIR` 未设置（例如 Docker / 无头模式）时，回退到 `~/.local/state/downloader/`。

> **注意：** 在 Linux 上，如果设置了 `$XDG_CONFIG_HOME` / `$XDG_STATE_HOME`，则会遵守它们；上面显示的路径为默认值。

---

### 常规设置 (General Settings)

| 键                    | 类型   | 说明                                                                                        | 默认值 |
| :--------------------- | :----- | :------------------------------------------------------------------------------------------------- | :------ |
| `default_download_dir` | string | 保存新下载的目录。如果为空，则默认为 `~/Downloads` 或当前目录。 | `""`    |
| `allow_remote_open_actions` | bool | 允许来自远程客户端的 `/open-file` 和 `/open-folder` API 请求。除非您信任您的网络和身份验证设置，否则请保持禁用。 | `false` |
| `warn_on_duplicate`    | bool   | 添加列表中已存在的下载时显示警告。                             | `true`  |
| `extension_prompt`     | bool   | 通过浏览器扩展程序添加下载时提示确认。                | `false` |
| `auto_resume`          | bool   | Downloader 启动时自动恢复暂停的下载。                                           | `false` |
| `skip_update_check`    | bool   | 禁用启动时自动检查新版本。                                               | `false` |
| `clipboard_monitor`    | bool   | 监视系统剪贴板中的 URL 并提示下载它们。                                   | `true`  |
| `theme`                | int    | UI 主题 (0=自适应, 1=亮色, 2=暗色)。                                                            | `0`     |
| `log_retention_count`  | int    | 保留最近日志文件的数量。                                                                | `5`     |

### 网络设置 (Connection Settings)

| 键                        | 类型   | 说明                                                                                           | 默认值 |
| :------------------------- | :----- | :---------------------------------------------------------------------------------------------------- | :------ |
| `max_connections_per_host` | int    | 允许连接到单个主机的最大并发连接数 (1-64)。                                       | `32`    |
| `max_concurrent_downloads` | int    | 同时运行的最大下载数 (需要重启生效)。                                | `3`     |
| `user_agent`               | string | HTTP 请求的自定义 User-Agent 字符串。留空以使用默认值。                                  | `""`    |
| `proxy_url`                | string | HTTP/HTTPS 代理 URL (例如 `http://127.0.0.1:8080`)。留空以使用系统设置。             | `""`    |
| `sequential_download`      | bool   | 严格按顺序下载文件片段 (流媒体模式)。可用于预览媒体，但可能较慢。 | `false` |
| `min_chunk_size`           | int64  | 下载分块的最小大小 (字节) (例如，`2097152` 表示 2MB)。                                  | `2MB`   |
| `worker_buffer_size`       | int    | 每个工作线程的 I/O 缓冲区大小 (字节) (例如，`524288` 表示 512KB)。                                       | `512KB` |

### 性能设置 (Performance Settings)

| 键                        | 类型     | 说明                                                                  | 默认值 |
| :------------------------- | :------- | :--------------------------------------------------------------------------- | :------ |
| `max_task_retries`         | int      | 在放弃之前重试失败分块的次数。                    | `3`     |
| `slow_worker_threshold`    | float    | 重启慢于平均速度此比例的工作线程 (0.0-1.0)。       | `0.3`   |
| `slow_worker_grace_period` | duration | 在检查工作线程速度之前等待的时间 (例如 `5s`)。                  | `5s`    |
| `stall_timeout`            | duration | 重启在这段时间内未接收到数据的工作线程 (例如 `3s`)。   | `3s`    |
| `speed_ema_alpha`          | float    | 用于计算速度的指数移动平均平滑因子 (0.0-1.0)。 | `0.3`   |

### 类别设置 (Category Settings)

| 键                    | 类型   | 说明                                                                                              | 默认值 |
| :--------------------- | :----- | :------------------------------------------------------------------------------------------------------- | :------ |
| `category_enabled`     | bool   | 启用根据文件类型类别将下载自动排序到子文件夹中。                     | `false` |
