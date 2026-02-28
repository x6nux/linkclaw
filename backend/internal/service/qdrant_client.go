package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// validateQdrantURL 验证 Qdrant URL 安全性（SSRF 防护）
func validateQdrantURL(raw string) error {
	if raw == "" {
		return nil // 允许空值，使用默认 localhost
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid qdrant url: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("qdrant url must be http or https")
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("qdrant url missing host")
	}
	hostLower := strings.ToLower(host)

	// 禁止 localhost
	if hostLower == "localhost" || hostLower == "127.0.0.1" || hostLower == "[::1]" || hostLower == "::1" {
		return fmt.Errorf("qdrant url cannot point to localhost")
	}

	// 禁止私有 IP (10.0.0.0/8)
	if strings.HasPrefix(hostLower, "10.") {
		return fmt.Errorf("qdrant url cannot point to private IP (10.x.x.x)")
	}
	// 禁止私有 IP (172.16.0.0/12)
	if matched, _ := regexp.MatchString(`^172\.(1[6-9]|2[0-9]|3[01])\.`, hostLower); matched {
		return fmt.Errorf("qdrant url cannot point to private IP (172.16-31.x.x)")
	}
	// 禁止私有 IP (192.168.0.0/16)
	if strings.HasPrefix(hostLower, "192.168.") {
		return fmt.Errorf("qdrant url cannot point to private IP (192.168.x.x)")
	}
	// 禁止回环地址 (127.0.0.0/8)
	if strings.HasPrefix(hostLower, "127.") {
		return fmt.Errorf("qdrant url cannot point to loopback address (127.x.x.x)")
	}
	// 禁止链路本地地址 (169.254.0.0/16)
	if strings.HasPrefix(hostLower, "169.254.") {
		return fmt.Errorf("qdrant url cannot point to link-local address (169.254.x.x)")
	}

	return nil
}

// QdrantClient Qdrant 向量数据库客户端
type QdrantClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type QdrantConfig struct {
	BaseURL string
	APIKey  string
}

func NewQdrantClient(cfg QdrantConfig) (*QdrantClient, error) {
	// 验证 URL 安全性
	if err := validateQdrantURL(cfg.BaseURL); err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:6333"
	}
	return &QdrantClient{
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Point 向量点
type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// SearchResult 搜索结果
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// CreateCollection 创建集合
func (c *QdrantClient) CreateCollection(ctx context.Context, name string, vectorSize int) error {
	payload := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     vectorSize,
			"distance": "Cosine",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		c.baseURL+"/collections/"+name, bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create collection: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpsertPoints 批量插入/更新向量点
func (c *QdrantClient) UpsertPoints(ctx context.Context, collection string, points []Point) error {
	payload := map[string]interface{}{"points": points}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		c.baseURL+"/collections/"+collection+"/points", bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert points: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Search 搜索相似向量
func (c *QdrantClient) Search(ctx context.Context, collection string, vector []float32, limit int, filter map[string]string) ([]SearchResult, error) {
	reqPayload := map[string]interface{}{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
	}

	if filter != nil && filter["company_id"] != "" {
		reqPayload["filter"] = map[string]interface{}{
			"must": []map[string]interface{}{
				{"key": "company_id", "match": map[string]interface{}{"value": filter["company_id"]}},
			},
		}
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/collections/"+collection+"/points/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search: %d - %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Result []SearchResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Result, nil
}

// DeletePoints 删除向量点
func (c *QdrantClient) DeletePoints(ctx context.Context, collection string, ids []string) error {
	payload := map[string]interface{}{"points": ids}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/collections/"+collection+"/points/delete", bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete points: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *QdrantClient) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("api-key", c.apiKey)
	}
	return c.httpClient.Do(req)
}
