package service

import (
	"errors"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"time"

	"gorm.io/gorm"
)

type TaskService struct {
	db *gorm.DB // 测试时注入,否则使用全局DB
}

func NewTaskService() *TaskService {
	return &TaskService{}
}

// getDB 获取数据库连接
func (s *TaskService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(userID int64, isAdmin bool, page, size int, filter map[string]interface{}) (map[string]interface{}, error) {
	db := s.getDB()

	var query *gorm.DB
	if isAdmin {
		query = db.Model(&models.TaskItem{}).Where("deleted = ?", 0)
	} else {
		query = db.Model(&models.TaskItem{}).Where("user_id = ? AND deleted = ?", userID, 0)
	}

	// 应用过滤条件
	if title, ok := filter["title"]; ok {
		query = query.Where("title LIKE ?", "%"+title.(string)+"%")
	}
	if itemType, ok := filter["item_type"]; ok {
		query = query.Where("item_type = ?", itemType)
	}
	if statusCode, ok := filter["status_code"]; ok {
		query = query.Where("status_code = ?", statusCode)
	}
	if priority, ok := filter["priority"]; ok {
		query = query.Where("priority = ?", priority)
	}
	if category, ok := filter["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if startDate, ok := filter["start_date"]; ok {
		query = query.Where("deadline >= ?", startDate)
	}
	if endDate, ok := filter["end_date"]; ok {
		query = query.Where("deadline <= ?", endDate)
	}
	// 今日待办(截止日期为今天) - 优化：避免使用 DATE() 函数
	if today, ok := filter["today"]; ok && today.(bool) {
		todayStart := time.Now().Format("2006-01-02")
		todayEnd := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		query = query.Where("deadline >= ? AND deadline < ? AND status_code = ?", todayStart, todayEnd, "TODO")
	}
	// 已逾期(截止日期早于当前时间且处于待办或进行中状态) - 优化：使用索引友好的查询
	if overdue, ok := filter["overdue"]; ok && overdue.(bool) {
		query = query.Where("deadline < NOW() AND status_code IN (?, ?)", "TODO", "DOING")
	}
	if createUser, ok := filter["create_user"]; ok {
		query = query.Where("user_id = ?", createUser)
	}

	var total int64
	query.Count(&total)

	var tasks []models.TaskItem
	offset := (page - 1) * size
	if err := query.Order("create_time DESC").Offset(offset).Limit(size).Find(&tasks).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"records": tasks,
		"total":   total,
		"size":    size,
		"current": page,
		"pages":   (total + int64(size) - 1) / int64(size),
	}, nil
}

