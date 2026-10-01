package middleware

import (
	"bytes"
	"log"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Logger_MiddlewareHttpLine_NoSensitiveBody 对应矩阵 tests/cases/app-logging.md
// 反例：访问日志含方法/路径/状态码/耗时/IP，不含请求体与密码等敏感信息
func TestLoggerMiddleware_HttpLineWithoutSensitiveBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	origWriter := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(origWriter) })

	r := gin.New()
	r.Use(Logger())
	r.POST("/api/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 200})
	})

	body := `{"username":"admin","password":"secret-password-123"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("期望恰 1 行访问日志，实际 %d 行: %q", len(lines), out)
	}
	// 标准库 log 默认前缀（日期时间）在行首，故只锚定行尾格式
	assert.Regexp(t, regexp.MustCompile(`\[HTTP\] POST /api/auth/login 200 \S+ \S+$`), lines[0])
	assert.NotContains(t, out, "secret-password-123")
	assert.NotContains(t, out, body)
}
