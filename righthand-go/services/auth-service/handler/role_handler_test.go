package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRoleHandler_ListRoles 测试获取角色列表
func TestRoleHandler_ListRoles(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.GET("/api/system/roles", handler.ListRoles)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/roles?pageNum=1&pageSize=10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRoleHandler_GetAllRoles 测试获取所有角色
func TestRoleHandler_GetAllRoles(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.GET("/api/system/roles/all", handler.GetAllRoles)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/roles/all", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRoleHandler_GetRole_InvalidID 测试获取角色详情（无效ID）
func TestRoleHandler_GetRole_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.GET("/api/system/roles/:id", handler.GetRole)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/roles/invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRoleHandler_CreateRole_InvalidData 测试创建角色（无效数据）
func TestRoleHandler_CreateRole_InvalidData(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.POST("/api/system/roles", handler.CreateRole)

	reqData := map[string]string{}
	body, _ := json.Marshal(reqData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRoleHandler_UpdateRole_InvalidID 测试更新角色（无效ID）
func TestRoleHandler_UpdateRole_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.PUT("/api/system/roles/:id", handler.UpdateRole)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/system/roles/invalid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRoleHandler_DeleteRole_InvalidID 测试删除角色（无效ID）
func TestRoleHandler_DeleteRole_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.DELETE("/api/system/roles/:id", handler.DeleteRole)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/system/roles/invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRoleHandler_AssignMenus_InvalidID 测试分配菜单（无效ID）
func TestRoleHandler_AssignMenus_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.POST("/api/system/roles/:id/menus", handler.AssignMenus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/roles/invalid/menus", bytes.NewBuffer([]byte(`{"menuIds":[1,2]}`)))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRoleHandler_GetRoleMenus_InvalidID 测试获取角色菜单（无效ID）
func TestRoleHandler_GetRoleMenus_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewRoleHandler()

	router.GET("/api/system/roles/:id/menus", handler.GetRoleMenus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/roles/invalid/menus", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
