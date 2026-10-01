package repository

import (
	"gorm.io/gorm"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
)

// MenuRepository 菜单数据访问层
type MenuRepository struct {
	db *gorm.DB
}

// NewMenuRepository 创建菜单仓库实例
func NewMenuRepository() *MenuRepository {
	return &MenuRepository{
		db: database.GetDB(),
	}
}

// GetAllMenus 获取所有菜单
func (r *MenuRepository) GetAllMenus() ([]models.SysMenu, error) {
	var menus []models.SysMenu
	err := r.db.Order("sort_order ASC, id ASC").Find(&menus).Error
	return menus, err
}

// GetMenusByIDs 根据ID列表获取菜单
func (r *MenuRepository) GetMenusByIDs(ids []int64) ([]models.SysMenu, error) {
	var menus []models.SysMenu
	err := r.db.Where("id IN ?", ids).Order("sort_order ASC, id ASC").Find(&menus).Error
	return menus, err
}

// GetMenuIDsByRoleID 获取角色的菜单ID列表
func (r *MenuRepository) GetMenuIDsByRoleID(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := r.db.Model(&models.SysRoleMenu{}).
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

// AssignMenusToRole 为角色分配菜单
func (r *MenuRepository) AssignMenusToRole(roleID int64, menuIDs []int64) error {
	tx := r.db.Begin()
	
	// 删除旧的菜单关联
	if err := tx.Where("role_id = ?", roleID).Delete(&models.SysRoleMenu{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 添加新的菜单关联
	for _, menuID := range menuIDs {
		roleMenu := models.SysRoleMenu{
			RoleID: roleID,
			MenuID: menuID,
		}
		if err := tx.Create(&roleMenu).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetMenusByRoleID 获取角色的菜单ID列表
func (r *MenuRepository) GetMenusByRoleID(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := r.db.Table("sys_role_menu").
		Where("role_id = ?", roleID).
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

// Create 创建菜单
func (r *MenuRepository) Create(menu *models.SysMenu) error {
	return r.db.Create(menu).Error
}

// Update 更新菜单
func (r *MenuRepository) Update(menu *models.SysMenu) error {
	return r.db.Model(&models.SysMenu{}).
		Where("id = ?", menu.ID).
		Updates(map[string]interface{}{
			"parent_id":    menu.ParentID,
			"menu_name":    menu.MenuName,
			"menu_type":    menu.MenuType,
			"path":         menu.Path,
			"component":    menu.Component,
			"permission":   menu.Permission,
			"icon":         menu.Icon,
			"sort_order":   menu.SortOrder,
			"product_code": menu.ProductCode,
		}).Error
}

// Delete 删除菜单
func (r *MenuRepository) Delete(id int64) error {
	tx := r.db.Begin()
	if err := tx.Delete(&models.SysMenu{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("menu_id = ?", id).Delete(&models.SysRoleMenu{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
