package api

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// FrontendProxy 反向代理前端请求到开发服务器
func FrontendProxy() gin.HandlerFunc {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	target, err := url.Parse(frontendURL)
	if err != nil {
		panic("invalid FRONTEND_URL: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("frontend proxy error: " + err.Error()))
	}

	return func(c *gin.Context) {
		// 跳过 API 路由
		if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
			strings.HasPrefix(c.Request.URL.Path, "/v1/") ||
			strings.HasPrefix(c.Request.URL.Path, "/health") ||
			strings.HasPrefix(c.Request.URL.Path, "/mcp") {
			c.Next()
			return
		}

		// 跳过 WebSocket 升级请求（由后端处理）
		if strings.Contains(c.GetHeader("Connection"), "Upgrade") {
			c.Next()
			return
		}

		// 代理到前端
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
