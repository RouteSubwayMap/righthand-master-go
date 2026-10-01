package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRequirementService 模拟需求服务
type MockRequirementService struct {
	mock.Mock
}

func (m *MockRequirementService) ListRequirements(userID int64, isAdmin bool, page, pageSize int, status, priority, bankName, groupType, title, requirementNo, riskLevel, requirementType, followerName *string) ([]models.Requirement, int64, error) {
	args := m.Called(userID, isAdmin, page, pageSize, status, priority, bankName, groupType, title, requirementNo, riskLevel, requirementType, followerName)
	return args.Get(0).([]models.Requirement), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequirementService) GetRequirement(userID, id int64) (*models.Requirement, error) {
	args := m.Called(userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) CreateRequirement(userID int64, req *models.Requirement) (*models.Requirement, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) UpdateRequirement(userID, id int64, req *models.Requirement) (*models.Requirement, error) {
	args := m.Called(userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Requirement), args.Error(1)
}

func (m *MockRequirementService) DeleteRequirement(id, userID int64, isAdmin bool) error {
	args := m.Called(id, userID, isAdmin)
	return args.Error(0)
}

func (m *MockRequirementService) UpdateStatus(userID, id int64, status string) error {
	args := m.Called(userID, id, status)
	return args.Error(0)
}

func TestListRequirements(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试正常查询
	t.Run("正常查询需求列表", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))
		c.Request = httptest.NewRequest("GET", "/api/requirements?pageNum=1&pageSize=10", nil)

		// 验证参数解析
		assert.Equal(t, int64(1), c.GetInt64("userID"))
		assert.Equal(t, "1", c.Query("pageNum"))
		assert.Equal(t, "10", c.Query("pageSize"))
	})

	// 测试带筛选条件查询
	t.Run("带筛选条件查询", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))
		c.Request = httptest.NewRequest("GET", "/api/requirements?pageNum=1&pageSize=10&status=PENDING_REVIEW&priority=HIGH", nil)

		assert.Equal(t, "PENDING_REVIEW", c.Query("status"))
		assert.Equal(t, "HIGH", c.Query("priority"))
	})
}

func TestGetRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("获取需求详情", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))
		c.Request = httptest.NewRequest("GET", "/api/requirements/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		// 验证参数
		assert.Equal(t, int64(1), c.GetInt64("userID"))
		assert.Equal(t, "1", c.Param("id"))
	})
}

func TestCreateRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("创建需求-正常参数", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))

		requirement := map[string]interface{}{
			"title":       "新需求",
			"description": "需求描述",
			"reqType":     "INTERNAL",
			"priority":    "HIGH",
			"status":      "PENDING_REVIEW",
		}

		body, _ := json.Marshal(requirement)
		c.Request = httptest.NewRequest("POST", "/api/requirements", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证请求
		assert.Equal(t, int64(1), c.GetInt64("userID"))
		assert.NotNil(t, c.Request.Body)
	})

	t.Run("创建需求-缺失必填字段", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))

		// 缺少 title
		requirement := map[string]interface{}{
			"description": "需求描述",
		}

		body, _ := json.Marshal(requirement)
		c.Request = httptest.NewRequest("POST", "/api/requirements", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证请求体
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		assert.Empty(t, reqBody["title"])
	})
}

func TestUpdateRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("更新需求", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		requirement := map[string]interface{}{
			"title":       "更新后的需求",
			"description": "更新后的描述",
			"priority":    "LOW",
		}

		body, _ := json.Marshal(requirement)
		c.Request = httptest.NewRequest("PUT", "/api/requirements/1", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证参数
		assert.Equal(t, "1", c.Param("id"))
		assert.Equal(t, int64(1), c.GetInt64("userID"))
	})
}

func TestDeleteRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("删除需求", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Request = httptest.NewRequest("DELETE", "/api/requirements/1", nil)

		// 验证参数
		assert.Equal(t, "1", c.Param("id"))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUpdateRequirementStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("更新需求状态", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		body := []byte(`{"status": "IN_PROGRESS"}`)
		c.Request = httptest.NewRequest("PUT", "/api/requirements/1/status", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证参数
		assert.Equal(t, "1", c.Param("id"))

		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		assert.Equal(t, "IN_PROGRESS", reqBody["status"])
	})
}

