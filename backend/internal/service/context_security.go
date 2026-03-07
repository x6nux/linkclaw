package service

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SecurityConfig 上下文扫描安全配置
type SecurityConfig struct {
	// 最大目录深度（默认 20）
	MaxDepth int `json:"max_depth"`
	// 最大文件数（默认 10000）
	MaxFiles int `json:"max_files"`
	// 最大总大小（默认 500MB）
	MaxTotalSize int64 `json:"max_total_size"`
	// 允许的路径前缀白名单
	AllowedPrefixes []string `json:"allowed_prefixes"`
	// 禁止的路径前缀黑名单
	ForbiddenPrefixes []string `json:"forbidden_prefixes"`
	// 禁止的文件名模式
	ForbiddenPatterns []string `json:"forbidden_patterns"`
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		MaxDepth:     20,
		MaxFiles:     10000,
		MaxTotalSize: 500 * 1024 * 1024, // 500MB
		ForbiddenPatterns: []string{
			// 二进制/资源文件
			"*.exe", "*.dll", "*.so", "*.dylib",
			"*.bin", "*.dat", "*.db", "*.sqlite",
			"*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp", "*.ico",
			"*.mp3", "*.mp4", "*.avi", "*.mov",
			"*.zip", "*.tar", "*.gz", "*.rar", "*.7z",
			// 构建产物
			"node_modules/*", "__pycache__/*", "*.pyc",
			"dist/*", "build/*", "target/*", "bin/*", "obj/*",
		},
	}
}

// ValidationResult 路径验证结果
type ValidationResult struct {
	Valid      bool   `json:"valid"`
	CleanPath  string `json:"clean_path"`
	RejectReason string `json:"reject_reason,omitempty"`
}

// ValidatePath 验证路径安全性
// 返回验证结果和清理后的路径
func (c *SecurityConfig) ValidatePath(basePath, targetPath string) ValidationResult {
	// 清理路径
	clean := filepath.Clean(targetPath)

	// 检查路径遍历
	if strings.Contains(clean, "..") {
		return ValidationResult{
			Valid:        false,
			CleanPath:    clean,
			RejectReason: "路径遍历不被允许",
		}
	}

	// 构建完整路径
	fullPath := clean
	if !filepath.IsAbs(clean) {
		fullPath = filepath.Join(basePath, clean)
	}

	// 解析为绝对路径
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return ValidationResult{
			Valid:        false,
			CleanPath:    clean,
			RejectReason: fmt.Sprintf("无效路径：%v", err),
		}
	}

	// 检查是否在 basePath 内
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return ValidationResult{
			Valid:        false,
			CleanPath:    clean,
			RejectReason: fmt.Sprintf("基准路径无效：%v", err),
		}
	}

	// 确保吸收后的路径在基准路径内
	if !strings.HasPrefix(absPath, absBase+string(os.PathSeparator)) && absPath != absBase {
		return ValidationResult{
			Valid:        false,
			CleanPath:    clean,
			RejectReason: "路径超出基准目录范围",
		}
	}

	// 检查白名单
	if len(c.AllowedPrefixes) > 0 {
		allowed := false
		for _, prefix := range c.AllowedPrefixes {
			if strings.HasPrefix(absPath, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return ValidationResult{
				Valid:        false,
				CleanPath:    clean,
				RejectReason: "路径不在允许的白名单内",
			}
		}
	}

	// 检查黑名单
	for _, prefix := range c.ForbiddenPrefixes {
		if strings.HasPrefix(absPath, prefix) {
			return ValidationResult{
				Valid:        false,
				CleanPath:    clean,
				RejectReason: "路径在禁止的黑名单内",
			}
		}
	}

	// 检查文件名模式
	filename := filepath.Base(absPath)
	for _, pattern := range c.ForbiddenPatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return ValidationResult{
				Valid:        false,
				CleanPath:    clean,
				RejectReason: fmt.Sprintf("文件类型不被允许：%s", pattern),
			}
		}
	}

	return ValidationResult{
		Valid:       true,
		CleanPath:   clean,
		RejectReason: "",
	}
}

// ScanStats 目录扫描统计
type ScanStats struct {
	TotalFiles   int     `json:"total_files"`
	TotalSize    int64   `json:"total_size"`
	MaxDepth     int     `json:"max_depth"`
	SkippedFiles int     `json:"skipped_files"`
	SkippedSize  int64   `json:"skipped_size"`
	Warnings     []string `json:"warnings,omitempty"`
}

// ScanDirectory 安全地扫描目录
// 返回扫描统计和错误
func (c *SecurityConfig) ScanDirectory(rootPath string, fn func(path string, entry fs.DirEntry, relPath string, depth int) error) (*ScanStats, error) {
	stats := &ScanStats{
		Warnings: make([]string, 0),
	}

	// 验证根路径
	result := c.ValidatePath(rootPath, rootPath)
	if !result.Valid {
		return stats, fmt.Errorf("根路径验证失败：%s", result.RejectReason)
	}

	err := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			stats.SkippedFiles++
			return nil // 跳过错误项
		}

		// 计算相对深度
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil
		}

		// 计算深度
		depth := len(strings.Split(relPath, string(os.PathSeparator)))

		// 检查深度限制
		if depth > c.MaxDepth {
			stats.SkippedFiles++
			stats.Warnings = append(stats.Warnings, fmt.Sprintf("跳过超出深度限制的路径：%s (深度：%d)", path, depth))
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件数限制
		if !entry.IsDir() {
			if stats.TotalFiles >= c.MaxFiles {
				stats.SkippedFiles++
				stats.Warnings = append(stats.Warnings, fmt.Sprintf("跳过超出文件数限制的文件：%s", path))
				return nil
			}
		}

		// 检查大小限制
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				size := info.Size()
				if stats.TotalSize+size > c.MaxTotalSize {
					stats.SkippedFiles++
					stats.SkippedSize += size
					stats.Warnings = append(stats.Warnings, fmt.Sprintf("跳过超出大小限制的文件：%s (%d bytes)", path, size))
					return nil
				}
			}
		}

		// 调用处理函数
		if err := fn(path, entry, relPath, depth); err != nil {
			return err
		}

		// 更新统计
		if !entry.IsDir() {
			stats.TotalFiles++
			if info, err := entry.Info(); err == nil {
				stats.TotalSize += info.Size()
			}
		}

		if depth > stats.MaxDepth {
			stats.MaxDepth = depth
		}

		return nil
	})

	return stats, err
}

// IsForbiddenPattern 检查文件名是否匹配禁止的模式
func (c *SecurityConfig) IsForbiddenPattern(filename string) bool {
	for _, pattern := range c.ForbiddenPatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
		// 也检查是否匹配目录模式
		if strings.Contains(pattern, "/") {
			if matched, _ := filepath.Match(pattern, filename); matched {
				return true
			}
		}
	}
	return false
}
