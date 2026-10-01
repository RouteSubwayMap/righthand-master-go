package models

import "time"

// BankParam 银行机构参数
type BankParam struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	BankType string `gorm:"column:bank_type;type:varchar(50);not null" json:"bankType"`
	BankName string `gorm:"column:bank_name;type:varchar(100);not null" json:"bankName"`
	BankCode string `gorm:"column:bank_code;type:varchar(50);not null" json:"bankCode"`
	Reserve1 string `gorm:"column:reserve1;type:varchar(200)" json:"reserve1"`
	Reserve2 string `gorm:"column:reserve2;type:varchar(200)" json:"reserve2"`
	Reserve3 string `gorm:"column:reserve3;type:varchar(200)" json:"reserve3"`
	Reserve4 string `gorm:"column:reserve4;type:varchar(200)" json:"reserve4"`
	Reserve5 string `gorm:"column:reserve5;type:varchar(200)" json:"reserve5"`
	Remark   string `gorm:"column:remark;type:varchar(500)" json:"remark"`

	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted    int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (BankParam) TableName() string {
	return "bank_param"
}
