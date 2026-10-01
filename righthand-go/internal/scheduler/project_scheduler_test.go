package scheduler

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"righthand-go/internal/config"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"righthand-go/services/task-service/dto"
	"righthand-go/services/task-service/service"

	"gorm.io/gorm"
)

// 项目调度器测试：真实 MySQL + 事务隔离（调度器所有读写走注入事务，回滚后零副作用）。
// 说明：Job 会对库中全部符合条件项目生效，因此断言只针对本用例创建的项目数据。

func setupSchedulerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败（需在 righthand-go 目录运行测试）: %v", err)
	}
	// 强约束：自动测试仅允许使用独立测试库（docs/ai/rules.md §9）
	cfg.Database.DBName = database.TestDBName
	if err := database.InitDB(&cfg.Database); err != nil {
		t.Fatalf("连接测试库 %s 失败，请先创建并导入初始化脚本（命令见 docs/ai/rules.md §9）: %v", database.TestDBName, err)
	}
	return database.GetDB()
}

func setupSchedulerTx(t *testing.T) (*ProjectScheduler, *service.ProjectService, *gorm.DB) {
	t.Helper()
	db := setupSchedulerTestDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("开启事务失败: %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return NewProjectScheduler(tx), service.NewProjectServiceWithDB(tx), tx
}

var schedKeySeq int

func schedKey(prefix string) string {
	schedKeySeq++
	return fmt.Sprintf("%s%d%d", prefix, time.Now().UnixNano()%100000000, schedKeySeq)
}

func createSchedProject(t *testing.T, svc *service.ProjectService, managerID int64, start, end string) int64 {
	t.Helper()
	key := schedKey("SCH")
	res, err := svc.CreateProject(managerID, false, &dto.ProjectCreateRequest{
		ProjectName: "调度测试-" + key,
		ProjectKey:  key,
		ManagerID:   managerID,
		StartDate:   start,
		EndDate:     end,
	})
	if err != nil {
		t.Fatalf("创建项目失败: %v", err)
	}
	return res["id"].(int64)
}

func TestProjectScheduler_CheckProjectDelays(t *testing.T) {
	s, svc, tx := setupSchedulerTx(t)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02")

	// 正例：过期未结项项目 → 置 AT_RISK + 通知 PM + 写变更历史
	pid := createSchedProject(t, svc, 2, threeDaysAgo, yesterday)
	s.checkProjectDelays()

	var p models.Project
	if err := tx.First(&p, pid).Error; err != nil {
		t.Fatalf("查询项目失败: %v", err)
	}
	if p.DeliveryStatus != "AT_RISK" {
		t.Fatalf("过期项目期望 delivery_status=AT_RISK，实际 %s", p.DeliveryStatus)
	}
	var msg models.Message
	if err := tx.Where("user_id = ? AND business_id = ?", 2, pid).First(&msg).Error; err != nil {
		t.Fatalf("期望 PM 收到延期预警消息: %v", err)
	}
	if msg.Title != "项目延期预警" || !strings.Contains(msg.Content, "已超过计划结束日期") {
		t.Fatalf("预警消息内容不符: %s / %s", msg.Title, msg.Content)
	}
	var logCount int64
	tx.Model(&models.ProjectChangeLog{}).Where("project_id = ? AND action = ?", pid, "RISK_CHANGE").Count(&logCount)
	if logCount != 1 {
		t.Fatalf("期望 1 条 RISK_CHANGE 日志，实际 %d", logCount)
	}

	// 幂等：再次执行不重复推送（条件已排除 AT_RISK）
	s.checkProjectDelays()
	var msgCount int64
	tx.Model(&models.Message{}).Where("user_id = ? AND business_id = ?", 2, pid).Count(&msgCount)
	if msgCount != 1 {
		t.Fatalf("重复执行期望仍为 1 条消息，实际 %d", msgCount)
	}

	// 反例：已结项项目不再预警
	pidClosed := createSchedProject(t, svc, 2, threeDaysAgo, yesterday)
	if _, err := svc.ChangeProjectStatus(pidClosed, 2, false, &dto.ProjectStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}
	if _, err := svc.ChangeProjectStatus(pidClosed, 2, false, &dto.ProjectStatusRequest{Action: "close"}); err != nil {
		t.Fatalf("结项失败: %v", err)
	}
	s.checkProjectDelays()
	var pClosed models.Project
	if err := tx.First(&pClosed, pidClosed).Error; err != nil {
		t.Fatalf("查询已结项项目失败: %v", err)
	}
	if pClosed.DeliveryStatus != "NORMAL" {
		t.Fatalf("已结项项目不应被置 AT_RISK，实际 %s", pClosed.DeliveryStatus)
	}
	tx.Model(&models.Message{}).Where("user_id = ? AND business_id = ?", 2, pidClosed).Count(&msgCount)
	if msgCount != 0 {
		t.Fatalf("已结项项目不应发预警，实际 %d 条", msgCount)
	}
}

func TestProjectScheduler_RemindWeeklySummaries(t *testing.T) {
	s, svc, tx := setupSchedulerTx(t)
	weekStart := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	weekEnd := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

	// 项目 A：进行中，第1周覆盖今天未提交、第2周已提交
	pid := createSchedProject(t, svc, 2, weekStart, time.Now().AddDate(0, 0, 10).Format("2006-01-02"))
	if _, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}
	if _, err := svc.CreateWeek(pid, 2, false, &dto.WeekCreateRequest{Title: "本周未提交", StartDate: weekStart, EndDate: weekEnd, OwnerID: 2}); err != nil {
		t.Fatalf("创建第1周失败: %v", err)
	}
	w2, err := svc.CreateWeek(pid, 2, false, &dto.WeekCreateRequest{Title: "本周已提交", StartDate: weekStart, EndDate: weekEnd, OwnerID: 2})
	if err != nil {
		t.Fatalf("创建第2周失败: %v", err)
	}
	w2ID := w2["id"].(int64)
	if _, err := svc.SubmitWeekSummary(pid, w2ID, 2, false, &dto.WeekSummaryRequest{Finished: "完成", NextPlan: "继续"}); err != nil {
		t.Fatalf("提交第2周总结失败: %v", err)
	}

	// 反例：非进行中项目（STARTING）不提醒
	pidStarting := createSchedProject(t, svc, 2, weekStart, time.Now().AddDate(0, 0, 10).Format("2006-01-02"))
	if _, err := svc.CreateWeek(pidStarting, 2, false, &dto.WeekCreateRequest{Title: "未开始项目周", StartDate: weekStart, EndDate: weekEnd, OwnerID: 2}); err != nil {
		t.Fatalf("创建周失败: %v", err)
	}

	// 第一次执行：仅未提交的第1周发送（已提交周被排除、非进行中项目被排除）
	s.remindWeeklySummaries()
	var msgCount int64
	tx.Model(&models.Message{}).Where("user_id = ? AND business_id = ? AND title = ?", 2, pid, "周总结提醒").Count(&msgCount)
	if msgCount != 1 {
		t.Fatalf("首次执行期望 1 条周总结提醒（仅未提交周），实际 %d", msgCount)
	}
	tx.Model(&models.Message{}).Where("user_id = ? AND business_id = ?", 2, pidStarting).Count(&msgCount)
	if msgCount != 0 {
		t.Fatalf("STARTING 项目不应提醒，实际 %d 条", msgCount)
	}
	var msg models.Message
	tx.Where("user_id = ? AND business_id = ? AND title = ?", 2, pid, "周总结提醒").First(&msg)
	if !strings.Contains(msg.Content, "尚未提交周总结") {
		t.Fatalf("提醒内容不符: %s", msg.Content)
	}

	// 再次执行：未提交周按周提醒（每周五重复）→ 新增 1 条
	s.remindWeeklySummaries()
	tx.Model(&models.Message{}).Where("user_id = ? AND business_id = ? AND title = ?", 2, pid, "周总结提醒").Count(&msgCount)
	if msgCount != 2 {
		t.Fatalf("重复执行期望累计 2 条（未提交周持续提醒），实际 %d", msgCount)
	}

	// 提交后再执行：不再新增
	var w1 models.ProjectWeeklyPlan
	tx.Where("project_id = ? AND title = ?", pid, "本周未提交").First(&w1)
	if _, err := svc.SubmitWeekSummary(pid, w1.ID, 2, false, &dto.WeekSummaryRequest{Finished: "完成", NextPlan: "继续"}); err != nil {
		t.Fatalf("提交第1周总结失败: %v", err)
	}
	s.remindWeeklySummaries()
	tx.Model(&models.Message{}).Where("user_id = ? AND business_id = ? AND title = ?", 2, pid, "周总结提醒").Count(&msgCount)
	if msgCount != 2 {
		t.Fatalf("全部提交后期望仍为 2 条，实际 %d", msgCount)
	}
}

