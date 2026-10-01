package service

import (
	"errors"
	"log"
	"righthand-go/internal/models"
	"righthand-go/services/auth-service/repository"

	"golang.org/x/crypto/bcrypt"
)

// DefaultPassword 新增用户初始密码与重置密码的默认值（SSOT）
const DefaultPassword = "88888888"

// UserService 用户管理服务
type UserService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
		roleRepo: repository.NewRoleRepository(),
	}
}

// ListUsers 查询用户列表
func (s *UserService) ListUsers(page, pageSize int, username, realName string, status *int, roleID *int64) ([]models.SysUser, int64, error) {
	return s.userRepo.List(page, pageSize, username, realName, status, roleID)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id int64) (*models.SysUser, error) {
	return s.userRepo.GetByID(id)
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *models.SysUser, roleIDs []int64) error {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(user.Username)
	if err != nil {
		return errors.New("查询用户失败")
	}
	if existingUser != nil {
		return errors.New("用户名已存在")
	}

	// 初始密码统一为默认密码（新增用户不接受调用方指定密码）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(DefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user.Password = string(hashedPassword)

	// 设置默认状态
	if user.Status == 0 {
		user.Status = 1
	}

	// 创建用户
	if err := s.userRepo.Create(user); err != nil {
		return errors.New("创建用户失败")
	}

	// 分配角色
	if len(roleIDs) > 0 {
		if err := s.roleRepo.AssignRolesToUser(user.ID, roleIDs); err != nil {
			return errors.New("分配角色失败")
		}
	}

	return nil
}
func (s *UserService) UpdateUser(user *models.SysUser, roleIDs []int64) error {
	log.Printf("[UpdateUser] 开始更新用户，ID: %d", user.ID)
	existingUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		log.Printf("[UpdateUser] 查询用户失败: %v", err)
		return errors.New("查询用户失败")
	}
	if existingUser == nil {
		log.Printf("[UpdateUser] 用户不存在，ID: %d", user.ID)
		return errors.New("用户不存在")
	}

	if user.Username != "" {
		existingUser.Username = user.Username
	}
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[UpdateUser] 密码加密失败: %v", err)
			return errors.New("密码加密失败")
		}
		existingUser.Password = string(hashedPassword)
	}
	if user.RealName != "" {
		existingUser.RealName = user.RealName
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Phone != "" {
		existingUser.Phone = user.Phone
	}
	existingUser.Status = user.Status

	if err := s.userRepo.Update(existingUser); err != nil {
		log.Printf("[UpdateUser] 更新用户失败: %v", err)
		return errors.New("更新用户失败")
	}

	if len(roleIDs) > 0 {
		if err := s.roleRepo.AssignRolesToUser(user.ID, roleIDs); err != nil {
			log.Printf("[UpdateUser] 更新角色失败: %v", err)
			return errors.New("更新角色失败")
		}
	}

	log.Printf("[UpdateUser] 更新成功，ID: %d", user.ID)
	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int64) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.New("查询用户失败")
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 逻辑删除用户
	return s.userRepo.Delete(id)
}

// AssignRoles 为用户分配角色
func (s *UserService) AssignRoles(userID int64, roleIDs []int64) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("查询用户失败")
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	return s.roleRepo.AssignRolesToUser(userID, roleIDs)
}

// GetUserRoles 获取用户的角色列表
func (s *UserService) GetUserRoles(userID int64) ([]models.SysRole, error) {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("查询用户失败")
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 获取用户角色
	return s.roleRepo.GetRolesByUserID(userID)
}

// ResetPassword 重置用户密码为默认密码 88888888
func (s *UserService) ResetPassword(userID int64) error {
	// 检查用户是否存在
	existingUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		log.Printf("[ResetPassword] 查询用户失败: %v", err)
		return errors.New("查询用户失败")
	}
	if existingUser == nil {
		log.Printf("[ResetPassword] 用户不存在，ID: %d", userID)
		return errors.New("用户不存在")
	}

	// 设置默认密码并加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(DefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ResetPassword] 密码加密失败: %v", err)
		return errors.New("密码加密失败")
	}

	existingUser.Password = string(hashedPassword)

	// 更新用户
	if err := s.userRepo.Update(existingUser); err != nil {
		log.Printf("[ResetPassword] 更新用户失败: %v", err)
		return errors.New("重置密码失败")
	}

	log.Printf("[ResetPassword] 重置密码成功，ID: %d", userID)
	return nil
}

// ChangePassword 修改用户密码（用户自行修改，需验证旧密码）
func (s *UserService) ChangePassword(username, oldPassword, newPassword string) error {
	// 根据用户名查询用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		log.Printf("[ChangePassword] 查询用户失败: %v", err)
		return errors.New("查询用户失败")
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("原密码不正确")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ChangePassword] 密码加密失败: %v", err)
		return errors.New("密码加密失败")
	}

	user.Password = string(hashedPassword)

	// 更新用户
	if err := s.userRepo.Update(user); err != nil {
		log.Printf("[ChangePassword] 更新用户失败: %v", err)
		return errors.New("修改密码失败")
	}

	log.Printf("[ChangePassword] 修改密码成功，用户: %s", username)
	return nil
}
