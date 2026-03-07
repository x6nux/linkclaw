package service

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/linkclaw/backend/internal/domain"
)

func TestExecuteGrep_BasicSearch(t *testing.T) {
	agent := &ContextSearchAgent{}

	// 创建测试目录
	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{
			ID:       "test-dir-1",
			Name:     "test",
			Path:     testDir,
			IsActive: true,
		},
	}

	// 创建测试文件
	testFile := testDir + "/test.go"
	testContent := `package main

func Hello() string {
	return "Hello, World!"
}

func main() {
	Hello()
}
`
	writeFile(t, testFile, testContent)

	// 执行 grep 搜索
	args, _ := json.Marshal(map[string]any{
		"pattern":     "func.*Hello",
		"max_results": 10,
	})

	result := agent.executeGrep(AgentToolCall{
		ID:        "test-1",
		Name:      "grep",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		t.Errorf("grep failed: %s", result.Content)
	}

	// 验证结果包含匹配
	if !strings.Contains(result.Content, "Hello") {
		t.Errorf("expected result to contain 'Hello', got: %s", result.Content)
	}
}

func TestExecuteGrep_MaxResultsLimit(t *testing.T) {
	agent := &ContextSearchAgent{}

	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{ID: "test-dir-1", Name: "test", Path: testDir, IsActive: true},
	}

	// 创建包含多行匹配的文件
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, "func TestFunc"+string(rune('A'+i%26))+"() {}")
	}
	writeFile(t, testDir+"/multi.go", strings.Join(lines, "\n"))

	args, _ := json.Marshal(map[string]any{
		"pattern":     "func",
		"max_results": 10,
	})

	result := agent.executeGrep(AgentToolCall{
		ID:        "test-2",
		Name:      "grep",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		t.Errorf("grep failed: %s", result.Content)
	}

	// 验证结果数量限制
	var matches []struct {
		DirName  string `json:"dir_name"`
		FilePath string `json:"file_path"`
		LineNum  int    `json:"line_num"`
	}
	json.Unmarshal([]byte(result.Content), &matches)

	if len(matches) > 10 {
		t.Errorf("expected max 10 results, got %d", len(matches))
	}
}

func TestExecuteReadChunk_BasicRead(t *testing.T) {
	agent := &ContextSearchAgent{}

	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{ID: "test-dir-1", Name: "test", Path: testDir, IsActive: true},
	}

	// 创建测试文件
	var lines []string
	for i := 1; i <= 100; i++ {
		lines = append(lines, "Line "+string(rune('A'+i%26)))
	}
	writeFile(t, testDir+"/chunk.txt", strings.Join(lines, "\n"))

	args, _ := json.Marshal(map[string]any{
		"path":   "chunk.txt",
		"offset": 10,
		"limit":  20,
	})

	result := agent.executeReadChunk(AgentToolCall{
		ID:        "test-3",
		Name:      "read_chunk",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		t.Errorf("read_chunk failed: %s", result.Content)
	}

	// 验证返回结构
	var chunkResult struct {
		FilePath   string   `json:"file_path"`
		Offset     int      `json:"offset"`
		Limit      int      `json:"limit"`
		TotalLines int      `json:"total_lines"`
		Lines      []string `json:"lines"`
		HasMore    bool     `json:"has_more"`
	}
	json.Unmarshal([]byte(result.Content), &chunkResult)

	if chunkResult.Offset != 10 {
		t.Errorf("expected offset 10, got %d", chunkResult.Offset)
	}
	if chunkResult.Limit != 20 {
		t.Errorf("expected limit 20, got %d", chunkResult.Limit)
	}
	if len(chunkResult.Lines) != 20 {
		t.Errorf("expected 20 lines, got %d", len(chunkResult.Lines))
	}
	if !chunkResult.HasMore {
		t.Error("expected HasMore to be true")
	}
}

func TestExecuteReadChunk_HasMoreFlag(t *testing.T) {
	agent := &ContextSearchAgent{}

	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{ID: "test-dir-1", Name: "test", Path: testDir, IsActive: true},
	}

	writeFile(t, testDir+"/small.txt", "Line1\nLine2\nLine3\nLine4\nLine5")

	// 读取最后 2 行
	args, _ := json.Marshal(map[string]any{
		"path":   "small.txt",
		"offset": 3,
		"limit":  10,
	})

	result := agent.executeReadChunk(AgentToolCall{
		ID:        "test-4",
		Name:      "read_chunk",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		t.Errorf("read_chunk failed: %s", result.Content)
	}

	var chunkResult struct {
		TotalLines int  `json:"total_lines"`
		HasMore    bool `json:"has_more"`
	}
	json.Unmarshal([]byte(result.Content), &chunkResult)

	if chunkResult.TotalLines != 5 {
		t.Errorf("expected 5 total lines, got %d", chunkResult.TotalLines)
	}
	if chunkResult.HasMore {
		t.Error("expected HasMore to be false (at end of file)")
	}
}

