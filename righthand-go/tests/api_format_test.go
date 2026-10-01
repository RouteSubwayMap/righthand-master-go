package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// APIResponse 通用API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageResult 分页结果结构(后端实际返回格式)
type PageResult struct {
	Total    int64       `json:"total"`
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
	List     interface{} `json:"list"`
}

// TestUserPage_ResponseFormat 测试用户分页API返回格式
func TestUserPage_ResponseFormat(t *testing.T) {
	t.Skip("需要真实数据库和认证token")

	router := setupTestRouter()

	// TODO: 添加JWT token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/users/page?pageNum=1&pageSize=10", nil)
	// req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证返回的是分页格式
	pageData, ok := response.Data.(map[string]interface{})
	if ok {
		assert.NotNil(t, pageData["list"], "应该返回list字段")
		assert.NotNil(t, pageData["total"], "应该返回total字段")
		assert.NotNil(t, pageData["pageNum"], "应该返回pageNum字段")
		assert.NotNil(t, pageData["pageSize"], "应该返回pageSize字段")
	}
}

// TestRoleAll_ResponseFormat 测试角色全部API返回格式
func TestRoleAll_ResponseFormat(t *testing.T) {
	t.Skip("需要真实数据库和认证token")

	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/roles/all", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证返回的是数组格式
	_, ok := response.Data.([]interface{})
	assert.True(t, ok, "应该返回角色数组")
}

// TestEmptyParams_ShouldNotReturn400 测试空参数不应该返回400
func TestEmptyParams_ShouldNotReturn400(t *testing.T) {
	t.Skip("需要真实数据库和认证token")

	router := setupTestRouter()

	testCases := []struct {
		path        string
		description string
	}{
		{
			"/api/system/users/page?pageNum=1&pageSize=10&username=&realName=",
			"用户查询带空字符串参数",
		},
		{
			"/api/system/users/page?pageNum=1&pageSize=10",
			"用户查询无额外参数",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tc.path, nil)
			router.ServeHTTP(w, req)

			// 不应该返回400
			assert.NotEqual(t, 400, w.Code, "%s 不应该返回400错误", tc.description)

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err == nil {
				t.Logf("%s - Code: %d, Message: %s", tc.description, response.Code, response.Message)
			}
		})
	}
}
