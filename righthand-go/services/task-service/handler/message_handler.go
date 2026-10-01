package handler

import (
	"errors"
	"strconv"

	"righthand-go/internal/utils"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

// MessageHandler 站内消息
type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		messageService: service.NewMessageService(),
	}
}

// ListMessages 消息列表（分页，可按已读状态过滤）
func (h *MessageHandler) ListMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	var isRead *bool
	switch c.Query("isRead") {
	case "true", "1":
		v := true
		isRead = &v
	case "false", "0":
		v := false
		isRead = &v
	}

	userID := c.GetInt64("userID")
	result, err := h.messageService.ListMessages(userID, isRead, page, size)
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询消息失败: "+err.Error())
		return
	}
	utils.Success(c, result)
}

// UnreadCount 未读消息数
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	userID := c.GetInt64("userID")
	count, err := h.messageService.UnreadCount(userID)
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询未读数失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{"count": count})
}

// MarkRead 标记单条消息已读
func (h *MessageHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的消息ID")
		return
	}
	userID := c.GetInt64("userID")
	if err := h.messageService.MarkRead(userID, id); err != nil {
		if errors.Is(err, service.ErrMessageNotFound) {
			utils.ErrorWithCode(c, 404, "消息不存在")
			return
		}
		utils.ErrorWithCode(c, 500, "标记已读失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "已读", nil)
}

// MarkAllRead 全部标记已读
func (h *MessageHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetInt64("userID")
	if err := h.messageService.MarkAllRead(userID); err != nil {
		utils.ErrorWithCode(c, 500, "标记全部已读失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "全部已读", nil)
}
