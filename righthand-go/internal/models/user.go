package models

import (
	"time"
)

// SysUser 系统用户
type SysUser struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username   string    `gorm:"column:username;type:varchar(50);uniqueIndex;not null" json:"username"`
	Password   string    `gorm:"column:password;type:varchar(100);not null" json:"-"`
	RealName   string    `gorm:"column:real_name;type:varchar(50)" json:"realName"`
	Email      string    `gorm:"column:email;type:varchar(100)" json:"email"`
	Phone      string    `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Avatar     string    `gorm:"column:avatar;type:varchar(255)" json:"avatar"`
	Status     int       `gorm:"column:status;type:tinyint;default:1" json:"status"`
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted    int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

// SysRole 系统角色
type SysRole struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleName    string    `gorm:"column:role_name;type:varchar(50);not null" json:"roleName"`
	RoleCode    string    `gorm:"column:role_code;type:varchar(50);not null" json:"roleCode"`
	Description string    `gorm:"column:description;type:varchar(255)" json:"description"`
	CreateTime  time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser  *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime  time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser  *int64    `gorm:"column:update_user" json:"updateUser"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

// SysUserRole 用户角色关联
type SysUserRole struct {
	UserID int64 `gorm:"column:user_id;primaryKey" json:"userId"`
	RoleID int64 `gorm:"column:role_id;primaryKey" json:"roleId"`
}

func (SysUserRole) TableName() string {
	return "sys_user_role"
}

// SysMenu 系统菜单
type SysMenu struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuName    string    `gorm:"column:menu_name;type:varchar(50);not null" json:"menuName"`
	ProductCode string    `gorm:"column:product_code;type:varchar(50);default:'auth';index" json:"productCode"`
	MenuType    int       `gorm:"column:menu_type;type:tinyint;not null" json:"menuType"`
	ParentID    int64     `gorm:"column:parent_id;default:0" json:"parentId"`
	Path        string    `gorm:"column:path;type:varchar(100)" json:"path"`
	Component   string    `gorm:"column:component;type:varchar(100)" json:"component"`
	Permission  string    `gorm:"column:permission;type:varchar(50)" json:"permission"`
	Icon        string    `gorm:"column:icon;type:varchar(50)" json:"icon"`
	SortOrder   int       `gorm:"column:sort_order;default:0" json:"sortOrder"`
	CreateTime  time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser  *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime  time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser  *int64    `gorm:"column:update_user" json:"updateUser"`

	// 非数据库字段，用于树形结构
	Children []SysMenu `gorm:"-" json:"children,omitempty"`
}

func (SysMenu) TableName() string {
	return "sys_menu"
}

// SysRoleMenu 角色菜单关联
type SysRoleMenu struct {
	RoleID int64 `gorm:"column:role_id;primaryKey" json:"roleId"`
	MenuID int64 `gorm:"column:menu_id;primaryKey" json:"menuId"`
}

func (SysRoleMenu) TableName() string {
	return "sys_role_menu"
}

// SysDict 数据字典
type SysDict struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	DictCode   string    `gorm:"column:dict_code;type:varchar(50);not null" json:"dictCode"`
	DictName   string    `gorm:"column:dict_name;type:varchar(100);not null" json:"dictName"`
	DictType   string    `gorm:"column:dict_type;type:varchar(50);not null;index:idx_dict_type_value" json:"dictType"`
	DictValue  string    `gorm:"column:dict_value;type:varchar(50);not null;index:idx_dict_type_value" json:"dictValue"`
	DictLabel  string    `gorm:"column:dict_label;type:varchar(100);not null" json:"dictLabel"`
	SortOrder  int       `gorm:"column:sort_order;default:0" json:"sortOrder"`
	Status     int       `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Remark     string    `gorm:"column:remark;type:varchar(255)" json:"remark"`
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted    int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (SysDict) TableName() string {
	return "sys_dict"
}
