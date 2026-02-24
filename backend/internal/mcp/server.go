package mcp

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

const (
	protocolVersion = "2024-11-05"
	sseKeepAlive    = 15 * time.Second
)

// Server 是 MCP SSE 服务端
type Server struct {
	agentRepo repository.AgentRepo
	handler   *Handler
	sessions  *SessionStore
	rdb       *redis.Client
}

func NewServer(
	agentRepo repository.AgentRepo,
	handler *Handler,
	rdb *redis.Client,
) *Server {
	return &Server{
		agentRepo: agentRepo,
		handler:   handler,
		sessions:  newSessionStore(),
		rdb:       rdb,
	}
}

// RegisterRoutes 注册 MCP 路由
func (s *Server) RegisterRoutes(r gin.IRouter) {
	// 旧版 SSE 传输（nanoclaw 主进程使用）
	r.GET("/mcp/sse", s.handleSSE)
	r.POST("/mcp/message", s.handleMessage)
	// Streamable HTTP 传输（Agent SDK 内联配置使用）
	r.POST("/mcp", s.handleHTTP)
	r.DELETE("/mcp", s.handleHTTPDelete)
}

// handleSSE 建立 SSE 连接
// GET /mcp/sse  Authorization: Bearer <api_key>
func (s *Server) handleSSE(c *gin.Context) {
	// 1. 验证 API Key
	agent, err := s.authenticateBearer(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 2. 创建 Session
	sessID := uuid.New().String()
	sess := newSession(sessID, agent)
	s.sessions.Set(sessID, sess)

	// 在 Redis 中标记 session 存活
	s.rdb.Set(c.Request.Context(),
		fmt.Sprintf("mcp:session:%s", sessID), agent.ID, 24*time.Hour)

	// 更新 agent 状态为 online
	s.agentRepo.UpdateStatus(c.Request.Context(), agent.ID, domain.StatusOnline)
	s.agentRepo.UpdateLastSeen(c.Request.Context(), agent.ID)

	// 3. SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	// 4. 发送 endpoint 事件（告知客户端消息端点）
	endpointEvent := fmt.Sprintf("event: endpoint\ndata: /mcp/message?session_id=%s\n\n", sessID)
	fmt.Fprint(c.Writer, endpointEvent)
	c.Writer.Flush()

	// 5. 进入事件循环
	ticker := time.NewTicker(sseKeepAlive)
	defer func() {
		ticker.Stop()
		sess.Close()
		s.sessions.Delete(sessID)
		s.rdb.Del(c.Request.Context(), fmt.Sprintf("mcp:session:%s", sessID))
		s.agentRepo.UpdateStatus(c.Request.Context(), agent.ID, domain.StatusOffline)
	}()

	for {
		select {
		case msg := <-sess.SendCh():
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			c.Writer.Flush()

		case <-ticker.C:
			fmt.Fprint(c.Writer, ": keepalive\n\n")
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return

		case <-sess.Done():
			return
		}
	}
}

// handleMessage 接收 JSON-RPC 请求
// POST /mcp/message?session_id=<id>
func (s *Server) handleMessage(c *gin.Context) {
	sessID := c.Query("session_id")
	sess, ok := s.sessions.Get(sessID)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResp(nil, ErrInvalidRequest, "session not found"))
		return
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResp(nil, ErrParseError, "parse error"))
		return
	}

	resp := s.handler.Handle(c.Request.Context(), sess, req)

	// 将响应通过 SSE 推回
	data, _ := json.Marshal(resp)
	sess.Send(string(data))

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// handleHTTP 处理 Streamable HTTP 传输的 JSON-RPC 请求
// POST /mcp  Authorization: Bearer <api_key>  Mcp-Session-Id: <optional>
func (s *Server) handleHTTP(c *gin.Context) {
	agent, err := s.authenticateBearer(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResp(nil, ErrParseError, "parse error"))
		return
	}

	// 获取或创建 session
	sessID := c.GetHeader("Mcp-Session-Id")
	var sess *Session

	if sessID != "" {
		sess, _ = s.sessions.Get(sessID)
	}

	// initialize 请求始终创建新 session
	if req.Method == "initialize" {
		sessID = uuid.New().String()
		sess = newSession(sessID, agent)
		s.sessions.Set(sessID, sess)
		s.rdb.Set(c.Request.Context(),
			fmt.Sprintf("mcp:session:%s", sessID), agent.ID, 24*time.Hour)
		s.agentRepo.UpdateStatus(c.Request.Context(), agent.ID, domain.StatusOnline)
		s.agentRepo.UpdateLastSeen(c.Request.Context(), agent.ID)
	}

	// session 丢失（后端重启等）时自动恢复，避免 SDK 无法重连
	if sess == nil {
		sessID = uuid.New().String()
		sess = newSession(sessID, agent)
		sess.Initialized = true // 跳过 initialize 握手
		s.sessions.Set(sessID, sess)
		s.rdb.Set(c.Request.Context(),
			fmt.Sprintf("mcp:session:%s", sessID), agent.ID, 24*time.Hour)
		s.agentRepo.UpdateLastSeen(c.Request.Context(), agent.ID)
	}

	// 通知（无 id）：仅确认
	if req.ID == nil {
		c.Writer.Header().Set("Mcp-Session-Id", sessID)
		c.Status(http.StatusAccepted)
		return
	}

	resp := s.handler.Handle(c.Request.Context(), sess, req)
	c.Writer.Header().Set("Mcp-Session-Id", sessID)
	c.JSON(http.StatusOK, resp)
}

// handleHTTPDelete 终止 streamable HTTP session
// DELETE /mcp  Mcp-Session-Id: <id>
func (s *Server) handleHTTPDelete(c *gin.Context) {
	sessID := c.GetHeader("Mcp-Session-Id")
	if sess, ok := s.sessions.Get(sessID); ok {
		s.agentRepo.UpdateStatus(c.Request.Context(), sess.Agent.ID, domain.StatusOffline)
		sess.Close()
		s.sessions.Delete(sessID)
		s.rdb.Del(c.Request.Context(), fmt.Sprintf("mcp:session:%s", sessID))
	}
	c.Status(http.StatusNoContent)
}

// authenticateBearer 解析并验证 Bearer API Key
func (s *Server) authenticateBearer(c *gin.Context) (*domain.Agent, error) {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("missing Bearer token")
	}
	key := strings.TrimPrefix(auth, "Bearer ")
	if key == "" {
		return nil, fmt.Errorf("empty API key")
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	agent, err := s.agentRepo.GetByAPIKeyHash(c.Request.Context(), hash)
	if err != nil || agent == nil {
		return nil, fmt.Errorf("invalid API key")
	}
	return agent, nil
}
