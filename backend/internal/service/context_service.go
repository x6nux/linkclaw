package service

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type ContextService struct {
	repo   repository.ContextRepo
	llmCli *ContextLLMClient
}

func NewContextService(repo repository.ContextRepo, llmCli *ContextLLMClient) *ContextService {
	return &ContextService{repo: repo, llmCli: llmCli}
}

// GetLLMMetrics 返回 LLM 客户端的指标收集器
func (s *ContextService) GetLLMMetrics() *ContextMetrics {
	if s.llmCli == nil {
		return nil
	}
	return s.llmCli.GetMetrics()
}

// GetLLMClient 返回 LLM 客户端（用于 Agent 工具执行）
func (s *ContextService) GetLLMClient() *ContextLLMClient {
	return s.llmCli
}

// GetRepo 返回仓库（用于 Agent 工具执行）
func (s *ContextService) GetRepo() repository.ContextRepo {
	return s.repo
}

// CreateContextSearchAgent 创建上下文搜索 Agent 实例
func (s *ContextService) CreateContextSearchAgent() *ContextSearchAgent {
	return NewContextSearchAgent(s.llmCli, s.repo)
}

func (s *ContextService) ListDirectories(ctx context.Context, companyID string) ([]*domain.ContextDirectory, error) {
	return s.repo.ListDirectories(ctx, companyID)
}

func (s *ContextService) GetDirectoryByID(ctx context.Context, id string) (*domain.ContextDirectory, error) {
	return s.repo.GetDirectoryByID(ctx, id)
}

type CreateDirectoryInput struct {
	CompanyID       string
	Name            string
	Path            string
	Description     string
	FilePatterns    string
	ExcludePatterns string
	MaxFileSize     int
}

func (s *ContextService) CreateDirectory(ctx context.Context, in CreateDirectoryInput) (*domain.ContextDirectory, error) {
	if in.MaxFileSize == 0 {
		in.MaxFileSize = 1024 * 1024 // 1MB default
	}
	now := time.Now()
	d := &domain.ContextDirectory{
		ID:              uuid.NewString(),
		CompanyID:       in.CompanyID,
		Name:            in.Name,
		Path:            in.Path,
		Description:     in.Description,
		IsActive:        true,
		FilePatterns:    in.FilePatterns,
		ExcludePatterns: in.ExcludePatterns,
		MaxFileSize:     in.MaxFileSize,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateDirectory(ctx, d); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}
	return d, nil
}

type UpdateDirectoryInput struct {
	Name            *string
	Path            *string
	Description     *string
	FilePatterns    *string
	ExcludePatterns *string
	MaxFileSize     *int
}

func (s *ContextService) UpdateDirectory(ctx context.Context, id string, in UpdateDirectoryInput) (*domain.ContextDirectory, error) {
	d, err := s.repo.GetDirectoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, nil
	}
	if in.Name != nil {
		d.Name = *in.Name
	}
	if in.Path != nil {
		d.Path = *in.Path
	}
	if in.Description != nil {
		d.Description = *in.Description
	}
	if in.FilePatterns != nil {
		d.FilePatterns = *in.FilePatterns
	}
	if in.ExcludePatterns != nil {
		d.ExcludePatterns = *in.ExcludePatterns
	}
	if in.MaxFileSize != nil {
		d.MaxFileSize = *in.MaxFileSize
	}
	d.UpdatedAt = time.Now()
	if err := s.repo.UpdateDirectory(ctx, d); err != nil {
		return nil, fmt.Errorf("update directory: %w", err)
	}
	return d, nil
}

func (s *ContextService) DeleteDirectory(ctx context.Context, id string) error {
	return s.repo.DeleteDirectory(ctx, id)
}

func (s *ContextService) ToggleDirectory(ctx context.Context, id string, isActive bool) (*domain.ContextDirectory, error) {
	d, err := s.repo.GetDirectoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, nil
	}
	d.IsActive = isActive
	d.UpdatedAt = time.Now()
	if err := s.repo.UpdateDirectory(ctx, d); err != nil {
		return nil, fmt.Errorf("toggle directory: %w", err)
	}
	return d, nil
}

type SearchInput struct {
	CompanyID    string
	AgentID      string
	Query        string
	DirectoryIDs []string
	// --- 统一参数 (v1.1) ---
	MaxResults   int     // 最大返回数 (默认 10)
	MinRelevance float64 // 最低相关性阈值 (默认 0.3)
	TimeoutMs    int     // 超时时间 (默认 30000ms)
	UseIndex     *bool   // 是否使用索引 (默认 true)
}

// SearchOutput 搜索输出 (向后兼容，逐步迁移到 domain.ContextSearchResponse)
type SearchOutput struct {
	Results     []*domain.ContextSearchResult
	Diagnostics *domain.SearchDiagnostics
	Error       *domain.SearchError
	LatencyMs   int // 实际耗时 (毫秒)
}

