package models

import "time"

// ProdNode 生产环境节点清单
type ProdNode struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	BankType      string `gorm:"column:bank_type;type:varchar(50)" json:"bankType"`
	BankName      string `gorm:"column:bank_name;type:varchar(100);not null" json:"bankName"`
	EnvType       string `gorm:"column:env_type;type:varchar(50);not null" json:"envType"`
	NodeName      string `gorm:"column:node_name;type:varchar(100);not null" json:"nodeName"`
	NodeIP        string `gorm:"column:node_ip;type:varchar(50);not null" json:"nodeIp"`
	Username      string `gorm:"column:username;type:varchar(50)" json:"username"`
	IsMutualTrust string `gorm:"column:is_mutual_trust;type:varchar(20)" json:"isMutualTrust"`

	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime"`
	CreateUser *int64    `gorm:"column:create_user" json:"createUser"`
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updateTime"`
	UpdateUser *int64    `gorm:"column:update_user" json:"updateUser"`
	Deleted    int       `gorm:"column:deleted;type:tinyint;default:0" json:"deleted"`
}

func (ProdNode) TableName() string {
	return "prod_node"
}
