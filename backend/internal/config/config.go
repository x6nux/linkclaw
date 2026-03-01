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
	ResetSecret string // 管理员密码重置密钥，从 RESET_SECRET 读取
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
