package service

import (
	"righthand-go/internal/database"
	"righthand-go/internal/models"

	"gorm.io/gorm"
)

type ProdNodeService struct {
	db *gorm.DB
}

func NewProdNodeService() *ProdNodeService {
	return &ProdNodeService{}
}

func (s *ProdNodeService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

// ListProdNodes 获取生产环境节点列表（分页 + 多条件查询）
func (s *ProdNodeService) ListProdNodes(page, size int, bankType, bankName, envType, nodeName, nodeIP, username string) (map[string]interface{}, error) {
	db := s.getDB()

	query := db.Model(&models.ProdNode{}).Where("deleted = ?", 0)

	if bankType != "" {
		query = query.Where("bank_type = ?", bankType)
	}
	if bankName != "" {
		query = query.Where("bank_name = ?", bankName)
	}
	if envType != "" {
		query = query.Where("env_type = ?", envType)
	}
	if nodeName != "" {
		query = query.Where("node_name LIKE ?", "%"+nodeName+"%")
	}
	if nodeIP != "" {
		query = query.Where("node_ip LIKE ?", "%"+nodeIP+"%")
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	var total int64
	query.Count(&total)

	var nodes []models.ProdNode
	offset := (page - 1) * size
	if err := query.Order("create_time DESC").Offset(offset).Limit(size).Find(&nodes).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"records": nodes,
		"total":   total,
		"size":    size,
		"current": page,
		"pages":   (total + int64(size) - 1) / int64(size),
	}, nil
}

// GetProdNode 获取节点详情
func (s *ProdNodeService) GetProdNode(id int64) (*models.ProdNode, error) {
	db := s.getDB()

	var node models.ProdNode
	if err := db.Where("id = ? AND deleted = ?", id, 0).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// CreateProdNode 创建节点
func (s *ProdNodeService) CreateProdNode(node *models.ProdNode) error {
	db := s.getDB()
	return db.Create(node).Error
}

// UpdateProdNode 更新节点
func (s *ProdNodeService) UpdateProdNode(id int64, updates map[string]interface{}) error {
	db := s.getDB()
	return db.Model(&models.ProdNode{}).
		Where("id = ? AND deleted = ?", id, 0).
		Updates(updates).Error
}

// DeleteProdNode 删除节点（逻辑删除）
func (s *ProdNodeService) DeleteProdNode(id int64) error {
	db := s.getDB()
	return db.Model(&models.ProdNode{}).
		Where("id = ?", id).
		Update("deleted", 1).Error
}

// BatchImportNodes 批量导入节点
func (s *ProdNodeService) BatchImportNodes(nodes []models.ProdNode) (int, error) {
	db := s.getDB()
	if len(nodes) == 0 {
		return 0, nil
	}
	if err := db.CreateInBatches(nodes, 100).Error; err != nil {
		return 0, err
	}
	return len(nodes), nil
}
