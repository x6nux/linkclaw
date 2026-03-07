package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	LLM         LLMConfig
	Agent       AgentConfig
	Context     ContextConfig      // 上下文搜索配置
	ResetSecret string // 管理员密码重置密钥，从 RESET_SECRET 读取
}

// ContextConfig 上下文搜索配置
type ContextConfig struct {
	EnableIndexing        bool    // 是否启用索引功能
	IndexThreshold        int     // 文件数超过此值启用索引 (默认 100)
	EnableIndexFallback   bool    // 索引失败时降级到全量扫描
	EnableLLMFallback     bool    // LLM 失败时降级到关键词匹配
	MaxConcurrentSearches int     // 最大并发搜索数 (默认 10)
	RateLimitPerAgent     int     // 每个 Agent 每秒最大请求数 (默认 5)
	AgentSearchEnabled    bool    // 是否启用 Agent 搜索
	AgentSearchRatio      float64 // Agent 搜索灰度比例 (0.0-1.0)
	SearchTimeoutMs       int     // 普通搜索超时 (默认 30000ms)
	AgentSearchTimeoutMs  int     // Agent 搜索超时 (默认 60000ms)
	MaxSearchTimeoutMs    int     // 最大允许超时 (默认 120000ms)
}

// AgentConfig 跨公司通信配置
// 注意：CompanySlug 用于启动时识别当前实例代表的公司， PartnerAPIKey 已废弃，改用数据库中的配对密钥
type AgentConfig struct {
	PartnerAPIKey string // [已废弃] 使用数据库中的 partner_api_keys 表
	CompanySlug   string // 本实例公司 slug（用于启动时识别）
}

type LLMConfig struct {
	EncryptKey string // 32 字节，从 LLM_ENCRYPT_KEY 读取
}

type ServerConfig struct {
	Port string
	Mode string // debug, release
}

type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	Expiry int // hours
}

// Validate 验证必要的环境变量已设置
func (c *Config) Validate() error {
	// 生产环境下关键配置不能使用默认值
	if c.Server.Mode == "release" {
		if c.JWT.Secret == "changeme-in-production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
		if c.Database.DSN == "" || c.Database.DSN == "postgres://postgres:postgres@localhost:5432/linkclaw?sslmode=disable" {
			return fmt.Errorf("DATABASE_URL must be set in production")
		}
	}

	// 检查 LLM_ENCRYPT_KEY 格式（应为 32 字节 = 64 十六进制字符）
	if c.LLM.EncryptKey != "" && len(c.LLM.EncryptKey) != 64 {
		return fmt.Errorf("LLM_ENCRYPT_KEY must be 64 characters (32 bytes hex), got %d", len(c.LLM.EncryptKey))
	}

	return nil
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			DSN:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/linkclaw?sslmode=disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "changeme-in-production"),
			Expiry: getEnvInt("JWT_EXPIRY_HOURS", 24),
		},
		LLM: LLMConfig{
			EncryptKey: getEnv("LLM_ENCRYPT_KEY", ""),
		},
		Agent: AgentConfig{
			PartnerAPIKey: getEnv("PARTNER_API_KEY", ""),
			CompanySlug:   getEnv("COMPANY_SLUG", ""),
		},
		Context: ContextConfig{
			EnableIndexing:        getEnvBool("CONTEXT_ENABLE_INDEXING", true),
			IndexThreshold:        getEnvInt("CONTEXT_INDEX_THRESHOLD", 100),
			EnableIndexFallback:   getEnvBool("CONTEXT_ENABLE_INDEX_FALLBACK", true),
			EnableLLMFallback:     getEnvBool("CONTEXT_ENABLE_LLM_FALLBACK", true),
			MaxConcurrentSearches: getEnvInt("CONTEXT_MAX_CONCURRENT_SEARCHES", 10),
			RateLimitPerAgent:     getEnvInt("CONTEXT_RATE_LIMIT_PER_AGENT", 5),
			AgentSearchEnabled:    getEnvBool("CONTEXT_AGENT_SEARCH_ENABLED", true),
			AgentSearchRatio:      getEnvFloat("CONTEXT_AGENT_SEARCH_RATIO", 1.0),
			SearchTimeoutMs:       getEnvInt("CONTEXT_SEARCH_TIMEOUT_MS", 30000),
			AgentSearchTimeoutMs:  getEnvInt("CONTEXT_AGENT_SEARCH_TIMEOUT_MS", 60000),
			MaxSearchTimeoutMs:    getEnvInt("CONTEXT_MAX_SEARCH_TIMEOUT_MS", 120000),
		},
		ResetSecret: getEnv("RESET_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
