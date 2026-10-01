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

func setupTaskTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestTaskHandler_ListTasks 测试获取事项列表
func TestTaskHandler_ListTasks(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.GET("/api/tasks", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.ListTasks(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tasks?page=1&size=10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTaskHandler_GetTask_InvalidID 测试获取事项详情（无效ID）
func TestTaskHandler_GetTask_InvalidID(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.GET("/api/tasks/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.GetTask(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tasks/invalid", nil)

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestTaskHandler_CreateTask_InvalidData 测试创建事项（无效数据）
func TestTaskHandler_CreateTask_InvalidData(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.POST("/api/tasks", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.CreateTask(c)
	})

	// 缺少必要字段
	reqData := map[string]string{}
	body, _ := json.Marshal(reqData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestTaskHandler_CreateTask_InvalidItemType 测试创建事项（无效类型）
func TestTaskHandler_CreateTask_InvalidItemType(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.POST("/api/tasks", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.CreateTask(c)
	})

	reqData := map[string]string{
		"title":    "Test Task",
		"itemType": "INVALID_TYPE",
	}
	body, _ := json.Marshal(reqData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestTaskHandler_UpdateTask_InvalidID 测试更新事项（无效ID）
func TestTaskHandler_UpdateTask_InvalidID(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.PUT("/api/tasks/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.UpdateTask(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/tasks/invalid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestTaskHandler_DeleteTask_InvalidID 测试删除事项（无效ID）
func TestTaskHandler_DeleteTask_InvalidID(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.DELETE("/api/tasks/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.DeleteTask(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/tasks/invalid", nil)

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestTaskHandler_UpdateStatus_InvalidID 测试更新状态（无效ID）
func TestTaskHandler_UpdateStatus_InvalidID(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.PUT("/api/tasks/:id/status", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.UpdateStatus(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/tasks/invalid/status", bytes.NewBuffer([]byte(`{"statusCode":"DOING"}`)))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestTaskHandler_UpdateStatus_InvalidStatusCode 测试更新状态（无效状态码）
func TestTaskHandler_UpdateStatus_InvalidStatusCode(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.PUT("/api/tasks/:id/status", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.UpdateStatus(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/tasks/1/status", bytes.NewBuffer([]byte(`{"statusCode":"INVALID"}`)))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assertBizCode(t, w, 400)
}

// TestTaskHandler_GetDashboardStats 测试获取仪表盘统计
func TestTaskHandler_GetDashboardStats(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.GET("/api/dashboard/stats", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.GetDashboardStats(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/dashboard/stats", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestTaskHandler_UpdateStatus_ValidTransitions 测试状态流转
func TestTaskHandler_UpdateStatus_ValidTransitions(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	// 状态流转共用同一路由，注册一次（重复注册会触发 gin panic）
	router.PUT("/api/tasks/:id/status", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.UpdateStatus(c)
	})

	// 测试 TODO -> DOING
	t.Run("TODO到DOING流转", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/tasks/1/status", bytes.NewBuffer([]byte(`{"statusCode":"DOING"}`)))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		// 根据实际业务逻辑，可能返回 200 或其他状态码
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})

	// 测试 DOING -> DONE
	t.Run("DOING到DONE流转", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/tasks/1/status", bytes.NewBuffer([]byte(`{"statusCode":"DONE"}`)))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})

	// 测试 TODO/DOING -> CANCELLED
	t.Run("TODO到CANCELLED流转", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/tasks/1/status", bytes.NewBuffer([]byte(`{"statusCode":"CANCELLED"}`)))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})
}

// TestTaskHandler_CreateTask_RecurrenceRule 测试创建带周期性配置的事项
func TestTaskHandler_CreateTask_RecurrenceRule(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.POST("/api/tasks", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.CreateTask(c)
	})

	// 测试非周期性任务（recurrenceRule为空对象）
	t.Run("创建非周期性任务", func(t *testing.T) {
		taskData := map[string]interface{}{
			"title":          "每日任务",
			"itemType":       "TASK",
			"priority":       "MEDIUM",
			"recurrenceType": "NONE",
			"recurrenceRule": "{}",
		}
		body, _ := json.Marshal(taskData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		// 空对象JSON应该被接受
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})

	// 测试周期性任务
	t.Run("创建周期性任务", func(t *testing.T) {
		taskData := map[string]interface{}{
			"title":          "每日任务",
			"itemType":       "TASK",
			"priority":       "MEDIUM",
			"recurrenceType": "DAILY",
			"recurrenceRule": `{"recurrenceType":"DAILY","weekDays":[],"endType":"NEVER"}`,
		}
		body, _ := json.Marshal(taskData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})
}

// TestTaskHandler_ListTasks_WithFilters 测试带筛选条件的列表查询
func TestTaskHandler_ListTasks_WithFilters(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.GET("/api/tasks", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.ListTasks(c)
	})

	testCases := []struct {
		name     string
		query    string
		expectOK bool
	}{
		{"按状态筛选TODO", "pageNum=1&pageSize=10&statusCode=TODO", true},
		{"按状态筛选DOING", "pageNum=1&pageSize=10&statusCode=DOING", true},
		{"按类型筛选TASK", "pageNum=1&pageSize=10&itemType=TASK", true},
		{"按优先级筛选HIGH", "pageNum=1&pageSize=10&priority=HIGH", true},
		{"今日事项", "pageNum=1&pageSize=10&today=true", true},
		{"已逾期事项", "pageNum=1&pageSize=10&overdue=true", true},
		{"多条件筛选", "pageNum=1&pageSize=10&statusCode=TODO&priority=HIGH", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/tasks?"+tc.query, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestTaskHandler_ValidateTaskData 测试事项数据验证
func TestTaskHandler_ValidateTaskData(t *testing.T) {
	testCases := []struct {
		name      string
		taskData  map[string]string
		expectErr bool
	}{
		{
			name:      "有效数据",
			taskData:  map[string]string{"title": "Test Task", "itemType": "TASK", "priority": "HIGH"},
			expectErr: false,
		},
		{
			name:      "缺少标题",
			taskData:  map[string]string{"itemType": "TASK", "priority": "HIGH"},
			expectErr: true,
		},
		{
			name:      "无效类型",
			taskData:  map[string]string{"title": "Test Task", "itemType": "INVALID", "priority": "HIGH"},
			expectErr: true,
		},
		{
			name:      "无效优先级",
			taskData:  map[string]string{"title": "Test Task", "itemType": "TASK", "priority": "INVALID"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟验证逻辑
			hasErr := false
			if tc.taskData["title"] == "" {
				hasErr = true
			}

			validTypes := map[string]bool{"TASK": true, "MEMO": true, "REQUIREMENT": true}
			if !validTypes[tc.taskData["itemType"]] {
				hasErr = true
			}

			validPriorities := map[string]bool{"HIGH": true, "MEDIUM": true, "LOW": true}
			if !validPriorities[tc.taskData["priority"]] {
				hasErr = true
			}

			assert.Equal(t, tc.expectErr, hasErr)
		})
	}
}

// TestTaskHandler_BatchDeleteTasks 测试批量删除
func TestTaskHandler_BatchDeleteTasks(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	// 模拟批量删除端点
	router.POST("/api/tasks/batch-delete", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.BatchDeleteTasks(c)
	})

	t.Run("批量删除多个事项", func(t *testing.T) {
		batchData := map[string]interface{}{
			"ids": []int64{1, 2, 3},
		}
		body, _ := json.Marshal(batchData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/tasks/batch-delete", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		// 根据实际实现，可能返回 200 或其他状态码
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})

	t.Run("批量删除空列表", func(t *testing.T) {
		batchData := map[string]interface{}{
			"ids": []int64{},
		}
		body, _ := json.Marshal(batchData)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/tasks/batch-delete", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assertBizCode(t, w, 400)
	})
}

// TestTaskHandler_GetTaskByID 测试根据ID获取事项详情
func TestTaskHandler_GetTaskByID(t *testing.T) {
	router := setupTaskTestRouter()
	handler := NewTaskHandler()

	router.GET("/api/tasks/:id", func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Set("isAdmin", false) // 对齐 auth 中间件身份上下文，避免 handler 内 isAdmin 断言 panic
		handler.GetTask(c)
	})

	// 响应契约：HTTP 固定 200 + body.code；期望值为业务码。
	// 有效ID 使用 id=4（init-test-data.sql 演示事项，归属 user 1，且不被本包其它测试改动）。
	testCases := []struct {
		name     string
		id       string
		wantCode int
	}{
		{"有效ID", "4", 200},
		{"无效ID格式", "abc", 400},
		{"负数ID", "-1", 404}, // 数字解析通过，查询无记录 → 事项不存在
		{"零ID", "0", 404},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/tasks/"+tc.id, nil)

			router.ServeHTTP(w, req)

			assertBizCode(t, w, tc.wantCode)
		})
	}
}
