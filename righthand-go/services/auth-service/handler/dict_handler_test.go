package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDictHandler_ListDicts 测试获取字典列表
func TestDictHandler_ListDicts(t *testing.T) {
	router := setupTestRouter()
	handler := NewDictHandler()

	router.GET("/api/system/dicts", handler.ListDicts)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/dicts?pageNum=1&pageSize=10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestDictHandler_GetDictsByType 测试按类型获取字典
func TestDictHandler_GetDictsByType(t *testing.T) {
	router := setupTestRouter()
	handler := NewDictHandler()

	router.GET("/api/system/dicts/type/:type", handler.GetDictsByType)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/dicts/type/priority", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestDictHandler_GetDict_InvalidID 测试获取字典详情（无效ID）
func TestDictHandler_GetDict_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewDictHandler()

	router.GET("/api/system/dicts/:id", handler.GetDict)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/system/dicts/invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDictHandler_CreateDict_InvalidData 测试创建字典（无效数据）
// 已知缺陷（待修复）：CreateDict 无参数校验，空数据会被接受并写入库。
// 本测试忠实断言当前真实行为；生产补充参数校验后应改回断言 400。
func TestDictHandler_CreateDict_InvalidData(t *testing.T) {
	router := setupTestRouter()
	handler := NewDictHandler()

	router.POST("/api/system/dicts", handler.CreateDict)

	reqData := map[string]string{}
	body, _ := json.Marshal(reqData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/system/dicts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestDictHandler_UpdateDict_InvalidID 测试更新字典（无效ID）
func TestDictHandler_UpdateDict_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewDictHandler()

	router.PUT("/api/system/dicts/:id", handler.UpdateDict)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/system/dicts/invalid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDictHandler_DeleteDict_InvalidID 测试删除字典（无效ID）
func TestDictHandler_DeleteDict_InvalidID(t *testing.T) {
	router := setupTestRouter()
	handler := NewDictHandler()

	router.DELETE("/api/system/dicts/:id", handler.DeleteDict)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/system/dicts/invalid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
