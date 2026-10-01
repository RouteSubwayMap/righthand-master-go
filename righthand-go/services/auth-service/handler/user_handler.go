package handler

import (
	"log"
	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/auth-service/dto"
	"righthand-go/services/auth-service/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// ListUsers 查询用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	username := c.Query("username")
	realName := c.Query("realName")

	// 处理 status 参数
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		statusVal, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &statusVal
		}
	}

	// 处理 roleId 参数
	var roleID *int64
	if roleIdStr := c.Query("roleId"); roleIdStr != "" {
		roleIdVal, err := strconv.ParseInt(roleIdStr, 10, 64)
		if err == nil {
			roleID = &roleIdVal
		}
	}

	users, total, err := h.userService.ListUsers(page, pageSize, username, realName, status, roleID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, utils.PageResult{
		Total:    total,
		PageNum:  page,
		PageSize: pageSize,
		List:     users,
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的用户ID")
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if user == nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	user := &models.SysUser{
		Username: req.Username,
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   req.Status,
	}

	if err := h.userService.CreateUser(user, req.RoleIDs); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "创建成功", nil)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的用户ID")
		return
	}

	var req dto.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[UpdateUser] 参数绑定失败: %v", err)
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user := &models.SysUser{
		ID:       id,
		Username: req.Username,
		Password: req.Password,
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   req.Status,
	}

	if err := h.userService.UpdateUser(user, req.RoleIDs); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// AssignRoles 为用户分配角色
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		RoleIDs []int64 `json:"roleIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.userService.AssignRoles(id, req.RoleIDs); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "分配角色成功", nil)
}

// GetUserRoles 获取用户的角色列表
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的用户ID")
		return
	}

	roles, err := h.userService.GetUserRoles(id)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, roles)
}

// ResetPassword 重置用户密码（管理员操作）
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.userService.ResetPassword(id); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "重置密码成功", nil)
}

// ChangePassword 修改密码（用户自行修改，需验证旧密码）
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if req.Username == "" || req.OldPassword == "" || req.NewPassword == "" {
		utils.BadRequest(c, "用户名、原密码和新密码不能为空")
		return
	}

	if err := h.userService.ChangePassword(req.Username, req.OldPassword, req.NewPassword); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "密码修改成功", nil)
}
