// Package config 文件分类配置
// 支持按文件类型自动分类下载文件到不同目录
package config

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"

	"go-downloader/internal/utils"
)

// Category 下载分类定义
type Category struct {
	// 分类名称
	Name string `json:"name" mapstructure:"name"`
	// 描述
	Description string `json:"description,omitempty" mapstructure:"description"`
	// 文件名匹配正则表达式
	Pattern string `json:"pattern" mapstructure:"pattern"`
	// 目标路径
	Path string `json:"path" mapstructure:"path"`
}

// Validate 验证分类配置
func (c *Category) Validate() error {
	if c == nil {
		return errors.New("分类不能为空")
	}
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("分类名称不能为空")
	}
	if strings.TrimSpace(c.Pattern) == "" {
		return errors.New("分类模式不能为空")
	}
	if _, err := regexp.Compile(strings.TrimSpace(c.Pattern)); err != nil {
		return err
	}
	if strings.TrimSpace(c.Path) == "" {
		return errors.New("分类路径不能为空")
	}
	return nil
}

// existingDirOrFallback 如果目录存在则返回该目录，否则返回回退目录
func existingDirOrFallback(dir, fallback string) string {
	trimmed := strings.TrimSpace(dir)
	if trimmed != "" {
		if info, err := os.Stat(trimmed); err == nil && info.IsDir() {
			return trimmed
		}
	}
	return fallback
}

// DefaultCategories 返回默认的下载分类集合
func DefaultCategories() []Category {
	downloadsDir := strings.TrimSpace(GetDownloadsDir())
	if downloadsDir == "" {
		downloadsDir = "."
	}

	videosDir := existingDirOrFallback(GetVideosDir(), downloadsDir)
	musicDir := existingDirOrFallback(GetMusicDir(), downloadsDir)
	documentsDir := existingDirOrFallback(GetDocumentsDir(), downloadsDir)
	picturesDir := existingDirOrFallback(GetPicturesDir(), downloadsDir)

	return []Category{
		{
			Name:        "视频",
			Description: "MP4、MKV、AVI 等视频文件",
			Pattern:     `(?i)\.(mp4|mkv|avi|mov|wmv|flv|webm|m4v|mpg|mpeg|3gp)$`,
			Path:        videosDir,
		},
		{
			Name:        "音乐",
			Description: "MP3、FLAC 等音频文件",
			Pattern:     `(?i)\.(mp3|flac|wav|aac|ogg|wma|m4a|opus)$`,
			Path:        musicDir,
		},
		{
			Name:        "压缩包",
			Description: "ZIP、RAR 等压缩文件",
			Pattern:     `(?i)\.(zip|rar|7z|tar|gz|bz2|xz|zst|tgz)$`,
			Path:        downloadsDir,
		},
		{
			Name:        "文档",
			Description: "PDF、Word、Excel 等文档文件",
			Pattern:     `(?i)\.(pdf|doc|docx|xls|xlsx|ppt|pptx|odt|ods|txt|rtf|csv|epub)$`,
			Path:        documentsDir,
		},
		{
			Name:        "程序",
			Description: "可执行文件和安装包",
			Pattern:     `(?i)\.(exe|msi|deb|rpm|appimage|dmg|pkg|flatpak|snap|sh|run|bin)$`,
			Path:        downloadsDir,
		},
		{
			Name:        "图片",
			Description: "JPEG、PNG 等图片文件",
			Pattern:     `(?i)\.(jpg|jpeg|png|gif|bmp|svg|webp|ico|tiff|psd)$`,
			Path:        picturesDir,
		},
	}
}

// 正则表达式缓存
var (
	patternCache = make(map[string]*regexp.Regexp)
	patternMu    sync.RWMutex
)

// getCompiledPattern 返回编译后的正则表达式，使用缓存避免重复编译
func getCompiledPattern(pattern string) *regexp.Regexp {
	patternMu.RLock()
	re, ok := patternCache[pattern]
	patternMu.RUnlock()
	if ok {
		return re
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		patternMu.Lock()
		patternCache[pattern] = nil
		patternMu.Unlock()
		return nil
	}

	patternMu.Lock()
	patternCache[pattern] = re
	patternMu.Unlock()
	return re
}

// GetCategoryForFile 返回文件匹配的分类
// 返回最后一个匹配的分类，以便用户添加的规则可以覆盖前面的默认规则
func GetCategoryForFile(filename string, categories []Category) (*Category, error) {
	if filename == "" || len(categories) == 0 {
		return nil, nil
	}

	var matched *Category

	for i := range categories {
		cat := &categories[i]
		if cat.Pattern == "" {
			continue
		}

		re := getCompiledPattern(cat.Pattern)
		if re != nil && re.MatchString(filename) {
			if matched != nil {
				utils.Debug("Config: 分类模式 %q 匹配 %q，覆盖之前的匹配 %q", cat.Pattern, filename, matched.Pattern)
			}
			matched = cat
		}
	}

	return matched, nil
}

// ResolveCategoryPath 返回分类的路径
func ResolveCategoryPath(cat *Category, defaultDownloadDir string) string {
	defaultPath := strings.TrimSpace(defaultDownloadDir)
	if cat == nil {
		return defaultPath
	}
	trimmed := strings.TrimSpace(cat.Path)
	if trimmed == "" {
		return defaultPath
	}
	return trimmed
}

// CategoryNames 返回分类名称列表
func CategoryNames(categories []Category) []string {
	names := make([]string, len(categories))
	for i, cat := range categories {
		names[i] = cat.Name
	}
	return names
}
