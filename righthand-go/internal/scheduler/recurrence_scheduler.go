package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"righthand-go/internal/models"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// RecurrenceScheduler 周期性任务调度器
type RecurrenceScheduler struct {
	db   *gorm.DB
	cron *cron.Cron
}

// RecurrenceRule 重复规则结构
type RecurrenceRule struct {
	WeekDays   []int `json:"weekDays,omitempty"`   // 每周的星期几 (1-7)
	DayOfMonth int   `json:"dayOfMonth,omitempty"` // 每月的几号 (1-31)
	Month      int   `json:"month,omitempty"`      // 每年的月份 (1-12)
}

// NewRecurrenceScheduler 创建调度器实例
func NewRecurrenceScheduler(db *gorm.DB) *RecurrenceScheduler {
	return &RecurrenceScheduler{
		db:   db,
		cron: cron.New(cron.WithSeconds()),
	}
}

// Start 启动调度器
func (s *RecurrenceScheduler) Start() {
	// 每30分钟执行一次，检查并生成周期性任务
	s.cron.AddFunc("0 */3 * * * *", s.generateRecurringMemos)
	s.cron.Start()
	log.Println("周期性任务调度器已启动")
}

// Stop 停止调度器
func (s *RecurrenceScheduler) Stop() {
	s.cron.Stop()
	log.Println("周期性任务调度器已停止")
}

// generateRecurringMemos 检查周期性任务，将已完成的已过期事项的截止时间推进到下一个符合规则的日期
func (s *RecurrenceScheduler) generateRecurringMemos() {
	log.Println("开始检查周期性备忘任务...")

	// 查询所有已完成且配置了周期性的任务（排除已删除的）
	var tasks []models.TaskItem
	err := s.db.Where(
		"recurrence_type != ? AND deleted = ? AND status_code = ?",
		"NONE", 0, "DONE",
	).Find(&tasks).Error

	if err != nil {
		log.Printf("查询周期性任务失败: %v", err)
		return
	}

	log.Printf("找到 %d 个周期性任务", len(tasks))

	for _, task := range tasks {
		if s.shouldUpdateDeadline(task) {
			if err := s.updateNextDeadline(task); err != nil {
				log.Printf("更新任务截止时间失败 (ID=%d): %v", task.ID, err)
			} else {
				log.Printf("成功更新任务截止时间 (ID=%d)", task.ID)
			}
		}
	}
}

// shouldUpdateDeadline 判断是否需要将截止时间推进到下一个周期
// 规则：
//   - 状态必须是已完成(DONE)
//   - 每天/每周：deadline的完整时间已过期（now > deadline）
//   - 每月：deadline的年月日已过期（仅比较日期，不比较时分秒）
func (s *RecurrenceScheduler) shouldUpdateDeadline(task models.TaskItem) bool {
	now := time.Now()

	// 检查结束条件：已过结束日期的任务不再更新
	if task.RecurrenceEndType == "DATE" && task.RecurrenceEndDate != nil {
		if now.After(*task.RecurrenceEndDate) {
			return false
		}
	}

	// 没有截止时间，不需要更新
	if task.Deadline == nil {
		return false
	}

	// 状态必须是已完成
	if task.StatusCode != "DONE" {
		return false
	}

	if task.RecurrenceType == "MONTHLY" {
		// 每月重复：仅比较年月日，不比较时分秒
		deadlineDate := time.Date(task.Deadline.Year(), task.Deadline.Month(), task.Deadline.Day(), 0, 0, 0, 0, task.Deadline.Location())
		todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if !deadlineDate.Before(todayDate) {
			return false // 截止日期还没过
		}
	} else {
		// 每天/每周重复：比较完整时间
		if !now.After(*task.Deadline) {
			return false // 截止时间还没过
		}
	}

	// 截止时间已过期且状态为已完成，需要更新
	return true
}

// isDateMatchRecurrenceRule 判断指定日期是否符合重复规则
func (s *RecurrenceScheduler) isDateMatchRecurrenceRule(task models.TaskItem, date time.Time) bool {
	switch task.RecurrenceType {
	case "DAILY":
		return true

	case "WEEKLY":
		var rule RecurrenceRule
		if err := json.Unmarshal([]byte(task.RecurrenceRule), &rule); err != nil {
			return false
		}
		// 转换：Go的Weekday (0=Sunday) -> 我们的格式 (1=Monday, 7=Sunday)
		goDay := int(date.Weekday())
		ourDay := goDay
		if goDay == 0 {
			ourDay = 7 // Sunday
		}
		for _, day := range rule.WeekDays {
			if day == ourDay {
				return true
			}
		}
		return false

	case "MONTHLY":
		var rule RecurrenceRule
		if err := json.Unmarshal([]byte(task.RecurrenceRule), &rule); err != nil {
			return false
		}
		if rule.DayOfMonth == 0 {
			// 月末：判断是否是该月最后一天
			return isLastDayOfMonth(date)
		}
		return date.Day() == rule.DayOfMonth

	case "YEARLY":
		var rule RecurrenceRule
		if err := json.Unmarshal([]byte(task.RecurrenceRule), &rule); err != nil {
			return false
		}
		return date.Month() == time.Month(rule.Month) && date.Day() == rule.DayOfMonth

	default:
		return false
	}
}

