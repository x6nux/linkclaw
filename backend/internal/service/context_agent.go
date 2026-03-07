package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

// AgentTool represents a tool definition for the AI agent
type AgentTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// AgentToolCall represents a tool call from the agent
type AgentToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// AgentMessage represents a message in the agent conversation
type AgentMessage struct {
	Role    string          `json:"role"`
	Content string          `json:"content,omitempty"`
	Tools   []AgentToolCall `json:"tool_calls,omitempty"`
}

// AgentToolResult represents the result of a tool execution
type AgentToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// ContextSearchAgent handles agent-based context search
type ContextSearchAgent struct {
	llmCli  *ContextLLMClient
	repo    repository.ContextRepo
	metrics *ContextMetrics
}

// NewContextSearchAgent creates a new context search agent
func NewContextSearchAgent(llmCli *ContextLLMClient, repo repository.ContextRepo) *ContextSearchAgent {
	return &ContextSearchAgent{
		llmCli:  llmCli,
		repo:    repo,
		metrics: NewContextMetrics(),
	}
}

// GetMetrics returns the metrics collector
func (a *ContextSearchAgent) GetMetrics() *ContextMetrics {
	return a.metrics
}

// ExecuteGrep 执行 grep 工具（供 MCP handler 调用）
func (a *ContextSearchAgent) ExecuteGrep(tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	return a.executeGrep(tc, dirs, maxSize)
}

// ExecuteReadChunk 执行 read_chunk 工具（供 MCP handler 调用）
func (a *ContextSearchAgent) ExecuteReadChunk(tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	return a.executeReadChunk(tc, dirs, maxSize)
}

// ExecuteListSymbols 执行 list_symbols 工具（供 MCP handler 调用）
func (a *ContextSearchAgent) ExecuteListSymbols(tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	return a.executeListSymbols(tc, dirs, maxSize)
}

// AgentSearchInput is the input for agent-based search
type AgentSearchInput struct {
	CompanyID    string
	Query        string
	DirectoryIDs []string // Optional: specific directory IDs to search (empty = all active)
	MaxTurns     int      // Maximum conversation turns (default 10)
	MaxFileSize  int64    // Maximum file size to read (default 1MB)
}

// AgentSearchOutput is the output from agent-based search
type AgentSearchOutput struct {
	Answer    string   `json:"answer"`
	Files     []string `json:"files_read,omitempty"`
	LatencyMs int      `json:"latency_ms"`
}

// Search executes the agent-based context search
func (a *ContextSearchAgent) Search(ctx context.Context, in AgentSearchInput) (*AgentSearchOutput, error) {
	start := time.Now()

	if in.MaxTurns == 0 {
		in.MaxTurns = 10
	}
	if in.MaxFileSize == 0 {
		in.MaxFileSize = 1024 * 1024 // 1MB
	}

	// Load directories from database
	dirs, err := a.loadDirectories(ctx, in.CompanyID, in.DirectoryIDs)
	if err != nil {
		return nil, fmt.Errorf("load directories: %w", err)
	}
	if len(dirs) == 0 {
		return &AgentSearchOutput{
			Answer:    "No active directories found.",
			LatencyMs: int(time.Since(start).Milliseconds()),
		}, nil
	}

	systemPrompt := a.buildSystemPrompt(dirs)
	messages := []AgentMessage{
		{Role: "user", Content: in.Query},
	}

	var filesRead []string
	tools := ContextSearchAgentTools

	for turn := 0; turn < in.MaxTurns; turn++ {
		response, toolCalls, err := a.llmCli.CallWithTools(ctx, systemPrompt, messages, tools)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// No tool calls means the agent has finished
		if len(toolCalls) == 0 {
			return &AgentSearchOutput{
				Answer:    response,
				Files:     filesRead,
				LatencyMs: int(time.Since(start).Milliseconds()),
			}, nil
		}

		// Execute tool calls and collect results
		var toolResults []AgentToolResult
		for _, tc := range toolCalls {
			result := a.executeTool(ctx, tc, dirs, in.MaxFileSize)
			toolResults = append(toolResults, result)

			// Track files read
			if tc.Name == "read_file" && !result.IsError {
				var args struct {
					Path string `json:"path"`
				}
				if json.Unmarshal(tc.Arguments, &args) == nil {
					filesRead = append(filesRead, args.Path)
				}
			}
		}

		// Add assistant message with tool calls and tool results to conversation
		messages = append(messages, AgentMessage{
			Role:    "assistant",
			Content: response,
			Tools:   toolCalls,
		})

		// Add tool results as user message
		resultsJSON, _ := json.Marshal(toolResults)
		messages = append(messages, AgentMessage{
			Role:    "user",
			Content: string(resultsJSON),
		})
	}

	return &AgentSearchOutput{
		Answer:    "Agent exceeded maximum turns without completing the search.",
		Files:     filesRead,
		LatencyMs: int(time.Since(start).Milliseconds()),
	}, nil
}

