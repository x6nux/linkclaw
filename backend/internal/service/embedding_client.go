package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/linkclaw/backend/internal/llm"
)

// EmbeddingClient 封装 Embedding API 调用
type EmbeddingClient struct {
	llmRouter  *llm.Router
	httpClient *http.Client
}

// NewEmbeddingClient 创建 Embedding 客户端
func NewEmbeddingClient(llmRouter *llm.Router) *EmbeddingClient {
	return &EmbeddingClient{
		llmRouter:  llmRouter,
		httpClient: &http.Client{},
	}
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// Generate 为文本生成 embedding 向量
func (c *EmbeddingClient) Generate(ctx context.Context, baseURL, model, apiKey, text string) ([]float32, error) {
	body, _ := json.Marshal(embeddingRequest{
		Model: model,
		Input: text,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(baseURL, "/")+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API %d: %s", resp.StatusCode, string(respBody))
	}

	// 读取原始响应体以便在解码失败时输出详细错误信息
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var result embeddingResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, fmt.Errorf("decode embedding response (raw: %s): %w", string(rawBody), err)
	}
	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	return result.Data[0].Embedding, nil
}
