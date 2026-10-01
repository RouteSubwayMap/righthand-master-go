package service

import (
	"righthand-go/internal/database"
	"righthand-go/internal/models"

	"gorm.io/gorm"
)

type BatchTaskService struct {
	db *gorm.DB
}

func NewBatchTaskService() *BatchTaskService {
	return &BatchTaskService{}
}

func (s *BatchTaskService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

// ListBatchTasks 获取批量任务列表（分页 + 多条件查询）
func (s *BatchTaskService) ListBatchTasks(page, size int, bankType, bankName, systemName, taskType, taskName, taskDesc, upstreamSystem, downstreamSystem string) (map[string]interface{}, error) {
	db := s.getDB()

	query := db.Model(&models.BatchTask{}).Where("deleted = ?", 0)

	if bankType != "" {
		query = query.Where("bank_type = ?", bankType)
	}
	if bankName != "" {
		query = query.Where("bank_name = ?", bankName)
	}
	if systemName != "" {
		query = query.Where("system_name = ?", systemName)
	}
	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	if taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}
	if taskDesc != "" {
		query = query.Where("task_desc LIKE ?", "%"+taskDesc+"%")
	}
	if upstreamSystem != "" {
		query = query.Where("upstream_system LIKE ?", "%"+upstreamSystem+"%")
	}
	if downstreamSystem != "" {
		query = query.Where("downstream_system LIKE ?", "%"+downstreamSystem+"%")
	}

	var total int64
	query.Count(&total)

	var tasks []models.BatchTask
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

// GetBatchTask 获取批量任务详情
func (s *BatchTaskService) GetBatchTask(id int64) (*models.BatchTask, error) {
	db := s.getDB()

	var task models.BatchTask
	if err := db.Where("id = ? AND deleted = ?", id, 0).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateBatchTask 创建批量任务
func (s *BatchTaskService) CreateBatchTask(task *models.BatchTask) error {
	db := s.getDB()
	return db.Create(task).Error
}

// UpdateBatchTask 更新批量任务
func (s *BatchTaskService) UpdateBatchTask(id int64, updates map[string]interface{}) error {
	db := s.getDB()
	return db.Model(&models.BatchTask{}).
		Where("id = ? AND deleted = ?", id, 0).
		Updates(updates).Error
}

// DeleteBatchTask 删除批量任务（逻辑删除）
func (s *BatchTaskService) DeleteBatchTask(id int64) error {
	db := s.getDB()
	return db.Model(&models.BatchTask{}).
		Where("id = ?", id).
		Update("deleted", 1).Error
}

// BatchImportTasks 批量导入任务
func (s *BatchTaskService) BatchImportTasks(tasks []models.BatchTask) (int, error) {
	db := s.getDB()
	if len(tasks) == 0 {
		return 0, nil
	}
	// 分批插入，每批100条
	if err := db.CreateInBatches(tasks, 100).Error; err != nil {
		return 0, err
	}
	return len(tasks), nil
}
