package repository

import (
	"righthand-go/internal/database"
	"righthand-go/internal/models"

	"gorm.io/gorm"
)

// RoleRepository 角色数据访问层
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库实例
func NewRoleRepository() *RoleRepository {
	return &RoleRepository{
		db: database.GetDB(),
	}
}

// GetRoleIDsByUserID 获取用户的角色ID列表
func (r *RoleRepository) GetRoleIDsByUserID(userID int64) ([]int64, error) {
	var roleIDs []int64
	err := r.db.Model(&models.SysUserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

// GetByID 根据ID查询角色
func (r *RoleRepository) GetByID(id int64) (*models.SysRole, error) {
	var role models.SysRole
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// List 查询角色列表
func (r *RoleRepository) List(page, pageSize int, roleName string) ([]models.SysRole, int64, error) {
	var roles []models.SysRole
	var total int64

	query := r.db.Model(&models.SysRole{})

	if roleName != "" {
		query = query.Where("role_name LIKE ?", "%"+roleName+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// Create 创建角色
func (r *RoleRepository) Create(role *models.SysRole) error {
	return r.db.Create(role).Error
}

// Update 更新角色
func (r *RoleRepository) Update(role *models.SysRole) error {
	return r.db.Model(&models.SysRole{}).
		Where("id = ?", role.ID).
		Updates(map[string]interface{}{
			"role_name":   role.RoleName,
			"role_code":   role.RoleCode,
			"description": role.Description,
		}).Error
}

// Delete 删除角色
func (r *RoleRepository) Delete(id int64) error {
	tx := r.db.Begin()
	if err := tx.Delete(&models.SysRole{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("role_id = ?", id).Delete(&models.SysUserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("role_id = ?", id).Delete(&models.SysRoleMenu{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// AssignRolesToUser 为用户分配角色
func (r *RoleRepository) AssignRolesToUser(userID int64, roleIDs []int64) error {
	tx := r.db.Begin()

	// 删除旧的角色关联
	if err := tx.Where("user_id = ?", userID).Delete(&models.SysUserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 添加新的角色关联
	for _, roleID := range roleIDs {
		userRole := models.SysUserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := tx.Create(&userRole).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetRolesByUserID 获取用户的角色列表
func (r *RoleRepository) GetRolesByUserID(userID int64) ([]models.SysRole, error) {
	var roles []models.SysRole
	err := r.db.Table("sys_role").
		Joins("INNER JOIN sys_user_role ON sys_role.id = sys_user_role.role_id").
		Where("sys_user_role.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}
