package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type IndexingService struct {
	codeIndexRepo repository.CodeIndexRepo
	companyRepo   repository.CompanyRepo
	embeddingCli  *EmbeddingClient
	qdrantCli     *QdrantClient
	chunker       *CodeChunker
}

const (
	CollectionCodeChunks = "code_chunks"
	VectorSize           = 1536 // OpenAI text-embedding-3-small
)

var (
	ErrIndexTaskNotFound     = errors.New("index task not found")
	ErrIndexTaskAccessDenied = errors.New("index task access denied")
)

func NewIndexingService(
	codeIndexRepo repository.CodeIndexRepo,
	companyRepo repository.CompanyRepo,
	embeddingCli *EmbeddingClient,
	qdrantCfg QdrantConfig,
	chunkSize, overlap int,
) (*IndexingService, error) {
	qdrantCli, err := NewQdrantClient(qdrantCfg)
	if err != nil {
		return nil, err
	}
	return &IndexingService{
		codeIndexRepo: codeIndexRepo,
		companyRepo:   companyRepo,
		embeddingCli:  embeddingCli,
		qdrantCli:     qdrantCli,
		chunker:       NewCodeChunker(chunkSize, overlap),
	}, nil
}

// IndexRepository 索引仓库
func (s *IndexingService) IndexRepository(ctx context.Context, companyID, repoURL, branch string) (*domain.IndexTask, error) {
	if branch == "" {
		branch = "main"
	}

	task := &domain.IndexTask{
		ID:            uuid.New().String(),
		CompanyID:     companyID,
		RepositoryURL: repoURL,
		Branch:        branch,
		Status:        domain.IndexStatusPending,
	}
	if err := s.codeIndexRepo.CreateIndexTask(ctx, task); err != nil {
		return nil, fmt.Errorf("create index task: %w", err)
	}

	// 传递 task ID 而非指针，避免竞态条件
	go s.runIndex(context.Background(), task.ID)
	return task, nil
}

// SearchCode 搜索代码
func (s *IndexingService) SearchCode(ctx context.Context, companyID, query string, limit int) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	if err := s.ensureCollection(ctx); err != nil {
		return nil, err
	}

	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("get company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	vector, err := s.embeddingCli.Generate(ctx, company.EmbeddingBaseURL, company.EmbeddingModel, company.EmbeddingApiKey, query)
	if err != nil {
		return nil, fmt.Errorf("generate embedding: %w", err)
	}

	results, err := s.qdrantCli.Search(ctx, CollectionCodeChunks, vector, limit, map[string]string{
		"company_id": companyID,
	})
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}

	out := make([]*SearchResult, 0, len(results))
	for i := range results {
		out = append(out, &results[i])
	}
	return out, nil
}

// ListIndexTasks 列出公司索引任务
func (s *IndexingService) ListIndexTasks(ctx context.Context, companyID string) ([]*domain.IndexTask, error) {
	return s.codeIndexRepo.ListIndexTasks(ctx, companyID)
}

// GetIndexStatus 获取索引状态
func (s *IndexingService) GetIndexStatus(ctx context.Context, taskID string) (*domain.IndexTask, error) {
	return s.codeIndexRepo.GetIndexTask(ctx, taskID)
}

// GrantTaskAccess 授予 Agent 对索引任务的读取权限
func (s *IndexingService) GrantTaskAccess(ctx context.Context, companyID, taskID, agentID string) error {
	task, err := s.getTaskForCompany(ctx, taskID, companyID)
	if err != nil {
		return err
	}

	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return fmt.Errorf("agent id is required")
	}

	if task.CreatedBy != nil && *task.CreatedBy == agentID {
		return nil
	}

	existing, err := s.codeIndexRepo.GetIndexTaskAgent(ctx, task.ID, agentID, companyID)
	if err != nil {
		return fmt.Errorf("get index task agent: %w", err)
	}
	if existing != nil {
		return nil
	}

	return s.codeIndexRepo.CreateIndexTaskAgent(ctx, &domain.IndexTaskAgent{
		IndexTaskID: task.ID,
		AgentID:     agentID,
		CompanyID:   companyID,
	})
}

// RevokeTaskAccess 撤销 Agent 对索引任务的读取权限
func (s *IndexingService) RevokeTaskAccess(ctx context.Context, companyID, taskID, agentID string) error {
	task, err := s.getTaskForCompany(ctx, taskID, companyID)
	if err != nil {
		return err
	}

	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return fmt.Errorf("agent id is required")
	}

	// 创建者始终保留访问权限
	if task.CreatedBy != nil && *task.CreatedBy == agentID {
		return nil
	}

	if err := s.codeIndexRepo.DeleteIndexTaskAgent(ctx, task.ID, agentID, companyID); err != nil {
		return fmt.Errorf("delete index task agent: %w", err)
	}
	return nil
}