func TestProjectScheduler_SummarizeHighRisks(t *testing.T) {
	s, svc, tx := setupSchedulerTx(t)
	today := time.Now().Format("2006-01-02")
	end := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	// 准备：风险=高 的项目
	pid := createSchedProject(t, svc, 2, today, end)
	if _, err := svc.ChangeProjectRisk(pid, 2, false, &dto.ProjectRiskRequest{RiskLevel: "HIGH", Reason: "单测置高"}); err != nil {
		t.Fatalf("置高风险失败: %v", err)
	}
	var p models.Project
	tx.First(&p, pid)
	projectName := p.ProjectName

	// ADMIN 查询：应包含 admin 用户（id=1）
	adminIDs, err := s.adminUserIDs()
	if err != nil {
		t.Fatalf("查询 ADMIN 失败: %v", err)
	}
	found := false
	for _, id := range adminIDs {
		if id == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("ADMIN 用户应包含 id=1，实际 %v", adminIDs)
	}

	// 正例：汇总一条站内信发给全部 ADMIN，内容包含高风险项目
	s.summarizeHighRisks()
	for _, uid := range adminIDs {
		var msg models.Message
		if err := tx.Where("user_id = ? AND title = ?", uid, "高风险项目汇总").Order("id DESC").First(&msg).Error; err != nil {
			t.Fatalf("ADMIN %d 未收到高风险汇总: %v", uid, err)
		}
		if !strings.Contains(msg.Content, projectName) {
			t.Fatalf("汇总内容未包含高风险项目「%s」: %s", projectName, msg.Content)
		}
	}

	// 反例：高风险与终态互斥（结项后不再纳入汇总）
	if _, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}
	if _, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "close"}); err != nil {
		t.Fatalf("结项失败: %v", err)
	}
	var before int64
	tx.Model(&models.Message{}).Where("user_id = ? AND title = ?", 1, "高风险项目汇总").Count(&before)
	s.summarizeHighRisks()
	var after int64
	tx.Model(&models.Message{}).Where("user_id = ? AND title = ?", 1, "高风险项目汇总").Count(&after)
	// 默认库中可能仍存在其他高风险项目（此时会再发汇总），
	// 断言改为：结项项目不再出现在最新汇总内容中
	if after > before {
		var latest models.Message
		tx.Where("user_id = ? AND title = ?", 1, "高风险项目汇总").Order("id DESC").First(&latest)
		if strings.Contains(latest.Content, projectName) {
			t.Fatalf("已结项项目不应再出现在高风险汇总: %s", latest.Content)
		}
	}
}
