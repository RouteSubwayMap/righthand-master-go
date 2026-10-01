package models

import (
	"time"
)

// Project 项目
type Project struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectName    string     `gorm:"column:project_name;type:varchar(200);not null" json:"projectName"`
	ProjectKey     string     `gorm:"column:project_key;type:varchar(50);not null" json:"projectKey"`
	Description    string     `gorm:"column:description;type:text" json:"description"`
	PlanDesc       string     `gorm:"column:plan_desc;type:text" json:"planDesc"`
	ManagerID      int64      `gorm:"column:manager_id;not null;index:idx_manager" json:"managerId"`
	StartDate      time.Time  `gorm:"column:start_date;type:date;not null" json:"startDate"`
	EndDate        time.Time  `gorm:"column:end_date;type:date;not null;index:idx_date_range,priority:2" json:"endDate"`
	Status         string     `gorm:"column:status;type:varchar(20);not null;default:STARTING;index:idx_status_risk,priority:1" json:"status"`                    // STARTING/IN_PROGRESS/PAUSED/CLOSED/TERMINATED
	RiskLevel      string     `gorm:"column:risk_level;type:varchar(20);not null;default:MEDIUM;index:idx_status_risk,priority:2" json:"riskLevel"`               // HIGH/MEDIUM/LOW
	DeliveryStatus string     `gorm:"column:delivery_status;type:varchar(20);not null;default:NORMAL" json:"deliveryStatus"`                                      // NORMAL/AT_RISK/CANNOT_FINISH
	RiskNote       string     `gorm:"column:risk_note;type:varchar(1000)" json:"riskNote"`
	ProgressPct    int        `gorm:"column:progress_pct;not null;default:0" json:"progressPct"`       // 仅由 RecalcProjectProgress 计算
	WeekCount      int        `gorm:"column:week_count;not null;default:0" json:"weekCount"`           // 非跳过周计划数（分母）
	DoneWeekCount  int        `gorm:"column:done_week_count;not null;default:0" json:"doneWeekCount"`   // 已完成周数（分子）
	PausedAt       *time.Time `gorm:"column:paused_at" json:"pausedAt"`
	ClosedAt       *time.Time `gorm:"column:closed_at" json:"closedAt"`
	CreateTime     time.Time  `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser     *int64     `gorm:"column:create_user" json:"createUser"`
	UpdateTime     time.Time  `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser     *int64     `gorm:"column:update_user" json:"updateUser"`
	Deleted        int        `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "project"
}
