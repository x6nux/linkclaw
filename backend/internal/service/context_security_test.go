package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSecurityConfig_ValidatePath(t *testing.T) {
	config := DefaultSecurityConfig()
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		basePath   string
		targetPath string
		wantValid  bool
		wantReject string
	}{
		{
			name:       "正常相对路径",
			basePath:   tmpDir,
			targetPath: "subdir/file.go",
			wantValid:  true,
		},
		{
			name:       "路径遍历攻击",
			basePath:   tmpDir,
			targetPath: "../etc/passwd",
			wantValid:  false,
			wantReject: "路径遍历",
		},
		{
			name:       "双重路径遍历",
			basePath:   tmpDir,
			targetPath: "subdir/../../etc/passwd",
			wantValid:  false,
			wantReject: "路径遍历",
		},
		{
			name:       "绝对路径超出基准",
			basePath:   tmpDir,
			targetPath: "/etc/passwd",
			wantValid:  false,
			wantReject: "超出基准目录",
		},
		{
			name:       "正常绝对路径",
			basePath:   tmpDir,
			targetPath: filepath.Join(tmpDir, "file.go"),
			wantValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ValidatePath(tt.basePath, tt.targetPath)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid=%v, want %v, reject=%s", result.Valid, tt.wantValid, result.RejectReason)
			}
			if !tt.wantValid && tt.wantReject != "" {
				if result.RejectReason == "" {
					t.Errorf("期望有拒绝原因")
				}
			}
		})
	}
}

func TestSecurityConfig_ScanDirectory(t *testing.T) {
	// 创建测试目录结构
	tmpDir := t.TempDir()

	// 创建深层目录
	deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e", "f")
	if err := os.MkdirAll(deepPath, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建测试文件
	for i := 0; i < 5; i++ {
		path := filepath.Join(tmpDir, "a", "b", "c", "file"+string(rune('0'+i))+".go")
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("package test"), 0644)
	}

	// 创建大文件
	largeFile := filepath.Join(tmpDir, "large.bin")
	os.WriteFile(largeFile, make([]byte, 1024*1024), 0644) // 1MB

	tests := []struct {
		name      string
		config    *SecurityConfig
		wantFiles int
		wantWarns int
	}{
		{
			name: "正常扫描",
			config: &SecurityConfig{
				MaxDepth:     20,
				MaxFiles:     100,
				MaxTotalSize: 10 * 1024 * 1024,
			},
			wantFiles: 6, // 5 个小文件 + 1 个大文件
			wantWarns: 0,
		},
		{
			name: "深度限制",
			config: &SecurityConfig{
				MaxDepth:     3,
				MaxFiles:     100,
				MaxTotalSize: 10 * 1024 * 1024,
			},
			wantFiles: 1, // 根目录的大文件会被扫描到
			wantWarns: 1,
		},
		{
			name: "文件数限制",
			config: &SecurityConfig{
				MaxDepth:     20,
				MaxFiles:     3,
				MaxTotalSize: 10 * 1024 * 1024,
			},
			wantFiles: 3,
			wantWarns: 1,
		},
		{
			name: "大小限制",
			config: &SecurityConfig{
				MaxDepth:     20,
				MaxFiles:     100,
				MaxTotalSize: 500 * 1024, // 500KB
			},
			wantFiles: 5, // 只扫描小文件
			wantWarns: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := tt.config.ScanDirectory(tmpDir, func(path string, entry os.DirEntry, relPath string, depth int) error {
				return nil
			})

			if err != nil {
				t.Errorf("ScanDirectory error: %v", err)
			}

			if stats.TotalFiles != tt.wantFiles {
				t.Errorf("TotalFiles=%d, want %d", stats.TotalFiles, tt.wantFiles)
			}

			if len(stats.Warnings) < tt.wantWarns {
				t.Errorf("Warnings=%d, want at least %d", len(stats.Warnings), tt.wantWarns)
			}
		})
	}
}

func TestSecurityConfig_ForbiddenPatterns(t *testing.T) {
	config := DefaultSecurityConfig()

	tests := []struct {
		name     string
		filename string
		wantBool bool
	}{
		{"可执行文件", "malware.exe", true},
		{"动态库", "lib.so", true},
		{"图片", "photo.jpg", true},
		{"压缩包", "archive.zip", true},
		{"Go 文件", "main.go", false},
		{"TypeScript 文件", "app.ts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.IsForbiddenPattern(tt.filename)
			if got != tt.wantBool {
				t.Errorf("IsForbiddenPattern(%s)=%v, want %v", tt.filename, got, tt.wantBool)
			}
		})
	}
}

func TestContextMetrics_RecordToolCall(t *testing.T) {
	m := NewContextMetrics()

	// 记录多次 tool call
	m.RecordToolCall("read_file", 100, false)
	m.RecordToolCall("read_file", 200, false)
	m.RecordToolCall("read_file", 50, true)
	m.RecordToolCall("list_files", 30, false)

	snapshot := m.GetSnapshot()

	if len(snapshot.ToolCalls) != 2 {
		t.Errorf("ToolCalls=%d, want 2", len(snapshot.ToolCalls))
	}

	// 检查 read_file 统计
	var readStats *ToolCallStats
	for _, s := range snapshot.ToolCalls {
		if s.Name == "read_file" {
			readStats = s
			break
		}
	}

	if readStats == nil {
		t.Fatal("read_file stats not found")
	}

	if readStats.Count != 3 {
		t.Errorf("Count=%d, want 3", readStats.Count)
	}
	if readStats.Success != 2 {
		t.Errorf("Success=%d, want 2", readStats.Success)
	}
	if readStats.Error != 1 {
		t.Errorf("Error=%d, want 1", readStats.Error)
	}
	if readStats.MinMs != 50 {
		t.Errorf("MinMs=%d, want 50", readStats.MinMs)
	}
	if readStats.MaxMs != 200 {
		t.Errorf("MaxMs=%d, want 200", readStats.MaxMs)
	}
}

