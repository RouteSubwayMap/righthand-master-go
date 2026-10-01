package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]string{"name": "test"}
	Success(c, testData)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "success", response.Message)
	assert.Equal(t, testData, response.Data)
}

func TestSuccessWithMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]string{"name": "test"}
	SuccessWithMessage(c, "操作成功", testData)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "操作成功", response.Message)
	assert.Equal(t, testData, response.Data)
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, "服务器错误")

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 500, response.Code)
	assert.Equal(t, "服务器错误", response.Message)
	assert.Nil(t, response.Data)
}

func TestErrorWithCode(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ErrorWithCode(c, 1001, "业务错误")

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1001, response.Code)
	assert.Equal(t, "业务错误", response.Message)
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c, "资源不存在")

	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 404, response.Code)
	assert.Equal(t, "资源不存在", response.Message)
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c, "未授权")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 401, response.Code)
	assert.Equal(t, "未授权", response.Message)
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c, "权限不足")

	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 403, response.Code)
	assert.Equal(t, "权限不足", response.Message)
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c, "请求参数错误")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 400, response.Code)
	assert.Equal(t, "请求参数错误", response.Message)
}

func TestPageResult(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testList := []map[string]interface{}{
		{"id": 1, "name": "item1"},
		{"id": 2, "name": "item2"},
	}

	pageResult := PageResult{
		Total:    2,
		PageNum:  1,
		PageSize: 10,
		List:     testList,
	}

	Success(c, pageResult)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response.Code)
	
	dataBytes, _ := json.Marshal(response.Data)
	var result PageResult
	err = json.Unmarshal(dataBytes, &result)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 1, result.PageNum)
	assert.Equal(t, 10, result.PageSize)
	assert.Len(t, result.List, 2)
}