// SearchTaskCode 在单个任务授权范围内执行代码搜索
func (s *IndexingService) SearchTaskCode(ctx context.Context, companyID, taskID, agentID, query string, limit int) ([]*SearchResult, error) {
	task, err := s.getTaskForCompany(ctx, taskID, companyID)
	if err != nil {
		return nil, err
	}

	authorized, err := s.isTaskReadableByAgent(ctx, task, agentID)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrIndexTaskAccessDenied
	}

	return s.SearchCode(ctx, companyID, query, limit)
}

// ListTaskAgents 列出有索引任务读取权限的 Agent
func (s *IndexingService) ListTaskAgents(ctx context.Context, companyID, taskID string) ([]*domain.IndexTaskAgent, error) {
	task, err := s.getTaskForCompany(ctx, taskID, companyID)
	if err != nil {
		return nil, err
	}

	agents, err := s.codeIndexRepo.ListIndexTaskAgents(ctx, task.ID, companyID)
	if err != nil {
		return nil, fmt.Errorf("list index task agents: %w", err)
	}
	task.AuthorizedAgents = agents
	return agents, nil
}

func (s *IndexingService) getTaskForCompany(ctx context.Context, taskID, companyID string) (*domain.IndexTask, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("task id is required")
	}

	task, err := s.codeIndexRepo.GetIndexTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("get index task: %w", err)
	}
	if task == nil || task.CompanyID != companyID {
		return nil, ErrIndexTaskNotFound
	}
	return task, nil
}

func (s *IndexingService) isTaskReadableByAgent(ctx context.Context, task *domain.IndexTask, agentID string) (bool, error) {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return false, nil
	}

	if task.CreatedBy != nil && *task.CreatedBy == agentID {
		return true, nil
	}

	for _, authorized := range task.AuthorizedAgents {
		if authorized != nil && authorized.AgentID == agentID {
			return true, nil
		}
	}

	granted, err := s.codeIndexRepo.GetIndexTaskAgent(ctx, task.ID, agentID, task.CompanyID)
	if err != nil {
		return false, fmt.Errorf("get index task agent: %w", err)
	}
	return granted != nil, nil
}

func (s *IndexingService) ensureCollection(ctx context.Context) error {
	if err := s.qdrantCli.CreateCollection(ctx, CollectionCodeChunks, VectorSize); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return nil
		}
		return fmt.Errorf("create qdrant collection: %w", err)
	}
	return nil
}