// loadDirectories loads directories from the repository
func (a *ContextSearchAgent) loadDirectories(ctx context.Context, companyID string, directoryIDs []string) ([]*domain.ContextDirectory, error) {
	if len(directoryIDs) > 0 {
		var dirs []*domain.ContextDirectory
		for _, id := range directoryIDs {
			d, err := a.repo.GetDirectoryByID(ctx, id)
			if err != nil {
				return nil, err
			}
			if d != nil && d.IsActive {
				dirs = append(dirs, d)
			}
		}
		return dirs, nil
	}
	return a.repo.ListActiveDirectories(ctx, companyID)
}

// executeTool executes a single tool call with request-level context
func (a *ContextSearchAgent) executeTool(ctx context.Context, tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	start := time.Now()
	var result AgentToolResult

	switch tc.Name {
	case "read_file":
		result = a.executeReadFile(ctx, tc, dirs, maxSize)
	case "list_files":
		result = a.executeListFiles(ctx, tc, dirs)
	case "search_index":
		result = a.executeSearchIndex(ctx, tc, dirs, maxSize)
	case "grep":
		result = a.executeGrep(tc, dirs, maxSize)
	case "read_chunk":
		result = a.executeReadChunk(tc, dirs, maxSize)
	case "list_symbols":
		result = a.executeListSymbols(tc, dirs, maxSize)
	default:
		result = AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Unknown tool: %s", tc.Name),
			IsError:    true,
		}
	}

	// Record tool call metrics
	a.metrics.RecordToolCall(tc.Name, time.Since(start).Milliseconds(), result.IsError)

	return result
}

// executeReadFile handles the read_file tool
func (a *ContextSearchAgent) executeReadFile(ctx context.Context, tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	var args struct {
		Path      string `json:"path"`
		StartLine int    `json:"start_line"`
		EndLine   int    `json:"end_line"`
	}
	if err := json.Unmarshal(tc.Arguments, &args); err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Invalid arguments: %v", err),
			IsError:    true,
		}
	}

	// Security: validate path using unified security config
	secConfig := DefaultSecurityConfig()
	var foundDir *domain.ContextDirectory
	var validPath string

	for _, dir := range dirs {
		result := secConfig.ValidatePath(dir.Path, args.Path)
		if result.Valid {
			foundDir = dir
			validPath = result.CleanPath
			break
		}
	}

	if foundDir == nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Path traversal not allowed or file not in configured directories",
			IsError:    true,
		}
	}

	fullPath := filepath.Join(foundDir.Path, validPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("File not found: %s", args.Path),
			IsError:    true,
		}
	}

	if info.Size() > maxSize {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("File too large: %d bytes (max: %d)", info.Size(), maxSize),
			IsError:    true,
		}
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Failed to read file: %v", err),
			IsError:    true,
		}
	}

	lines := strings.Split(string(content), "\n")
	start := 1
	end := len(lines)

	if args.StartLine > 0 {
		start = args.StartLine
	}
	if args.EndLine > 0 && args.EndLine < end {
		end = args.EndLine
	}

	if start > len(lines) {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Start line exceeds file length",
			IsError:    true,
		}
	}

	result := fmt.Sprintf("[%s] %s:\n", foundDir.Name, args.Path)
	result += strings.Join(lines[start-1:end], "\n")
	return AgentToolResult{
		ToolCallID: tc.ID,
		Content:    result,
	}
}

