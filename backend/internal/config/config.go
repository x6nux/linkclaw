package config

import (
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
	ResetSecret string // 管理员密码重置密钥，从 RESET_SECRET 读取
}

// AgentConfig 跨公司通信配置
type AgentConfig struct {
	PartnerAPIKey string // 接受跨公司消息的密钥
	CompanySlug   string // 本实例公司 slug
	MCPPublicURL  string // 对外暴露的 MCP 端点（跨公司返回）
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
			MCPPublicURL:  getEnv("MCP_PUBLIC_URL", ""),
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