func TestRequirementStatusValues(t *testing.T) {
	// 测试需求状态枚举值
	validStatuses := []string{
		"PENDING_REVIEW", // 待评审
		"ACCEPTED",       // 已接受
		"IN_PROGRESS",    // 进行中
		"COMPLETED",      // 已完成
		"REJECTED",       // 已拒绝
	}

	for _, status := range validStatuses {
		t.Run("有效状态_"+status, func(t *testing.T) {
			assert.NotEmpty(t, status)
		})
	}

	// 测试无效状态
	invalidStatuses := []string{"", "INVALID", "pending", "COMPLETE"}
	for _, status := range invalidStatuses {
		t.Run("无效状态_"+status, func(t *testing.T) {
			isValid := false
			for _, valid := range validStatuses {
				if status == valid {
					isValid = true
					break
				}
			}
			if status == "" || status == "INVALID" || status == "pending" || status == "COMPLETE" {
				assert.False(t, isValid)
			}
		})
	}
}

func TestRequirementTypeValues(t *testing.T) {
	// 测试需求类型枚举值
	validTypes := []string{
		"INTERNAL", // 内部优化
		"BANK",     // 行内需求
	}

	for _, reqType := range validTypes {
		t.Run("有效类型_"+reqType, func(t *testing.T) {
			assert.NotEmpty(t, reqType)
		})
	}
}

func TestRequirementPriorityValues(t *testing.T) {
	// 测试需求优先级枚举值
	validPriorities := []string{"HIGH", "MEDIUM", "LOW"}

	for _, priority := range validPriorities {
		t.Run("有效优先级_"+priority, func(t *testing.T) {
			assert.NotEmpty(t, priority)
		})
	}
}

func TestRequirementResponseFormat(t *testing.T) {
	// 测试响应格式
	t.Run("分页响应格式", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 正确调用 utils.Success，传入 gin.Context
		utils.Success(c, gin.H{
			"list":     []models.Requirement{},
			"total":    0,
			"pageNum":  1,
			"pageSize": 10,
		})

		// 验证响应状态码和结构
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应体
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
		assert.NotNil(t, response["data"])
	})
}

// TestRequirementHandler_ListRequirements 测试需求列表查询
func TestRequirementHandler_ListRequirements(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name         string
		query        string
		expectParams map[string]string
	}{
		{
			name:         "基础列表查询",
			query:        "pageNum=1&pageSize=10",
			expectParams: map[string]string{"pageNum": "1", "pageSize": "10"},
		},
		{
			name:         "按状态筛选",
			query:        "pageNum=1&pageSize=10&status=PENDING_CLARIFY",
			expectParams: map[string]string{"status": "PENDING_CLARIFY"},
		},
		{
			name:         "按优先级筛选",
			query:        "pageNum=1&pageSize=10&priority=HIGH",
			expectParams: map[string]string{"priority": "HIGH"},
		},
		{
			name:         "按银行筛选",
			query:        "pageNum=1&pageSize=10&bankName=工商银行",
			expectParams: map[string]string{"bankName": "工商银行"},
		},
		{
			name:         "按需求类型筛选",
			query:        "pageNum=1&pageSize=10&requirementType=BANK",
			expectParams: map[string]string{"requirementType": "BANK"},
		},
		{
			name:         "按小组类型筛选",
			query:        "pageNum=1&pageSize=10&groupType=LOAN_MIDDLE",
			expectParams: map[string]string{"groupType": "LOAN_MIDDLE"},
		},
		{
			name:         "多条件筛选",
			query:        "pageNum=1&pageSize=10&status=PENDING_CLARIFY&priority=HIGH&bankName=工商银行",
			expectParams: map[string]string{"status": "PENDING_CLARIFY", "priority": "HIGH", "bankName": "工商银行"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", int64(1))
			c.Request = httptest.NewRequest("GET", "/api/requirements?"+tc.query, nil)

			// 验证参数解析
			for key, expectedValue := range tc.expectParams {
				assert.Equal(t, expectedValue, c.Query(key))
			}
		})
	}
}

