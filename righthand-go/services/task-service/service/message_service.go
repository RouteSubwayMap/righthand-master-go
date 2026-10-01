package service

import (
	"errors"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"time"

	"gorm.io/gorm"
)

// ErrMessageNotFound 消息不存在（或不属于当前用户）
var ErrMessageNotFound = errors.New("消息不存在")

// MessageService 站内消息服务
type MessageService struct {
	db *gorm.DB // 测试时注入,否则使用全局DB
}

func NewMessageService() *MessageService {
	return &MessageService{}
}

// NewMessageServiceWithDB 使用指定 DB 构造（调度器/测试注入）
func NewMessageServiceWithDB(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

// getDB 获取数据库连接
func (s *MessageService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

// CreateMessage 创建站内信（内部调用，如项目调度器；businessID 关联项目ID）
func (s *MessageService) CreateMessage(userID int64, title, content, msgType string, businessID *int64) error {
	msg := &models.Message{
		UserID:     userID,
		Title:      title,
		Content:    content,
		Type:       msgType,
		BusinessID: businessID,
	}
	return s.getDB().Create(msg).Error
}

// ListMessages 查询当前用户消息（分页，可按已读状态过滤，按时间倒序）
func (s *MessageService) ListMessages(userID int64, isRead *bool, page, size int) (map[string]interface{}, error) {
	db := s.getDB()
	query := db.Model(&models.Message{}).Where("user_id = ? AND deleted = ?", userID, 0)
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var list []models.Message
	offset := (page - 1) * size
	if err := query.Order("create_time DESC, id DESC").Offset(offset).Limit(size).Find(&list).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":    total,
		"pageNum":  page,
		"pageSize": size,
		"list":     list,
	}, nil
}

// UnreadCount 未读消息数
func (s *MessageService) UnreadCount(userID int64) (int64, error) {
	var count int64
	err := s.getDB().Model(&models.Message{}).
		Where("user_id = ? AND is_read = ? AND deleted = ?", userID, false, 0).
		Count(&count).Error
	return count, err
}

// MarkRead 标记单条消息已读（仅能操作本人的消息）
func (s *MessageService) MarkRead(userID, messageID int64) error {
	now := time.Now()
	result := s.getDB().Model(&models.Message{}).
		Where("id = ? AND user_id = ? AND deleted = ?", messageID, userID, 0).
		Updates(map[string]interface{}{"is_read": true, "read_time": &now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMessageNotFound
	}
	return nil
}

// MarkAllRead 全部标记已读
func (s *MessageService) MarkAllRead(userID int64) error {
	now := time.Now()
	return s.getDB().Model(&models.Message{}).
		Where("user_id = ? AND is_read = ? AND deleted = ?", userID, false, 0).
		Updates(map[string]interface{}{"is_read": true, "read_time": &now}).Error
}
