package models

import (
	"time"
)

// ProjectStage 项目阶段
type ProjectStage struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID   int64     `gorm:"column:project_id;not null;index:idx_project_start,priority:1" json:"projectId"`
	StageName   string    `gorm:"column:stage_name;type:varchar(100);not null" json:"stageName"` // 同一项目未删除范围内唯一（服务层校验）
	Description string    `gorm:"column:description;type:varchar(500)" json:"description"`
	StartDate   time.Time `gorm:"column:start_date;type:date;not null;index:idx_project_start,priority:2" json:"startDate"`
	EndDate     time.Time `gorm:"column:end_date;type:date;not null" json:"endDate"`
	CreateTime  time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser  *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime  time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser  *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted     int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

// TableName 指定表名
func (ProjectStage) TableName() string {
	return "project_stage"
}
