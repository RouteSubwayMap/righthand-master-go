package handler

import (
	"log"
	"righthand-go/internal/utils"
	"righthand-go/services/auth-service/dto"
	"righthand-go/services/auth-service/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.Success(c, response)
}

// GetInfo 获取用户信息
func (h *AuthHandler) GetInfo(c *gin.Context) {
	// 从上下文获取用户ID（由中间件设置）
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	userInfo, err := h.authService.GetUserInfo(userID.(int64))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, userInfo)
}

// GetMenus 获取用户菜单
func (h *AuthHandler) GetMenus(c *gin.Context) {
	// 从上下文获取用户ID和管理员标识
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	isAdmin, _ := c.Get("isAdmin")

	menus, err := h.authService.GetUserMenus(userID.(int64), isAdmin != nil && isAdmin.(bool))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	// 打印菜单树结构用于调试
	for _, menu := range menus {
		if menu.MenuType == 0 { // 目录菜单
			childCount := 0
			if menu.Children != nil {
				childCount = len(menu.Children)
			}
			log.Printf("目录菜单: %s (ID:%d), 子菜单数量: %d", menu.MenuName, menu.ID, childCount)
		}
	}

	utils.Success(c, menus)
}

// Logout 退出登录
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT是无状态的，客户端删除token即可
	utils.Success(c, nil)
}
