package api

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

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
