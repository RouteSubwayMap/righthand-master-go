package service

import (
	"errors"
	"righthand-go/internal/models"
	"righthand-go/services/auth-service/repository"
)

// RoleService 角色管理服务
type RoleService struct {
	roleRepo *repository.RoleRepository
	menuRepo *repository.MenuRepository
}

// NewRoleService 创建角色服务实例
func NewRoleService() *RoleService {
	return &RoleService{
		roleRepo: repository.NewRoleRepository(),
		menuRepo: repository.NewMenuRepository(),
	}
}

// ListRoles 查询角色列表
func (s *RoleService) ListRoles(page, pageSize int, roleName string) ([]models.SysRole, int64, error) {
	return s.roleRepo.List(page, pageSize, roleName)
}

// GetRoleByID 根据ID获取角色
func (s *RoleService) GetRoleByID(id int64) (*models.SysRole, error) {
	return s.roleRepo.GetByID(id)
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(role *models.SysRole) error {
	return s.roleRepo.Create(role)
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(role *models.SysRole) error {
	existingRole, err := s.roleRepo.GetByID(role.ID)
	if err != nil {
		return errors.New("查询角色失败")
	}
	if existingRole == nil {
		return errors.New("角色不存在")
	}

	return s.roleRepo.Update(role)
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id int64) error {
	existingRole, err := s.roleRepo.GetByID(id)
	if err != nil {
		return errors.New("查询角色失败")
	}
	if existingRole == nil {
		return errors.New("角色不存在")
	}

	return s.roleRepo.Delete(id)
}

// AssignMenus 为角色分配菜单
func (s *RoleService) AssignMenus(roleID int64, menuIDs []int64) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.New("查询角色失败")
	}
	if role == nil {
		return errors.New("角色不存在")
	}

	return s.menuRepo.AssignMenusToRole(roleID, menuIDs)
}

// GetRoleMenus 获取角色的菜单ID列表
func (s *RoleService) GetRoleMenus(roleID int64) ([]int64, error) {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, errors.New("查询角色失败")
	}
	if role == nil {
		return nil, errors.New("角色不存在")
	}

	return s.menuRepo.GetMenusByRoleID(roleID)
}
