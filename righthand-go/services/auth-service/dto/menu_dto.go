package dto

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	ParentID    int64  `json:"parentId"`
	MenuName    string `json:"menuName"`
	MenuType    int    `json:"menuType"`
	Path        string `json:"path"`
	Component   string `json:"component"`
	Permission  string `json:"permission"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
	ProductCode string `json:"productCode"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	ParentID    int64  `json:"parentId"`
	MenuName    string `json:"menuName"`
	MenuType    int    `json:"menuType"`
	Path        string `json:"path"`
	Component   string `json:"component"`
	Permission  string `json:"permission"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sortOrder"`
	ProductCode string `json:"productCode"`
	// 以下字段用于兼容前端传递的额外数据
	ID         int64  `json:"id"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	CreateUser *int64 `json:"createUser"`
	UpdateUser *int64 `json:"updateUser"`
}
