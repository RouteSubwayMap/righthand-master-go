package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"righthand-go/internal/utils"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "未提供认证token")
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Unauthorized(c, "token格式错误")
			c.Abort()
			return
		}

		// 解析token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.Unauthorized(c, "token无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("realName", claims.RealName)
		c.Set("isAdmin", claims.IsAdmin)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger 日志中间件（访问日志：方法/路径/状态码/耗时/客户端IP；不记录请求体与Token，遵守日志安全红线）
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 记录响应信息
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		latency := time.Since(startTime)

		log.Printf("[HTTP] %s %s %d %s %s", method, path, statusCode, latency, c.ClientIP())
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, err any) {
		log.Printf("[PANIC] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		utils.Error(c, "服务器内部错误")
	})
}