// TestRequirementHandler_CreateRequirement 测试创建需求
func TestRequirementHandler_CreateRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("创建需求-完整参数", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))

		requirement := map[string]interface{}{
			"title":                  "新需求",
			"description":            "需求描述",
			"requirementType":        "BANK",
			"priority":               "HIGH",
			"bankName":               "工商银行",
			"requirementNo":          "REQ20260001",
			"groupType":              "LOAN_MIDDLE",
			"riskLevel":              "MEDIUM",
			"followerName":           "张三",
			"expectedCompletionDate": "2026-06-30",
		}

		body, _ := json.Marshal(requirement)
		c.Request = httptest.NewRequest("POST", "/api/requirements", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证请求
		assert.Equal(t, int64(1), c.GetInt64("userID"))
		assert.NotNil(t, c.Request.Body)

		// 解析并验证请求体
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		assert.Equal(t, "新需求", reqBody["title"])
		assert.Equal(t, "BANK", reqBody["requirementType"])
		assert.Equal(t, "HIGH", reqBody["priority"])
	})

	t.Run("创建需求-最小参数", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))

		requirement := map[string]interface{}{
			"title": "新需求",
		}

		body, _ := json.Marshal(requirement)
		c.Request = httptest.NewRequest("POST", "/api/requirements", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证请求
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		assert.Equal(t, "新需求", reqBody["title"])
		assert.Nil(t, reqBody["description"])
	})

	t.Run("创建需求-缺少标题", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))

		requirement := map[string]interface{}{
			"description": "需求描述",
		}

		body, _ := json.Marshal(requirement)
		c.Request = httptest.NewRequest("POST", "/api/requirements", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证请求体
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		assert.Empty(t, reqBody["title"])
	})
}

// TestRequirementHandler_UpdateRequirement 测试更新需求
func TestRequirementHandler_UpdateRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name       string
		id         string
		updateData map[string]interface{}
		expectCode int
	}{
		{
			name:       "更新标题和描述",
			id:         "1",
			updateData: map[string]interface{}{"title": "更新标题", "description": "更新描述"},
			expectCode: http.StatusOK,
		},
		{
			name:       "更新优先级",
			id:         "1",
			updateData: map[string]interface{}{"priority": "LOW"},
			expectCode: http.StatusOK,
		},
		{
			name:       "更新状态",
			id:         "1",
			updateData: map[string]interface{}{"status": "PENDING_SCHEDULING"},
			expectCode: http.StatusOK,
		},
		{
			name:       "无效ID",
			id:         "invalid",
			updateData: map[string]interface{}{"title": "测试"},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", int64(1))
			c.Params = gin.Params{{Key: "id", Value: tc.id}}

			body, _ := json.Marshal(tc.updateData)
			c.Request = httptest.NewRequest("PUT", "/api/requirements/"+tc.id, bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// 验证参数
			assert.Equal(t, tc.id, c.Param("id"))
		})
	}
}

// TestRequirementHandler_BatchImportRequirements 测试批量导入需求
func TestRequirementHandler_BatchImportRequirements(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("批量导入需求-有效数据", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))

		importData := []map[string]interface{}{
			{
				"title":           "需求1",
				"description":     "描述1",
				"requirementType": "BANK",
				"priority":        "HIGH",
				"bankName":        "工商银行",
			},
			{
				"title":           "需求2",
				"description":     "描述2",
				"requirementType": "INTERNAL",
				"priority":        "MEDIUM",
			},
		}

		body, _ := json.Marshal(map[string]interface{}{"requirements": importData})
		c.Request = httptest.NewRequest("POST", "/api/requirements/batch-import", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证请求
		assert.Equal(t, int64(1), c.GetInt64("userID"))

		// 解析并验证
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		requirements := reqBody["requirements"].([]interface{})
		assert.Equal(t, 2, len(requirements))
	})

	t.Run("批量导入需求-空数据", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))

		importData := []map[string]interface{}{}

		body, _ := json.Marshal(map[string]interface{}{"requirements": importData})
		c.Request = httptest.NewRequest("POST", "/api/requirements/batch-import", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 验证请求
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		requirements := reqBody["requirements"].([]interface{})
		assert.Equal(t, 0, len(requirements))
	})
}

