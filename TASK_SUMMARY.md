# 任务总结: 配置系统迁移到 YAML

## 问题
用户要求所有配置应该在 `config.yaml` 而不是 `settings.json`。

## 解决方案

### 1. 添加 Viper 依赖
```bash
go get github.com/spf13/viper
```

### 2. 创建新的配置系统 (`internal/config/config.go`)
- 使用 Viper 管理 YAML 配置
- 配置文件路径：`%APPDATA%\downloader\config.yaml` (Windows)
- 自动创建默认配置文件

### 3. 兼容层 (`internal/config/settings.go`)
- `Settings` = `Config` (类型别名)
- `LoadSettings()` → 调用 `InitConfig()` + `GetConfig()`
- `SaveSettings()` → 调用 `SaveConfig()`
- 保持旧代码无需修改

### 4. 更新 Category 结构体
添加 `mapstructure` 标签支持 YAML 解析：
```go
type Category struct {
    Name        string `json:"name" mapstructure:"name"`
    Description string `json:"description,omitempty" mapstructure:"description"`
    Pattern     string `json:"pattern" mapstructure:"pattern"`
    Path        string `json:"path" mapstructure:"path"`
}
```

## 配置文件示例

```yaml
general:
    default_download_dir: C:\Users\Admin\Downloads
    port: 8080
    local_only: true
    open_on_start: true
    auto_run: false
    warn_on_duplicate: true
    download_complete_notification: true
    clipboard_monitor: true
    theme: 0
    log_retention_count: 5

network:
    max_connections_per_host: 32
    max_concurrent_downloads: 3
    min_chunk_size: 2097152        # 2 MB
    worker_buffer_size: 524288     # 512 KB
    connect_timeout: 30s
    response_timeout: 60s
    request_timeout: 5m
    max_redirects: 10
    user_agent: ""
    proxy_url: ""
    sequential_download: false

performance:
    max_task_retries: 3
    slow_worker_threshold: 0.3
    slow_worker_grace_period: 5s
    stall_timeout: 3s
    speed_ema_alpha: 0.3

categories:
    category_enabled: false
    categories:
        - name: 视频
          description: MP4、MKV、AVI 等视频文件
          pattern: (?i)\.(mp4|mkv|avi|mov|wmv|flv|webm|m4v|mpg|mpeg|3gp)$
          path: C:\Users\Admin\Videos
        - name: 音乐
          description: MP3、FLAC 等音频文件
          pattern: (?i)\.(mp3|flac|wav|aac|ogg|wma|m4a|opus)$
          path: C:\Users\Admin\Music
        # ... 其他分类
```

## 关键改进

1. **友好的时间格式**: `30s`, `60s`, `5m` 而不是纳秒整数
2. **自动创建**: 首次运行自动生成默认配置
3. **兼容性**: 旧代码无需修改，通过类型别名平滑过渡
4. **跨平台**: 
   - Windows: `%APPDATA%\downloader\config.yaml`
   - Linux/macOS: `~/.config/downloader/config.yaml`

## 构建结果
- ✅ 编译成功
- ✅ 配置文件自动生成
- ✅ 时间格式友好
- ✅ 旧代码兼容