// executeListFiles handles the list_files tool
func (a *ContextSearchAgent) executeListFiles(ctx context.Context, tc AgentToolCall, dirs []*domain.ContextDirectory) AgentToolResult {
	var args struct {
		Pattern string `json:"pattern"`
	}
	json.Unmarshal(tc.Arguments, &args)

	var files []string
	for _, dir := range dirs {
		err := filepath.WalkDir(dir.Path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(dir.Path, path)
			if err != nil {
				return nil
			}

			// Apply pattern filter if specified
			if args.Pattern != "" {
				matched, err := filepath.Match(args.Pattern, rel)
				if err != nil || !matched {
					// Also try matching just the filename
					matched, _ = filepath.Match(args.Pattern, filepath.Base(rel))
					if !matched {
						return nil
					}
				}
			}

			files = append(files, fmt.Sprintf("%s:%s", dir.Name, rel))
			return nil
		})
		if err != nil {
			return AgentToolResult{
				ToolCallID: tc.ID,
				Content:    fmt.Sprintf("Failed to list files: %v", err),
				IsError:    true,
			}
		}
	}

	if len(files) == 0 {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "No files found.",
		}
	}

	resultJSON, _ := json.Marshal(files)
	return AgentToolResult{
		ToolCallID: tc.ID,
		Content:    string(resultJSON),
	}
}

// executeGrep handles the grep tool - search for regex pattern in file contents
func (a *ContextSearchAgent) executeGrep(tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	var args struct {
		Pattern     string `json:"pattern"`
		FilePattern string `json:"file_pattern"`
		IgnoreCase  bool   `json:"ignore_case"`
		MaxResults  int    `json:"max_results"`
	}
	if err := json.Unmarshal(tc.Arguments, &args); err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Invalid arguments: %v", err),
			IsError:    true,
		}
	}

	if args.Pattern == "" {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Pattern is required",
			IsError:    true,
		}
	}

	// 设置默认值和上限
	if args.MaxResults <= 0 {
		args.MaxResults = 50
	}
	if args.MaxResults > 200 {
		args.MaxResults = 200
	}

	// 编译正则表达式
	flags := ""
	if args.IgnoreCase {
		flags = "(?i)"
	}
	regex, err := regexp.Compile(flags + args.Pattern)
	if err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Invalid regex pattern: %v", err),
			IsError:    true,
		}
	}

	type GrepMatch struct {
		DirName   string `json:"dir_name"`
		FilePath  string `json:"file_path"`
		LineNum   int    `json:"line_num"`
		Line      string `json:"line"`
		MatchText string `json:"match_text"`
	}

	var matches []GrepMatch
	for _, dir := range dirs {
		err := filepath.WalkDir(dir.Path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(dir.Path, path)
			if err != nil {
				return nil
			}

			// 应用文件模式过滤
			if args.FilePattern != "" {
				matched, err := filepath.Match(args.FilePattern, rel)
				if err != nil || !matched {
					matched, _ = filepath.Match(args.FilePattern, filepath.Base(rel))
					if !matched {
						return nil
					}
				}
			}

			// 检查文件大小
			info, err := os.Stat(path)
			if err != nil || info.Size() > maxSize {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			lines := strings.Split(string(content), "\n")
			for i, line := range lines {
				if len(matches) >= args.MaxResults {
					break
				}
				if match := regex.FindString(line); match != "" {
					matches = append(matches, GrepMatch{
						DirName:   dir.Name,
						FilePath:  rel,
						LineNum:   i + 1,
						Line:      truncateLine(line, 200),
						MatchText: match,
					})
				}
			}
			return nil
		})
		if err != nil {
			continue
		}
	}

	if len(matches) == 0 {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "No matches found.",
		}
	}

	resultJSON, _ := json.Marshal(matches)
	return AgentToolResult{
		ToolCallID: tc.ID,
		Content:    string(resultJSON),
	}
}

