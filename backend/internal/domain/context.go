package domain

import "time"

// ── 搜索配置与开关 ─────────────────────────────────────────

// ContextSearchConfig 上下文搜索配置
type ContextSearchConfig struct {
	// --- 索引相关 ---
	EnableIndexing      bool // 是否启用索引功能
	IndexThreshold      int  // 文件数超过此值启用索引 (默认 100)
	EnableIndexFallback bool // 索引失败时降级到全量扫描

	// --- 降级策略 ---
	EnableLLMFallback bool // LLM 失败时降级到关键词匹配

	// --- 限流保护 ---
	MaxConcurrentSearches int // 最大并发搜索数 (默认 10)
	RateLimitPerAgent     int // 每个 Agent 每秒最大请求数 (默认 5)

	// --- 灰度发布 ---
	AgentSearchEnabled bool    // 是否启用 Agent 搜索
	AgentSearchRatio   float64 // Agent 搜索灰度比例 (0.0-1.0)

	// --- 超时控制 ---
	SearchTimeoutMs      int // 普通搜索超时 (默认 30000ms)
	AgentSearchTimeoutMs int // Agent 搜索超时 (默认 60000ms)
	MaxSearchTimeoutMs   int // 最大允许超时 (默认 120000ms)
}

// DefaultContextSearchConfig 返回默认搜索配置
func DefaultContextSearchConfig() *ContextSearchConfig {
	return &ContextSearchConfig{
		EnableIndexing:        true,
		IndexThreshold:        100,
		EnableIndexFallback:   true,
		EnableLLMFallback:     true,
		MaxConcurrentSearches: 10,
		RateLimitPerAgent:     5,
		AgentSearchEnabled:    true,
		AgentSearchRatio:      1.0,
		SearchTimeoutMs:       30000,
		AgentSearchTimeoutMs:  60000,
		MaxSearchTimeoutMs:    120000,
	}
}

// ── 搜索错误码 ─────────────────────────────────────────

// SearchErrorCode 搜索错误码
type SearchErrorCode string

const (
	// --- 通用错误 ---
	ErrInvalidRequest SearchErrorCode = "INVALID_REQUEST"
	ErrUnauthorized   SearchErrorCode = "UNAUTHORIZED"
	ErrForbidden      SearchErrorCode = "FORBIDDEN"
	ErrNotFound       SearchErrorCode = "NOT_FOUND"

	// --- 搜索特定错误 ---
	ErrNoDirectories SearchErrorCode = "NO_DIRECTORIES"
	ErrIndexNotReady SearchErrorCode = "INDEX_NOT_READY"
	ErrTimeout       SearchErrorCode = "TIMEOUT"

	// --- LLM 相关错误 ---
	ErrLLMUnavailable     SearchErrorCode = "LLM_UNAVAILABLE"
	ErrLLMRatelimited     SearchErrorCode = "LLM_RATE_LIMITED"
	ErrLLMContextTooLarge SearchErrorCode = "LLM_CONTEXT_TOO_LARGE"

	// --- Agent 特定错误 ---
	ErrAgentMaxTurns   SearchErrorCode = "AGENT_MAX_TURNS"
	ErrAgentToolFailed SearchErrorCode = "AGENT_TOOL_FAILED"
)

// ── 目录管理 ─────────────────────────────────────────

