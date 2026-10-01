package service

import (
	"testing"
	"time"

	"righthand-go/internal/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDBForTask 创建测试数据库(Task相关)
func setupTestDBForTask(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法打开测试数据库: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&models.TaskItem{}, &models.Contact{})
	if err != nil {
		t.Fatalf("无法迁移表结构: %v", err)
	}

	return db
}

// TestTaskService_CreateAndGet 测试创建和获取事项
func TestTaskService_CreateAndGet(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &TaskService{db: db}

	now := time.Now()
	task := &models.TaskItem{
		UserID:     1,
		Title:      "测试任务",
		ItemType:   "TASK",
		StatusCode: "TODO",
		Priority:   "HIGH",
		Deadline:   &now,
		CreateUser: func() *int64 { i := int64(1); return &i }(),
	}

	// 创建
	err := service.CreateTask(task)
	assert.NoError(t, err)
	assert.Greater(t, task.ID, int64(0))

	// 获取
	retrieved, err := service.GetTask(task.ID, 1, false)
	assert.NoError(t, err)
	assert.Equal(t, "测试任务", retrieved.Title)
	assert.Equal(t, "TASK", retrieved.ItemType)
	assert.Equal(t, "TODO", retrieved.StatusCode)
}

// TestTaskService_ListTasks 测试获取事项列表
func TestTaskService_ListTasks(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &TaskService{db: db}

	// 创建测试数据
	tasks := []models.TaskItem{
		{UserID: 1, Title: "任务1", ItemType: "TASK", StatusCode: "TODO", Priority: "HIGH"},
		{UserID: 1, Title: "任务2", ItemType: "MEMO", StatusCode: "DOING", Priority: "MEDIUM"},
		{UserID: 1, Title: "任务3", ItemType: "TASK", StatusCode: "DONE", Priority: "LOW"},
		{UserID: 2, Title: "其他用户任务", ItemType: "TASK", StatusCode: "TODO"},
	}

	for _, task := range tasks {
		db.Create(&task)
	}

	// 测试查询所有
	result, err := service.ListTasks(1, false, 1, 10, map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result["total"])

	// 测试按类型过滤
	filter := map[string]interface{}{"item_type": "TASK"}
	result, err = service.ListTasks(1, false, 1, 10, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result["total"])

	// 测试按状态过滤
	filter = map[string]interface{}{"status_code": "TODO"}
	result, err = service.ListTasks(1, false, 1, 10, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result["total"])
}

// TestTaskService_UpdateTask 测试更新事项
func TestTaskService_UpdateTask(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &TaskService{db: db}

	// 创建测试数据
	task := models.TaskItem{
		UserID:     1,
		Title:      "原任务",
		StatusCode: "TODO",
	}
	db.Create(&task)

	// 更新
	updates := map[string]interface{}{
		"title":       "更新后的任务",
		"status_code": "DOING",
		"update_user": int64(1),
	}

	err := service.UpdateTask(task.ID, 1, updates)
	assert.NoError(t, err)

	// 验证更新
	var updated models.TaskItem
	db.First(&updated, task.ID)
	assert.Equal(t, "更新后的任务", updated.Title)
	assert.Equal(t, "DOING", updated.StatusCode)
}

// TestTaskService_DeleteTask 测试删除事项
func TestTaskService_DeleteTask(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &TaskService{db: db}

	// 创建测试数据
	task := models.TaskItem{
		UserID:  1,
		Title:   "待删除任务",
		Deleted: 0,
	}
	db.Create(&task)

	// 删除
	err := service.DeleteTask(task.ID, 1)
	assert.NoError(t, err)

	// 验证逻辑删除
	var deleted models.TaskItem
	db.First(&deleted, task.ID)
	assert.Equal(t, 1, deleted.Deleted)
}

