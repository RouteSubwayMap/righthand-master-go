package service

import (
	"errors"
	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/auth-service/dto"
	"righthand-go/services/auth-service/repository"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	menuRepo *repository.MenuRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
		roleRepo: repository.NewRoleRepository(),
		menuRepo: repository.NewMenuRepository(),
	}
}

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 查询用户
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 获取用户角色
	roleIDs, err := s.roleRepo.GetRoleIDsByUserID(user.ID)
	if err != nil {
		return nil, errors.New("获取用户角色失败")
	}

	// 检查是否为管理员
	isAdmin := false
	var roles []models.SysRole
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(roleID)
		if err != nil {
			continue
		}
		if role != nil {
			roles = append(roles, *role)
			if role.RoleCode == "ADMIN" {
				isAdmin = true
			}
		}
	}

	// 构建角色字符串
	var roleCodes []string
	for _, role := range roles {
		roleCodes = append(roleCodes, role.RoleCode)
	}
	rolesStr := strings.Join(roleCodes, ",")

	// 生成Token
	token, err := utils.GenerateToken(user.ID, user.Username, user.RealName, isAdmin, rolesStr)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	return &dto.LoginResponse{
		Token: token,
		User:  user,
		Roles: roles,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID int64) (*dto.UserInfoResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}

	// 获取用户角色
	roleIDs, err := s.roleRepo.GetRoleIDsByUserID(userID)
	if err != nil {
		return nil, errors.New("获取用户角色失败")
	}

	var roles []models.SysRole
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetByID(roleID)
		if err != nil {
			continue
		}
		if role != nil {
			roles = append(roles, *role)
		}
	}

	// 获取用户可访问的产品列表
	var menuIDs []int64
	for _, roleID := range roleIDs {
		ids, err := s.menuRepo.GetMenuIDsByRoleID(roleID)
		if err != nil {
			continue
		}
		menuIDs = append(menuIDs, ids...)
	}

	// 去重
	menuIDMap := make(map[int64]bool)
	for _, id := range menuIDs {
		menuIDMap[id] = true
	}
	var uniqueMenuIDs []int64
	for id := range menuIDMap {
		uniqueMenuIDs = append(uniqueMenuIDs, id)
	}

	// 获取菜单并提取产品编码
	menus, err := s.menuRepo.GetMenusByIDs(uniqueMenuIDs)
	if err != nil {
		return nil, errors.New("获取菜单失败")
	}

	productMap := make(map[string]bool)
	for _, menu := range menus {
		if menu.MenuType != 2 && menu.ProductCode != "" { // 排除按钮
			productMap[menu.ProductCode] = true
		}
	}

	var products []string
	for product := range productMap {
		products = append(products, product)
	}

	return &dto.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		RealName: user.RealName,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Roles:    roles,
		Products: products,
	}, nil
}

// GetUserMenus 获取用户菜单
func (s *AuthService) GetUserMenus(userID int64, isAdmin bool) ([]models.SysMenu, error) {
	if isAdmin {
		// 管理员返回所有菜单（树形结构）
		menus, err := s.menuRepo.GetAllMenus()
		if err != nil {
			return nil, err
		}
		return buildMenuTree(menus, 0), nil
	}

	// 普通用户返回分配的菜单
	roleIDs, err := s.roleRepo.GetRoleIDsByUserID(userID)
	if err != nil {
		return nil, errors.New("获取用户角色失败")
	}

	var menuIDs []int64
	for _, roleID := range roleIDs {
		ids, err := s.menuRepo.GetMenuIDsByRoleID(roleID)
		if err != nil {
			continue
		}
		menuIDs = append(menuIDs, ids...)
	}

	if len(menuIDs) == 0 {
		return []models.SysMenu{}, nil
	}

	// 获取菜单列表并构建树形结构
	menus, err := s.menuRepo.GetMenusByIDs(menuIDs)
	if err != nil {
		return nil, err
	}

	return buildMenuTree(menus, 0), nil
}
