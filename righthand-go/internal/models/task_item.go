package models

import (
	"time"
)

// TaskItem 统一事项表（支持任务/备忘/需求三种类型）
type TaskItem struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"column:user_id;not null;index" json:"userId"`
	
	// 基本信息
	Title      string     `gorm:"column:title;type:varchar(200);not null" json:"title"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	ItemType   string     `gorm:"column:item_type;type:varchar(20);default:'TASK'" json:"itemType"` // TASK/MEMO/REQUIREMENT
	
	// 分类和标签
	Category   string     `gorm:"column:category;type:varchar(50)" json:"category"`
	Priority   string     `gorm:"column:priority;type:varchar(20);default:'MEDIUM'" json:"priority"` // HIGH/MEDIUM/LOW
	Tags       string     `gorm:"column:tags;type:varchar(500)" json:"tags"` // 逗号分隔
	
	// 时间信息
	StartTime  *time.Time `gorm:"column:start_time" json:"startTime"`
	Deadline   *time.Time `gorm:"column:deadline" json:"deadline"`
	CompletedTime *time.Time `gorm:"column:completed_time" json:"completedTime"`
	
	// 简化状态（固定流程：TODO -> DOING -> DONE/CANCELLED）
	StatusCode string     `gorm:"column:status_code;type:varchar(20);default:'TODO'" json:"statusCode"` // TODO/DOING/DONE/CANCELLED
	Progress   int        `gorm:"column:progress;default:0" json:"progress"` // 0-100
	
	// 周期性任务字段
	RecurrenceType     string     `gorm:"column:recurrence_type;type:varchar(20);default:'NONE'" json:"recurrenceType"` // NONE/DAILY/WEEKLY/MONTHLY/YEARLY
	RecurrenceRule     string     `gorm:"column:recurrence_rule;type:json" json:"recurrenceRule"` // JSON格式存储重复规则
	RecurrenceEndType  string     `gorm:"column:recurrence_end_type;type:varchar(20);default:'NEVER'" json:"recurrenceEndType"` // NEVER/DATE/COUNT
	RecurrenceEndDate  *time.Time `gorm:"column:recurrence_end_date" json:"recurrenceEndDate"`
	RecurrenceCount    int        `gorm:"column:recurrence_count;default:10" json:"recurrenceCount"`
	ParentRecurrenceID *int64     `gorm:"column:parent_recurrence_id;index" json:"parentRecurrenceId"`
	
	// 审计字段
	CreateTime time.Time  `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64     `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time  `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64     `gorm:"column:update_user" json:"updateUser"`
	Deleted    int        `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (TaskItem) TableName() string {
	return "task_item"
}