// executeReadChunk handles the read_chunk tool - read a chunk of a large file
func (a *ContextSearchAgent) executeReadChunk(tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	var args struct {
		Path   string `json:"path"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
	if err := json.Unmarshal(tc.Arguments, &args); err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Invalid arguments: %v", err),
			IsError:    true,
		}
	}

	if args.Path == "" {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Path is required",
			IsError:    true,
		}
	}

	// 设置默认值和上限
	if args.Limit <= 0 {
		args.Limit = 100
	}
	if args.Limit > 500 {
		args.Limit = 500
	}
	if args.Offset < 0 {
		args.Offset = 0
	}

	// 安全：防止路径遍历
	cleanPath := filepath.Clean(args.Path)
	if strings.Contains(cleanPath, "..") {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Path traversal not allowed",
			IsError:    true,
		}
	}

	// 在配置的目录中查找文件
	for _, dir := range dirs {
		fullPath := filepath.Join(dir.Path, cleanPath)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.Size() > maxSize {
			return AgentToolResult{
				ToolCallID: tc.ID,
				Content:    fmt.Sprintf("File too large: %d bytes (max: %d)", info.Size(), maxSize),
				IsError:    true,
			}
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		totalLines := len(lines)

		if args.Offset >= totalLines {
			return AgentToolResult{
				ToolCallID: tc.ID,
				Content:    fmt.Sprintf("Offset %d exceeds file length %d", args.Offset, totalLines),
				IsError:    true,
			}
		}

		end := args.Offset + args.Limit
		if end > totalLines {
			end = totalLines
		}

		type ChunkResult struct {
			DirName    string   `json:"dir_name"`
			FilePath   string   `json:"file_path"`
			Offset     int      `json:"offset"`
			Limit      int      `json:"limit"`
			TotalLines int      `json:"total_lines"`
			Lines      []string `json:"lines"`
			HasMore    bool     `json:"has_more"`
		}

		result := ChunkResult{
			DirName:    dir.Name,
			FilePath:   args.Path,
			Offset:     args.Offset,
			Limit:      args.Limit,
			TotalLines: totalLines,
			Lines:      lines[args.Offset:end],
			HasMore:    end < totalLines,
		}

		resultJSON, _ := json.Marshal(result)
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    string(resultJSON),
		}
	}

	return AgentToolResult{
		ToolCallID: tc.ID,
		Content:    fmt.Sprintf("File not found: %s", args.Path),
		IsError:    true,
	}
}

// executeListSymbols handles the list_symbols tool - list symbols in a file
func (a *ContextSearchAgent) executeListSymbols(tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	var args struct {
		Path       string `json:"path"`
		SymbolType string `json:"symbol_type"`
	}
	if err := json.Unmarshal(tc.Arguments, &args); err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Invalid arguments: %v", err),
			IsError:    true,
		}
	}

	if args.Path == "" {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Path is required",
			IsError:    true,
		}
	}

	if args.SymbolType == "" {
		args.SymbolType = "all"
	}

	// 安全：防止路径遍历
	cleanPath := filepath.Clean(args.Path)
	if strings.Contains(cleanPath, "..") {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Path traversal not allowed",
			IsError:    true,
		}
	}

	// 在配置的目录中查找文件
	for _, dir := range dirs {
		fullPath := filepath.Join(dir.Path, cleanPath)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.Size() > maxSize {
			return AgentToolResult{
				ToolCallID: tc.ID,
				Content:    fmt.Sprintf("File too large: %d bytes (max: %d)", info.Size(), maxSize),
				IsError:    true,
			}
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		symbols := extractSymbols(string(content), filepath.Ext(args.Path), args.SymbolType)
		symbolsJSON, _ := json.Marshal(symbols)
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    string(symbolsJSON),
		}
	}

	return AgentToolResult{
		ToolCallID: tc.ID,
		Content:    fmt.Sprintf("File not found: %s", args.Path),
		IsError:    true,
	}
}

// truncateLine 截断单行到指定长度
func truncateLine(line string, maxLen int) string {
	line = strings.TrimSpace(line)
	if len(line) <= maxLen {
		return line
	}
	return line[:maxLen] + "..."
}

// executeSearchIndex handles the search_index tool
// 基于文件摘要索引进行快速召回，避免读取完整文件内容
func (a *ContextSearchAgent) executeSearchIndex(ctx context.Context, tc AgentToolCall, dirs []*domain.ContextDirectory, maxSize int64) AgentToolResult {
	var args struct {
		Query        string   `json:"query"`
		DirectoryIDs []string `json:"directory_ids"`
		MaxResults   int      `json:"max_results"`
	}
	if err := json.Unmarshal(tc.Arguments, &args); err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Invalid arguments: %v", err),
			IsError:    true,
		}
	}

	if args.Query == "" {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "Query is required",
			IsError:    true,
		}
	}

	if args.MaxResults <= 0 {
		args.MaxResults = 10
	}
	if args.MaxResults > 50 {
		args.MaxResults = 50
	}

	// 过滤目录
	var targetDirs []*domain.ContextDirectory
	dirIDSet := make(map[string]bool)
	for _, id := range args.DirectoryIDs {
		dirIDSet[id] = true
	}
	for _, d := range dirs {
		if len(args.DirectoryIDs) == 0 || dirIDSet[d.ID] {
			targetDirs = append(targetDirs, d)
		}
	}

	if len(targetDirs) == 0 {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "No directories to search",
			IsError:    true,
		}
	}

	// 使用请求级 context 而非 context.Background()
	var allSummaries []*domain.ContextFileSummary
	dirMap := make(map[string]string)

	for _, d := range targetDirs {
		summaries, err := a.repo.ListFileSummaries(ctx, d.ID)
		if err != nil {
			continue
		}
		for _, s := range summaries {
			allSummaries = append(allSummaries, s)
			dirMap[s.FilePath] = d.ID
		}
	}

	if len(allSummaries) == 0 {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    "No file summaries available. Index may not be built yet.",
			IsError:    true,
		}
	}

	// 使用 LLM 对摘要进行语义搜索
	files := make([]FileContent, 0, len(allSummaries))
	for _, s := range allSummaries {
		files = append(files, FileContent{
			FilePath: s.FilePath,
			Content:  s.Summary,
			Language: s.Language,
		})
	}

	llmResults, err := a.llmCli.SemanticSearchFromSummaries(ctx, args.Query, files)
	if err != nil {
		return AgentToolResult{
			ToolCallID: tc.ID,
			Content:    fmt.Sprintf("Search failed: %v", err),
			IsError:    true,
		}
	}

	// 裁剪结果数量
	if len(llmResults) > args.MaxResults {
		llmResults = llmResults[:args.MaxResults]
	}

	// 构建返回结果
	type SearchResult struct {
		FilePath    string  `json:"file_path"`
		Language    string  `json:"language"`
		Summary     string  `json:"summary"`
		Relevance   float64 `json:"relevance"`
		Reason      string  `json:"reason"`
		LineCount   int     `json:"line_count"`
		DirectoryID string  `json:"directory_id"`
	}

	results := make([]SearchResult, 0, len(llmResults))
	for _, r := range llmResults {
		results = append(results, SearchResult{
			FilePath:    r.FilePath,
			Language:    r.Language,
			Summary:     r.Summary,
			Relevance:   r.Relevance,
			Reason:      r.Reason,
			LineCount:   r.LineCount,
			DirectoryID: dirMap[r.FilePath],
		})
	}

	resultJSON, _ := json.Marshal(results)
	return AgentToolResult{
		ToolCallID: tc.ID,
		Content:    string(resultJSON),
	}
}

// buildSystemPrompt constructs the system prompt with available directories
func (a *ContextSearchAgent) buildSystemPrompt(dirs []*domain.ContextDirectory) string {
	var sb strings.Builder
	sb.WriteString(`You are a code search assistant. Your task is to help users find relevant code in their codebase.

You have access to the following tools:
- search_index: Fast semantic search over file summaries/index (args: query, directory_ids?, max_results?) - Use FIRST for quick relevance matching
- read_file: Read file content (args: path, start_line?, end_line?)
- list_files: List files matching pattern (args: pattern?)
- grep: Search for regex pattern in file contents (args: pattern, file_pattern?, ignore_case?, max_results?)
- read_chunk: Read a chunk of a large file (args: path, offset?, limit?)
- list_symbols: List symbols in a file (args: path, symbol_type?)

Available directories:
`)
	for _, d := range dirs {
		desc := d.Description
		if desc == "" {
			desc = "no description"
		}
		sb.WriteString(fmt.Sprintf("- %s: %s (%s)\n", d.Name, d.Path, desc))
	}

	sb.WriteString(`
Search strategy:
1. FIRST use search_index for fast semantic recall from file summaries
2. If index results are insufficient, use list_files to discover more files
3. Use grep to search for specific patterns across files
4. Use list_symbols to understand file structure quickly
5. Use read_file or read_chunk to examine file contents in detail
6. Provide a comprehensive natural language answer

Always respond in the same language as the user's question.`)

	return sb.String()
}

// Pre-defined tool schemas for the context search agent
var (
	ReadFileTool = AgentTool{
		Name:        "read_file",
		Description: "Read the content of a file at the specified path. Use this to examine file contents when you need to understand code, find specific implementations, or verify details.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The file path relative to the directory root",
				},
				"start_line": map[string]any{
					"type":        "integer",
					"description": "Optional start line number (1-indexed)",
				},
				"end_line": map[string]any{
					"type":        "integer",
					"description": "Optional end line number (inclusive)",
				},
			},
			"required": []string{"path"},
		},
	}

	ListFilesTool = AgentTool{
		Name:        "list_files",
		Description: "List files in the directory, optionally filtered by a glob pattern. Use this to discover what files exist before reading them.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "Optional glob pattern to filter files (e.g., '*.go', '**/*.ts')",
				},
			},
		},
	}

	GrepTool = AgentTool{
		Name:        "grep",
		Description: "Search for a regex pattern in file contents. Returns matching lines with file paths and line numbers.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "Regex pattern to search for",
				},
				"file_pattern": map[string]any{
					"type":        "string",
					"description": "Optional glob pattern to filter files (e.g., '*.go')",
				},
				"ignore_case": map[string]any{
					"type":        "boolean",
					"description": "Case insensitive search (default: false)",
				},
				"max_results": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results (default: 50, max: 200)",
				},
			},
			"required": []string{"pattern"},
		},
	}

	ReadChunkTool = AgentTool{
		Name:        "read_chunk",
		Description: "Read a specific chunk/segment of a large file by offset and limit. Useful for paginated reading of large files.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The file path relative to the directory root",
				},
				"offset": map[string]any{
					"type":        "integer",
					"description": "Chunk offset (0-indexed)",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Number of lines to read (default: 100, max: 500)",
				},
			},
			"required": []string{"path"},
		},
	}

	ListSymbolsTool = AgentTool{
		Name:        "list_symbols",
		Description: "List symbols (functions, classes, variables, imports) in a file. Helps understand code structure without reading full content.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The file path relative to the directory root",
				},
				"symbol_type": map[string]any{
					"type":        "string",
					"description": "Filter by symbol type",
					"enum":        []string{"function", "class", "interface", "variable", "import", "all"},
				},
			},
			"required": []string{"path"},
		},
	}

	SearchIndexTool = AgentTool{
		Name:        "search_index",
		Description: "Fast semantic search over file summaries/index. Returns relevant files with relevance scores and summaries. Use this FIRST before reading files.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query (natural language question or keywords)",
				},
				"directory_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Optional directory IDs to search (empty = all active directories)",
				},
				"max_results": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results (default: 10, max: 50)",
				},
			},
			"required": []string{"query"},
		},
	}

	// B4 任务实现：以下工具待实现
	// GrepTool, ReadChunkTool, ListSymbolsTool 定义保留供 B4 使用

	ContextSearchAgentTools = []AgentTool{SearchIndexTool, ReadFileTool, ListFilesTool, GrepTool, ReadChunkTool, ListSymbolsTool}
)

