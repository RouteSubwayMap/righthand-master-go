package models

import (
	"time"
)

// Contact 联系人
type Contact struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"column:user_id;not null;index" json:"userId"`
	Name         string    `gorm:"column:name;type:varchar(50);not null" json:"name"`
	Gender       *int      `gorm:"column:gender;type:tinyint" json:"gender"`
	Phone        string    `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Email        string    `gorm:"column:email;type:varchar(100)" json:"email"`
	ContactGroup string    `gorm:"column:contact_group;type:varchar(50)" json:"contactGroup"`
	Company      string    `gorm:"column:company;type:varchar(100)" json:"company"`
	Position     string    `gorm:"column:position;type:varchar(100)" json:"position"`
	Remark       string    `gorm:"column:remark;type:varchar(255)" json:"remark"`
	Avatar       string    `gorm:"column:avatar;type:varchar(255)" json:"avatar"`
	CreateTime   time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser   *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime   time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser   *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted      int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (Contact) TableName() string {
	return "contact"
}
