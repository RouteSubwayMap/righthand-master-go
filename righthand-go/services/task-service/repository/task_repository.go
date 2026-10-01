package repository

import (
	"gorm.io/gorm"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
)

// TaskRepository 任务数据访问层
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓库实例
func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		db: database.GetDB(),
	}
}

// List 查询任务列表
func (r *TaskRepository) List(userID int64, page, pageSize int, status *int, priority string) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	query := r.db.Model(&models.Task{}).Where("user_id = ? AND deleted = ?", userID, 0)

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetByID 根据ID查询任务
func (r *TaskRepository) GetByID(id, userID int64) (*models.Task, error) {
	var task models.Task
	err := r.db.Where("id = ? AND user_id = ? AND deleted = ?", id, userID, 0).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// Create 创建任务
func (r *TaskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

// Update 更新任务
func (r *TaskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

// Delete 删除任务（逻辑删除）
func (r *TaskRepository) Delete(id, userID int64) error {
	return r.db.Model(&models.Task{}).Where("id = ? AND user_id = ?", id, userID).Update("deleted", 1).Error
}
