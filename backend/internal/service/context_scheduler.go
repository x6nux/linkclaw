package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

// ContextScheduler handles background indexing tasks for context directories
type ContextScheduler struct {
	repo        repository.ContextRepo
	llmCli      *ContextLLMClient
	scanChan    chan string // directory ID scan queue
	ticker      *time.Ticker
	mu          sync.RWMutex
	running     bool
	scanRunning map[string]bool // directory ID -> scanning status
}

// NewContextScheduler creates a new context scheduler
func NewContextScheduler(repo repository.ContextRepo, llmCli *ContextLLMClient) *ContextScheduler {
	return &ContextScheduler{
		repo:        repo,
		llmCli:      llmCli,
		scanChan:    make(chan string, 100),
		scanRunning: make(map[string]bool),
	}
}

// Start starts the background scheduler
func (s *ContextScheduler) Start(ctx context.Context) {
	// 每 5 分钟检查一次需要扫描的目录
	s.ticker = time.NewTicker(5 * time.Minute)
	s.running = true

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.running = false
				s.ticker.Stop()
				return
			case dirID := <-s.scanChan:
				s.processDirectory(ctx, dirID)
			case <-s.ticker.C:
				s.scheduleScan(ctx)
			}
		}
	}()

	log.Println("ContextScheduler started")
}

// Stop stops the scheduler
func (s *ContextScheduler) Stop() {
	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.scanChan)
}

// TriggerScan manually triggers a scan for a specific directory
func (s *ContextScheduler) TriggerScan(dirID string) error {
	if !s.running {
		return fmt.Errorf("scheduler not running")
	}
	select {
	case s.scanChan <- dirID:
		return nil
	default:
		return fmt.Errorf("scan queue full")
	}
}

// scheduleScan scans all active directories that need re-indexing
func (s *ContextScheduler) scheduleScan(ctx context.Context) {
	dirs, err := s.repo.ListAllActiveDirectories(ctx)
	if err != nil {
		log.Printf("Failed to list directories: %v", err)
		return
	}

	now := time.Now()
	for _, d := range dirs {
		// 如果从未扫描过或超过 1 小时未扫描，则加入扫描队列
		shouldScan := d.LastIndexedAt == nil || now.Sub(*d.LastIndexedAt) > time.Hour

		if shouldScan {
			select {
			case s.scanChan <- d.ID:
			default:
				log.Printf("Scan queue full, skipping directory %s", d.ID)
			}
		}
	}
}

// processDirectory processes a single directory: scans files, computes hashes, updates summaries
func (s *ContextScheduler) processDirectory(ctx context.Context, dirID string) {
	s.mu.Lock()
	if s.scanRunning[dirID] {
		s.mu.Unlock()
		log.Printf("Scan already running for directory %s", dirID)
		return
	}
	s.scanRunning[dirID] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.scanRunning, dirID)
		s.mu.Unlock()
	}()

	log.Printf("Starting scan for directory %s", dirID)

	// 获取目录信息
	dir, err := s.repo.GetDirectoryByID(ctx, dirID)
	if err != nil {
		log.Printf("Failed to get directory %s: %v", dirID, err)
		return
	}
	if dir == nil {
		log.Printf("Directory %s not found", dirID)
		return
	}

	// 1. 扫描文件系统
	files, err := s.scanFileSystem(dir)
	if err != nil {
		log.Printf("Failed to scan directory %s: %v", dirID, err)
		// 标记目录为失效
		dir.IsActive = false
		dir.UpdatedAt = time.Now()
		_ = s.repo.UpdateDirectory(ctx, dir)
		return
	}

	// 2. 获取现有文件总结
	existingSummaries, err := s.repo.ListFileSummaries(ctx, dirID)
	if err != nil {
		log.Printf("Failed to list summaries for directory %s: %v", dirID, err)
		return
	}

	// 构建 hash 映射用于增量更新
	existingHashes := make(map[string]*domain.ContextFileSummary)
	for _, summary := range existingSummaries {
		existingHashes[summary.FilePath] = summary
	}

	// 3. 增量处理文件
	var newCount, updatedCount, deletedCount int
	filePaths := make(map[string]bool)

	for _, file := range files {
		filePaths[file.FilePath] = true

		// 检查 hash 是否变化
		if existing, ok := existingHashes[file.FilePath]; ok {
			if existing.ContentHash == file.ContentHash {
				// Hash 未变化，跳过
				delete(existingHashes, file.FilePath)
				continue
			}
			// Hash 变化，更新总结
			if err := s.updateFileSummary(ctx, dirID, file); err != nil {
				log.Printf("Failed to update summary for %s: %v", file.FilePath, err)
				continue
			}
			updatedCount++
			delete(existingHashes, file.FilePath)
		} else {
			// 新文件，创建总结
			if err := s.createFileSummary(ctx, dirID, file); err != nil {
				log.Printf("Failed to create summary for %s: %v", file.FilePath, err)
				continue
			}
			newCount++
		}
	}

	// 4. 删除已不存在的文件
	for filePath := range existingHashes {
		if err := s.repo.DeleteFileSummary(ctx, dirID, filePath); err != nil {
			log.Printf("Failed to delete summary for %s: %v", filePath, err)
		}
		deletedCount++
	}

	// 5. 更新目录元数据
	now := time.Now()
	dir.LastIndexedAt = &now
	dir.FileCount = len(files)
	dir.UpdatedAt = now
	if err := s.repo.UpdateDirectory(ctx, dir); err != nil {
		log.Printf("Failed to update directory metadata: %v", err)
	}

	log.Printf("Directory %s scan complete: %d new, %d updated, %d deleted", dirID, newCount, updatedCount, deletedCount)
}

