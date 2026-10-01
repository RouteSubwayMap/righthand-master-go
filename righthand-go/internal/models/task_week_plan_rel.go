package models

import (
	"time"
)

// TaskWeekPlanRel 任务-周计划多周关联表（一个任务可同时出现在多个周清单，物理删除）
type TaskWeekPlanRel struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID     int64     `gorm:"column:task_id;not null;uniqueIndex:uk_task_week" json:"taskId"`
	WeekPlanID int64     `gorm:"column:week_plan_id;not null;uniqueIndex:uk_task_week;index:idx_week_plan" json:"weekPlanId"`
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64    `gorm:"column:create_user" json:"createUser"`
}

// TableName 指定表名
func (TaskWeekPlanRel) TableName() string {
	return "task_week_plan_rel"
}
