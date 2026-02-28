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
	companyRepo   repository.CompanyRepo
	embeddingCli  *EmbeddingClient
}

// NewEmbeddingWorker 创建 worker
func NewEmbeddingWorker(memoryRepo repository.MemoryRepo, companyRepo repository.CompanyRepo, embeddingCli *EmbeddingClient) *EmbeddingWorker {
	return &EmbeddingWorker{
		memoryRepo:   memoryRepo,
		companyRepo:  companyRepo,
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
		company, err := w.companyRepo.GetByID(ctx, m.CompanyID)
		if err != nil {
			log.Printf("embedding worker get company error (id=%s): %v", m.ID, err)
			continue
		}
		if company == nil {
			log.Printf("embedding worker company not found (id=%s): %v", m.ID, err)
			continue
		}

		vec, err := w.embeddingCli.Generate(ctx, company.EmbeddingBaseURL, company.EmbeddingModel, company.EmbeddingApiKey, m.Content)
		if err != nil {
			log.Printf("embedding worker generate error (id=%s): %v", m.ID, err)
			continue
		}
		if err := w.memoryRepo.UpdateEmbedding(ctx, m.ID, vec); err != nil {
			log.Printf("embedding worker update error (id=%s): %v", m.ID, err)
		}
	}
}
