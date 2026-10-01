package dto

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleName    string `json:"roleName" binding:"required"`
	RoleCode    string `json:"roleCode" binding:"required"`
	Description string `json:"description"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	RoleName    string `json:"roleName" binding:"required"`
	RoleCode    string `json:"roleCode"`
	Description string `json:"description"`
	// 以下字段用于兼容前端传递的额外数据
	ID         int64  `json:"id"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	CreateUser *int64 `json:"createUser"`
	UpdateUser *int64 `json:"updateUser"`
}