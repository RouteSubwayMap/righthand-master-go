package service

import (
	"righthand-go/internal/database"
	"righthand-go/internal/models"

	"gorm.io/gorm"
)

type BankParamService struct {
	db *gorm.DB
}

func NewBankParamService() *BankParamService {
	return &BankParamService{}
}

func (s *BankParamService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

// ListBankParams 获取银行机构参数列表（分页 + 多条件查询）
func (s *BankParamService) ListBankParams(page, size int, bankType, bankName, bankCode string) (map[string]interface{}, error) {
	db := s.getDB()

	query := db.Model(&models.BankParam{}).Where("deleted = ?", 0)

	if bankType != "" {
		query = query.Where("bank_type = ?", bankType)
	}
	if bankName != "" {
		query = query.Where("bank_name = ?", bankName)
	}
	if bankCode != "" {
		query = query.Where("bank_code LIKE ?", "%"+bankCode+"%")
	}

	var total int64
	query.Count(&total)

	var params []models.BankParam
	offset := (page - 1) * size
	if err := query.Order("create_time DESC").Offset(offset).Limit(size).Find(&params).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"records": params,
		"total":   total,
		"size":    size,
		"current": page,
		"pages":   (total + int64(size) - 1) / int64(size),
	}, nil
}

// GetBankParam 获取银行机构参数详情
func (s *BankParamService) GetBankParam(id int64) (*models.BankParam, error) {
	db := s.getDB()

	var param models.BankParam
	if err := db.Where("id = ? AND deleted = ?", id, 0).First(&param).Error; err != nil {
		return nil, err
	}
	return &param, nil
}

// CreateBankParam 创建银行机构参数
func (s *BankParamService) CreateBankParam(param *models.BankParam) error {
	db := s.getDB()
	return db.Create(param).Error
}

// UpdateBankParam 更新银行机构参数
func (s *BankParamService) UpdateBankParam(id int64, updates map[string]interface{}) error {
	db := s.getDB()
	return db.Model(&models.BankParam{}).
		Where("id = ? AND deleted = ?", id, 0).
		Updates(updates).Error
}

// DeleteBankParam 删除银行机构参数（逻辑删除）
func (s *BankParamService) DeleteBankParam(id int64) error {
	db := s.getDB()
	return db.Model(&models.BankParam{}).
		Where("id = ?", id).
		Update("deleted", 1).Error
}

// BatchImportParams 批量导入银行机构参数
func (s *BankParamService) BatchImportParams(params []models.BankParam) (int, error) {
	db := s.getDB()
	if len(params) == 0 {
		return 0, nil
	}
	if err := db.CreateInBatches(params, 100).Error; err != nil {
		return 0, err
	}
	return len(params), nil
}
