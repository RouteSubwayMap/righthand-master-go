package models

import (
	"time"
)

// ProjectMember 项目成员
type ProjectMember struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID  int64     `gorm:"column:project_id;not null;index:idx_project_user,priority:1" json:"projectId"`
	UserID     int64     `gorm:"column:user_id;not null;index:idx_project_user,priority:2" json:"userId"`
	MemberRole string    `gorm:"column:member_role;type:varchar(20);not null;default:MEMBER" json:"memberRole"` // PM/MEMBER
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted    int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "project_member"
}
