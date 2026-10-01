package models

import (
	"time"
)

// ProjectWeeklyPlan 项目周计划
type ProjectWeeklyPlan struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID  int64      `gorm:"column:project_id;not null;index:idx_project_sort,priority:1" json:"projectId"`
	WeekNo     int        `gorm:"column:week_no;not null" json:"weekNo"` // 项目内周序号（服务端 max+1 生成）
	Title      string     `gorm:"column:title;type:varchar(200);not null" json:"title"`
	PlanDesc   string     `gorm:"column:plan_desc;type:text" json:"planDesc"`
	StartDate  time.Time  `gorm:"column:start_date;type:date;not null" json:"startDate"`
	EndDate    time.Time  `gorm:"column:end_date;type:date;not null;index:idx_date" json:"endDate"`
	OwnerID    int64      `gorm:"column:owner_id;not null;index:idx_owner_status,priority:1" json:"ownerId"`
	Status     string     `gorm:"column:status;type:varchar(20);not null;default:NOT_STARTED;index:idx_owner_status,priority:2" json:"status"` // NOT_STARTED/IN_PROGRESS/DONE/SKIPPED
	Remark     string     `gorm:"column:remark;type:varchar(1000)" json:"remark"`
	Summary    *string    `gorm:"column:summary;type:text" json:"summary"` // JSON: finished/nextPlan/risks/support
	SummaryBy  *int64     `gorm:"column:summary_by" json:"summaryBy"`
	SummaryAt  *time.Time `gorm:"column:summary_at" json:"summaryAt"` // 非空即锁定
	SortOrder  int        `gorm:"column:sort_order;not null;default:0;index:idx_project_sort,priority:2" json:"sortOrder"`
	CreateTime time.Time  `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64     `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time  `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64     `gorm:"column:update_user" json:"updateUser"`
	Deleted    int        `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

// TableName 指定表名
func (ProjectWeeklyPlan) TableName() string {
	return "project_weekly_plan"
}