// IndexedFile represents a file with its content hash
type IndexedFile struct {
	FilePath    string
	Content     string
	ContentHash string
	Language    string
	LineCount   int
}

// scanFileSystem scans the directory and returns indexed files
func (s *ContextScheduler) scanFileSystem(dir *domain.ContextDirectory) ([]IndexedFile, error) {
	patterns := parsePatterns(dir.FilePatterns)
	excludes := parsePatterns(dir.ExcludePatterns)
	maxSize := int64(dir.MaxFileSize)

	var files []IndexedFile
	err := filepath.WalkDir(dir.Path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(dir.Path, path)
		if !matchAny(rel, patterns) {
			return nil
		}
		if len(excludes) > 0 && matchAny(rel, excludes) {
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

		contentStr := string(content)
		hash := computeHash(content)
		language := detectLanguage(rel)
		lines := strings.Count(contentStr, "\n")
		if contentStr != "" {
			lines++
		}

		files = append(files, IndexedFile{
			FilePath:    rel,
			Content:     contentStr,
			ContentHash: hash,
			Language:    language,
			LineCount:   lines,
		})
		return nil
	})
	return files, err
}

// computeHash computes SHA256 hash of content
func computeHash(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

// createFileSummary creates a summary for a new file
func (s *ContextScheduler) createFileSummary(ctx context.Context, dirID string, file IndexedFile) error {
	// 调用 LLM 生成总结
	summary, lineCount, err := s.llmCli.SummarizeFile(ctx, file.Content, file.Language)
	if err != nil {
		// LLM 失败时使用内容片段作为降级总结
		summary = generateFallbackSummary(file.Content, file.Language)
		lineCount = file.LineCount
	}

	now := time.Now()
	fs := &domain.ContextFileSummary{
		ID:           uuid.NewString(),
		DirectoryID:  dirID,
		FilePath:     file.FilePath,
		ContentHash:  file.ContentHash,
		Summary:      summary,
		Language:     file.Language,
		LineCount:    lineCount,
		SummarizedAt: now,
	}
	return s.repo.CreateFileSummary(ctx, fs)
}

// updateFileSummary updates summary for an existing file
func (s *ContextScheduler) updateFileSummary(ctx context.Context, dirID string, file IndexedFile) error {
	summary, lineCount, err := s.llmCli.SummarizeFile(ctx, file.Content, file.Language)
	if err != nil {
		summary = generateFallbackSummary(file.Content, file.Language)
		lineCount = file.LineCount
	}

	now := time.Now()
	fs := &domain.ContextFileSummary{
		ID:           uuid.NewString(),
		DirectoryID:  dirID,
		FilePath:     file.FilePath,
		ContentHash:  file.ContentHash,
		Summary:      summary,
		Language:     file.Language,
		LineCount:    lineCount,
		SummarizedAt: now,
	}
	return s.repo.CreateFileSummary(ctx, fs) // UPSERT semantics
}

// generateFallbackSummary generates a simple summary when LLM is unavailable
func generateFallbackSummary(content, language string) string {
	lines := strings.Count(content, "\n")
	if content != "" {
		lines++
	}
	return fmt.Sprintf("File: %s, Lines: %d", language, lines)
}
