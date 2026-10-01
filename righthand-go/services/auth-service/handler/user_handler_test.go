package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUserHandler_ListUsers 测试获取用户列表
func TestUserHandler_ListUsers(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.GET("/api/system/users", handler.ListUsers)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/users?pageNum=1&pageSize=10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUserHandler_GetUser_InvalidID 测试获取用户详情（无效ID）
func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.GET("/api/system/users/:id", handler.GetUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/users/invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserHandler_GetUser_NotFound 测试获取用户详情（用户不存在）
func TestUserHandler_GetUser_NotFound(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.GET("/api/system/users/:id", handler.GetUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/users/99999", nil)

	router.ServeHTTP(w, req)

	// 对齐 handler 实现：用户不存在时 utils.NotFound 返回真实 HTTP 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestUserHandler_CreateUser_InvalidData 测试创建用户（无效数据）
func TestUserHandler_CreateUser_InvalidData(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.POST("/api/system/users", handler.CreateUser)

	// 缺少必要字段
	reqData := map[string]string{}
	body, _ := json.Marshal(reqData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserHandler_UpdateUser_InvalidID 测试更新用户（无效ID）
func TestUserHandler_UpdateUser_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.PUT("/api/system/users/:id", handler.UpdateUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/system/users/invalid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserHandler_DeleteUser_InvalidID 测试删除用户（无效ID）
func TestUserHandler_DeleteUser_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.DELETE("/api/system/users/:id", handler.DeleteUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/system/users/invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserHandler_AssignRoles_InvalidID 测试分配角色（无效ID）
func TestUserHandler_AssignRoles_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.POST("/api/system/users/:id/roles", handler.AssignRoles)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/users/invalid/roles", bytes.NewBuffer([]byte(`{"roleIds":[1,2]}`)))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUserHandler_GetUserRoles_InvalidID 测试获取用户角色（无效ID）
func TestUserHandler_GetUserRoles_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.GET("/api/system/users/:id/roles", handler.GetUserRoles)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/users/invalid/roles", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
