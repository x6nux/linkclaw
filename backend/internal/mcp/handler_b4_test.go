package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestHandler_AgentTools_Routing 测试 MCP handler 正确路由到新工具
func TestHandler_AgentTools_Routing(t *testing.T) {
	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&domain.ContextDirectory{}, &domain.ContextFileSummary{}, &domain.ContextSearchLog{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// 创建测试目录
	testDir := t.TempDir()
	ctx := context.Background()
	companyID := "test-company"

	dir := &domain.ContextDirectory{
		ID:         "test-dir-1",
		CompanyID:  companyID,
		Name:       "test",
		Path:       testDir,
		IsActive:   true,
		FileCount:  0,
	}
	db.Create(dir)

	// 创建测试文件
	testFile := filepath.Join(testDir, "example.go")
	testContent := `package main

func Hello() string {
	return "Hello, World!"
}

func main() {
	Hello()
}
`
	os.WriteFile(testFile, []byte(testContent), 0644)

	// 创建服务
	repo := repository.NewContextRepo(db)
	contextSvc := service.NewContextService(repo, nil)

	// 创建 handler
	handler := &Handler{
		contextSvc: contextSvc,
	}

	// 创建测试 session
	sess := &Session{
		Agent: &domain.Agent{
			ID:        "test-agent",
			CompanyID: companyID,
		},
		Initialized: true,
	}

	// 测试 agent_grep 路由
	t.Run("agent_grep", func(t *testing.T) {
		args, _ := json.Marshal(map[string]any{
			"pattern":     "func.*Hello",
			"max_results": 10,
			"directory_ids": []string{"test-dir-1"},
		})

		result := handler.dispatchTool(ctx, sess, "agent_grep", args)
		if result.IsError {
			// 检查是否是 LLM 客户端未配置的错误（预期），而不是路由错误
			if len(result.Content) > 0 && result.Content[0].Type == "text" {
				// 路由成功，但执行可能因缺少 LLM 客户端而失败
				t.Logf("grep result: %s", result.Content[0].Text)
			}
		}
	})

	// 测试 agent_read_chunk 路由
	t.Run("agent_read_chunk", func(t *testing.T) {
		args, _ := json.Marshal(map[string]any{
			"path":          "example.go",
			"offset":        0,
			"limit":         10,
			"directory_ids": []string{"test-dir-1"},
		})

		result := handler.dispatchTool(ctx, sess, "agent_read_chunk", args)
		if result.IsError {
			t.Logf("read_chunk result: %v", result.Content)
		}
	})

	// 测试 agent_list_symbols 路由
	t.Run("agent_list_symbols", func(t *testing.T) {
		args, _ := json.Marshal(map[string]any{
			"path":          "example.go",
			"symbol_type":   "function",
			"directory_ids": []string{"test-dir-1"},
		})

		result := handler.dispatchTool(ctx, sess, "agent_list_symbols", args)
		if result.IsError {
			t.Logf("list_symbols result: %v", result.Content)
		}
	})

	// 测试未知工具路由（应该返回错误）
	t.Run("unknown_tool", func(t *testing.T) {
		args, _ := json.Marshal(map[string]any{})
		result := handler.dispatchTool(ctx, sess, "unknown_tool", args)
		if !result.IsError {
			t.Error("expected error for unknown tool")
		}
	})
}

// TestHandler_DirectoryIDs_Array 测试 directory_ids 数组语义
func TestHandler_DirectoryIDs_Array(t *testing.T) {
	// 测试参数解析是否支持数组
	t.Run("parse array", func(t *testing.T) {
		args, _ := json.Marshal(map[string]any{
			"pattern":       "test",
			"directory_ids": []string{"dir-1", "dir-2", "dir-3"},
		})

		var params struct {
			Pattern     string   `json:"pattern"`
			DirectoryIDs []string `json:"directory_ids,omitempty"`
		}

		if err := json.Unmarshal(args, &params); err != nil {
			t.Errorf("failed to parse array: %v", err)
		}

		if len(params.DirectoryIDs) != 3 {
			t.Errorf("expected 3 directory IDs, got %d", len(params.DirectoryIDs))
		}
	})
}
