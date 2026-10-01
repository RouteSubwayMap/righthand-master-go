package handler

import (
	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/auth-service/dto"
	"righthand-go/services/auth-service/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色管理处理器
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建角色处理器实例
func NewRoleHandler() *RoleHandler {
	return &RoleHandler{
		roleService: service.NewRoleService(),
	}
}

// ListRoles 查询角色列表
func (h *RoleHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	roleName := c.Query("roleName")

	roles, total, err := h.roleService.ListRoles(page, pageSize, roleName)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, utils.PageResult{
		Total:    total,
		PageNum:  page,
		PageSize: pageSize,
		List:     roles,
	})
}

// GetAllRoles 获取所有角色（不分页）
func (h *RoleHandler) GetAllRoles(c *gin.Context) {
	roles, _, err := h.roleService.ListRoles(1, 1000, "")
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, roles)
}

// GetRole 获取角色详情
func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的角色ID")
		return
	}

	role, err := h.roleService.GetRoleByID(id)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if role == nil {
		utils.NotFound(c, "角色不存在")
		return
	}

	utils.Success(c, role)
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	role := &models.SysRole{
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
	}

	if err := h.roleService.CreateRole(role); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "创建成功", nil)
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的角色ID")
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	role := &models.SysRole{
		ID:          id,
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
	}

	if err := h.roleService.UpdateRole(role); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的角色ID")
		return
	}

	if err := h.roleService.DeleteRole(id); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// AssignMenus 为角色分配菜单
func (h *RoleHandler) AssignMenus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的角色ID")
		return
	}

	var req struct {
		MenuIDs []int64 `json:"menuIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.roleService.AssignMenus(id, req.MenuIDs); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "分配菜单成功", nil)
}

// GetRoleMenus 获取角色的菜单ID列表
func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的角色ID")
		return
	}

	menuIDs, err := h.roleService.GetRoleMenus(id)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, menuIDs)
}
