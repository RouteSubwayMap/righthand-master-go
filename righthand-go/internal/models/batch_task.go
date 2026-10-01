package models

import "time"

// BatchTask 批量任务清单
type BatchTask struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	BankType        string    `gorm:"column:bank_type;type:varchar(50);not null" json:"bankType"`
	BankName        string    `gorm:"column:bank_name;type:varchar(100);not null" json:"bankName"`
	SystemName      string    `gorm:"column:system_name;type:varchar(50);not null" json:"systemName"`
	TaskType        string    `gorm:"column:task_type;type:varchar(50);not null" json:"taskType"`
	TaskName        string    `gorm:"column:task_name;type:varchar(200);not null" json:"taskName"`
	TaskDesc        string    `gorm:"column:task_desc;type:varchar(500)" json:"taskDesc"`
	ExecTime        string    `gorm:"column:exec_time;type:varchar(100)" json:"execTime"`
	UpstreamSystem  string    `gorm:"column:upstream_system;type:varchar(200)" json:"upstreamSystem"`
	UpstreamInfo    string    `gorm:"column:upstream_info;type:varchar(500)" json:"upstreamInfo"`
	DownstreamSystem string   `gorm:"column:downstream_system;type:varchar(200)" json:"downstreamSystem"`
	DownstreamInfo  string    `gorm:"column:downstream_info;type:varchar(500)" json:"downstreamInfo"`
	Remark          string    `gorm:"column:remark;type:varchar(500)" json:"remark"`

	CreateTime time.Time  `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64     `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time  `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64     `gorm:"column:update_user" json:"updateUser"`
	Deleted    int        `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (BatchTask) TableName() string {
	return "batch_task"
}