func TestContextMetrics_RecordFailure(t *testing.T) {
	m := NewContextMetrics()

	m.RecordFailure("path_traversal")
	m.RecordFailure("path_traversal")
	m.RecordFailure("file_too_large")

	snapshot := m.GetSnapshot()

	if len(snapshot.Failures) != 2 {
		t.Errorf("Failures=%d, want 2", len(snapshot.Failures))
	}
	if snapshot.Failures["path_traversal"] != 2 {
		t.Errorf("path_traversal=%d, want 2", snapshot.Failures["path_traversal"])
	}
	if snapshot.Failures["file_too_large"] != 1 {
		t.Errorf("file_too_large=%d, want 1", snapshot.Failures["file_too_large"])
	}
}

func TestContextMetrics_Reset(t *testing.T) {
	m := NewContextMetrics()

	m.RecordToolCall("read_file", 100, false)
	m.RecordFailure("test")
	m.RecordTokens(1000, 500)
	m.RecordCost(100)

	m.Reset()

	snapshot := m.GetSnapshot()

	if len(snapshot.ToolCalls) != 0 {
		t.Errorf("ToolCalls=%d, want 0", len(snapshot.ToolCalls))
	}
	if len(snapshot.Failures) != 0 {
		t.Errorf("Failures=%d, want 0", len(snapshot.Failures))
	}
	if snapshot.InputTokens != 0 {
		t.Errorf("InputTokens=%d, want 0", snapshot.InputTokens)
	}
	if snapshot.TotalCost != 0 {
		t.Errorf("TotalCost=%d, want 0", snapshot.TotalCost)
	}
}

func TestTokenBucket_Allow(t *testing.T) {
	bucket := NewTokenBucket(10, 1) // 10 tokens max, 1 token/sec refill

	ctx := context.Background()

	// 应该允许前 10 次请求
	for i := 0; i < 10; i++ {
		if !bucket.Allow(ctx, "test") {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// 第 11 次应该被拒绝
	if bucket.Allow(ctx, "test") {
		t.Error("Request 11 should be rejected")
	}
}

func TestTokenBucket_Wait(t *testing.T) {
	bucket := NewTokenBucket(1, 10) // 1 token max, 10 tokens/sec refill

	ctx := context.Background()

	// 第一次应该立即允许
	if !bucket.Allow(ctx, "test") {
		t.Error("First request should be allowed")
	}

	// 第二次应该被拒绝
	if bucket.Allow(ctx, "test") {
		t.Error("Second request should be rejected initially")
	}

	// Wait 应该等待令牌补充
	go func() {
		time.Sleep(200 * time.Millisecond)
		bucket.mu.Lock()
		bucket.tokens = 1 // 手动补充令牌
		bucket.mu.Unlock()
	}()

	err := bucket.Wait(ctx, "test")
	if err != nil {
		t.Errorf("Wait error: %v", err)
	}
}

func TestRateLimiterManager(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalRPS = 10
	config.GlobalConcurrency = 5
	config.DirectoryRPS = 5
	config.DirectoryConcurrency = 2

	manager := NewRateLimiterManager(config)
	ctx := context.Background()

	// 测试全局限流
	for i := 0; i < 5; i++ {
		if err := manager.AcquireGlobal(ctx); err != nil {
			t.Errorf("AcquireGlobal %d error: %v", i, err)
		}
	}

	// 测试目录限流
	dirID := "test-dir"
	for i := 0; i < 2; i++ {
		if err := manager.AcquireDirectory(ctx, dirID); err != nil {
			t.Errorf("AcquireDirectory %d error: %v", i, err)
		}
	}

	// 测试 token 预算
	if !manager.CheckTokenBudget(50000) {
		t.Error("50000 tokens should be within budget")
	}
	if manager.CheckTokenBudget(200000) {
		t.Error("200000 tokens should exceed budget")
	}

	// 测试超时获取
	if timeout := manager.GetTimeout("search"); timeout != 30*time.Second {
		t.Errorf("Search timeout=%v, want 30s", timeout)
	}
	if timeout := manager.GetTimeout("llm_call"); timeout != 60*time.Second {
		t.Errorf("LLM call timeout=%v, want 60s", timeout)
	}
}

func TestCostBudget(t *testing.T) {
	budget := NewCostBudget(1000000, 0.8, true) // 1 dollar, warn at 80%

	// 记录 50% 预算
	allowed, warned, err := budget.CheckAndRecord(500000)
	if !allowed {
		t.Error("Should allow 50% usage")
	}
	if warned {
		t.Error("Should not warn at 50%")
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 记录到 90%，应该预警
	allowed, warned, err = budget.CheckAndRecord(400000)
	if !allowed {
		t.Error("Should allow 90% usage")
	}
	if !warned {
		t.Error("Should warn at 90%")
	}

	// 尝试超出预算，应该拒绝
	allowed, warned, err = budget.CheckAndRecord(200000)
	if allowed {
		t.Error("Should reject over-budget request")
	}
	if err == nil {
		t.Error("Should return error for over-budget")
	}

	// 检查剩余预算
	remaining := budget.GetRemaining()
	if remaining != 100000 {
		t.Errorf("Remaining=%d, want 100000", remaining)
	}

	// 检查使用比例
	ratio := budget.GetUsageRatio()
	if ratio < 0.89 || ratio > 0.91 {
		t.Errorf("Usage ratio=%.2f, want ~0.9", ratio)
	}
}