// Symbol 表示代码中的符号
type Symbol struct {
	Name       string `json:"name"`
	Type       string `json:"type"`       // function, class, interface, variable, import
	Line       int    `json:"line"`       // 行号
	Signature  string `json:"signature"`  // 函数签名/声明
	Exported   bool   `json:"exported"`   // 是否导出
}

// extractSymbols 从代码中提取符号（基于正则的轻量级解析）
func extractSymbols(content, ext, symbolType string) []Symbol {
	var symbols []Symbol

	// 根据文件类型选择解析策略
	switch ext {
	case ".go":
		symbols = extractGoSymbols(content, symbolType)
	case ".ts", ".tsx", ".js", ".jsx":
		symbols = extractJSSymbols(content, symbolType)
	case ".py":
		symbols = extractPythonSymbols(content, symbolType)
	default:
		symbols = extractGenericSymbols(content, symbolType)
	}

	return symbols
}

// extractGoSymbols 提取 Go 代码符号
func extractGoSymbols(content, symbolType string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// 函数定义: func (recv) Name(args) return
	funcRegex := regexp.MustCompile(`^func\s+(?:\([^)]+\)\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	// 类型定义: type Name struct/interface
	typeRegex := regexp.MustCompile(`^type\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+(struct|interface)`)
	// 变量定义: var Name = 或 Name :=
	varRegex := regexp.MustCompile(`^(?:var\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:=|:=)`)
	// import
	importRegex := regexp.MustCompile(`^import\s+(?:\(([^)]+)\)|"([^"]+)")`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if symbolType == "all" || symbolType == "function" {
			if matches := funcRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "function",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  true,
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "class" {
			if matches := typeRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      matches[2],
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  true,
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "variable" {
			if matches := varRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "variable",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  true,
				})
			}
		}

		if symbolType == "all" || symbolType == "import" {
			if matches := importRegex.FindStringSubmatch(line); matches != nil {
				importPath := matches[1]
				if importPath == "" {
					importPath = matches[2]
				}
				symbols = append(symbols, Symbol{
					Name:      importPath,
					Type:      "import",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  false,
				})
			}
		}
	}

	return symbols
}

