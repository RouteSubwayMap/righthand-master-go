package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"righthand-go/internal/config"
	"righthand-go/internal/database"
	authHandler "righthand-go/services/auth-service/handler"
)

// setupIntegrationRouter 设置集成测试路由
func setupIntegrationRouter() *gin.Engine {
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
	authGroup := router.Group("/api/auth")
	{
		authHandler := authHandler.NewAuthHandler()
		authGroup.POST("/login", authHandler.Login)
	}

	systemGroup := router.Group("/api/system")
	{
		userHandler := authHandler.NewUserHandler()
		systemGroup.GET("/users/page", userHandler.ListUsers)

		roleHandler := authHandler.NewRoleHandler()
		systemGroup.GET("/roles", roleHandler.ListRoles)
		systemGroup.GET("/roles/all", roleHandler.GetAllRoles)

		menuHandler := authHandler.NewMenuHandler()
		systemGroup.GET("/menus/tree", menuHandler.GetMenuTree)

		dictHandler := authHandler.NewDictHandler()
		systemGroup.GET("/dicts", dictHandler.ListDicts)
	}

	return router
}

// TestIntegration_UserList_API 测试用户列表API
func TestIntegration_UserList_API(t *testing.T) {
	router := setupIntegrationRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/users/page?pageNum=1&pageSize=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应结构
	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "list")
	assert.Contains(t, data, "total")
	assert.Contains(t, data, "pageNum")
	assert.Contains(t, data, "pageSize")

	t.Logf("✅ 用户列表API测试通过,总数: %.0f", data["total"])
}

// TestIntegration_RoleList_API 测试角色列表API
func TestIntegration_RoleList_API(t *testing.T) {
	router := setupIntegrationRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/roles?pageNum=1&pageSize=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "list")
	assert.Contains(t, data, "total")

	t.Logf("✅ 角色列表API测试通过,总数: %.0f", data["total"])
}

// TestIntegration_RoleAll_API 测试获取所有角色API
func TestIntegration_RoleAll_API(t *testing.T) {
	router := setupIntegrationRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/roles/all", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])

	// 应该返回数组
	data := response["data"].([]interface{})
	assert.Greater(t, len(data), 0)

	t.Logf("✅ 获取所有角色API测试通过,角色数: %d", len(data))
}

// TestIntegration_MenuTree_API 测试菜单树API
func TestIntegration_MenuTree_API(t *testing.T) {
	router := setupIntegrationRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/menus/tree", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])

	// 应该返回数组
	data := response["data"].([]interface{})
	assert.Greater(t, len(data), 0)

	t.Logf("✅ 菜单树API测试通过,顶级菜单数: %d", len(data))
}

// TestIntegration_DictList_API 测试字典列表API
func TestIntegration_DictList_API(t *testing.T) {
	router := setupIntegrationRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/dicts?pageNum=1&pageSize=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "list")
	assert.Contains(t, data, "total")

	t.Logf("✅ 字典列表API测试通过,总数: %.0f", data["total"])
}

// TestIntegration_Login_API 测试登录API
func TestIntegration_Login_API(t *testing.T) {
	router := setupIntegrationRouter()

	// 准备登录数据
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginData)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "token")
	assert.NotEmpty(t, data["token"])

	t.Logf("✅ 登录API测试通过")
}

// TestIntegration_EmptyParams_ShouldNotReturn400 测试空参数不应返回400
func TestIntegration_EmptyParams_ShouldNotReturn400(t *testing.T) {
	router := setupIntegrationRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/users/page?pageNum=1&pageSize=10&username=&realName=", nil)
	router.ServeHTTP(w, req)

	// 不应该返回400错误
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Logf("✅ 空参数处理测试通过")
}

// TestIntegration_Menu_ChineseEncoding 测试菜单中文编码
func TestIntegration_Menu_ChineseEncoding(t *testing.T) {
	router := setupIntegrationRouter()

	// 获取菜单树
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/menus/tree", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, float64(200), response["code"])
	
	// 验证响应是有效的UTF-8
	assert.True(t, utf8.Valid(w.Body.Bytes()), 
		"API响应应该是有效的UTF-8编码")

	data := response["data"].([]interface{})
	t.Logf("✅ 菜单树API测试通过,顶级菜单数: %d", len(data))
	
	// 检查是否有中文菜单名称
	chineseMenuCount := 0
	for _, item := range data {
		menu := item.(map[string]interface{})
		if menuName, ok := menu["menuName"].(string); ok {
			// 检查是否包含非ASCII字符
			for _, char := range menuName {
				if char > 127 {
					chineseMenuCount++
					t.Logf("发现中文菜单: %s (ID: %v)", menuName, menu["id"])
					
					// 验证UTF-8字节正确性
					menuBytes := []byte(menuName)
					assert.True(t, utf8.Valid(menuBytes),
						"菜单名称 '%s' 应该是有效的UTF-8编码", menuName)
					break
				}
			}
		}
	}
	
	if chineseMenuCount > 0 {
		t.Logf("✅ 发现 %d 个中文菜单,编码正常", chineseMenuCount)
	} else {
		t.Logf("⚠️  未发现中文菜单(可能是测试环境数据)")
	}
}
