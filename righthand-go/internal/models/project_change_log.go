package models

import (
	"time"
)

// ProjectChangeLog 项目级变更日志（仅记录关键变更，无软删除/审计更新字段）
type ProjectChangeLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID  int64     `gorm:"column:project_id;not null;index:idx_project_time,priority:1" json:"projectId"`
	Action     string    `gorm:"column:action;type:varchar(50);not null" json:"action"` // CREATE/UPDATE/STATUS/RISK_CHANGE/DELETE/MEMBER_ADD/MEMBER_REMOVE/WEEK_*/SUMMARY_*
	Detail     string    `gorm:"column:detail;type:varchar(1000)" json:"detail"`
	OperatorID int64     `gorm:"column:operator_id;not null" json:"operatorId"`
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP;index:idx_project_time,priority:2" json:"createTime"`
}

// TableName 指定表名
func (ProjectChangeLog) TableName() string {
	return "project_change_log"
}
