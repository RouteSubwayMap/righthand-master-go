package handler

import (
	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/auth-service/dto"
	"righthand-go/services/auth-service/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单管理处理器
type MenuHandler struct {
	menuService *service.MenuService
}

// NewMenuHandler 创建菜单处理器实例
func NewMenuHandler() *MenuHandler {
	return &MenuHandler{
		menuService: service.NewMenuService(),
	}
}

// GetMenus 获取所有菜单
func (h *MenuHandler) GetMenus(c *gin.Context) {
	menus, err := h.menuService.GetAllMenus()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, menus)
}

// GetMenu 获取菜单详情
func (h *MenuHandler) GetMenu(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的菜单ID")
		return
	}

	menu, err := h.menuService.GetMenuByID(id)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if menu == nil {
		utils.NotFound(c, "菜单不存在")
		return
	}

	utils.Success(c, menu)
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var req dto.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	menu := &models.SysMenu{
		ParentID:    req.ParentID,
		MenuName:    req.MenuName,
		MenuType:    req.MenuType,
		Path:        req.Path,
		Component:   req.Component,
		Permission:  req.Permission,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		ProductCode: req.ProductCode,
	}

	if err := h.menuService.CreateMenu(menu); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "创建成功", nil)
}

// UpdateMenu 更新菜单
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的菜单ID")
		return
	}

	var req dto.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	menu := &models.SysMenu{
		ID:          id,
		ParentID:    req.ParentID,
		MenuName:    req.MenuName,
		MenuType:    req.MenuType,
		Path:        req.Path,
		Component:   req.Component,
		Permission:  req.Permission,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
		ProductCode: req.ProductCode,
	}

	if err := h.menuService.UpdateMenu(menu); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteMenu 删除菜单
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的菜单ID")
		return
	}

	if err := h.menuService.DeleteMenu(id); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GetMenuTree 获取菜单树形结构
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
	tree, err := h.menuService.GetMenuTree()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, tree)
}
