package dto

// CreateUserRequest 创建用户请求
// 初始密码不接受调用方指定，由 service 层统一使用默认密码（DefaultPassword）
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required"`
	RealName string  `json:"realName" binding:"required"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Status   int     `json:"status"`
	RoleIDs  []int64 `json:"roleIds"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	RealName string  `json:"realName"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Status   int     `json:"status"`
	RoleIDs  []int64 `json:"roleIds"`
	// 以下字段用于兼容前端传递的额外数据
	ID         int64  `json:"id"`
	Avatar     string `json:"avatar"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	CreateUser *int64 `json:"createUser"`
	UpdateUser *int64 `json:"updateUser"`
	Deleted    int    `json:"deleted"`
}