// GetTask 获取事项详情
func (s *TaskService) GetTask(id int64, userID int64, isAdmin bool) (*models.TaskItem, error) {
	db := s.getDB()

	var task models.TaskItem
	var err error
	if isAdmin {
		err = db.Where("id = ? AND deleted = ?", id, 0).First(&task).Error
	} else {
		err = db.Where("id = ? AND user_id = ? AND deleted = ?", id, userID, 0).First(&task).Error
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// CreateTask 鍒涘缓浜嬮」
func (s *TaskService) CreateTask(task *models.TaskItem) error {
	db := s.getDB()
	return db.Create(task).Error
}

// UpdateTask 鏇存柊浜嬮」
func (s *TaskService) UpdateTask(id int64, userID int64, updates map[string]interface{}) error {
	db := s.getDB()
	result := db.Model(&models.TaskItem{}).
		Where("id = ? AND deleted = ?", id, 0).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到匹配的记录或无权限更新")
	}
	return nil
}

// DeleteTask 鍒犻櫎浜嬮」(閫昏緫鍒犻櫎)
func (s *TaskService) DeleteTask(id int64, userID int64) error {
	db := s.getDB()
	return db.Model(&models.TaskItem{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("deleted", 1).Error
}

// BatchDeleteTasks 批量删除事项(閫昏緫鍒犻櫎)
func (s *TaskService) BatchDeleteTasks(ids []int64, userID int64) error {
	db := s.getDB()
	return db.Model(&models.TaskItem{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Update("deleted", 1).Error
}

// GetDashboardStats 获取仪表盘统计
func (s *TaskService) GetDashboardStats(userID int64, isAdmin bool) (map[string]interface{}, error) {
	db := s.getDB()

	// 优化：使用单次查询获取所有统计，减少数据库往返次数
	var stats struct {
		TodoCount      int64
		TodayTodoCount int64
		DoingCount     int64
		DoneCount      int64
		OverdueCount   int64
	}

	// 优化后的查询：使用 SUM(CASE WHEN ...) 在一次查询中获取所有计数
	// 同时优化日期比较，使用 deadline >= AND deadline < 替代 DATE(deadline) = CURDATE()
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	var query *gorm.DB
	if isAdmin {
		query = db.Model(&models.TaskItem{}).Where("deleted = ?", 0)
	} else {
		query = db.Model(&models.TaskItem{}).Where("user_id = ? AND deleted = ?", userID, 0)
	}

	err := query.
		Select(`
			SUM(CASE WHEN status_code = 'TODO' THEN 1 ELSE 0 END) as todo_count,
			SUM(CASE WHEN status_code = 'TODO' AND deadline >= ? AND deadline < ? THEN 1 ELSE 0 END) as today_todo_count,
			SUM(CASE WHEN status_code = 'DOING' THEN 1 ELSE 0 END) as doing_count,
			SUM(CASE WHEN status_code = 'DONE' THEN 1 ELSE 0 END) as done_count,
			SUM(CASE WHEN deadline < NOW() AND status_code IN ('TODO', 'DOING') THEN 1 ELSE 0 END) as overdue_count
		`, today, tomorrow).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"todayTodoCount": stats.TodayTodoCount,
		"todoCount":      stats.TodoCount,
		"doingCount":     stats.DoingCount,
		"doneCount":      stats.DoneCount,
		"overdueCount":   stats.OverdueCount,
	}, nil
}

// TransferTask 转交事项
func (s *TaskService) TransferTask(id int64, userID int64, targetUserID int64) error {
	db := s.getDB()
	result := db.Model(&models.TaskItem{}).
		Where("id = ? AND user_id = ? AND deleted = ?", id, userID, 0).
		Updates(map[string]interface{}{
			"user_id":    targetUserID,
			"create_user": targetUserID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到匹配的记录或无权限转交")
	}
	return nil
}

// SetTaskRequirement 设置事项与需求的关联（先删旧关联，再建新关联）
func (s *TaskService) SetTaskRequirement(taskID, requirementID, userID int64) error {
	db := s.getDB()

	// 先删除该事项的旧关联
	if err := db.Where("task_id = ?", taskID).Delete(&models.TaskRequirementRel{}).Error; err != nil {
		return err
	}

	// 查询需求编号（冗余存储，便于查询）
	var reqNo *string
	var req models.Requirement
	if err := db.Where("id = ? AND deleted = ?", requirementID, 0).First(&req).Error; err == nil {
		reqNo = req.RequirementNo
	}

	// 插入新关联
	rel := &models.TaskRequirementRel{
		TaskID:         taskID,
		RequirementID:  requirementID,
		RequirementNo:  reqNo,
		CreateUser:     &userID,
		UpdateUser:     &userID,
	}
	return db.Create(rel).Error
}

// RemoveTaskRequirement 删除事项与需求的关联
func (s *TaskService) RemoveTaskRequirement(taskID int64) error {
	db := s.getDB()
	return db.Where("task_id = ?", taskID).Delete(&models.TaskRequirementRel{}).Error
}

// GetTaskRequirement 获取事项关联的需求信息（ID、编号、标题）
func (s *TaskService) GetTaskRequirement(taskID int64) (requirementID *int64, requirementNo *string, requirementTitle *string, err error) {
	db := s.getDB()
	var rel models.TaskRequirementRel
	if e := db.Where("task_id = ?", taskID).First(&rel).Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, e
	}
	requirementID = &rel.RequirementID
	requirementNo = rel.RequirementNo

	// 查询需求标题
	var req models.Requirement
	if e := db.Where("id = ? AND deleted = ?", rel.RequirementID, 0).First(&req).Error; e == nil {
		requirementTitle = &req.Title
	}
	return requirementID, requirementNo, requirementTitle, nil
}

// ListTasksByRequirementID 根据需求ID查询关联的事项列表
func (s *TaskService) ListTasksByRequirementID(requirementID int64) ([]map[string]interface{}, error) {
	db := s.getDB()
	var tasks []models.TaskItem
	err := db.Where("id IN (SELECT task_id FROM task_requirement_rel WHERE requirement_id = ?)", requirementID).
		Where("deleted = ?", 0).
		Order("deadline DESC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		item := map[string]interface{}{
			"id":         t.ID,
			"title":      t.Title,
			"itemType":   t.ItemType,
			"priority":   t.Priority,
			"statusCode": t.StatusCode,
			"deadline":   t.Deadline,
		}
		result = append(result, item)
	}
	return result, nil
}
