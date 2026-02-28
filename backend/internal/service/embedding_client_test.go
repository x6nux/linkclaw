package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/linkclaw/backend/internal/llm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库连接
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	dsn := fmt.Sprintf(
		"host=localhost port=5432 user=linkclaw password=linkclaw_dev_pass dbname=linkclaw sslmode=disable",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("无法连接测试数据库 (dev DB): %v", dsn)
		return nil, nil
	}

	return db, func() {
		// 清理函数（当前无需特殊清理）
	}
}

// mockLLMProvider 创建用于测试的 LLM Provider
func mockLLMProvider(baseURL string, apiKey string) *llm.Provider {
	return &llm.Provider{
		ID:        uuid.New().String(),
		CompanyID: uuid.New().String(),
		Name:      "Test Embedding Provider",
		Type:      llm.ProviderOpenAI, // 使用 openai 类型绕过数据库约束
		BaseURL:   baseURL,
		APIKeyEnc: apiKey,
		Models:    []string{"text-embedding-3-small"},
		Weight:    100,
		IsActive:  true,
	}
}

func TestEmbeddingClient_Generate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := llm.NewRepository(db)
	router := llm.NewRouter(repo, "test_enc_key")
	client := NewEmbeddingClient(router)

	// =========================
	// 测试用例 1: 正常情况测试
	// =========================
	t.Run("正常响应", func(t *testing.T) {
		expectedEmbedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

		// 创建 Mock 服务器
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("读取请求体失败：%v", err)
			}
			t.Logf("请求体：%s", string(body))

			var req embeddingRequest
			if err := json.Unmarshal(body, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
				return
			}

			// 返回有效响应
			response := embeddingResponse{
				Data: []struct {
					Embedding []float32 `json:"embedding"`
				}{
					{Embedding: expectedEmbedding},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		// 创建 Provider
		provider := mockLLMProvider(mockServer.URL, "test-api-key")
		if err := repo.CreateProvider(context.Background(), provider); err != nil {
			t.Fatalf("创建 Provider 失败：%v", err)
		}
		defer repo.DeleteProvider(context.Background(), provider.ID)

		// 调用 Generate
		embedding, err := client.Generate(context.Background(), mockServer.URL, "text-embedding-3-small", "test-api-key", "Hello, World!")

		if err != nil {
			t.Errorf("Generate 返回错误：%v", err)
		}
		if len(embedding) != len(expectedEmbedding) {
			t.Errorf("向量长度不匹配：期望 %d, 得到 %d", len(expectedEmbedding), len(embedding))
		}
		for i, v := range expectedEmbedding {
			if embedding[i] != v {
				t.Errorf("向量 [%d] 不匹配：期望 %f, 得到 %f", i, v, embedding[i])
			}
		}
		t.Logf("成功生成 embedding: %v", embedding)
	})

	// =========================
	// 测试用例 2: 空响应测试
	// =========================
	t.Run("空响应", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 返回空 data 数组
			response := embeddingResponse{
				Data: []struct {
					Embedding []float32 `json:"embedding"`
				}{},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer mockServer.Close()

		_, err := client.Generate(context.Background(), mockServer.URL, "text-embedding-3-small", "test-api-key", "test")

		if err == nil {
			t.Error("期望返回错误，但得到 nil")
		}
		t.Logf("空响应错误 (符合预期): %v", err)
	})

	// =========================
	// 测试用例 3: 错误状态码测试
	// =========================
	t.Run("错误状态码", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid_api_key",
			})
		}))
		defer mockServer.Close()

		_, err := client.Generate(context.Background(), mockServer.URL, "text-embedding-3-small", "wrong-key", "test")

		if err == nil {
			t.Error("期望返回错误，但得到 nil")
		}
		t.Logf("错误状态码错误 (符合预期): %v", err)
	})

	// =========================
	// 测试用例 4: 无效 JSON 响应测试
	// =========================
	t.Run("无效 JSON 响应", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// 返回格式错误的 JSON（不完整的 JSON 数据）
			w.Write([]byte("{\"data\": [{\"embedding\": "))
		}))
		defer mockServer.Close()

		_, err := client.Generate(context.Background(), mockServer.URL, "text-embedding-3-small", "test-api-key", "test")

		if err == nil {
			t.Error("期望返回错误，但得到 nil")
		}
		t.Logf("无效 JSON 错误 (符合预期): %v", err)
	})
}

// TestReadEnvironment 测试从环境读取 dev 配置
func TestReadEnvironment(t *testing.T) {
	envVars := map[string]string{
		"POSTGRES_PASSWORD": "linkclaw_dev_pass",
		"LLM_ENCRYPT_KEY":   "9b8723613d92626bde52252fb028c656ac1e9e0821771859967d8fb8978424b1",
	}

	for key, expected := range envVars {
		actual := os.Getenv(key)
		if actual != expected {
			t.Logf("环境变量 %s: 期望 '%s', 实际 '%s'", key, expected, actual)
		} else {
			t.Logf("环境变量 %s: 正确配置", key)
		}
	}
}
