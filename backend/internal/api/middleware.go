package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

const (
	ctxAgent     = "agent"
	ctxCompanyID = "company_id"
)

// AuthMiddleware 支持三种认证：
// 1. Authorization: Bearer <api_key>   → AI Agent（API Key）
// 2. Authorization: Bearer <jwt_token> → 人类用户（JWT）
// 3. x-api-key: <api_key>             → Anthropic SDK 兼容（AI Agent）
func AuthMiddleware(agentRepo repository.AgentRepo, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		} else if xKey := c.GetHeader("x-api-key"); xKey != "" {
			token = xKey
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		// 尝试 JWT 解析（人类用户）
		if agent := tryJWT(c, token, agentRepo, jwtSecret); agent != nil {
			c.Set(ctxAgent, agent)
			c.Set(ctxCompanyID, agent.CompanyID)
			c.Next()
			return
		}

		// 尝试 API Key（AI Agent）
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
		agent, err := agentRepo.GetByAPIKeyHash(c.Request.Context(), hash)
		if err != nil || agent == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(ctxAgent, agent)
		c.Set(ctxCompanyID, agent.CompanyID)
		c.Next()
	}
}

func tryJWT(c *gin.Context, tokenStr string, agentRepo repository.AgentRepo, secret string) *domain.Agent {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	agentID, _ := claims["sub"].(string)
	if agentID == "" {
		return nil
	}
	agent, err := agentRepo.GetByID(c.Request.Context(), agentID)
	if err != nil || agent == nil || !agent.IsHuman {
		return nil
	}
	return agent
}

func currentAgent(c *gin.Context) *domain.Agent {
	v, _ := c.Get(ctxAgent)
	a, _ := v.(*domain.Agent)
	return a
}

func currentCompanyID(c *gin.Context) string {
	v, _ := c.Get(ctxCompanyID)
	s, _ := v.(string)
	return s
}

// ChairmanOnly 仅允许董事长访问
func ChairmanOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		agent := currentAgent(c)
		if agent == nil || agent.RoleType != domain.RoleChairman {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "chairman only"})
			return
		}
		c.Next()
	}
}

func AuditMiddleware(auditRepo repository.AuditRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if auditRepo == nil {
			return
		}
		agent := currentAgent(c)
		companyID := currentCompanyID(c)
		if agent == nil || companyID == "" {
			return
		}

		action := auditActionFromMethod(c.Request.Method)
		if action == "" {
			return
		}

		resourceType := auditResourceType(c)
		resourceID := auditResourceID(c)
		ipAddress := optionalString(c.ClientIP())
		userAgent := optionalString(c.Request.UserAgent())
		requestID := optionalString(c.GetHeader("X-Request-ID"))

		details, err := json.Marshal(map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"statusCode": c.Writer.Status(),
		})
		if err != nil {
			details = []byte("{}")
		}

		log := &domain.AuditLog{
			CompanyID:    companyID,
			AgentID:      agent.ID,
			AgentName:    agent.Name,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			RequestID:    requestID,
			Details:      details,
		}

		go func(log *domain.AuditLog) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = auditRepo.CreateAuditLog(ctx, log)
		}(log)
	}
}

func SensitiveAccessMiddleware(auditRepo repository.AuditRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auditRepo == nil {
			c.Next()
			return
		}

		agent := currentAgent(c)
		companyID := currentCompanyID(c)
		if agent == nil || companyID == "" {
			c.Next()
			return
		}

		action, targetResource, targetAgentID, ok := detectSensitiveAction(c)
		if !ok {
			c.Next()
			return
		}

		justification := strings.TrimSpace(c.GetHeader("X-Justification"))
		if justification == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Justification header"})
			return
		}

		approvalID := optionalString(c.GetHeader("X-Approval-ID"))
		ipAddress := optionalString(c.ClientIP())
		userAgent := optionalString(c.Request.UserAgent())
		requestID := optionalString(c.GetHeader("X-Request-ID"))
		justificationPtr := &justification

		c.Next()

		log := &domain.SensitiveAccessLog{
			CompanyID:      companyID,
			AgentID:        agent.ID,
			AgentName:      agent.Name,
			Action:         action,
			TargetResource: targetResource,
			TargetAgentID:  targetAgentID,
			IPAddress:      ipAddress,
			UserAgent:      userAgent,
			RequestID:      requestID,
			Justification:  justificationPtr,
			ApprovalID:     approvalID,
		}

		go func(log *domain.SensitiveAccessLog) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = auditRepo.CreateSensitiveAccessLog(ctx, log)
		}(log)
	}
}

func auditActionFromMethod(method string) domain.AuditAction {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return domain.AuditActionRead
	case http.MethodPost:
		return domain.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return domain.AuditActionUpdate
	case http.MethodDelete:
		return domain.AuditActionDelete
	default:
		return ""
	}
}

func auditResourceType(c *gin.Context) string {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	path = strings.Trim(path, "/")
	if path == "" {
		return "unknown"
	}
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[0] == "api" {
		return parts[2]
	}
	return parts[0]
}

func auditResourceID(c *gin.Context) *string {
	if id := strings.TrimSpace(c.Param("id")); id != "" {
		return &id
	}
	return nil
}

func optionalString(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}

func detectSensitiveAction(c *gin.Context) (domain.SensitiveAction, string, *string, bool) {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	lowerPath := strings.ToLower(path)
	method := c.Request.Method

	switch {
	case strings.Contains(lowerPath, "/payroll"):
		if method == http.MethodGet {
			return domain.SensitiveActionViewPayroll, "payroll", nil, true
		}
		return domain.SensitiveActionEditPayroll, "payroll", nil, true
	case strings.Contains(lowerPath, "/contracts"):
		if method == http.MethodGet {
			return domain.SensitiveActionViewContracts, "contracts", nil, true
		}
		return domain.SensitiveActionEditContracts, "contracts", nil, true
	case strings.Contains(lowerPath, "/personal"):
		if method == http.MethodGet {
			return domain.SensitiveActionViewPersonalInfo, "personal_info", nil, true
		}
		return domain.SensitiveActionEditPersonalInfo, "personal_info", nil, true
	case method == http.MethodDelete && isDeleteAgentPath(lowerPath):
		return domain.SensitiveActionDeleteAgent, "agent", auditResourceID(c), true
	case strings.Contains(lowerPath, "/export"):
		return domain.SensitiveActionExportAllData, "export", nil, true
	default:
		return "", "", nil, false
	}
}

func isDeleteAgentPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "agents"
}
