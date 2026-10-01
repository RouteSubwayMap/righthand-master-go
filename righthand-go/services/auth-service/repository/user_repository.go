package repository

import (
	"righthand-go/internal/database"
	"righthand-go/internal/models"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: database.GetDB(),
	}
}

// GetByUsername 根据用户名查询用户
func (r *UserRepository) GetByUsername(username string) (*models.SysUser, error) {
	var user models.SysUser
	err := r.db.Where("username = ? AND deleted = ?", username, 0).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(id int64) (*models.SysUser, error) {
	var user models.SysUser
	err := r.db.Where("id = ? AND deleted = ?", id, 0).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// List 查询用户列表
func (r *UserRepository) List(page, pageSize int, username, realName string, status *int, roleID *int64) ([]models.SysUser, int64, error) {
	var users []models.SysUser
	var total int64

	query := r.db.Model(&models.SysUser{}).Where("deleted = ?", 0)

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if realName != "" {
		query = query.Where("real_name LIKE ?", "%"+realName+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if roleID != nil && *roleID > 0 {
		query = query.Joins("JOIN sys_user_role ON sys_user.id = sys_user_role.user_id").
			Where("sys_user_role.role_id = ?", *roleID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *models.SysUser) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(user *models.SysUser) error {
	return r.db.Save(user).Error
}

// Delete 删除用户（逻辑删除）
func (r *UserRepository) Delete(id int64) error {
	return r.db.Model(&models.SysUser{}).Where("id = ?", id).Update("deleted", 1).Error
}