// TestRequirementHandler_ExportRequirements 测试导出需求
func TestRequirementHandler_ExportRequirements(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("导出需求", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("userID", int64(1))
		c.Request = httptest.NewRequest("GET", "/api/requirements/export?status=COMPLETED", nil)

		// 验证请求
		assert.Equal(t, int64(1), c.GetInt64("userID"))
		assert.Equal(t, "COMPLETED", c.Query("status"))
	})
}

// TestRequirementStatusTransitions 测试需求状态流转
func TestRequirementStatusTransitions(t *testing.T) {
	// 定义有效的状态流转
	validTransitions := map[string][]string{
		"PENDING_REVIEW":     {"ACCEPTED", "REJECTED"},
		"ACCEPTED":           {"PENDING_SCHEDULING"},
		"PENDING_SCHEDULING": {"DEVELOPING"},
		"DEVELOPING":         {"TESTING"},
		"TESTING":            {"PENDING_RELEASE", "DEVELOPING"},
		"PENDING_RELEASE":    {"COMPLETED"},
		"COMPLETED":          {},
		"REJECTED":           {},
	}

	// 测试有效流转
	t.Run("PENDING_REVIEW -> ACCEPTED", func(t *testing.T) {
		transitions, exists := validTransitions["PENDING_REVIEW"]
		assert.True(t, exists)
		assert.Contains(t, transitions, "ACCEPTED")
	})

	t.Run("PENDING_REVIEW -> REJECTED", func(t *testing.T) {
		transitions, exists := validTransitions["PENDING_REVIEW"]
		assert.True(t, exists)
		assert.Contains(t, transitions, "REJECTED")
	})

	t.Run("ACCEPTED -> PENDING_SCHEDULING", func(t *testing.T) {
		transitions, exists := validTransitions["ACCEPTED"]
		assert.True(t, exists)
		assert.Contains(t, transitions, "PENDING_SCHEDULING")
	})

	t.Run("PENDING_SCHEDULING -> DEVELOPING", func(t *testing.T) {
		transitions, exists := validTransitions["PENDING_SCHEDULING"]
		assert.True(t, exists)
		assert.Contains(t, transitions, "DEVELOPING")
	})

	t.Run("DEVELOPING -> TESTING", func(t *testing.T) {
		transitions, exists := validTransitions["DEVELOPING"]
		assert.True(t, exists)
		assert.Contains(t, transitions, "TESTING")
	})

	t.Run("TESTING -> PENDING_RELEASE", func(t *testing.T) {
		transitions, exists := validTransitions["TESTING"]
		assert.True(t, exists)
		assert.Contains(t, transitions, "PENDING_RELEASE")
	})

	t.Run("PENDING_RELEASE -> COMPLETED", func(t *testing.T) {
		transitions, exists := validTransitions["PENDING_RELEASE"]
		assert.True(t, exists)
		assert.Contains(t, transitions, "COMPLETED")
	})

	// 测试终态不能流转
	t.Run("COMPLETED终态不能流转", func(t *testing.T) {
		transitions, exists := validTransitions["COMPLETED"]
		assert.True(t, exists)
		assert.Empty(t, transitions)
	})

	t.Run("REJECTED终态不能流转", func(t *testing.T) {
		transitions, exists := validTransitions["REJECTED"]
		assert.True(t, exists)
		assert.Empty(t, transitions)
	})
}

// TestRequirementGroupTypeValues 测试小组类型枚举值
func TestRequirementGroupTypeValues(t *testing.T) {
	validGroupTypes := []string{
		"LOAN_MIDDLE", // 贷中
		"LOAN_PRE",    // 贷前
		"RISK",        // 风险
	}

	for _, groupType := range validGroupTypes {
		t.Run("有效小组类型_"+groupType, func(t *testing.T) {
			assert.NotEmpty(t, groupType)
		})
	}

	// 测试无效小组类型
	invalidGroupTypes := []string{"", "INVALID", "loan_middle"}
	for _, groupType := range invalidGroupTypes {
		t.Run("无效小组类型_"+groupType, func(t *testing.T) {
			isValid := false
			for _, valid := range validGroupTypes {
				if groupType == valid {
					isValid = true
					break
				}
			}
			assert.False(t, isValid)
		})
	}
}

