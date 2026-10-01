package service

import (
	"righthand-go/internal/database"
	"righthand-go/internal/models"

	"gorm.io/gorm"
)

type ContactService struct {
	db *gorm.DB // 测试时注入,否则使用全局DB
}

func NewContactService() *ContactService {
	return &ContactService{}
}

// getDB 获取数据库连接
func (s *ContactService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

// ListContacts 获取联系人列表
func (s *ContactService) ListContacts(userID int64, page, size int, keyword, group, position string) (map[string]interface{}, error) {
	db := s.getDB()

	query := db.Model(&models.Contact{}).Where("deleted = ?", 0)

	if keyword != "" {
		query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if group != "" {
		query = query.Where("contact_group = ?", group)
	}
	if position != "" {
		query = query.Where("position LIKE ?", "%"+position+"%")
	}

	var total int64
	query.Count(&total)

	var contacts []models.Contact
	offset := (page - 1) * size
	if err := query.Order("create_time DESC").Offset(offset).Limit(size).Find(&contacts).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"records": contacts,
		"total":   total,
		"size":    size,
		"current": page,
		"pages":   (total + int64(size) - 1) / int64(size),
	}, nil
}

// GetContact 获取联系人详情
func (s *ContactService) GetContact(id int64, userID int64) (*models.Contact, error) {
	db := s.getDB()

	var contact models.Contact
	if err := db.Where("id = ? AND deleted = ?", id, 0).First(&contact).Error; err != nil {
		return nil, err
	}

	return &contact, nil
}

// CreateContact 鍒涘缓鑱旂郴浜?
func (s *ContactService) CreateContact(contact *models.Contact) error {
	db := s.getDB()
	return db.Create(contact).Error
}

// UpdateContact 鏇存柊鑱旂郴浜?
func (s *ContactService) UpdateContact(id int64, userID int64, updates map[string]interface{}) error {
	db := s.getDB()
	return db.Model(&models.Contact{}).
		Where("id = ? AND user_id = ? AND deleted = ?", id, userID, 0).
		Updates(updates).Error
}

// DeleteContact 鍒犻櫎鑱旂郴浜?閫昏緫鍒犻櫎)
func (s *ContactService) DeleteContact(id int64, userID int64) error {
	db := s.getDB()
	return db.Model(&models.Contact{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("deleted", 1).Error
}

// GetContactGroups 获取联系人分组列表
func (s *ContactService) GetContactGroups(userID int64) ([]string, error) {
	db := s.getDB()

	var groups []string
	err := db.Model(&models.Contact{}).
		Where("deleted = ? AND contact_group IS NOT NULL AND contact_group != ?", 0, "").
		Distinct("contact_group").
		Pluck("contact_group", &groups).Error

	if err != nil {
		return nil, err
	}

	return groups, nil
}
