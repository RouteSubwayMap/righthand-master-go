package models

import (
	"time"
)

// Task 任务
type Task struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"column:user_id;not null;index" json:"userId"`
	Title       string    `gorm:"column:title;type:varchar(100);not null" json:"title"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Status      int       `gorm:"column:status;type:tinyint;default:0" json:"status"`
	Priority    string    `gorm:"column:priority;type:varchar(20);default:'MEDIUM'" json:"priority"`
	StartTime   *time.Time `gorm:"column:start_time" json:"startTime"`
	EndTime     *time.Time `gorm:"column:end_time" json:"endTime"`
	Progress    int       `gorm:"column:progress;default:0" json:"progress"`
	Tags        string    `gorm:"column:tags;type:json" json:"tags"`
	Category    string    `gorm:"column:category;type:varchar(50)" json:"category"`
	CreateTime  time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser  *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime  time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser  *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted     int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (Task) TableName() string {
	return "task_item"
}

// Message 消息
type Message struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"column:user_id;not null;index" json:"userId"`
	Title      string     `gorm:"column:title;type:varchar(200);not null" json:"title"`
	Content    string     `gorm:"column:content;type:text" json:"content"`
	Type       string     `gorm:"column:type;type:varchar(50)" json:"type"`
	IsRead     bool       `gorm:"column:is_read;default:false" json:"isRead"`
	BusinessID *int64     `gorm:"column:business_id" json:"businessId"` // 关联业务 ID（如项目 ID）
	CreateTime time.Time  `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	ReadTime   *time.Time `gorm:"column:read_time" json:"readTime"`
	CreateUser *int64     `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time  `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64     `gorm:"column:update_user" json:"updateUser"`
	Deleted    int        `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (Message) TableName() string {
	return "sys_message"
}