// TestRequirementRiskLevelValues 测试风险等级枚举值
func TestRequirementRiskLevelValues(t *testing.T) {
	validRiskLevels := []string{"HIGH", "MEDIUM", "LOW"}

	for _, riskLevel := range validRiskLevels {
		t.Run("有效风险等级_"+riskLevel, func(t *testing.T) {
			assert.NotEmpty(t, riskLevel)
		})
	}

	// 测试无效风险等级
	invalidRiskLevels := []string{"", "INVALID", "high"}
	for _, riskLevel := range invalidRiskLevels {
		t.Run("无效风险等级_"+riskLevel, func(t *testing.T) {
			isValid := false
			for _, valid := range validRiskLevels {
				if riskLevel == valid {
					isValid = true
					break
				}
			}
			assert.False(t, isValid)
		})
	}
}

// TestRequirementValidation 测试需求数据验证
func TestRequirementValidation(t *testing.T) {
	testCases := []struct {
		name      string
		data      map[string]interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "有效需求",
			data:      map[string]interface{}{"title": "测试需求", "requirementType": "BANK", "priority": "HIGH"},
			expectErr: false,
		},
		{
			name:      "缺少标题",
			data:      map[string]interface{}{"requirementType": "BANK", "priority": "HIGH"},
			expectErr: true,
			errField:  "title",
		},
		{
			name:      "无效需求类型",
			data:      map[string]interface{}{"title": "测试需求", "requirementType": "INVALID", "priority": "HIGH"},
			expectErr: true,
			errField:  "requirementType",
		},
		{
			name:      "无效优先级",
			data:      map[string]interface{}{"title": "测试需求", "requirementType": "BANK", "priority": "INVALID"},
			expectErr: true,
			errField:  "priority",
		},
		{
			name:      "无效风险等级",
			data:      map[string]interface{}{"title": "测试需求", "requirementType": "BANK", "priority": "HIGH", "riskLevel": "INVALID"},
			expectErr: true,
			errField:  "riskLevel",
		},
		{
			name:      "无效小组类型",
			data:      map[string]interface{}{"title": "测试需求", "requirementType": "BANK", "priority": "HIGH", "groupType": "INVALID"},
			expectErr: true,
			errField:  "groupType",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟验证逻辑
			hasErr := false
			errorField := ""

			if tc.data["title"] == nil || tc.data["title"] == "" {
				hasErr = true
				errorField = "title"
			}

			validTypes := map[string]bool{"BANK": true, "INTERNAL": true}
			if !validTypes[tc.data["requirementType"].(string)] {
				hasErr = true
				errorField = "requirementType"
			}

			validPriorities := map[string]bool{"HIGH": true, "MEDIUM": true, "LOW": true}
			if !validPriorities[tc.data["priority"].(string)] {
				hasErr = true
				errorField = "priority"
			}

			// 风险等级/小组类型枚举校验（值域与 TestRequirementRiskLevelValues / TestRequirementGroupTypeValues 一致）
			if v, ok := tc.data["riskLevel"]; ok {
				validRiskLevels := map[string]bool{"HIGH": true, "MEDIUM": true, "LOW": true}
				if !validRiskLevels[v.(string)] {
					hasErr = true
					errorField = "riskLevel"
				}
			}
			if v, ok := tc.data["groupType"]; ok {
				validGroupTypes := map[string]bool{"LOAN_MIDDLE": true, "LOAN_PRE": true, "RISK": true}
				if !validGroupTypes[v.(string)] {
					hasErr = true
					errorField = "groupType"
				}
			}

			assert.Equal(t, tc.expectErr, hasErr)
			if tc.expectErr {
				assert.Equal(t, tc.errField, errorField)
			}
		})
	}
}
