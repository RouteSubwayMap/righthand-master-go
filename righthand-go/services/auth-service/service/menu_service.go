package service

import (
	"errors"
	"righthand-go/internal/models"
	"righthand-go/services/auth-service/repository"
	"time"
)

// MenuService 菜单管理服务
type MenuService struct {
	menuRepo *repository.MenuRepository
}

// NewMenuService 创建菜单服务实例
func NewMenuService() *MenuService {
	return &MenuService{
		menuRepo: repository.NewMenuRepository(),
	}
}

// GetAllMenus 获取所有菜单（树形结构）
func (s *MenuService) GetAllMenus() ([]models.SysMenu, error) {
	return s.menuRepo.GetAllMenus()
}

// GetMenuByID 根据ID获取菜单
func (s *MenuService) GetMenuByID(id int64) (*models.SysMenu, error) {
	menus, err := s.menuRepo.GetMenusByIDs([]int64{id})
	if err != nil {
		return nil, err
	}
	if len(menus) == 0 {
		return nil, nil
	}
	return &menus[0], nil
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(menu *models.SysMenu) error {
	menu.CreateTime = time.Now()
	menu.UpdateTime = time.Now()
	return s.menuRepo.Create(menu)
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(menu *models.SysMenu) error {
	existingMenu, err := s.GetMenuByID(menu.ID)
	if err != nil {
		return errors.New("查询菜单失败")
	}
	if existingMenu == nil {
		return errors.New("菜单不存在")
	}

	return s.menuRepo.Update(menu)
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(id int64) error {
	existingMenu, err := s.GetMenuByID(id)
	if err != nil {
		return errors.New("查询菜单失败")
	}
	if existingMenu == nil {
		return errors.New("菜单不存在")
	}

	return s.menuRepo.Delete(id)
}

// GetMenuTree 获取菜单树形结构
func (s *MenuService) GetMenuTree() ([]models.SysMenu, error) {
	// 获取所有菜单
	menus, err := s.menuRepo.GetAllMenus()
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	return buildMenuTree(menus, 0), nil
}

// buildMenuTree 递归构建菜单树
func buildMenuTree(menus []models.SysMenu, parentID int64) []models.SysMenu {
	var tree []models.SysMenu
	for i := range menus {
		if menus[i].ParentID == parentID {
			// 递归查找子菜单
			children := buildMenuTree(menus, menus[i].ID)
			if len(children) > 0 {
				menus[i].Children = children
			}
			tree = append(tree, menus[i])
		}
	}
	return tree
}
