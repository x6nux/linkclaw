package service

import (
	"context"
	"log"
	"time"

	"github.com/linkclaw/backend/internal/repository"
)

const (
	workerInterval = 5 * time.Second
	workerBatch    = 10
)

// EmbeddingWorker 后台异步生成 embedding
type EmbeddingWorker struct {
	memoryRepo    repository.MemoryRepo
	embeddingCli  *EmbeddingClient
}

// NewEmbeddingWorker 创建 worker
func NewEmbeddingWorker(memoryRepo repository.MemoryRepo, embeddingCli *EmbeddingClient) *EmbeddingWorker {
	return &EmbeddingWorker{
		memoryRepo:   memoryRepo,
		embeddingCli: embeddingCli,
	}
}

// Start 启动 worker 循环（在 goroutine 中调用）
func (w *EmbeddingWorker) Start(ctx context.Context) {
	log.Println("embedding worker started")
	ticker := time.NewTicker(workerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("embedding worker stopped")
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *EmbeddingWorker) process(ctx context.Context) {
	mems, err := w.memoryRepo.ListPendingEmbedding(ctx, workerBatch)
	if err != nil {
		log.Printf("embedding worker list error: %v", err)
		return
	}
	for _, m := range mems {
		vec, err := w.embeddingCli.Generate(ctx, m.CompanyID, m.Content)
		if err != nil {
			log.Printf("embedding worker generate error (id=%s): %v", m.ID, err)
			continue
		}
		if err := w.memoryRepo.UpdateEmbedding(ctx, m.ID, vec); err != nil {
			log.Printf("embedding worker update error (id=%s): %v", m.ID, err)
		}
	}
}
