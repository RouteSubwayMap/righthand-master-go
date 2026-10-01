package service

import (
	"errors"
	"righthand-go/internal/models"
	"righthand-go/services/auth-service/repository"
	"time"
)

// DictService 字典管理服务
type DictService struct {
	dictRepo *repository.DictRepository
}

// NewDictService 创建字典服务实例
func NewDictService() *DictService {
	return &DictService{
		dictRepo: repository.NewDictRepository(),
	}
}

// ListDicts 查询字典列表
func (s *DictService) ListDicts(page, pageSize int, dictType string) ([]models.SysDict, int64, error) {
	return s.dictRepo.List(page, pageSize, dictType)
}

// GetDictsByType 根据类型获取字典列表
func (s *DictService) GetDictsByType(dictType string) ([]models.SysDict, error) {
	return s.dictRepo.GetByType(dictType)
}

// GetDictByID 根据ID获取字典
func (s *DictService) GetDictByID(id int64) (*models.SysDict, error) {
	return s.dictRepo.GetByID(id)
}

// CreateDict 创建字典
func (s *DictService) CreateDict(dict *models.SysDict) error {
	dict.CreateTime = time.Now()
	dict.UpdateTime = time.Now()
	return s.dictRepo.Create(dict)
}

// UpdateDict 更新字典
func (s *DictService) UpdateDict(dict *models.SysDict) error {
	existingDict, err := s.dictRepo.GetByID(dict.ID)
	if err != nil {
		return errors.New("查询字典失败")
	}
	if existingDict == nil {
		return errors.New("字典不存在")
	}

	return s.dictRepo.Update(dict)
}

// DeleteDict 删除字典
func (s *DictService) DeleteDict(id int64) error {
	existingDict, err := s.dictRepo.GetByID(id)
	if err != nil {
		return errors.New("查询字典失败")
	}
	if existingDict == nil {
		return errors.New("字典不存在")
	}

	return s.dictRepo.Delete(id)
}
