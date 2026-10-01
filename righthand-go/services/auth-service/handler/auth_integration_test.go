package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"righthand-go/internal/config"
	"righthand-go/internal/database"
	"righthand-go/internal/middleware"
)

// setupIntegrationTest 设置集成测试环境(使用真实数据库)
func setupIntegrationTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 初始化数据库（强约束：自动测试仅允许使用独立测试库，见 docs/ai/rules.md §9）
	cfg, err := config.LoadConfig("../../configs")
	if err != nil {
		panic(err)
	}
	cfg.Database.DBName = database.TestDBName
	database.InitDB(&cfg.Database)

	// 注册路由
	authHandler := NewAuthHandler()
	
	// 登录接口(无需认证)
	router.POST("/api/auth/login", authHandler.Login)
	
	// 需要JWT认证的路由
	authGroup := router.Group("/api/auth")
	authGroup.Use(middleware.JWTAuth())
	{
		authGroup.GET("/info", authHandler.GetInfo)
		authGroup.POST("/logout", authHandler.Logout)
	}

	return router
}

// TestLoginHandler_Success_Integration 测试登录成功(集成测试)
func TestLoginHandler_Success_Integration(t *testing.T) {
	router := setupIntegrationTest()

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
	assert.Equal(t, http.StatusOK, w.Code, "登录应该成功")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证返回token
	assert.NotNil(t, response["data"], "应该返回数据")
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["token"], "应该返回token")

	t.Logf("登录成功, Token: %v", data["token"])
}

// TestLoginHandler_InvalidCredentials_Integration 测试登录失败（无效凭证）
func TestLoginHandler_InvalidCredentials_Integration(t *testing.T) {
	router := setupIntegrationTest()

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

	// 验证响应 - 应该返回401而不是500
	assert.Equal(t, http.StatusUnauthorized, w.Code, "密码错误应该返回401")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(401), response["code"], "错误码应该是401")
	t.Logf("响应: %s", w.Body.String())
}

// TestLoginHandler_MissingFields_Integration 测试登录失败（缺少字段）
func TestLoginHandler_MissingFields_Integration(t *testing.T) {
	router := setupIntegrationTest()

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
	assert.Equal(t, http.StatusBadRequest, w.Code, "缺少字段应该返回400")
	t.Logf("响应: %s", w.Body.String())
}

// TestGetUserInfo_Integration 测试获取用户信息(集成测试)
func TestGetUserInfo_Integration(t *testing.T) {
	router := setupIntegrationTest()

	// 先登录获取token
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginData)

	loginW := httptest.NewRecorder()
	loginReq, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	loginReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(loginW, loginReq)

	var loginResponse map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	token := loginResponse["data"].(map[string]interface{})["token"].(string)

	// 使用token获取用户信息
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/auth/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code, "获取用户信息应该成功")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(200), response["code"])
	t.Logf("用户信息: %s", w.Body.String())
}

// TestLogout_Integration 测试退出登录(集成测试)
func TestLogout_Integration(t *testing.T) {
	router := setupIntegrationTest()

	// 先登录获取token
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginData)

	loginW := httptest.NewRecorder()
	loginReq, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	loginReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(loginW, loginReq)

	var loginResponse map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	token := loginResponse["data"].(map[string]interface{})["token"].(string)

	// 退出登录
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code, "退出登录应该成功")
	t.Logf("响应: %s", w.Body.String())
}
