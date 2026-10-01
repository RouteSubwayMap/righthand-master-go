package models

import (
	"time"
	"database/sql"
)

// Requirement 需求模型(基于Java版表结构)
type Requirement struct {
	ID                int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Title             string          `gorm:"type:varchar(200);not null" json:"title"`
	BankName          *string         `gorm:"type:varchar(100)" json:"bankName"`
	RequirementNo     *string         `gorm:"type:varchar(50)" json:"requirementNo"`
	Priority          *string         `gorm:"type:varchar(20);default:'MEDIUM'" json:"priority"` // HIGH/MEDIUM/LOW
	RiskLevel         *string         `gorm:"type:varchar(20);default:'MEDIUM'" json:"riskLevel"` // HIGH/MEDIUM/LOW
	RequirementType   *string         `gorm:"type:varchar(50);default:'INTERNAL'" json:"requirementType"` // INTERNAL/BANK
	Status            *string         `gorm:"type:varchar(50);default:'PENDING_CLARIFY';index" json:"status"` // PENDING_CLARIFY/PENDING_SCHEDULING/DEVELOPING/TESTING/COMPLETED
	ScheduleDate      *string         `gorm:"type:varchar(50)" json:"scheduleDate"`
	MainSystem        *string         `gorm:"type:varchar(100)" json:"mainSystem"`
	Developer         *string         `gorm:"type:varchar(100)" json:"developer"`
	InvolvedSystems   *string         `gorm:"type:varchar(500)" json:"involvedSystems"`
	FollowerID        *int64          `json:"followerId"`
	FollowerName      *string         `gorm:"type:varchar(100)" json:"followerName"`
	SubmitterID       *int64          `json:"submitterId"`
	SubmitterName     *string         `gorm:"type:varchar(100)" json:"submitterName"`
	SubmitTime        *time.Time      `json:"submitTime"`
	CompletionDate    *time.Time      `json:"completionDate"`
	EstimatedWorkload sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"estimatedWorkload"`
	GroupType         *string         `gorm:"type:varchar(50)" json:"groupType"` // LOAN_MIDDLE/LOAN_PRE/RISK
	LastModifierID    *int64          `json:"lastModifierId"`
	LastModifierName  *string         `gorm:"type:varchar(100)" json:"lastModifierName"`
	Description       *string         `gorm:"type:text" json:"description"`
	Remark            *string         `gorm:"type:varchar(1000)" json:"remark"`
	UserID            int64           `gorm:"not null;index" json:"userId"`
	Deleted           *int            `gorm:"default:0" json:"deleted"`
	CreateTime        *time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser        *int64          `json:"createUser"`
	UpdateTime        *time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser        *int64          `json:"updateUser"`
}

// TableName 指定表名
func (Requirement) TableName() string {
	return "requirement"
}
