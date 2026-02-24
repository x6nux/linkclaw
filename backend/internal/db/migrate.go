package db

import (
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations 在启动时执行所有待运行的 SQL 迁移文件。
// 使用 schema_migrations 表跟踪已执行的迁移版本，保证幂等性。
func RunMigrations(db *gorm.DB) error {
	// 1. 确保 schema_migrations 表存在
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(20) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// 2. 读取已执行的版本
	applied, err := getAppliedVersions(db)
	if err != nil {
		return fmt.Errorf("get applied versions: %w", err)
	}

	// 3. 读取迁移文件并排序
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// 4. 执行未运行的迁移
	for _, f := range files {
		version := extractVersion(f)
		if applied[version] {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		log.Printf("[migrate] 执行迁移: %s", f)
		if err := db.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("execute migration %s: %w", f, err)
		}

		if err := db.Exec(
			"INSERT INTO schema_migrations (version) VALUES (?)", version,
		).Error; err != nil {
			return fmt.Errorf("record migration %s: %w", f, err)
		}

		log.Printf("[migrate] 完成: %s", f)
	}

	return nil
}

// getAppliedVersions 查询已执行的迁移版本
func getAppliedVersions(db *gorm.DB) (map[string]bool, error) {
	var versions []string
	if err := db.Raw("SELECT version FROM schema_migrations").Scan(&versions).Error; err != nil {
		// 表刚创建时可能为空，不算错误
		if strings.Contains(err.Error(), "does not exist") {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	m := make(map[string]bool, len(versions))
	for _, v := range versions {
		m[v] = true
	}
	return m, nil
}

// extractVersion 从文件名提取版本号（如 "001_init_schema.sql" → "001"）
func extractVersion(filename string) string {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return filename
}