// ContextDirectory 上下文目录配置
type ContextDirectory struct {
	ID              string     `json:"id"`
	CompanyID       string     `json:"company_id"`
	Name            string     `json:"name"`
	Path            string     `json:"path"`
	Description     string     `json:"description,omitempty"`
	IsActive        bool       `json:"is_active"`
	FilePatterns    string     `json:"file_patterns,omitempty"`
	ExcludePatterns string     `json:"exclude_patterns,omitempty"`
	MaxFileSize     int        `json:"max_file_size"`
	LastIndexedAt   *time.Time `json:"last_indexed_at,omitempty"`
	FileCount       int        `json:"file_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ContextFileSummary 文件总结缓存
type ContextFileSummary struct {
	ID           string    `json:"id"`
	DirectoryID  string    `json:"directory_id"`
	FilePath     string    `json:"file_path"`
	ContentHash  string    `json:"content_hash"`
	Summary      string    `json:"summary"`
	Language     string    `json:"language,omitempty"`
	LineCount    int       `json:"line_count"`
	SummarizedAt time.Time `json:"summarized_at"`
}

// ContextSearchLog 搜索历史
type ContextSearchLog struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"company_id"`
	AgentID      string    `json:"agent_id,omitempty"`
	Query        string    `json:"query"`
	DirectoryIDs string    `json:"directory_ids,omitempty"`
	ResultsCount int       `json:"results_count"`
	LatencyMs    int       `json:"latency_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

// ContextSearchResult 搜索结果
type ContextSearchResult struct {
	FilePath    string  `json:"file_path"`
	Language    string  `json:"language,omitempty"`
	Summary     string  `json:"summary"`
	Relevance   float64 `json:"relevance"` // 0-1 相关性分数
	Reason      string  `json:"reason"`    // LLM 解释为什么相关
	LineCount   int     `json:"line_count,omitempty"`
	DirectoryID string  `json:"directory_id"`
	// --- 新增字段 (v1.1) ---
	ContentHash string `json:"content_hash,omitempty"` // 内容哈希（用于缓存）
	IndexedAt   string `json:"indexed_at,omitempty"`   // 索引时间（如有）
}

// ── 统一请求参数 ─────────────────────────────────────────

// BaseSearchRequest 基础搜索请求参数
type BaseSearchRequest struct {
	Query        string   `json:"query" binding:"required"` // 必填：自然语言查询
	DirectoryIDs []string `json:"directory_ids"`            // 可选：限制搜索范围
	// --- 新增统一参数 ---
	MaxResults   int     `json:"max_results,omitempty"`   // 可选：最大返回数 (默认 10)
	MinRelevance float64 `json:"min_relevance,omitempty"` // 可选：最低相关性阈值 0-1 (默认 0.3)
	TimeoutMs    int     `json:"timeout_ms,omitempty"`    // 可选：超时时间 (默认 30000)
	UseIndex     *bool   `json:"use_index,omitempty"`     // 可选：是否使用索引 (默认 true)
}

// AgentSearchRequest Agent 搜索请求参数
type AgentSearchRequest struct {
	BaseSearchRequest
	MaxTurns    int      `json:"max_turns,omitempty"`    // 可选：Agent 最大轮数 (默认 5)
	EnableTools []string `json:"enable_tools,omitempty"` // 可选：启用的工具列表
}

// ── 统一响应结构 ─────────────────────────────────────────

// SearchResponseMeta 响应元数据
type SearchResponseMeta struct {
	OK        bool   `json:"ok"`
	RequestID string `json:"request_id"`
	LatencyMs int    `json:"latency_ms"`
}

// SearchDiagnostics 搜索诊断信息
type SearchDiagnostics struct {
	DirectoriesScanned int    `json:"directories_scanned"`
	FilesAnalyzed      int    `json:"files_analyzed"`
	IndexUsed          bool   `json:"index_used"`
	FallbackReason     string `json:"fallback_reason,omitempty"` // 降级原因（如有）
	FilesListed        int    `json:"files_listed,omitempty"`    // Agent 搜索：列出文件数
	FilesRead          int    `json:"files_read,omitempty"`      // Agent 搜索：读取文件数
}

// SearchError 错误响应
type SearchError struct {
	Code    SearchErrorCode `json:"code"`
	Message string          `json:"message"`
	Details map[string]any  `json:"details,omitempty"`
}

// ContextSearchResponse 普通搜索/MCP 搜索 响应
type ContextSearchResponse struct {
	SearchResponseMeta
	Results     []*ContextSearchResult `json:"results"`
	Total       int                    `json:"total"`
	Diagnostics *SearchDiagnostics     `json:"diagnostics,omitempty"`
	Error       *SearchError           `json:"error,omitempty"`
}

// ToolCallTrace Agent 工具调用 trace
type ToolCallTrace struct {
	ToolName      string `json:"tool_name"`
	Arguments     any    `json:"arguments"`
	ResultSummary string `json:"result_summary"`
	IsError       bool   `json:"is_error"`
}

// AgentSearchTrace Agent 搜索中间过程
type AgentSearchTrace struct {
	Turns     int             `json:"turns"`
	ToolCalls []ToolCallTrace `json:"tool_calls"`
}

// AgentSearchResponse Agent 搜索 响应
type AgentSearchResponse struct {
	SearchResponseMeta
	Answer      string             `json:"answer"`
	FilesRead   []string           `json:"files_read"`
	Trace       *AgentSearchTrace  `json:"trace,omitempty"`
	Diagnostics *SearchDiagnostics `json:"diagnostics,omitempty"`
	Error       *SearchError       `json:"error,omitempty"`
}
