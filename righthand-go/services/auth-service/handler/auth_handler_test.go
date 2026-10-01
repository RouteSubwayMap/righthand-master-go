package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestLoginHandler_Success 测试登录成功
func TestLoginHandler_Success(t *testing.T) {
	// 跳过需要数据库的测试
	t.Skip("Skipping test that requires database connection")
	
	router := setupTestRouter()
	handler := NewAuthHandler()
	
	router.POST("/api/auth/login", handler.Login)

	// 准备请求数据
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginData)

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 注意：实际测试需要数据库连接，这里只是示例
	t.Logf("Response: %s", w.Body.String())
}

// TestLoginHandler_InvalidCredentials 测试登录失败（无效凭证）
func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// 跳过需要数据库的测试
	t.Skip("Skipping test that requires database connection")
	
	router := setupTestRouter()
	handler := NewAuthHandler()
	
	router.POST("/api/auth/login", handler.Login)

	// 准备错误的请求数据
	loginData := map[string]string{
		"username": "admin",
		"password": "wrong_password",
	}
	body, _ := json.Marshal(loginData)

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	t.Logf("Response: %s", w.Body.String())
}

// TestLoginHandler_MissingFields 测试登录失败（缺少字段）
func TestLoginHandler_MissingFields(t *testing.T) {
	router := setupTestRouter()
	handler := NewAuthHandler()
	
	router.POST("/api/auth/login", handler.Login)

	// 准备不完整的请求数据
	loginData := map[string]string{
		"username": "admin",
		// 缺少password
	}
	body, _ := json.Marshal(loginData)

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetUserInfo_Handler 测试获取用户信息
func TestGetUserInfo_Handler(t *testing.T) {
	// 跳过需要数据库的测试
	t.Skip("Skipping test that requires database connection")
	
	router := setupTestRouter()
	handler := NewAuthHandler()
	
	// 添加JWT认证中间件的模拟
	router.GET("/api/auth/info", func(c *gin.Context) {
		// 模拟已认证的用户
		c.Set("userID", int64(1))
		c.Set("username", "admin")
		handler.GetInfo(c)
	})

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/auth/info", nil)

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	t.Logf("Response: %s", w.Body.String())
}

// TestLogout_Handler 测试退出登录
func TestLogout_Handler(t *testing.T) {
	router := setupTestRouter()
	handler := NewAuthHandler()
	
	router.POST("/api/auth/logout", handler.Logout)

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/logout", nil)

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	t.Logf("Response: %s", w.Body.String())
}
