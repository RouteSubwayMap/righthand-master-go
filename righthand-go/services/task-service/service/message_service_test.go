package service

import (
	"errors"
	"fmt"
	"testing"

	"righthand-go/internal/models"

	"gorm.io/gorm"
)

// setupMessageServiceTx 真库事务隔离（复用 project 测试的 MySQL 连接辅助）
func setupMessageServiceTx(t *testing.T) (*MessageService, *gorm.DB) {
	t.Helper()
	db := setupProjectTestDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("开启事务失败: %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return NewMessageServiceWithDB(tx), tx
}

func TestMessageService_UnreadCountAndReadFlow(t *testing.T) {
	svc, tx := setupMessageServiceTx(t)
	const testUser int64 = 987654
	const otherUser int64 = 987655
	bizID := int64(424242)

	// 基线：新用户未读数为 0
	count, err := svc.UnreadCount(testUser)
	if err != nil {
		t.Fatalf("查询未读数失败: %v", err)
	}
	if count != 0 {
		t.Fatalf("新用户未读数期望 0，实际 %d", count)
	}

	// 正例：创建 2 条未读消息
	for i := 1; i <= 2; i++ {
		if err := svc.CreateMessage(testUser, "周总结提醒", fmt.Sprintf("第%d周尚未提交周总结", i), "PROJECT", &bizID); err != nil {
			t.Fatalf("创建消息失败: %v", err)
		}
	}
	count, _ = svc.UnreadCount(testUser)
	if count != 2 {
		t.Fatalf("未读数期望 2，实际 %d", count)
	}

	// 取一条消息标记已读 → 未读减 1
	var one models.Message
	if err := tx.Where("user_id = ? AND title = ?", testUser, "周总结提醒").First(&one).Error; err != nil {
		t.Fatalf("查询消息失败: %v", err)
	}
	if err = svc.MarkRead(testUser, one.ID); err != nil {
		t.Fatalf("标记已读失败: %v", err)
	}
	count, _ = svc.UnreadCount(testUser)
	if count != 1 {
		t.Fatalf("标记一条已读后期望 1，实际 %d", count)
	}

	// 反例：越权标记他人消息 → ErrMessageNotFound
	if err = svc.CreateMessage(otherUser, "他人消息", "内容", "PROJECT", nil); err != nil {
		t.Fatalf("创建他人消息失败: %v", err)
	}
	var otherMsg models.Message
	if err := tx.Where("user_id = ?", otherUser).First(&otherMsg).Error; err != nil {
		t.Fatalf("查询他人消息失败: %v", err)
	}
	err = svc.MarkRead(testUser, otherMsg.ID)
	if !errors.Is(err, ErrMessageNotFound) {
		t.Fatalf("越权标记期望 ErrMessageNotFound，实际 %v", err)
	}

	// 反例：标记不存在的消息 → ErrMessageNotFound
	err = svc.MarkRead(testUser, 999999999)
	if !errors.Is(err, ErrMessageNotFound) {
		t.Fatalf("不存在消息期望 ErrMessageNotFound，实际 %v", err)
	}

	// 列表：总数 2；按已读过滤（true 1 条 / false 1 条）；按时间倒序
	result, err := svc.ListMessages(testUser, nil, 1, 10)
	if err != nil {
		t.Fatalf("查询消息列表失败: %v", err)
	}
	if total, ok := result["total"].(int64); !ok || total != 2 {
		t.Fatalf("消息总数期望 2，实际 %v", result["total"])
	}
	list, ok := result["list"].([]models.Message)
	if !ok || len(list) != 2 {
		t.Fatalf("列表期望 2 条 models.Message，实际 %T/%d", result["list"], len(list))
	}
	readTrue, readFalse := true, false
	rTrue, _ := svc.ListMessages(testUser, &readTrue, 1, 10)
	if l := rTrue["list"].([]models.Message); len(l) != 1 || !l[0].IsRead {
		t.Fatalf("isRead=true 过滤期望 1 条已读，实际 %d 条", len(l))
	}
	rFalse, _ := svc.ListMessages(testUser, &readFalse, 1, 10)
	if l := rFalse["list"].([]models.Message); len(l) != 1 || l[0].IsRead {
		t.Fatalf("isRead=false 过滤期望 1 条未读，实际 %d 条", len(l))
	}

	// 正例：全部已读 → 未读数 0；幂等再次调用无错
	if err = svc.MarkAllRead(testUser); err != nil {
		t.Fatalf("全部已读失败: %v", err)
	}
	count, _ = svc.UnreadCount(testUser)
	if count != 0 {
		t.Fatalf("全部已读后期望 0，实际 %d", count)
	}
	if err = svc.MarkAllRead(testUser); err != nil {
		t.Fatalf("重复全部已读应幂等无错: %v", err)
	}

	// 越权用户的已读数不受影响（仅操作本人消息）
	if err = svc.MarkRead(otherUser, otherMsg.ID); err != nil {
		t.Fatalf("本人消息标记失败: %v", err)
	}
	count, _ = svc.UnreadCount(otherUser)
	if count != 0 {
		t.Fatalf("他人未读数期望 0，实际 %d", count)
	}
}