func TestExecuteListSymbols_GoFunctions(t *testing.T) {
	agent := &ContextSearchAgent{}

	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{ID: "test-dir-1", Name: "test", Path: testDir, IsActive: true},
	}

	testContent := `package main

import "fmt"

func Hello() string {
	return "Hello"
}

func main() {
	fmt.Println(Hello())
}
`
	writeFile(t, testDir+"/main.go", testContent)

	args, _ := json.Marshal(map[string]any{
		"path":       "main.go",
		"symbol_type": "function",
	})

	result := agent.executeListSymbols(AgentToolCall{
		ID:        "test-5",
		Name:      "list_symbols",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		t.Errorf("list_symbols failed: %s", result.Content)
	}

	// 验证结果包含函数
	var symbols []Symbol
	json.Unmarshal([]byte(result.Content), &symbols)

	foundHello := false
	foundMain := false
	for _, s := range symbols {
		if s.Name == "Hello" {
			foundHello = true
		}
		if s.Name == "main" {
			foundMain = true
		}
	}

	if !foundHello {
		t.Error("expected to find Hello function")
	}
	if !foundMain {
		t.Error("expected to find main function")
	}
}

func TestExecuteListSymbols_TypeScript(t *testing.T) {
	agent := &ContextSearchAgent{}

	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{ID: "test-dir-1", Name: "test", Path: testDir, IsActive: true},
	}

	testContent := `import React from 'react'

export class Counter {
	value: number = 0
}

export interface Props {
	name: string
}

function greet(person: string) {
	return 'Hello ' + person
}
`
	writeFile(t, testDir+"/app.ts", testContent)

	args, _ := json.Marshal(map[string]any{
		"path":       "app.ts",
		"symbol_type": "all",
	})

	result := agent.executeListSymbols(AgentToolCall{
		ID:        "test-6",
		Name:      "list_symbols",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		t.Errorf("list_symbols failed: %s", result.Content)
	}

	var symbols []Symbol
	json.Unmarshal([]byte(result.Content), &symbols)

	foundCounter := false
	foundGreet := false
	for _, s := range symbols {
		if s.Name == "Counter" {
			foundCounter = true
		}
		if s.Name == "greet" {
			foundGreet = true
		}
	}

	if !foundCounter {
		t.Error("expected to find Counter class")
	}
	if !foundGreet {
		t.Error("expected to find greet function")
	}
}

func TestExecuteGrep_InvalidRegex(t *testing.T) {
	agent := &ContextSearchAgent{}

	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{ID: "test-dir-1", Name: "test", Path: testDir, IsActive: true},
	}

	// 无效的正则表达式
	args, _ := json.Marshal(map[string]any{
		"pattern": "[invalid(regex",
	})

	result := agent.executeGrep(AgentToolCall{
		ID:        "test-7",
		Name:      "grep",
		Arguments: args,
	}, dirs, 1024*1024)

	if !result.IsError {
		t.Error("expected error for invalid regex")
	}
	if !strings.Contains(result.Content, "Invalid regex") {
		t.Errorf("expected 'Invalid regex' error, got: %s", result.Content)
	}
}

func TestExecuteReadChunk_PathTraversal(t *testing.T) {
	agent := &ContextSearchAgent{}

	testDir := t.TempDir()
	dirs := []*domain.ContextDirectory{
		{ID: "test-dir-1", Name: "test", Path: testDir, IsActive: true},
	}

	// 尝试路径遍历攻击
	args, _ := json.Marshal(map[string]any{
		"path": "../../../etc/passwd",
	})

	result := agent.executeReadChunk(AgentToolCall{
		ID:        "test-8",
		Name:      "read_chunk",
		Arguments: args,
	}, dirs, 1024*1024)

	if !result.IsError {
		t.Error("expected error for path traversal")
	}
	if !strings.Contains(result.Content, "Path traversal not allowed") {
		t.Errorf("expected path traversal error, got: %s", result.Content)
	}
}

func TestExtractSymbols_Python(t *testing.T) {
	testContent := `
import os
from typing import List

class DataProcessor:
	def __init__(self):
		pass

	def process(self, data: List[str]) -> str:
		return ",".join(data)

async def fetch_data(url: str):
	pass
`

	symbols := extractPythonSymbols(testContent, "all")

	foundClass := false
	foundFunc := false
	foundImport := false

	for _, s := range symbols {
		if s.Name == "DataProcessor" {
			foundClass = true
		}
		if s.Name == "fetch_data" {
			foundFunc = true
		}
		if s.Type == "import" {
			foundImport = true
		}
	}

	if !foundClass {
		t.Error("expected to find DataProcessor class")
	}
	if !foundFunc {
		t.Error("expected to find fetch_data function")
	}
	if !foundImport {
		t.Error("expected to find imports")
	}
}

// 辅助函数
func writeFile(t *testing.T, path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}
