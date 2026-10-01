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

func setupContactTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestContactHandler_ListContacts 测试获取联系人列表
func TestContactHandler_ListContacts(t *testing.T) {
	router := setupContactTestRouter()
	handler := NewContactHandler()

	router.GET("/api/contacts", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.ListContacts(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contacts?pageNum=1&pageSize=10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestContactHandler_GetContact_InvalidID 测试获取联系人详情（无效ID）
func TestContactHandler_GetContact_InvalidID(t *testing.T) {
	router := setupContactTestRouter()
	handler := NewContactHandler()

	router.GET("/api/contacts/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.GetContact(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contacts/invalid", nil)

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestContactHandler_CreateContact_InvalidData 测试创建联系人（无效数据）
func TestContactHandler_CreateContact_InvalidData(t *testing.T) {
	router := setupContactTestRouter()
	handler := NewContactHandler()

	router.POST("/api/contacts", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.CreateContact(c)
	})

	// 缺少必要字段
	reqData := map[string]string{}
	body, _ := json.Marshal(reqData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestContactHandler_UpdateContact_InvalidID 测试更新联系人（无效ID）
func TestContactHandler_UpdateContact_InvalidID(t *testing.T) {
	router := setupContactTestRouter()
	handler := NewContactHandler()

	router.PUT("/api/contacts/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.UpdateContact(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/contacts/invalid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestContactHandler_DeleteContact_InvalidID 测试删除联系人（无效ID）
func TestContactHandler_DeleteContact_InvalidID(t *testing.T) {
	router := setupContactTestRouter()
	handler := NewContactHandler()

	router.DELETE("/api/contacts/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		handler.DeleteContact(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/contacts/invalid", nil)

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestContactHandler_GetContactGroups 测试获取联系人分组
func TestContactHandler_GetContactGroups(t *testing.T) {
	router := setupContactTestRouter()
	handler := NewContactHandler()

	router.GET("/api/contacts/groups", handler.GetContactGroups)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contacts/groups", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}