package scheduler

import (
	"fmt"
	"log"
	"righthand-go/internal/models"
	"righthand-go/services/task-service/service"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// ProjectScheduler 项目自动化调度器（M3）
// 全部 Job 固定在凌晨 01:00 执行：每日任务 cron `0 0 1 * * *`，周五总结提醒 `0 0 1 * * 5`
type ProjectScheduler struct {
	db             *gorm.DB
	cron           *cron.Cron
	messageService *service.MessageService
	projectService *service.ProjectService
}

// NewProjectScheduler 创建项目调度器实例
func NewProjectScheduler(db *gorm.DB) *ProjectScheduler {
	return &ProjectScheduler{
		db:             db,
		cron:           cron.New(cron.WithSeconds()),
		messageService: service.NewMessageServiceWithDB(db),
		projectService: service.NewProjectServiceWithDB(db),
	}
}

// Start 启动调度器
func (s *ProjectScheduler) Start() {
	// 每日 01:00：项目延期预警（项目级置位 + 周级提醒）+ 高风险汇总
	s.cron.AddFunc("0 0 1 * * *", s.runDailyJobs)
	// 每周五 01:00：周总结提醒
	s.cron.AddFunc("0 0 1 * * 5", s.remindWeeklySummaries)
	s.cron.Start()
	log.Println("项目自动化调度器已启动（每日/每周五 01:00 执行）")
}

// Stop 停止调度器
func (s *ProjectScheduler) Stop() {
	s.cron.Stop()
	log.Println("项目自动化调度器已停止")
}

// ManualTrigger 手动触发全部 Job（用于测试/联调）
func (s *ProjectScheduler) ManualTrigger() {
	log.Println("手动触发项目自动化任务...")
	s.remindWeeklySummaries()
	s.runDailyJobs()
	log.Println("项目自动化任务执行完成")
}

// runDailyJobs 每日 01:00 任务集合
func (s *ProjectScheduler) runDailyJobs() {
	log.Println("[项目调度] 开始执行每日任务...")
	s.checkProjectDelays()
	s.remindWeekPlanDelays()
	s.summarizeHighRisks()
	log.Println("[项目调度] 每日任务执行完成")
}

const dateLayout = "2006-01-02"

// remindWeeklySummaries 周总结提醒（周五 01:00）
// 条件：进行中项目 + 今天落在某周起止区间内 + 该周未提交总结 -> 站内信给周负责人
func (s *ProjectScheduler) remindWeeklySummaries() {
	todayStr := time.Now().Format(dateLayout)
	var projects []models.Project
	if err := s.db.Where("status = ? AND deleted = ?", "IN_PROGRESS", 0).Find(&projects).Error; err != nil {
		log.Printf("[项目调度] 查询进行中项目失败: %v", err)
		return
	}
	sent := 0
	for i := range projects {
		p := &projects[i]
		var weeks []models.ProjectWeeklyPlan
		if err := s.db.Where("project_id = ? AND deleted = ? AND start_date <= ? AND end_date >= ?", p.ID, 0, todayStr, todayStr).
			Find(&weeks).Error; err != nil {
			log.Printf("[项目调度] 查询项目(ID=%d)本周周计划失败: %v", p.ID, err)
			continue
		}
		for j := range weeks {
			w := &weeks[j]
			if w.SummaryAt != nil {
				continue
			}
			content := fmt.Sprintf("项目「%s」第%d周（%s ~ %s）尚未提交周总结，请及时填写。",
				p.ProjectName, w.WeekNo, w.StartDate.Format(dateLayout), w.EndDate.Format(dateLayout))
			pid := p.ID
			if err := s.messageService.CreateMessage(w.OwnerID, "周总结提醒", content, "PROJECT", &pid); err != nil {
				log.Printf("[项目调度] 发送周总结提醒失败（项目ID=%d 周ID=%d）: %v", p.ID, w.ID, err)
				continue
			}
			sent++
		}
	}
	log.Printf("[项目调度] 周总结提醒完成，共发送 %d 条", sent)
}

// checkProjectDelays 项目延期预警（每日 01:00）
// 条件：end_date 已过 且 状态未结项/终止 且 交付状态 != 有延期风险
// 动作：置「有延期风险」+ 站内信给 PM + 变更历史（条件本身保证每条仅推一次）
func (s *ProjectScheduler) checkProjectDelays() {
	todayStr := time.Now().Format(dateLayout)
	var projects []models.Project
	if err := s.db.Where("deleted = ? AND status NOT IN ? AND delivery_status <> ? AND end_date < ?",
		0, []string{"CLOSED", "TERMINATED"}, "AT_RISK", todayStr).Find(&projects).Error; err != nil {
		log.Printf("[项目调度] 查询延期项目失败: %v", err)
		return
	}
	for i := range projects {
		p := &projects[i]
		if err := s.db.Model(&models.Project{}).Where("id = ?", p.ID).
			Updates(map[string]interface{}{"delivery_status": "AT_RISK"}).Error; err != nil {
			log.Printf("[项目调度] 更新项目(ID=%d)交付状态失败: %v", p.ID, err)
			continue
		}
		content := fmt.Sprintf("项目「%s」已超过计划结束日期（%s）但仍未结项，系统已自动标记「有延期风险」，请关注。",
			p.ProjectName, p.EndDate.Format(dateLayout))
		pid := p.ID
		if err := s.messageService.CreateMessage(p.ManagerID, "项目延期预警", content, "PROJECT", &pid); err != nil {
			log.Printf("[项目调度] 发送延期预警失败（项目ID=%d）: %v", p.ID, err)
		}
		if err := s.projectService.AddChangeLog(p.ID, "RISK_CHANGE",
			"交付状态：正常 → 有延期风险（系统自动：项目已超过计划结束日期）", 0); err != nil {
			log.Printf("[项目调度] 写变更历史失败（项目ID=%d）: %v", p.ID, err)
		}
	}
	log.Printf("[项目调度] 项目延期预警完成，共处理 %d 个项目", len(projects))
}

// remindWeekPlanDelays 周计划延期提醒（每日 01:00，不落库）
// 条件：周 end_date 已过 且 周状态为未开始/进行中 -> 站内信给周负责人与 PM
func (s *ProjectScheduler) remindWeekPlanDelays() {
	todayStr := time.Now().Format(dateLayout)
	var weeks []models.ProjectWeeklyPlan
	if err := s.db.Where("deleted = ? AND status IN ? AND end_date < ?",
		0, []string{"NOT_STARTED", "IN_PROGRESS"}, todayStr).Find(&weeks).Error; err != nil {
		log.Printf("[项目调度] 查询逾期周计划失败: %v", err)
		return
	}
	sent := 0
	for i := range weeks {
		w := &weeks[i]
		var p models.Project
		if err := s.db.Where("id = ? AND deleted = ?", w.ProjectID, 0).First(&p).Error; err != nil {
			continue
		}
		// 终态项目不再提醒
		if p.Status == "CLOSED" || p.Status == "TERMINATED" {
			continue
		}
		content := fmt.Sprintf("项目「%s」第%d周（%s ~ %s）已逾期且未完成，请尽快处理。",
			p.ProjectName, w.WeekNo, w.StartDate.Format(dateLayout), w.EndDate.Format(dateLayout))
		pid := p.ID
		if err := s.messageService.CreateMessage(w.OwnerID, "周计划延期提醒", content, "PROJECT", &pid); err != nil {
			log.Printf("[项目调度] 发送周计划延期提醒失败（周ID=%d）: %v", w.ID, err)
		} else {
			sent++
		}
		// 负责人与管理人相同时不重复发送
		if p.ManagerID != w.OwnerID {
			if err := s.messageService.CreateMessage(p.ManagerID, "周计划延期提醒", content, "PROJECT", &pid); err != nil {
				log.Printf("[项目调度] 发送周计划延期提醒失败（周ID=%d 经理）: %v", w.ID, err)
			} else {
				sent++
			}
		}
	}
	log.Printf("[项目调度] 周计划延期提醒完成，共发送 %d 条", sent)
}

// summarizeHighRisks 高风险汇总（每日 01:00）
// 条件：风险=高 或 交付状态=预计无法完成（未终结）-> 汇总一条站内信发给全部 ADMIN 用户
func (s *ProjectScheduler) summarizeHighRisks() {
	var projects []models.Project
	if err := s.db.Where("deleted = ? AND status NOT IN ? AND (risk_level = ? OR delivery_status = ?)",
		0, []string{"CLOSED", "TERMINATED"}, "HIGH", "CANNOT_FINISH").Find(&projects).Error; err != nil {
		log.Printf("[项目调度] 查询高风险项目失败: %v", err)
		return
	}
	if len(projects) == 0 {
		log.Println("[项目调度] 无高风险项目，跳过汇总")
		return
	}
	lines := make([]string, 0, len(projects))
	for i := range projects {
		p := &projects[i]
		lines = append(lines, fmt.Sprintf("「%s」风险：%s / 交付：%s", p.ProjectName, riskLevelText(p.RiskLevel), deliveryStatusText(p.DeliveryStatus)))
	}
	content := fmt.Sprintf("当前共有 %d 个高风险项目：\n%s", len(projects), strings.Join(lines, "\n"))

	adminIDs, err := s.adminUserIDs()
	if err != nil {
		log.Printf("[项目调度] 查询ADMIN用户失败: %v", err)
		return
	}
	for _, uid := range adminIDs {
		if err := s.messageService.CreateMessage(uid, "高风险项目汇总", content, "PROJECT", nil); err != nil {
			log.Printf("[项目调度] 发送高风险汇总失败（用户ID=%d）: %v", uid, err)
		}
	}
	log.Printf("[项目调度] 高风险汇总完成，共发送给 %d 个ADMIN", len(adminIDs))
}

// adminUserIDs 查询全部 ADMIN 角色用户ID
func (s *ProjectScheduler) adminUserIDs() ([]int64, error) {
	var ids []int64
	err := s.db.Table("sys_user u").
		Joins("JOIN sys_user_role ur ON ur.user_id = u.id").
		Joins("JOIN sys_role r ON r.id = ur.role_id").
		Where("r.role_code = ? AND u.deleted = ?", "ADMIN", 0).
		Distinct().Pluck("u.id", &ids).Error
	return ids, err
}

func riskLevelText(v string) string {
	switch v {
	case "HIGH":
		return "高"
	case "LOW":
		return "低"
	default:
		return "中"
	}
}

func deliveryStatusText(v string) string {
	switch v {
	case "AT_RISK":
		return "有延期风险"
	case "CANNOT_FINISH":
		return "预计无法完成"
	default:
		return "正常"
	}
}
