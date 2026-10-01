package service

import (
	"righthand-go/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试用的SQLite数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法打开测试数据库: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&models.Requirement{})
	if err != nil {
		t.Fatalf("无法迁移表结构: %v", err)
	}

	return db
}

// TestRequirementService_Integration_CreateAndGet 集成测试: 创建并获取需求
func TestRequirementService_Integration_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db // 注入测试数据库

	// 创建测试需求
	now := time.Now()
	req := &models.Requirement{
		UserID:          1,
		Title:           "测试需求-集成测试",
		Description:     stringPtrTest("这是一个集成测试需求"),
		BankName:        stringPtrTest("工商银行"),
		RequirementNo:   stringPtrTest("REQ-INTEGRATION-001"),
		Priority:        stringPtrTest("HIGH"),
		RiskLevel:       stringPtrTest("MEDIUM"),
		RequirementType: stringPtrTest("INTERNAL"),
		Status:          stringPtrTest("PENDING_CLARIFY"),
		SubmitterID:     int64PtrTest(1),
		SubmitterName:   stringPtrTest("测试用户"),
		SubmitTime:      &now,
	}

	// 创建需求
	err := service.CreateRequirement(req)
	assert.NoError(t, err)
	assert.NotZero(t, req.ID)

	// 获取需求
	retrieved, err := service.GetRequirementByID(req.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "测试需求-集成测试", retrieved.Title)
	assert.Equal(t, "工商银行", *retrieved.BankName)
	assert.Equal(t, "REQ-INTEGRATION-001", *retrieved.RequirementNo)
	assert.Equal(t, "HIGH", *retrieved.Priority)
	assert.Equal(t, "PENDING_CLARIFY", *retrieved.Status)
}

// TestRequirementService_Integration_UpdateStatus 集成测试: 更新需求状态
func TestRequirementService_Integration_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 创建测试需求
	now := time.Now()
	req := &models.Requirement{
		UserID:      1,
		Title:       "状态流转测试",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(1),
		SubmitTime:  &now,
	}

	err := service.CreateRequirement(req)
	assert.NoError(t, err)

	// 更新状态为待排期
	err = service.UpdateStatus(req.ID, 1, "PENDING_SCHEDULING")
	assert.NoError(t, err)

	// 验证状态已更新
	updated, err := service.GetRequirementByID(req.ID)
	assert.NoError(t, err)
	assert.Equal(t, "PENDING_SCHEDULING", *updated.Status)

	// 继续流转到开发中
	err = service.UpdateStatus(req.ID, 1, "DEVELOPING")
	assert.NoError(t, err)

	updated, err = service.GetRequirementByID(req.ID)
	assert.NoError(t, err)
	assert.Equal(t, "DEVELOPING", *updated.Status)

	// 流转到测试中
	err = service.UpdateStatus(req.ID, 1, "TESTING")
	assert.NoError(t, err)

	updated, err = service.GetRequirementByID(req.ID)
	assert.NoError(t, err)
	assert.Equal(t, "TESTING", *updated.Status)

	// 完成需求
	err = service.UpdateStatus(req.ID, 1, "COMPLETED")
	assert.NoError(t, err)

	updated, err = service.GetRequirementByID(req.ID)
	assert.NoError(t, err)
	assert.Equal(t, "COMPLETED", *updated.Status)
	assert.NotNil(t, updated.CompletionDate)
}