// TestTaskService_GetDashboardStats 测试仪表盘统计
func TestTaskService_GetDashboardStats(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &TaskService{db: db}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// 创建测试数据
	tasks := []models.TaskItem{
		{UserID: 1, Title: "待办1", StatusCode: "TODO"},
		{UserID: 1, Title: "待办2", StatusCode: "TODO"},
		{UserID: 1, Title: "进行中1", StatusCode: "DOING"},
		{UserID: 1, Title: "已完成1", StatusCode: "DONE"},
		{UserID: 1, Title: "已逾期1", StatusCode: "TODO", Deadline: &yesterday},
		{UserID: 1, Title: "未逾期", StatusCode: "TODO", Deadline: &tomorrow},
		{UserID: 2, Title: "其他用户", StatusCode: "TODO"}, // 不应计入统计
	}

	for _, task := range tasks {
		db.Create(&task)
	}

	// 获取统计
	stats, err := service.GetDashboardStats(1, false)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(3), stats["todo"]) // 2个待办 + 1个已逾期
	assert.Equal(t, int64(1), stats["doing"])
	assert.Equal(t, int64(1), stats["done"])
	assert.Equal(t, int64(1), stats["overdue"]) // 只有昨天的算逾期
}

// TestContactService_CreateAndGet 测试创建和获取联系人
func TestContactService_CreateAndGet(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &ContactService{db: db}

	contact := &models.Contact{
		UserID:       1,
		Name:         "张三",
		Phone:        "13800138000",
		Email:        "zhangsan@example.com",
		ContactGroup: "研发",
		CreateUser:   func() *int64 { i := int64(1); return &i }(),
	}

	// 创建
	err := service.CreateContact(contact)
	assert.NoError(t, err)
	assert.Greater(t, contact.ID, int64(0))

	// 获取
	retrieved, err := service.GetContact(contact.ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, "张三", retrieved.Name)
	assert.Equal(t, "13800138000", retrieved.Phone)
}

// TestContactService_ListContacts 测试获取联系人列表
func TestContactService_ListContacts(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &ContactService{db: db}

	// 创建测试数据
	contacts := []models.Contact{
		{UserID: 1, Name: "张三", Phone: "13800000001", ContactGroup: "研发"},
		{UserID: 1, Name: "李四", Phone: "13800000002", ContactGroup: "产品经理"},
		{UserID: 1, Name: "王五", Phone: "13800000003", ContactGroup: "研发"},
		{UserID: 2, Name: "其他用户", Phone: "13800000004"},
	}

	for _, contact := range contacts {
		db.Create(&contact)
	}

	// 测试查询所有
	result, err := service.ListContacts(1, 1, 10, "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result["total"])

	// 测试按分组过滤
	result, err = service.ListContacts(1, 1, 10, "", "研发", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result["total"])

	// 测试关键字搜索
	result, err = service.ListContacts(1, 1, 10, "张三", "", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result["total"])
}

// TestContactService_UpdateContact 测试更新联系人
func TestContactService_UpdateContact(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &ContactService{db: db}

	// 创建测试数据
	contact := models.Contact{
		UserID: 1,
		Name:   "原名",
		Phone:  "13800000000",
	}
	db.Create(&contact)

	// 更新
	updates := map[string]interface{}{
		"name":  "新名字",
		"phone": "13900000000",
	}

	err := service.UpdateContact(contact.ID, 1, updates)
	assert.NoError(t, err)

	// 验证更新
	var updated models.Contact
	db.First(&updated, contact.ID)
	assert.Equal(t, "新名字", updated.Name)
	assert.Equal(t, "13900000000", updated.Phone)
}

// TestContactService_DeleteContact 测试删除联系人
func TestContactService_DeleteContact(t *testing.T) {
	db := setupTestDBForTask(t)
	service := &ContactService{db: db}

	// 创建测试数据
	contact := models.Contact{
		UserID:  1,
		Name:    "待删除",
		Deleted: 0,
	}
	db.Create(&contact)

	// 删除
	err := service.DeleteContact(contact.ID, 1)
	assert.NoError(t, err)

	// 验证逻辑删除
	var deleted models.Contact
	db.First(&deleted, contact.ID)
	assert.Equal(t, 1, deleted.Deleted)
}
