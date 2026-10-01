package repository

import (
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"time"

	"gorm.io/gorm"
)

// DictRepository 字典数据访问层
type DictRepository struct {
	db *gorm.DB
}

// NewDictRepository 创建字典仓库实例
func NewDictRepository() *DictRepository {
	return &DictRepository{
		db: database.GetDB(),
	}
}

// List 查询字典列表
func (r *DictRepository) List(page, pageSize int, dictType string) ([]models.SysDict, int64, error) {
	var dicts []models.SysDict
	var total int64

	query := r.db.Model(&models.SysDict{}).Where("deleted = ?", 0)

	if dictType != "" {
		query = query.Where("dict_type = ?", dictType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("sort_order ASC, id DESC").Find(&dicts).Error
	if err != nil {
		return nil, 0, err
	}

	return dicts, total, nil
}

// GetByType 根据类型查询字典列表
func (r *DictRepository) GetByType(dictType string) ([]models.SysDict, error) {
	var dicts []models.SysDict
	err := r.db.Where("dict_type = ? AND status = ? AND deleted = ?", dictType, 1, 0).
		Order("sort_order ASC").
		Find(&dicts).Error
	return dicts, err
}

// GetByID 根据ID查询字典
func (r *DictRepository) GetByID(id int64) (*models.SysDict, error) {
	var dict models.SysDict
	err := r.db.Where("id = ? AND deleted = ?", id, 0).First(&dict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dict, nil
}

// Create 创建字典
func (r *DictRepository) Create(dict *models.SysDict) error {
	return r.db.Create(dict).Error
}

// Update 更新字典
func (r *DictRepository) Update(dict *models.SysDict) error {
	return r.db.Model(&models.SysDict{}).Where("id = ?", dict.ID).Updates(map[string]interface{}{
		"dict_code":   dict.DictCode,
		"dict_name":   dict.DictName,
		"dict_type":   dict.DictType,
		"dict_value":  dict.DictValue,
		"dict_label":  dict.DictLabel,
		"sort_order":  dict.SortOrder,
		"status":      dict.Status,
		"remark":      dict.Remark,
		"update_time": time.Now(),
	}).Error
}

// Delete 删除字典（逻辑删除）
func (r *DictRepository) Delete(id int64) error {
	return r.db.Model(&models.SysDict{}).Where("id = ?", id).Update("deleted", 1).Error
}