// isLastDayOfMonth 判断是否是该月最后一天
func isLastDayOfMonth(date time.Time) bool {
	next := date.AddDate(0, 0, 1)
	return next.Month() != date.Month()
}

// updateNextDeadline 将任务的截止时间推进到下一个符合重复规则的日期，保持时分秒不变
func (s *RecurrenceScheduler) updateNextDeadline(task models.TaskItem) error {
	now := time.Now()
	oldDeadline := *task.Deadline

	// 计算下一个符合规则的截止时间
	newDeadline, err := s.calculateNextDeadline(task, oldDeadline, now)
	if err != nil {
		return err
	}

	// 检查新截止时间是否超过结束日期
	if task.RecurrenceEndType == "DATE" && task.RecurrenceEndDate != nil {
		if newDeadline.After(*task.RecurrenceEndDate) {
			log.Printf("任务(ID=%d)的新截止时间已超过结束日期，不再更新", task.ID)
			return nil
		}
	}

	log.Printf("任务(ID=%d) 截止时间从 %s 更新为 %s",
		task.ID,
		oldDeadline.Format("2006-01-02 15:04:05"),
		newDeadline.Format("2006-01-02 15:04:05"))

	return s.db.Model(&task).Updates(map[string]interface{}{
		"deadline":    newDeadline,
		"status_code": "TODO",
	}).Error
}

// calculateNextDeadline 计算下一个符合重复规则的截止时间
// 从原截止时间的下一天开始逐日查找，找到第一个既符合规则又在当前时间之后的日期
// 时分秒保持与原截止时间一致
func (s *RecurrenceScheduler) calculateNextDeadline(task models.TaskItem, oldDeadline time.Time, now time.Time) (time.Time, error) {
	// 从截止时间 +1天 开始（AddDate 保持时分秒不变）
	next := oldDeadline.AddDate(0, 0, 1)

	// 安全限制，最多查找2年（防止规则配置错误导致死循环）
	maxIterations := 366 * 2

	for i := 0; i < maxIterations; i++ {
		// 日期符合重复规则 且 该时间点在当前时间之后
		if s.isDateMatchRecurrenceRule(task, next) && next.After(now) {
			return next, nil
		}
		next = next.AddDate(0, 0, 1)
	}

	return time.Time{}, fmt.Errorf("在2年内无法找到符合规则的下一个日期")
}

// ManualTrigger 手动触发检查（用于测试）
func (s *RecurrenceScheduler) ManualTrigger() {
	log.Println("手动触发周期性任务检查...")
	s.generateRecurringMemos()
}

// GetPendingRecurrences 获取待处理的周期性任务列表（用于前端展示）
func (s *RecurrenceScheduler) GetPendingRecurrences(userID int64) ([]models.TaskItem, error) {
	var tasks []models.TaskItem
	err := s.db.Where(
		"user_id = ? AND recurrence_type != ? AND deleted = ?",
		userID, "NONE", 0,
	).Find(&tasks).Error
	return tasks, err
}

// FormatRecurrenceDescription 格式化重复规则描述
func FormatRecurrenceDescription(task models.TaskItem) string {
	typeMap := map[string]string{
		"DAILY":   "每天",
		"WEEKLY":  "每周",
		"MONTHLY": "每月",
		"YEARLY":  "每年",
	}

	desc := typeMap[task.RecurrenceType]
	if desc == "" {
		return "不重复"
	}

	var rule RecurrenceRule
	if err := json.Unmarshal([]byte(task.RecurrenceRule), &rule); err != nil {
		return desc
	}

	switch task.RecurrenceType {
	case "WEEKLY":
		weekMap := map[int]string{
			1: "周一", 2: "周二", 3: "周三", 4: "周四",
			5: "周五", 6: "周六", 7: "周日",
		}
		days := ""
		for i, day := range rule.WeekDays {
			if i > 0 {
				days += "、"
			}
			days += weekMap[day]
		}
		desc += fmt.Sprintf("(%s)", days)

	case "MONTHLY":
		if rule.DayOfMonth == 0 {
			desc += "(月末)"
		} else {
			desc += fmt.Sprintf("(%d号)", rule.DayOfMonth)
		}

	case "YEARLY":
		desc += fmt.Sprintf("(%d月%d号)", rule.Month, rule.DayOfMonth)
	}

	// 添加结束条件
	if task.RecurrenceEndType == "DATE" && task.RecurrenceEndDate != nil {
		desc += fmt.Sprintf("，直到%s", task.RecurrenceEndDate.Format("2006-01-02"))
	} else if task.RecurrenceEndType == "COUNT" {
		desc += fmt.Sprintf("，共%d次", task.RecurrenceCount)
	}

	return desc
}