// runIndex 执行索引（并发安全：使用 task ID 而非指针）
func (s *IndexingService) runIndex(ctx context.Context, taskID string) {
	// 从数据库重新加载 task，避免竞态条件
	task, err := s.codeIndexRepo.GetIndexTask(ctx, taskID)
	if err != nil || task == nil {
		return
	}

	// 获取公司配置
	company, err := s.companyRepo.GetByID(ctx, task.CompanyID)
	if err != nil || company == nil {
		task.Status = domain.IndexStatusFailed
		task.ErrorMessage = err.Error()
		completed := time.Now()
		task.CompletedAt = &completed
		_ = s.codeIndexRepo.UpdateIndexTask(ctx, task)
		return
	}

	task.Status = domain.IndexStatusRunning
	now := time.Now()
	task.StartedAt = &now
	if err := s.codeIndexRepo.UpdateIndexTask(ctx, task); err != nil {
		return
	}

	if err := s.ensureCollection(ctx); err != nil {
		task.Status = domain.IndexStatusFailed
		task.ErrorMessage = err.Error()
		completed := time.Now()
		task.CompletedAt = &completed
		_ = s.codeIndexRepo.UpdateIndexTask(ctx, task)
		return
	}

	markFailed := func(err error) {
		task.Status = domain.IndexStatusFailed
		task.ErrorMessage = err.Error()
		completed := time.Now()
		task.CompletedAt = &completed
		_ = s.codeIndexRepo.UpdateIndexTask(ctx, task)
	}

	if s.embeddingCli == nil || s.embeddingCli.httpClient == nil {
		markFailed(fmt.Errorf("embedding client not configured"))
		return
	}

	repoDir, err := os.MkdirTemp("", "linkclaw-index-*")
	if err != nil {
		markFailed(fmt.Errorf("create temp dir: %w", err))
		return
	}
	defer os.RemoveAll(repoDir)

	cloneArgs := []string{"clone", "--depth", "1"}
	if task.Branch != "" {
		cloneArgs = append(cloneArgs, "--branch", task.Branch)
	}
	cloneArgs = append(cloneArgs, task.RepositoryURL, repoDir)
	if out, err := exec.CommandContext(ctx, "git", cloneArgs...).CombinedOutput(); err != nil {
		markFailed(fmt.Errorf("git clone failed: %w: %s", err, strings.TrimSpace(string(out))))
		return
	}

	allowedExt := map[string]struct{}{
		".go": {}, ".js": {}, ".jsx": {}, ".ts": {}, ".tsx": {},
		".py": {}, ".java": {}, ".rs": {}, ".md": {}, ".json": {},
		".yaml": {}, ".yml": {}, ".rb": {}, ".php": {}, ".cs": {},
		".cpp": {}, ".cc": {}, ".c": {}, ".h": {}, ".hpp": {},
		".swift": {}, ".kt": {}, ".kts": {}, ".scala": {}, ".sql": {},
	}
	skipDirs := map[string]struct{}{
		".git": {}, "node_modules": {}, "vendor": {}, "dist": {},
		"build": {}, "target": {}, ".next": {}, "__pycache__": {},
	}

	files := make([]string, 0, 256)
	err = filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if _, skip := skipDirs[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		// 跳过符号链接，防止符号链接攻击
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, ok := allowedExt[ext]; !ok {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() > 1024*1024 {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		markFailed(fmt.Errorf("walk repository: %w", err))
		return
	}

	task.TotalFiles = len(files)
	task.IndexedFiles = 0
	if err := s.codeIndexRepo.UpdateIndexTask(ctx, task); err != nil {
		return
	}

	embeddingModel := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}
	envBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")), "/")

	type embeddingResponse struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	generateEmbedding := func(text string) ([]float32, error) {
		baseURL := company.EmbeddingBaseURL
		if envBaseURL != "" {
			baseURL = envBaseURL
		}
		if baseURL == "" {
			return nil, fmt.Errorf("empty embedding base URL")
		}

		reqBody, err := json.Marshal(map[string]string{
			"model": company.EmbeddingModel,
			"input": text,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal embedding request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/embeddings", bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("create embedding request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+company.EmbeddingApiKey)

		resp, err := s.embeddingCli.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("embedding request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("embedding API %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		}

		var out embeddingResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, fmt.Errorf("decode embedding response: %w", err)
		}
		if len(out.Data) == 0 || len(out.Data[0].Embedding) == 0 {
			return nil, fmt.Errorf("empty embedding response")
		}

		return out.Data[0].Embedding, nil
	}

	const flushSize = 64
	points := make([]Point, 0, flushSize)
	flushPoints := func() error {
		if len(points) == 0 {
			return nil
		}
		if err := s.qdrantCli.UpsertPoints(ctx, CollectionCodeChunks, points); err != nil {
			return fmt.Errorf("upsert vectors: %w", err)
		}
		points = points[:0]
		return nil
	}

	for _, absPath := range files {
		relPath, err := filepath.Rel(repoDir, absPath)
		if err != nil {
			markFailed(fmt.Errorf("get relative path %s: %w", absPath, err))
			return
		}
		relPath = filepath.ToSlash(relPath)

		contentBytes, err := os.ReadFile(absPath)
		if err != nil {
			markFailed(fmt.Errorf("read file %s: %w", relPath, err))
			return
		}

		fileChunk := s.chunker.ChunkFile(relPath, string(contentBytes))
		for idx, chunk := range fileChunk.Chunks {
			chunkText := strings.TrimSpace(chunk.Content)
			if chunkText == "" {
				continue
			}

			vector, err := generateEmbedding(chunkText)
			if err != nil {
				markFailed(fmt.Errorf("generate embedding %s#%d: %w", relPath, idx, err))
				return
			}
			if len(vector) != VectorSize {
				markFailed(fmt.Errorf("embedding dimension mismatch %s#%d: got %d want %d", relPath, idx, len(vector), VectorSize))
				return
			}

			chunkID := uuid.New().String()
			if err := s.codeIndexRepo.CreateChunk(ctx, &domain.CodeChunk{
				ID:          uuid.New().String(),
				CompanyID:   task.CompanyID,
				FilePath:    relPath,
				ChunkIndex:  idx,
				Content:     chunk.Content,
				StartLine:   chunk.StartLine,
				EndLine:     chunk.EndLine,
				Language:    fileChunk.Language,
				Symbols:     strings.Join(chunk.Symbols, ","),
				EmbeddingID: chunkID,
			}); err != nil {
				markFailed(fmt.Errorf("save chunk %s#%d: %w", relPath, idx, err))
				return
			}

			points = append(points, Point{
				ID:     chunkID,
				Vector: vector,
				Payload: map[string]interface{}{
					"company_id":  task.CompanyID,
					"file_path":   relPath,
					"chunk_index": idx,
					"content":     chunk.Content,
					"start_line":  chunk.StartLine,
					"end_line":    chunk.EndLine,
					"language":    fileChunk.Language,
					"symbols":     strings.Join(chunk.Symbols, ","),
				},
			})
			if len(points) >= flushSize {
				if err := flushPoints(); err != nil {
					markFailed(err)
					return
				}
			}
		}

		task.IndexedFiles++
		if err := s.codeIndexRepo.UpdateIndexTask(ctx, task); err != nil {
			return
		}
	}

	if err := flushPoints(); err != nil {
		markFailed(err)
		return
	}

	task.Status = domain.IndexStatusCompleted
	completed := time.Now()
	task.CompletedAt = &completed
	task.ErrorMessage = ""
	_ = s.codeIndexRepo.UpdateIndexTask(ctx, task)
}