func (s *ContextService) Search(ctx context.Context, in SearchInput) (*SearchOutput, error) {
	start := time.Now()

	// --- 超时控制 ---
	timeout := in.TimeoutMs
	if timeout <= 0 {
		timeout = 30000 // 默认 30s
	}
	if timeout > 120000 {
		timeout = 120000 // 最大 120s
	}
	searchCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// --- 加载目录 ---
	var dirs []*domain.ContextDirectory
	if len(in.DirectoryIDs) > 0 {
		for _, id := range in.DirectoryIDs {
			d, err := s.repo.GetDirectoryByID(searchCtx, id)
			if err != nil {
				return nil, err
			}
			if d != nil && d.CompanyID == in.CompanyID && d.IsActive {
				dirs = append(dirs, d)
			}
		}
	} else {
		var err error
		dirs, err = s.repo.ListActiveDirectories(searchCtx, in.CompanyID)
		if err != nil {
			return nil, err
		}
	}

	// 检查是否有可搜索的目录
	if len(dirs) == 0 {
		latency := int(time.Since(start).Milliseconds())
		return &SearchOutput{
			Results: []*domain.ContextSearchResult{},
			Error: &domain.SearchError{
				Code:    domain.ErrNoDirectories,
				Message: "没有可搜索的目录。请先配置上下文目录。",
				Details: map[string]any{"company_id": in.CompanyID},
			},
			LatencyMs: latency,
		}, nil
	}

	diagnostics := &domain.SearchDiagnostics{
		DirectoriesScanned: len(dirs),
		IndexUsed:          in.UseIndex == nil || *in.UseIndex,
	}

	// --- 索引优先搜索 ---
	useIndex := in.UseIndex == nil || *in.UseIndex
	var results []*domain.ContextSearchResult
	var err error

	if useIndex {
		results, err = s.searchFromIndex(searchCtx, in.Query, dirs)
		if err != nil {
			// 索引搜索失败，记录降级原因
			diagnostics.FallbackReason = fmt.Sprintf("索引搜索失败：%v", err)
			diagnostics.IndexUsed = false
			results = nil
		}
	}

	// --- 降级策略：索引结果不足或禁用索引时，使用全文搜索 ---
	if len(results) < 3 || !useIndex {
		if diagnostics.FallbackReason == "" && !useIndex {
			diagnostics.FallbackReason = "索引被禁用"
		}
		fileResults, fileErr := s.searchFromFileContent(searchCtx, in.Query, dirs)
		if fileErr != nil {
			if results == nil {
				return nil, fmt.Errorf("file content search: %w", fileErr)
			}
			// 已有索引结果，记录错误但不中断
			diagnostics.FallbackReason += fmt.Sprintf("; 全文搜索失败：%v", fileErr)
		} else {
			// 合并结果，避免重复
			if len(results) > 0 && len(fileResults) > 0 {
				existingPaths := make(map[string]bool)
				for _, r := range results {
					existingPaths[r.FilePath] = true
				}
				for _, r := range fileResults {
					if !existingPaths[r.FilePath] {
						results = append(results, r)
					}
				}
			} else if len(fileResults) > 0 {
				results = fileResults
			}
		}
		diagnostics.FilesAnalyzed = len(results)
	}

	// --- 过滤和限制结果 ---
	results = s.filterAndLimitResults(results, in.MaxResults, in.MinRelevance)

	// --- 记录搜索日志 ---
	dirIDs := make([]string, 0, len(dirs))
	for _, d := range dirs {
		dirIDs = append(dirIDs, d.ID)
	}
	latency := int(time.Since(start).Milliseconds())
	log := &domain.ContextSearchLog{
		ID:           uuid.NewString(),
		CompanyID:    in.CompanyID,
		AgentID:      in.AgentID,
		Query:        in.Query,
		DirectoryIDs: strings.Join(dirIDs, ","),
		ResultsCount: len(results),
		LatencyMs:    latency,
		CreatedAt:    time.Now(),
	}
	_ = s.repo.CreateSearchLog(searchCtx, log) // 日志失败不影响主流程

	return &SearchOutput{
		Results:     results,
		Diagnostics: diagnostics,
		Error:       nil,
		LatencyMs:   latency,
	}, nil
}

// filterAndLimitResults 过滤结果并限制数量
func (s *ContextService) filterAndLimitResults(results []*domain.ContextSearchResult, maxResults int, minRelevance float64) []*domain.ContextSearchResult {
	if minRelevance <= 0 {
		minRelevance = 0.3 // 默认阈值
	}
	maxR := maxResults
	if maxR <= 0 {
		maxR = 10 // 默认 10 条
	}

	filtered := make([]*domain.ContextSearchResult, 0, len(results))
	for _, r := range results {
		if r.Relevance >= minRelevance {
			filtered = append(filtered, r)
		}
	}

	// 限制数量
	if len(filtered) > maxR {
		filtered = filtered[:maxR]
	}

	return filtered
}