// TestRequirementService_Integration_ListRequirements 集成测试: 查询需求列表
func TestRequirementService_Integration_ListRequirements(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 创建多个测试需求
	now := time.Now()
	requirements := []models.Requirement{
		{
			UserID:      1,
			Title:       "需求1",
			Status:      stringPtrTest("PENDING_CLARIFY"),
			Priority:    stringPtrTest("HIGH"),
			BankName:    stringPtrTest("工商银行"),
			GroupType:   stringPtrTest("LOAN_MIDDLE"),
			SubmitterID: int64PtrTest(1),
			SubmitTime:  &now,
		},
		{
			UserID:      1,
			Title:       "需求2",
			Status:      stringPtrTest("DEVELOPING"),
			Priority:    stringPtrTest("MEDIUM"),
			BankName:    stringPtrTest("建设银行"),
			GroupType:   stringPtrTest("RISK"),
			SubmitterID: int64PtrTest(1),
			SubmitTime:  &now,
		},
		{
			UserID:      1,
			Title:       "需求3",
			Status:      stringPtrTest("COMPLETED"),
			Priority:    stringPtrTest("LOW"),
			BankName:    stringPtrTest("农业银行"),
			GroupType:   stringPtrTest("LOAN_PRE"),
			SubmitterID: int64PtrTest(1),
			SubmitTime:  &now,
		},
	}

	for i := range requirements {
		err := service.CreateRequirement(&requirements[i])
		assert.NoError(t, err)
	}

	// 测试用例1: 查询所有需求
	t.Run("查询所有需求", func(t *testing.T) {
		list, total, err := service.ListRequirements(1, false, 1, 10, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, list, 3)
	})

	// 测试用例2: 按状态筛选
	t.Run("按状态筛选", func(t *testing.T) {
		status := "DEVELOPING"
		list, total, err := service.ListRequirements(1, false, 1, 10, &status, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
		assert.Equal(t, "需求2", list[0].Title)
	})

	// 测试用例3: 按优先级筛选
	t.Run("按优先级筛选", func(t *testing.T) {
		priority := "HIGH"
		list, total, err := service.ListRequirements(1, false, 1, 10, nil, &priority, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
		assert.Equal(t, "需求1", list[0].Title)
	})

	// 测试用例4: 按银行筛选
	t.Run("按银行筛选", func(t *testing.T) {
		bankName := "工商"
		list, total, err := service.ListRequirements(1, false, 1, 10, nil, nil, &bankName, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
		assert.Equal(t, "工商银行", *list[0].BankName)
	})

	// 测试用例5: 按小组类型筛选
	t.Run("按小组类型筛选", func(t *testing.T) {
		groupType := "RISK"
		list, total, err := service.ListRequirements(1, false, 1, 10, nil, nil, nil, &groupType, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
		assert.Equal(t, "RISK", *list[0].GroupType)
	})

	// 测试用例6: 分页测试
	t.Run("分页测试", func(t *testing.T) {
		list, total, err := service.ListRequirements(1, false, 1, 2, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, list, 2)

		list2, total2, err := service.ListRequirements(1, false, 2, 2, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total2)
		assert.Len(t, list2, 1)
	})
}

// TestRequirementService_Integration_UpdateRequirement 集成测试: 更新需求
func TestRequirementService_Integration_UpdateRequirement(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 创建测试需求
	now := time.Now()
	req := &models.Requirement{
		UserID:      1,
		Title:       "原始标题",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(1),
		SubmitTime:  &now,
	}

	err := service.CreateRequirement(req)
	assert.NoError(t, err)

	// 更新需求
	req.Title = "更新后的标题"
	req.Description = stringPtrTest("更新的描述")
	req.Priority = stringPtrTest("HIGH")
	updateUser := int64(1)
	req.UpdateUser = &updateUser

	err = service.UpdateRequirement(req, false)
	assert.NoError(t, err)

	// 验证更新
	updated, err := service.GetRequirementByID(req.ID)
	assert.NoError(t, err)
	assert.Equal(t, "更新后的标题", updated.Title)
	assert.Equal(t, "更新的描述", *updated.Description)
	assert.Equal(t, "HIGH", *updated.Priority)
}

// TestRequirementService_Integration_DeleteRequirement 集成测试: 删除需求
func TestRequirementService_Integration_DeleteRequirement(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 创建测试需求
	now := time.Now()
	req := &models.Requirement{
		UserID:      1,
		Title:       "待删除需求",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(1),
		SubmitTime:  &now,
	}

	err := service.CreateRequirement(req)
	assert.NoError(t, err)

	// 删除需求
	err = service.DeleteRequirement(req.ID, 1, false)
	assert.NoError(t, err)

	// 验证删除(逻辑删除,应该查不到)
	_, err = service.GetRequirementByID(req.ID)
	assert.Error(t, err)
}

// TestRequirementService_Integration_ExportRequirements 集成测试: 导出需求
func TestRequirementService_Integration_ExportRequirements(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 创建测试数据
	now := time.Now()
	requirements := []models.Requirement{
		{
			UserID:        1,
			Title:         "导出测试1",
			Status:        stringPtrTest("DEVELOPING"),
			Priority:      stringPtrTest("HIGH"),
			BankName:      stringPtrTest("工商银行"),
			GroupType:     stringPtrTest("LOAN_MIDDLE"),
			RequirementNo: stringPtrTest("REQ-EXP-001"),
			SubmitterID:   int64PtrTest(1),
			SubmitTime:    &now,
		},
		{
			UserID:        1,
			Title:         "导出测试2",
			Status:        stringPtrTest("TESTING"),
			Priority:      stringPtrTest("MEDIUM"),
			BankName:      stringPtrTest("建设银行"),
			GroupType:     stringPtrTest("RISK"),
			RequirementNo: stringPtrTest("REQ-EXP-002"),
			SubmitterID:   int64PtrTest(1),
			SubmitTime:    &now,
		},
	}

	for i := range requirements {
		err := service.CreateRequirement(&requirements[i])
		assert.NoError(t, err)
	}

	// 测试用例1: 导出所有需求
	t.Run("导出所有需求", func(t *testing.T) {
		list, err := service.ExportRequirements(1, false, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, list, 2)
	})

	// 测试用例2: 按状态导出
	t.Run("按状态导出", func(t *testing.T) {
		status := "DEVELOPING"
		list, err := service.ExportRequirements(1, false, &status, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "导出测试1", list[0].Title)
	})

	// 测试用例3: 按银行导出
	t.Run("按银行导出", func(t *testing.T) {
		bankName := "建设"
		list, err := service.ExportRequirements(1, false, nil, nil, &bankName, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "建设银行", *list[0].BankName)
	})

	// 测试用例4: 按需求编号导出
	t.Run("按需求编号导出", func(t *testing.T) {
		requirementNo := "REQ-EXP-002"
		list, err := service.ExportRequirements(1, false, nil, nil, nil, nil, nil, &requirementNo, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "导出测试2", list[0].Title)
		assert.Equal(t, "REQ-EXP-002", *list[0].RequirementNo)
	})
}

// TestRequirementService_Integration_GetDetailShared 集成测试: 需求详情为团队共享视图（用例矩阵 tests/cases/requirement-shared-view.md）
func TestRequirementService_Integration_GetDetailShared(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 需求A: 用户1创建; 需求B: 用户2创建且follower_id为NULL; 需求D: 用户2创建且已软删除
	now := time.Now()
	reqA := &models.Requirement{
		UserID:      1,
		Title:       "需求A-用户1创建",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(1),
		SubmitTime:  &now,
	}
	reqB := &models.Requirement{
		UserID:      2,
		Title:       "需求B-用户2创建",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(2),
		SubmitTime:  &now,
	}
	reqD := &models.Requirement{
		UserID:      2,
		Title:       "需求D-已软删除",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(2),
		SubmitTime:  &now,
		Deleted:     intPtrTest(1),
	}

	for _, r := range []*models.Requirement{reqA, reqB, reqD} {
		assert.NoError(t, service.CreateRequirement(r))
	}

	t.Run("GetDetail_NonOwnerNonFollower_Success", func(t *testing.T) {
		retrieved, err := service.GetRequirementByID(reqB.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, reqB.ID, retrieved.ID)
		assert.Equal(t, "需求B-用户2创建", retrieved.Title)
	})

	t.Run("GetDetail_OwnRequirement_Success", func(t *testing.T) {
		retrieved, err := service.GetRequirementByID(reqA.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, reqA.ID, retrieved.ID)
		assert.Equal(t, "需求A-用户1创建", retrieved.Title)
	})

	t.Run("GetDetail_SoftDeleted_NotFound", func(t *testing.T) {
		_, err := service.GetRequirementByID(reqD.ID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("GetDetail_NotExist_NotFound", func(t *testing.T) {
		_, err := service.GetRequirementByID(999999)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("GetDetail_InvalidID_Negative_NotFound", func(t *testing.T) {
		_, err := service.GetRequirementByID(-1)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

// TestRequirementService_Integration_SharedView 集成测试: 列表/导出为团队共享视图（用例矩阵 tests/cases/requirement-shared-view.md）
func TestRequirementService_Integration_SharedView(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 需求A: 用户1创建; 需求B: 用户2创建(状态DEVELOPING, 编号REQ-B); 需求E: 用户2创建且已软删除
	now := time.Now()
	reqA := &models.Requirement{
		UserID:      1,
		Title:       "需求A-用户1创建",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(1),
		SubmitTime:  &now,
	}
	reqB := &models.Requirement{
		UserID:        2,
		Title:         "需求B-用户2创建",
		Status:        stringPtrTest("DEVELOPING"),
		Priority:      stringPtrTest("MEDIUM"),
		RequirementNo: stringPtrTest("REQ-B"),
		SubmitterID:   int64PtrTest(2),
		SubmitTime:    &now,
	}
	reqE := &models.Requirement{
		UserID:      2,
		Title:       "需求E-已软删除",
		Status:      stringPtrTest("PENDING_CLARIFY"),
		Priority:    stringPtrTest("MEDIUM"),
		SubmitterID: int64PtrTest(2),
		SubmitTime:  &now,
		Deleted:     intPtrTest(1),
	}

	for _, r := range []*models.Requirement{reqA, reqB, reqE} {
		assert.NoError(t, service.CreateRequirement(r))
	}

	t.Run("List_AnyUserSeesAll_Success", func(t *testing.T) {
		list, total, err := service.ListRequirements(1, false, 1, 10, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, list, 2)
	})

	t.Run("List_StatusFilter_AcrossUsers_Visible", func(t *testing.T) {
		status := "DEVELOPING"
		list, total, err := service.ListRequirements(1, false, 1, 10, &status, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
		assert.Equal(t, "需求B-用户2创建", list[0].Title)
	})

	t.Run("Export_AnyUserExportsAll_Success", func(t *testing.T) {
		list, err := service.ExportRequirements(1, false, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, list, 2)
		hasOtherUser := false
		for _, r := range list {
			if r.Title == "需求B-用户2创建" {
				hasOtherUser = true
			}
		}
		assert.True(t, hasOtherUser, "导出应包含用户2创建的需求B")
	})

	t.Run("Export_RequirementNoOfOtherUser_Found", func(t *testing.T) {
		requirementNo := "REQ-B"
		list, err := service.ExportRequirements(1, false, nil, nil, nil, nil, nil, &requirementNo, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "需求B-用户2创建", list[0].Title)
	})

	t.Run("Export_Deleted_Excluded", func(t *testing.T) {
		list, err := service.ExportRequirements(1, false, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, list, 2)
		for _, r := range list {
			assert.NotEqual(t, "需求E-已软删除", r.Title)
		}
	})
}

// TestRequirementService_Integration_DefaultValues 集成测试: 默认值设置
func TestRequirementService_Integration_DefaultValues(t *testing.T) {
	db := setupTestDB(t)
	service := NewRequirementService()
	service.db = db

	// 创建最小化需求(只提供必填字段)
	req := &models.Requirement{
		UserID: 1,
		Title:  "最小化需求",
	}

	err := service.CreateRequirement(req)
	assert.NoError(t, err)

	// 验证默认值
	retrieved, err := service.GetRequirementByID(req.ID)
	assert.NoError(t, err)
	assert.Equal(t, "MEDIUM", *retrieved.Priority)
	assert.Equal(t, "MEDIUM", *retrieved.RiskLevel)
	assert.Equal(t, "INTERNAL", *retrieved.RequirementType)
	assert.Equal(t, "PENDING_CLARIFY", *retrieved.Status)
	assert.NotNil(t, retrieved.Deleted)
	assert.Equal(t, 0, *retrieved.Deleted)
}

// 辅助函数(用于测试)
func stringPtrTest(s string) *string {
	return &s
}

func int64PtrTest(i int64) *int64 {
	return &i
}

func intPtrTest(i int) *int {
	return &i
}