// extractJSSymbols 提取 JS/TS 代码符号
func extractJSSymbols(content, symbolType string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// 函数声明: function name, const name =, export function
	funcRegex := regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`)
	// 箭头函数: const name =, let name =
	arrowFuncRegex := regexp.MustCompile(`(?:export\s+)?(?:const|let)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>`)
	// class
	classRegex := regexp.MustCompile(`(?:export\s+)?class\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`)
	// interface (TypeScript)
	interfaceRegex := regexp.MustCompile(`(?:export\s+)?interface\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`)
	// import
	importRegex := regexp.MustCompile(`import\s+.*?from\s+['"]([^'"]+)['"]`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if symbolType == "all" || symbolType == "function" {
			if matches := funcRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "function",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  strings.Contains(line, "export"),
				})
				continue
			}
			if matches := arrowFuncRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "function",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  strings.Contains(line, "export"),
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "class" {
			if matches := classRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "class",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  strings.Contains(line, "export"),
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "interface" {
			if matches := interfaceRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "interface",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  strings.Contains(line, "export"),
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "import" {
			if matches := importRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "import",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  false,
				})
			}
		}
	}

	return symbols
}

// extractPythonSymbols 提取 Python 代码符号
func extractPythonSymbols(content, symbolType string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// def name
	funcRegex := regexp.MustCompile(`^def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	// async def name
	asyncFuncRegex := regexp.MustCompile(`^async\s+def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	// class name
	classRegex := regexp.MustCompile(`^class\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	// import
	importRegex := regexp.MustCompile(`^(?:from\s+\S+\s+)?import\s+(.+)`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if symbolType == "all" || symbolType == "function" {
			if matches := asyncFuncRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "function",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  !strings.HasPrefix(matches[1], "_"),
				})
				continue
			}
			if matches := funcRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "function",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  !strings.HasPrefix(matches[1], "_"),
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "class" {
			if matches := classRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "class",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  !strings.HasPrefix(matches[1], "_"),
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "import" {
			if matches := importRegex.FindStringSubmatch(line); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "import",
					Line:      i + 1,
					Signature: truncateLine(line, 100),
					Exported:  false,
				})
			}
		}
	}

	return symbols
}

// extractGenericSymbols 通用符号提取（基于缩进和关键字）
func extractGenericSymbols(content, symbolType string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// 通用函数模式
	funcRegex := regexp.MustCompile(`(?:function|def|fn|func|public|private|protected)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*[\(\{]`)
	// 通用类模式
	classRegex := regexp.MustCompile(`(?:class|struct|interface|type)\s+([a-zA-Z_][a-zA-Z0-9_]*)`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if symbolType == "all" || symbolType == "function" {
			if matches := funcRegex.FindStringSubmatch(trimmed); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "function",
					Line:      i + 1,
					Signature: truncateLine(trimmed, 100),
					Exported:  true,
				})
				continue
			}
		}

		if symbolType == "all" || symbolType == "class" {
			if matches := classRegex.FindStringSubmatch(trimmed); matches != nil {
				symbols = append(symbols, Symbol{
					Name:      matches[1],
					Type:      "class",
					Line:      i + 1,
					Signature: truncateLine(trimmed, 100),
					Exported:  true,
				})
			}
		}
	}

	return symbols
}