// searchFromIndex 从文件摘要索引中召回结果
func (s *ContextService) searchFromIndex(ctx context.Context, query string, dirs []*domain.ContextDirectory) ([]*domain.ContextSearchResult, error) {
	var allSummaries []*domain.ContextFileSummary
	dirMap := make(map[string]string) // filePath -> directoryID

	for _, d := range dirs {
		summaries, err := s.repo.ListFileSummaries(ctx, d.ID)
		if err != nil {
			continue
		}
		for _, summary := range summaries {
			allSummaries = append(allSummaries, summary)
			dirMap[summary.FilePath] = d.ID
		}
	}

	if len(allSummaries) == 0 {
		return nil, nil
	}

	// 将摘要转换为 FileContent 供 LLM 分析
	files := make([]FileContent, 0, len(allSummaries))
	for _, s := range allSummaries {
		files = append(files, FileContent{
			FilePath: s.FilePath,
			Content:  s.Summary,
			Language: s.Language,
		})
	}

	// 调用 LLM 对摘要进行语义搜索
	llmResults, err := s.llmCli.SemanticSearchFromSummaries(ctx, query, files)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.ContextSearchResult, 0, len(llmResults))
	summaryMap := make(map[string]*domain.ContextFileSummary)
	for _, s := range allSummaries {
		summaryMap[s.FilePath] = s
	}

	for _, r := range llmResults {
		summary := summaryMap[r.FilePath]
		results = append(results, &domain.ContextSearchResult{
			FilePath:    r.FilePath,
			Language:    r.Language,
			Summary:     summary.Summary,
			Relevance:   r.Relevance,
			Reason:      r.Reason,
			LineCount:   summary.LineCount,
			DirectoryID: dirMap[r.FilePath],
			// --- 新增字段 ---
			ContentHash: summary.ContentHash,
			IndexedAt:   summary.SummarizedAt.Format(time.RFC3339),
		})
	}

	sortByRelevanceForResults(results)
	return results, nil
}

// searchFromFileContent 读取文件内容进行全文搜索（降级方案）
func (s *ContextService) searchFromFileContent(ctx context.Context, query string, dirs []*domain.ContextDirectory) ([]*domain.ContextSearchResult, error) {
	var allFiles []FileContent
	dirMap := make(map[string]string)

	for _, d := range dirs {
		files, err := s.readDirectoryFiles(d)
		if err != nil {
			continue
		}
		for _, f := range files {
			allFiles = append(allFiles, f)
			dirMap[f.FilePath] = d.ID
		}
	}

	if len(allFiles) == 0 {
		return []*domain.ContextSearchResult{}, nil
	}

	llmResults, err := s.llmCli.SemanticSearch(ctx, query, allFiles, "")
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	results := make([]*domain.ContextSearchResult, 0, len(llmResults))
	for _, r := range llmResults {
		results = append(results, &domain.ContextSearchResult{
			FilePath:    r.FilePath,
			Language:    r.Language,
			Summary:     r.Summary,
			Relevance:   r.Relevance,
			Reason:      r.Reason,
			LineCount:   r.LineCount,
			DirectoryID: dirMap[r.FilePath],
		})
	}

	return results, nil
}

func (s *ContextService) readDirectoryFiles(d *domain.ContextDirectory) ([]FileContent, error) {
	// 使用安全配置扫描目录
	secConfig := DefaultSecurityConfig()

	patterns := parsePatterns(d.FilePatterns)
	excludes := parsePatterns(d.ExcludePatterns)
	maxSize := int64(d.MaxFileSize)

	var files []FileContent

	_, err := secConfig.ScanDirectory(d.Path, func(path string, entry fs.DirEntry, relPath string, depth int) error {
		if entry.IsDir() {
			return nil
		}

		// 检查文件模式
		if !matchAny(relPath, patterns) {
			return nil
		}
		// 检查排除模式
		if len(excludes) > 0 && matchAny(relPath, excludes) {
			return nil
		}

		info, err := entry.Info()
		if err != nil || (maxSize > 0 && info.Size() > maxSize) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		files = append(files, FileContent{
			FilePath: relPath,
			Content:  string(content),
			Language: detectLanguage(relPath),
		})
		return nil
	})

	return files, err
}

func parsePatterns(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func matchAny(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if matched, _ := filepath.Match(p, path); matched {
			return true
		}
		if matched, _ := filepath.Match(p, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".sql":
		return "sql"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return ""
	}
}

func sortByRelevanceForResults(results []*domain.ContextSearchResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})
}
