package dto

import "righthand-go/internal/models"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string          `json:"token"`
	User  *models.SysUser `json:"user"`
	Roles []models.SysRole `json:"roles"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID       int64            `json:"id"`
	Username string           `json:"username"`
	RealName string           `json:"realName"`
	Email    string           `json:"email"`
	Phone    string           `json:"phone"`
	Avatar   string           `json:"avatar"`
	Roles    []models.SysRole `json:"roles"`
	Products []string         `json:"products"`
}
