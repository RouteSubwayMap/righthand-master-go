package models

import (
	"time"
)

// TaskRequirementRel 事项-需求关联表（一个事项关联一个需求）
type TaskRequirementRel struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID        int64     `gorm:"column:task_id;not null;uniqueIndex:uk_task_requirement" json:"taskId"`
	RequirementID int64     `gorm:"column:requirement_id;not null;uniqueIndex:uk_task_requirement" json:"requirementId"`
	RequirementNo *string   `gorm:"column:requirement_no;type:varchar(50)" json:"requirementNo"`
	CreateTime    time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser    *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime    time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser    *int64    `gorm:"column:update_user" json:"updateUser"`
}

// TableName 指定表名
func (TaskRequirementRel) TableName() string {
	return "task_requirement_rel"
}
