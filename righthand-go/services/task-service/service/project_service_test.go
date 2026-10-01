package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"righthand-go/internal/config"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"righthand-go/services/task-service/dto"

	"gorm.io/gorm"
)

/* ============================== 测试基础设施 ============================== */
// 项目模块测试使用真实 MySQL（不允许降级 SQLite：仓库既有 sqlite 迁移与 MySQL 方言不一致）。
// 所有用例在事务内执行，结束后整体回滚，确保对开发库零副作用。

func setupProjectTestDB(t *testing.T) *gorm.DB {
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
	db := database.GetDB()
	if db == nil {
		t.Fatalf("数据库未初始化")
	}
	return db
}

// setupProjectServiceTx 开启事务并返回注入式服务；测试结束回滚
func setupProjectServiceTx(t *testing.T) (*ProjectService, *gorm.DB) {
	t.Helper()
	db := setupProjectTestDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("开启事务失败: %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return NewProjectServiceWithDB(tx), tx
}

var testKeySeq int

func uniqueKey(prefix string) string {
	testKeySeq++
	return fmt.Sprintf("%s%d%d", prefix, time.Now().UnixNano()%100000000, testKeySeq)
}

func strPtr(s string) *string { return &s }

// assertBizCode 断言业务错误码与错误文案片段
func assertBizCode(t *testing.T, err error, code int, msgPart string) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望业务错误 code=%d，实际 err=nil", code)
	}
	var be *BizError
	if !errors.As(err, &be) {
		t.Fatalf("期望 *BizError，实际 %T: %v", err, err)
	}
	if be.Code != code {
		t.Fatalf("期望业务码 %d，实际 %d（%s）", code, be.Code, be.Message)
	}
	if msgPart != "" && !strings.Contains(be.Message, msgPart) {
		t.Fatalf("错误信息期望包含 %q，实际 %q", msgPart, be.Message)
	}
}

// createTestProject 创建测试项目（PM=创建人=managerID）
func createTestProject(t *testing.T, svc *ProjectService, managerID int64, startDate, endDate string) map[string]interface{} {
	t.Helper()
	key := uniqueKey("TPRJ")
	res, err := svc.CreateProject(managerID, false, &dto.ProjectCreateRequest{
		ProjectName: "测试项目-" + key,
		ProjectKey:  key,
		ManagerID:   managerID,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		t.Fatalf("创建测试项目失败: %v", err)
	}
	return res
}

// createTestWeek 创建周计划
func createTestWeek(t *testing.T, svc *ProjectService, projectID, ownerID int64, title, start, end string) map[string]interface{} {
	t.Helper()
	res, err := svc.CreateWeek(projectID, ownerID, false, &dto.WeekCreateRequest{
		Title:     title,
		PlanDesc:  title + "-计划内容",
		StartDate: start,
		EndDate:   end,
		OwnerID:   ownerID,
	})
	if err != nil {
		t.Fatalf("创建周计划失败: %v", err)
	}
	return res
}

func mustInt64(t *testing.T, m map[string]interface{}, key string) int64 {
	t.Helper()
	v, ok := m[key]
	if !ok || v == nil {
		t.Fatalf("结果缺少字段 %s: %#v", key, m)
	}
	id, ok := v.(int64)
	if !ok {
		t.Fatalf("字段 %s 期望 int64，实际 %T", key, v)
	}
	return id
}

func mustStr(t *testing.T, m map[string]interface{}, key string) string {
	t.Helper()
	v, ok := m[key]
	if !ok || v == nil {
		t.Fatalf("结果缺少字段 %s: %#v", key, m)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("字段 %s 期望 string，实际 %T", key, v)
	}
	return s
}

/* ============================== 进度自动计算 ============================== */

func TestProjectService_RecalcProgress(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-03-01")
	pid := mustInt64(t, res, "id")

	// 场景一：5 周（3 完成 / 1 进行中 / 1 未开始）→ 60%
	for i := 1; i <= 5; i++ {
		createTestWeek(t, svc, pid, 2, fmt.Sprintf("第%d周", i), "2026-01-05", "2026-01-11")
	}
	if err := tx.Model(&models.ProjectWeeklyPlan{}).Where("project_id = ? AND week_no IN ?", pid, []int{1, 2, 3}).
		Update("status", "DONE").Error; err != nil {
		t.Fatalf("预置周状态失败: %v", err)
	}
	if err := svc.RecalcProjectProgress(pid); err != nil {
		t.Fatalf("重算进度失败: %v", err)
	}
	p, err := svc.getProject(pid)
	if err != nil {
		t.Fatalf("查询项目失败: %v", err)
	}
	if p.ProgressPct != 60 || p.WeekCount != 5 || p.DoneWeekCount != 3 {
		t.Fatalf("5周3完成期望 progress=60/weekCount=5/done=3，实际 %d/%d/%d", p.ProgressPct, p.WeekCount, p.DoneWeekCount)
	}

	// 场景二：跳过 1 个未开始周 → 分母 4，3/4 = 75%（SKIPPED 不计入分子分母）
	if err := tx.Model(&models.ProjectWeeklyPlan{}).Where("project_id = ? AND week_no = ?", pid, 5).
		Update("status", "SKIPPED").Error; err != nil {
		t.Fatalf("预置跳过状态失败: %v", err)
	}
	if err := svc.RecalcProjectProgress(pid); err != nil {
		t.Fatalf("重算进度失败: %v", err)
	}
	p, _ = svc.getProject(pid)
	if p.ProgressPct != 75 || p.WeekCount != 4 || p.DoneWeekCount != 3 {
		t.Fatalf("跳过1周后期望 75/4/3，实际 %d/%d/%d", p.ProgressPct, p.WeekCount, p.DoneWeekCount)
	}

	// 边界：全部跳过 → 分母 0 → 进度 0
	if err := tx.Model(&models.ProjectWeeklyPlan{}).Where("project_id = ?", pid).
		Update("status", "SKIPPED").Error; err != nil {
		t.Fatalf("预置全跳过失败: %v", err)
	}
	if err := svc.RecalcProjectProgress(pid); err != nil {
		t.Fatalf("重算进度失败: %v", err)
	}
	p, _ = svc.getProject(pid)
	if p.ProgressPct != 0 || p.WeekCount != 0 || p.DoneWeekCount != 0 {
		t.Fatalf("全跳过期望 0/0/0，实际 %d/%d/%d", p.ProgressPct, p.WeekCount, p.DoneWeekCount)
	}
}

// TestProjectService_ProgressOnlyFromService 进度只由服务端计算：编辑项目不改动进度
func TestProjectService_ProgressOnlyFromService(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid := mustInt64(t, res, "id")
	week := createTestWeek(t, svc, pid, 2, "第1周", "2026-01-05", "2026-01-11")
	wid := mustInt64(t, week, "id")

	// 周流转 start → done 触发重算：1/1 = 100%
	if _, err := svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("周开始失败: %v", err)
	}
	if _, err := svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{Action: "done"}); err != nil {
		t.Fatalf("周完成失败: %v", err)
	}
	p, _ := svc.getProject(pid)
	if p.ProgressPct != 100 {
		t.Fatalf("周完成后期望进度 100，实际 %d", p.ProgressPct)
	}

	// 编辑项目（DTO 无 progress 字段，服务端不受客户端影响）
	updated, err := svc.UpdateProject(pid, 2, false, &dto.ProjectUpdateRequest{Description: strPtr("仅改描述")})
	if err != nil {
		t.Fatalf("编辑项目失败: %v", err)
	}
	if got := updated["progressPct"]; got != 100 {
		t.Fatalf("编辑后期望进度仍为 100，实际 %v", got)
	}
	p, _ = svc.getProject(pid)
	if p.WeekCount != 1 || p.DoneWeekCount != 1 {
		t.Fatalf("编辑后期望 weekCount/done=1/1，实际 %d/%d", p.WeekCount, p.DoneWeekCount)
	}
}

/* ============================== 项目状态机 ============================== */

func TestProjectService_TransitionRules(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid := mustInt64(t, res, "id")

	// 反例：STARTING 状态不允许「结项」
	_, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "close"})
	assertBizCode(t, err, 400, "当前状态不允许执行「结项」")

	// 反例：未知 action
	_, err = svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "fly"})
	assertBizCode(t, err, 400, "不支持的状态操作")

	// 反例：非 PM 非创建人流转 → 403
	_, err = svc.ChangeProjectStatus(pid, 3, false, &dto.ProjectStatusRequest{Action: "start"})
	assertBizCode(t, err, 403, "无权限")

	// 正例：start → IN_PROGRESS
	sres, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "start"})
	if err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}
	if mustStr(t, sres, "status") != "IN_PROGRESS" {
		t.Fatalf("start 后期望 IN_PROGRESS，实际 %v", sres["status"])
	}

	// 反例：IN_PROGRESS 再 start
	_, err = svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "start"})
	assertBizCode(t, err, 400, "当前状态不允许执行「开始」")

	// 正例：pause → PAUSED（记录 paused_at）
	sres, err = svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "pause"})
	if err != nil {
		t.Fatalf("暂停失败: %v", err)
	}
	if mustStr(t, sres, "status") != "PAUSED" || sres["pausedAt"] == nil {
		t.Fatalf("pause 后期望 PAUSED 且 pausedAt 非空，实际 %v / %v", sres["status"], sres["pausedAt"])
	}

	// 正例：resume → IN_PROGRESS（清空 paused_at）
	sres, err = svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "resume"})
	if err != nil {
		t.Fatalf("恢复失败: %v", err)
	}
	if mustStr(t, sres, "status") != "IN_PROGRESS" || sres["pausedAt"] != nil {
		t.Fatalf("resume 后期望 IN_PROGRESS 且 pausedAt 为空，实际 %v / %v", sres["status"], sres["pausedAt"])
	}

	// 反例：终止必须填写原因
	_, err = svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "terminate", Reason: "  "})
	assertBizCode(t, err, 400, "终止项目必须填写原因")

	// 正例：close → CLOSED（记录 closed_at）
	sres, err = svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "close"})
	if err != nil {
		t.Fatalf("结项失败: %v", err)
	}
	if mustStr(t, sres, "status") != "CLOSED" || sres["closedAt"] == nil {
		t.Fatalf("close 后期望 CLOSED 且 closedAt 非空，实际 %v / %v", sres["status"], sres["closedAt"])
	}

	// 反例：终态不可再流转
	_, err = svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "pause"})
	assertBizCode(t, err, 400, "终态")
}

func TestProjectService_TerminateAndRiskChange(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid := mustInt64(t, res, "id")
	if _, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}

	// 反例：非 PM 改风险 → 403
	_, err := svc.ChangeProjectRisk(pid, 3, false, &dto.ProjectRiskRequest{RiskLevel: "HIGH", Reason: "测试"})
	assertBizCode(t, err, 403, "无权限修改项目风险")

	// 反例：空原因
	_, err = svc.ChangeProjectRisk(pid, 2, false, &dto.ProjectRiskRequest{RiskLevel: "HIGH", Reason: ""})
	assertBizCode(t, err, 400, "必须填写原因")

	// 反例：未检测到任何变更
	_, err = svc.ChangeProjectRisk(pid, 2, false, &dto.ProjectRiskRequest{Reason: "例行检查"})
	assertBizCode(t, err, 400, "未检测到任何变更")

	// 正例：风险 HIGH + 交付 CANNOT_FINISH
	rres, err := svc.ChangeProjectRisk(pid, 2, false, &dto.ProjectRiskRequest{
		RiskLevel: "HIGH", DeliveryStatus: "CANNOT_FINISH", Reason: "联调严重延期",
	})
	if err != nil {
		t.Fatalf("变更风险失败: %v", err)
	}
	if mustStr(t, rres, "riskLevel") != "HIGH" || mustStr(t, rres, "deliveryStatus") != "CANNOT_FINISH" {
		t.Fatalf("风险变更未生效: %v / %v", rres["riskLevel"], rres["deliveryStatus"])
	}

	// 正例：terminate 带原因 → risk_note 写入 + closed_at 记录
	tres, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "terminate", Reason: "资源不足放弃"})
	if err != nil {
		t.Fatalf("终止失败: %v", err)
	}
	if mustStr(t, tres, "status") != "TERMINATED" || mustStr(t, tres, "riskNote") != "资源不足放弃" || tres["closedAt"] == nil {
		t.Fatalf("terminate 后字段不符: status=%v riskNote=%v closedAt=%v", tres["status"], tres["riskNote"], tres["closedAt"])
	}

	// 反例：终态不可改风险
	_, err = svc.ChangeProjectRisk(pid, 2, false, &dto.ProjectRiskRequest{RiskLevel: "LOW", Reason: "回滚"})
	assertBizCode(t, err, 400, "终态")
}

/* ============================== 周计划权限与总结锁定 ============================== */

func TestProjectService_WeekPermissionsAndSummaryLock(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid := mustInt64(t, res, "id")
	week := createTestWeek(t, svc, pid, 2, "第1周", "2026-01-05", "2026-01-11")
	wid := mustInt64(t, week, "id")

	// 反例：跨项目 weekId → 404
	res2 := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid2 := mustInt64(t, res2, "id")
	_, err := svc.UpdateWeek(pid2, wid, 2, false, &dto.WeekUpdateRequest{Title: strPtr("越权")})
	assertBizCode(t, err, 404, "周计划不存在")

	// 反例：非 PM 非 owner（user 3）编辑/删除/流转/提交总结 → 403
	_, err = svc.UpdateWeek(pid, wid, 3, false, &dto.WeekUpdateRequest{Title: strPtr("x")})
	assertBizCode(t, err, 403, "无权限编辑该周计划")
	err = svc.DeleteWeek(pid, wid, 3, false)
	assertBizCode(t, err, 403, "无权限删除该周计划")
	_, err = svc.ChangeWeekStatus(pid, wid, 3, false, &dto.WeekStatusRequest{Action: "start"})
	assertBizCode(t, err, 403, "无权限变更该周状态")
	_, err = svc.SubmitWeekSummary(pid, wid, 3, false, &dto.WeekSummaryRequest{Finished: "x", NextPlan: "y"})
	assertBizCode(t, err, 403, "仅周负责人可提交总结")

	// 反例：总结必填项
	_, err = svc.SubmitWeekSummary(pid, wid, 2, false, &dto.WeekSummaryRequest{Finished: "", NextPlan: "下周计划"})
	assertBizCode(t, err, 400, "「本周完成」为必填项")
	_, err = svc.SubmitWeekSummary(pid, wid, 2, false, &dto.WeekSummaryRequest{Finished: "本周完成", NextPlan: ""})
	assertBizCode(t, err, 400, "「下周计划」为必填项")

	// 正例：owner 提交总结 → 锁定；校验 summary JSON 结构与 summaryBy
	sres, err := svc.SubmitWeekSummary(pid, wid, 2, false, &dto.WeekSummaryRequest{Summary: &dto.WeekSummaryContent{
		Finished: "完成登录联调", NextPlan: "开始支付联调", Risks: "支付环境未就绪", Support: "需要测试账号",
	}})
	if err != nil {
		t.Fatalf("提交总结失败: %v", err)
	}
	if sres["summaryAt"] == nil {
		t.Fatalf("提交后 summaryAt 应为非空")
	}
	if sbPtr, ok := sres["summaryBy"].(*int64); !ok || sbPtr == nil || *sbPtr != 2 {
		t.Fatalf("summaryBy 期望 2，实际 %v", sres["summaryBy"])
	}
	summary, ok := sres["summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("summary 期望对象，实际 %T", sres["summary"])
	}
	if summary["finished"] != "完成登录联调" || summary["nextPlan"] != "开始支付联调" ||
		summary["risks"] != "支付环境未就绪" || summary["support"] != "需要测试账号" {
		t.Fatalf("summary JSON 结构不符: %#v", summary)
	}

	// 反例：重复提交（幂等保护）
	_, err = svc.SubmitWeekSummary(pid, wid, 2, false, &dto.WeekSummaryRequest{Finished: "再来", NextPlan: "再来"})
	assertBizCode(t, err, 400, "请勿重复提交")

	// 反例：锁定后编辑/删除不可行
	_, err = svc.UpdateWeek(pid, wid, 2, false, &dto.WeekUpdateRequest{Title: strPtr("锁定后改")})
	assertBizCode(t, err, 400, "需先解锁才能编辑")
	err = svc.DeleteWeek(pid, wid, 2, false)
	assertBizCode(t, err, 400, "需先解锁才能删除")

	// 反例：锁定后直接置 NOT_STARTED 被拒
	_, err = svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{Status: "NOT_STARTED"})
	assertBizCode(t, err, 400, "该周总结已提交锁定")

	// 反例：非 PM 解锁 → 403；未锁定解锁 → 400
	_, err = svc.UnlockWeekSummary(pid, wid, 3, false)
	assertBizCode(t, err, 403, "仅项目经理/管理员可解锁总结")

	// 正例：PM 解锁 → 锁定标记清空、总结内容保留
	ures, err := svc.UnlockWeekSummary(pid, wid, 2, false)
	if err != nil {
		t.Fatalf("解锁失败: %v", err)
	}
	if ures["summaryAt"] != nil || ures["summary"] == nil {
		t.Fatalf("解锁后期望 summaryAt 为空且内容保留，实际 %v / %v", ures["summaryAt"], ures["summary"])
	}

	// 正例：解锁后可编辑
	ures, err = svc.UpdateWeek(pid, wid, 2, false, &dto.WeekUpdateRequest{Title: strPtr("解锁后新标题")})
	if err != nil {
		t.Fatalf("解锁后编辑失败: %v", err)
	}
	if mustStr(t, ures, "title") != "解锁后新标题" {
		t.Fatalf("标题未更新: %v", ures["title"])
	}

	// 反例：再次解锁 → 400
	_, err = svc.UnlockWeekSummary(pid, wid, 2, false)
	assertBizCode(t, err, 400, "未锁定，无需解锁")

	// 反例：周状态流转非法参数
	_, err = svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{})
	assertBizCode(t, err, 400, "缺少目标状态")
	_, err = svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{Status: "UNKNOWN"})
	assertBizCode(t, err, 400, "非法的周状态")
	_, err = svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{Action: "fly"})
	assertBizCode(t, err, 400, "不支持的状态操作")

	// 正例：周状态流转 start → done 并重算进度（1/2 = 50%）
	week2 := createTestWeek(t, svc, pid, 2, "第2周", "2026-01-12", "2026-01-18")
	wid2 := mustInt64(t, week2, "id")
	if _, err = svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("周开始失败: %v", err)
	}
	if _, err = svc.ChangeWeekStatus(pid, wid, 2, false, &dto.WeekStatusRequest{Action: "done"}); err != nil {
		t.Fatalf("周完成失败: %v", err)
	}
	p, _ := svc.getProject(pid)
	if p.ProgressPct != 50 || p.WeekCount != 2 || p.DoneWeekCount != 1 {
		t.Fatalf("1/2 周完成期望 50/2/1，实际 %d/%d/%d", p.ProgressPct, p.WeekCount, p.DoneWeekCount)
	}
	// 反例：DONE 状态不允许再执行「开始」
	_, err = svc.ChangeWeekStatus(pid, wid2, 2, false, &dto.WeekStatusRequest{Action: "done"})
	assertBizCode(t, err, 400, "当前状态不允许执行")
}

/* ============================== Key 唯一与成员管理 ============================== */

func TestProjectService_KeyDuplicateAndMembers(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid := mustInt64(t, res, "id")
	key := mustStr(t, res, "projectKey")

	// 反例：project_key 重复 → 400
	_, err := svc.CreateProject(2, false, &dto.ProjectCreateRequest{
		ProjectName: "重复Key项目", ProjectKey: key, ManagerID: 2, StartDate: "2026-01-05", EndDate: "2026-02-01",
	})
	assertBizCode(t, err, 400, "已存在")

	// 反例：缺必填 → 400
	_, err = svc.CreateProject(2, false, &dto.ProjectCreateRequest{ProjectName: "缺必填", ProjectKey: uniqueKey("TPRJ")})
	assertBizCode(t, err, 400, "必填项")

	// 反例：起止日期反序 → 400
	_, err = svc.CreateProject(2, false, &dto.ProjectCreateRequest{
		ProjectName: "反序", ProjectKey: uniqueKey("TPRJ"), ManagerID: 2, StartDate: "2026-02-01", EndDate: "2026-01-01",
	})
	assertBizCode(t, err, 400, "结束日期不能早于开始日期")

	// 正例：创建时 PM 自动入成员（角色 PM）
	members, err := svc.ListMembers(pid)
	if err != nil {
		t.Fatalf("查询成员失败: %v", err)
	}
	if len(members) != 1 || members[0]["userId"] != int64(2) || members[0]["memberRole"] != "PM" {
		t.Fatalf("初始成员期望 PM(2)，实际 %#v", members)
	}
	pmMemberID := members[0]["id"].(int64)

	// 反例：非 PM 添加成员 → 403；空成员 → 400
	_, _, err = svc.AddMember(pid, 3, false, &dto.ProjectMemberRequest{UserID: 5})
	assertBizCode(t, err, 403, "仅项目经理可管理成员")
	_, _, err = svc.AddMember(pid, 2, false, &dto.ProjectMemberRequest{UserID: 0})
	assertBizCode(t, err, 400, "请选择要添加的成员")

	// 正例：添加成员 5
	m, already, err := svc.AddMember(pid, 2, false, &dto.ProjectMemberRequest{UserID: 5})
	if err != nil || already {
		t.Fatalf("添加成员失败: err=%v already=%v", err, already)
	}
	if m["userId"] != int64(5) || m["memberRole"] != "MEMBER" {
		t.Fatalf("成员字段不符: %#v", m)
	}
	m5ID := m["id"].(int64)

	// 幂等：重复添加同一成员 → already=true 且返回同一记录
	m2, already2, err := svc.AddMember(pid, 2, false, &dto.ProjectMemberRequest{UserID: 5})
	if err != nil || !already2 {
		t.Fatalf("重复添加期望幂等 already=true: err=%v already=%v", err, already2)
	}
	if m2["id"].(int64) != m5ID {
		t.Fatalf("幂等返回的成员 id 不一致: %v vs %v", m2["id"], m5ID)
	}

	// 反例：移除项目经理 → 400；成员不存在 → 404
	err = svc.RemoveMember(pid, pmMemberID, 2, false)
	assertBizCode(t, err, 400, "项目经理不可移除")
	err = svc.RemoveMember(pid, 999999999, 2, false)
	assertBizCode(t, err, 404, "成员不存在")

	// 反例：跨项目 memberId → 404
	res2 := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid2 := mustInt64(t, res2, "id")
	err = svc.RemoveMember(pid2, m5ID, 2, false)
	assertBizCode(t, err, 404, "成员不存在")

	// 正例：移除成员 5
	if err = svc.RemoveMember(pid, m5ID, 2, false); err != nil {
		t.Fatalf("移除成员失败: %v", err)
	}
	members, _ = svc.ListMembers(pid)
	if len(members) != 1 {
		t.Fatalf("移除后期望仅剩 PM，实际 %d 人", len(members))
	}
}

/* ============================== 删除与级联 ============================== */

func TestProjectService_DeleteCascade(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)

	// 场景一：STARTING 项目删除成功，级联软删周/成员、物理删关联
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid := mustInt64(t, res, "id")
	week := createTestWeek(t, svc, pid, 2, "第1周", "2026-01-05", "2026-01-11")
	wid := mustInt64(t, week, "id")
	taskRes, err := svc.CreateTaskInWeek(pid, wid, 2, false, &dto.WeekTaskCreateRequest{Title: "级联任务"})
	if err != nil {
		t.Fatalf("周内建任务失败: %v", err)
	}
	taskID := mustInt64(t, taskRes, "id")

	if err = svc.DeleteProject(pid, 2, false); err != nil {
		t.Fatalf("删除项目失败: %v", err)
	}
	// 详情 404
	_, err = svc.getProject(pid)
	assertBizCode(t, err, 404, "项目不存在")
	// 周计划软删（行保留 deleted=1），任务关联物理删除
	var weekCnt, relCnt, memberActive, logCnt, taskCnt int64
	tx.Model(&models.ProjectWeeklyPlan{}).Where("project_id = ?", pid).Count(&weekCnt)
	tx.Model(&models.ProjectWeeklyPlan{}).Where("project_id = ? AND deleted = ?", pid, 0).Count(&relCnt)
	if weekCnt != 1 || relCnt != 0 {
		t.Fatalf("周计划级联期望 1 条软删 0 条有效，实际 %d/%d", weekCnt, relCnt)
	}
	tx.Model(&models.TaskWeekPlanRel{}).Where("week_plan_id = ?", wid).Count(&relCnt)
	if relCnt != 0 {
		t.Fatalf("任务关联期望物理删除，实际残留 %d 条", relCnt)
	}
	tx.Model(&models.ProjectMember{}).Where("project_id = ? AND deleted = ?", pid, 0).Count(&memberActive)
	if memberActive != 0 {
		t.Fatalf("成员期望全部软删，实际有效 %d 条", memberActive)
	}
	tx.Model(&models.ProjectChangeLog{}).Where("project_id = ?", pid).Count(&logCnt)
	if logCnt == 0 {
		t.Fatalf("变更历史期望保留 DELETE 记录")
	}
	// 事项本体共享核心表，不随项目删除
	tx.Model(&models.TaskItem{}).Where("id = ? AND deleted = ?", taskID, 0).Count(&taskCnt)
	if taskCnt != 1 {
		t.Fatalf("任务事项本体应保留，实际 %d", taskCnt)
	}

	// 反例：非 PM 删除 → 403
	resB := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pidB := mustInt64(t, resB, "id")
	err = svc.DeleteProject(pidB, 3, false)
	assertBizCode(t, err, 403, "无权限删除该项目")

	// 反例：IN_PROGRESS 项目不可删除 → 400
	resC := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pidC := mustInt64(t, resC, "id")
	if _, err = svc.ChangeProjectStatus(pidC, 2, false, &dto.ProjectStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}
	err = svc.DeleteProject(pidC, 2, false)
	assertBizCode(t, err, 400, "仅「启动」状态的项目允许删除")
}

/* ============================== 列表筛选与排序 ============================== */

func TestProjectService_ListFilters(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)
	resA := createTestProject(t, svc, 2, "2026-01-05", "2026-01-30")
	pidA := mustInt64(t, resA, "id")
	resB := createTestProject(t, svc, 2, "2026-02-01", "2026-02-28")
	pidB := mustInt64(t, resB, "id")
	if _, err := svc.ChangeProjectStatus(pidA, 2, false, &dto.ProjectStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}

	containsPID := func(list []map[string]interface{}, id int64) bool {
		for _, item := range list {
			if item["id"] == id {
				return true
			}
		}
		return false
	}
	fetch := func(filter map[string]interface{}) []map[string]interface{} {
		t.Helper()
		result, err := svc.ListProjects(2, false, 1, 200, filter)
		if err != nil {
			t.Fatalf("查询列表失败: %v", err)
		}
		list, ok := result["list"].([]map[string]interface{})
		if !ok {
			t.Fatalf("list 类型不符: %T", result["list"])
		}
		return list
	}

	// 状态单选：IN_PROGRESS 含 A 不含 B
	list := fetch(map[string]interface{}{"status": "IN_PROGRESS"})
	if !containsPID(list, pidA) || containsPID(list, pidB) {
		t.Fatalf("状态筛选 IN_PROGRESS 不符：含A=%v 含B=%v", containsPID(list, pidA), containsPID(list, pidB))
	}
	// 状态多选：STARTING,IN_PROGRESS 两者都含
	list = fetch(map[string]interface{}{"status": "STARTING, IN_PROGRESS"})
	if !containsPID(list, pidA) || !containsPID(list, pidB) {
		t.Fatalf("状态多选筛选不符")
	}
	// keyword 按 key 精确命中
	list = fetch(map[string]interface{}{"keyword": mustStr(t, resA, "projectKey")})
	if len(list) != 1 || !containsPID(list, pidA) {
		t.Fatalf("keyword 筛选期望仅命中 A，实际 %d 条", len(list))
	}
	// 参与人筛选：user3 加入 A 作为成员 → 仅命中 A；user2 为经理 → 命中 A/B
	if _, _, err := svc.AddMember(pidA, 2, false, &dto.ProjectMemberRequest{UserID: 3}); err != nil {
		t.Fatalf("添加成员失败: %v", err)
	}
	list = fetch(map[string]interface{}{"participant_id": "3", "keyword": "TPRJ"})
	if !containsPID(list, pidA) || containsPID(list, pidB) {
		t.Fatalf("参与人筛选不符：user3 为 A 成员应仅命中 A")
	}
	list = fetch(map[string]interface{}{"participant_id": "2", "keyword": "TPRJ"})
	if !containsPID(list, pidA) || !containsPID(list, pidB) {
		t.Fatalf("参与人筛选不符：user2 为经理应命中两个项目")
	}
	// 默认排序 end_date 升序：A(01-30) 在 B(02-28) 之前
	idxA, idxB := -1, -1
	for i, item := range list {
		if item["id"] == pidA {
			idxA = i
		}
		if item["id"] == pidB {
			idxB = i
		}
	}
	if idxA < 0 || idxB < 0 || idxA > idxB {
		t.Fatalf("默认 end_date 升序不符：idxA=%d idxB=%d", idxA, idxB)
	}
}

/* ============================== 任务-周关联全流程 ============================== */

func TestProjectService_TaskWeekFlow(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-01-05", "2026-02-27")
	pid := mustInt64(t, res, "id")
	w1 := createTestWeek(t, svc, pid, 2, "第1周", "2026-01-05", "2026-01-11")
	w2 := createTestWeek(t, svc, pid, 2, "第2周", "2026-01-12", "2026-01-18")
	w3 := createTestWeek(t, svc, pid, 2, "第3周", "2026-01-19", "2026-01-25")
	w1ID := mustInt64(t, w1, "id")
	w2ID := mustInt64(t, w2, "id")
	w3ID := mustInt64(t, w3, "id")

	// 反例：非授权用户（user3）在周内建任务 → 403；标题必填 → 400
	_, err := svc.CreateTaskInWeek(pid, w1ID, 3, false, &dto.WeekTaskCreateRequest{Title: "越权任务"})
	assertBizCode(t, err, 403, "仅周负责人或项目经理可在该周新建任务")
	_, err = svc.CreateTaskInWeek(pid, w1ID, 2, false, &dto.WeekTaskCreateRequest{Title: "  "})
	assertBizCode(t, err, 400, "任务标题为必填项")

	// 正例：周内新建任务 → weekNos [1]、deadline 格式、默认指派周负责人
	tres, err := svc.CreateTaskInWeek(pid, w1ID, 2, false, &dto.WeekTaskCreateRequest{Title: "任务A", Deadline: "2026-01-10"})
	if err != nil {
		t.Fatalf("周内建任务失败: %v", err)
	}
	taskID := mustInt64(t, tres, "id")
	if tres["userId"] != int64(2) {
		t.Fatalf("默认指派人期望周负责人(2)，实际 %v", tres["userId"])
	}
	if mustStr(t, tres, "deadline") != "2026-01-10 00:00:00" {
		t.Fatalf("deadline 格式不符: %v", tres["deadline"])
	}
	weekNos, ok := tres["weekNos"].([]int)
	if !ok || len(weekNos) != 1 || weekNos[0] != 1 {
		t.Fatalf("weekNos 期望 [1]，实际 %v", tres["weekNos"])
	}

	// 幂等：同一任务重复加入同一周 → already=true
	_, already, err := svc.LinkTaskToWeek(pid, w1ID, taskID, 2, false)
	if err != nil || !already {
		t.Fatalf("重复关联期望幂等 already=true: err=%v already=%v", err, already)
	}

	// 正例：加入第二周（一个任务可挂多周）
	_, already2, err := svc.LinkTaskToWeek(pid, w2ID, taskID, 2, false)
	if err != nil || already2 {
		t.Fatalf("加入第二周失败: err=%v already=%v", err, already2)
	}

	// 反例：关联不存在的事项 → 404
	_, _, err = svc.LinkTaskToWeek(pid, w1ID, 999999999, 2, false)
	assertBizCode(t, err, 404, "事项不存在")

	// 「全部」视图按任务去重且带全部所属周
	all, err := svc.ListProjectTasks(pid, 1, 10, nil)
	if err != nil {
		t.Fatalf("查询项目任务失败: %v", err)
	}
	allList := all["list"].([]map[string]interface{})
	if len(allList) != 1 {
		t.Fatalf("全部视图期望按任务去重 1 条，实际 %d 条", len(allList))
	}
	weekNos = allList[0]["weekNos"].([]int)
	if len(weekNos) != 2 || weekNos[0] != 1 || weekNos[1] != 2 {
		t.Fatalf("多周关联期望 weekNos [1 2]，实际 %v", weekNos)
	}

	// 按周过滤：第1周可见；第3周为空
	byW1, _ := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{"weekPlanId": strconv.FormatInt(w1ID, 10)})
	if len(byW1["list"].([]map[string]interface{})) != 1 {
		t.Fatalf("第1周过滤期望 1 条")
	}
	byW3, _ := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{"weekPlanId": strconv.FormatInt(w3ID, 10)})
	if len(byW3["list"].([]map[string]interface{})) != 0 {
		t.Fatalf("第3周过滤期望 0 条")
	}

	// 正例：从第2周移出；重复移出 → 404
	if err = svc.UnlinkTaskFromWeek(pid, w2ID, taskID, 2, false); err != nil {
		t.Fatalf("移出第2周失败: %v", err)
	}
	err = svc.UnlinkTaskFromWeek(pid, w2ID, taskID, 2, false)
	assertBizCode(t, err, 404, "该任务不在此周计划中")

	// 正例：移动到第3周 → weekNos [3]
	mres, err := svc.MoveTaskToWeek(pid, taskID, 2, false, &dto.TaskMoveRequest{WeekPlanID: w3ID})
	if err != nil {
		t.Fatalf("移动任务失败: %v", err)
	}
	weekNos = mres["weekNos"].([]int)
	if len(weekNos) != 1 || weekNos[0] != 3 {
		t.Fatalf("移动后期望 weekNos [3]，实际 %v", weekNos)
	}

	// 反例：移动到不存在周 → 404；无关联任务移动 → 404
	_, err = svc.MoveTaskToWeek(pid, taskID, 2, false, &dto.TaskMoveRequest{WeekPlanID: 999999999})
	assertBizCode(t, err, 404, "目标周计划不存在")
	tres2, err := svc.CreateTaskInWeek(pid, w1ID, 2, false, &dto.WeekTaskCreateRequest{Title: "任务B"})
	if err != nil {
		t.Fatalf("创建任务B失败: %v", err)
	}
	taskBID := mustInt64(t, tres2, "id")
	if err = svc.UnlinkTaskFromWeek(pid, w1ID, taskBID, 2, false); err != nil {
		t.Fatalf("移出任务B失败: %v", err)
	}
	_, err = svc.MoveTaskToWeek(pid, taskBID, 2, false, &dto.TaskMoveRequest{WeekPlanID: w3ID})
	assertBizCode(t, err, 404, "不在本项目任一周期中")

	// 锁定链路：第3周提交总结后，新增/移出/移动均被拒
	if _, err = svc.SubmitWeekSummary(pid, w3ID, 2, false, &dto.WeekSummaryRequest{Finished: "完成", NextPlan: "继续"}); err != nil {
		t.Fatalf("提交第3周总结失败: %v", err)
	}
	_, err = svc.CreateTaskInWeek(pid, w3ID, 2, false, &dto.WeekTaskCreateRequest{Title: "锁定后新增"})
	assertBizCode(t, err, 400, "不可新增任务")
	err = svc.UnlinkTaskFromWeek(pid, w3ID, taskID, 2, false)
	assertBizCode(t, err, 400, "不可移出任务")
	_, err = svc.MoveTaskToWeek(pid, taskID, 2, false, &dto.TaskMoveRequest{WeekPlanID: w1ID})
	assertBizCode(t, err, 400, "第3周总结已锁定")

	// 正例：解锁后删除第3周 → 该周关联物理清理，全部视图为空
	if _, err = svc.UnlockWeekSummary(pid, w3ID, 2, false); err != nil {
		t.Fatalf("解锁失败: %v", err)
	}
	if err = svc.DeleteWeek(pid, w3ID, 2, false); err != nil {
		t.Fatalf("删除第3周失败: %v", err)
	}
	finalAll, err := svc.ListProjectTasks(pid, 1, 10, nil)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(finalAll["list"].([]map[string]interface{})) != 0 {
		t.Fatalf("删除周后该周任务关联应清理，实际仍有 %d 条", len(finalAll["list"].([]map[string]interface{})))
	}
}

/* ============================== 任务列表筛选（负责人/状态） ============================== */

// TestProjectService_ListProjectTasks_Filters 对应矩阵 tests/cases/project-task-filter.md
func TestProjectService_ListProjectTasks_Filters(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-09-01", "2026-09-30")
	pid := mustInt64(t, res, "id")
	w1 := createTestWeek(t, svc, pid, 2, "第1周", "2026-09-01", "2026-09-07")
	w2 := createTestWeek(t, svc, pid, 2, "第2周", "2026-09-08", "2026-09-14")
	w1ID := mustInt64(t, w1, "id")
	w2ID := mustInt64(t, w2, "id")

	// 造数：T1(user3, DOING, W1+W2)、T2(user2, 待办, W1)、T3(user2, DONE, W2)
	u3 := int64(3)
	t1, err := svc.CreateTaskInWeek(pid, w1ID, 2, false, &dto.WeekTaskCreateRequest{Title: "任务T1", UserID: &u3})
	if err != nil {
		t.Fatalf("创建 T1 失败: %v", err)
	}
	t1ID := mustInt64(t, t1, "id")
	if _, already, err := svc.LinkTaskToWeek(pid, w2ID, t1ID, 2, false); err != nil || already {
		t.Fatalf("T1 关联第2周失败: err=%v already=%v", err, already)
	}
	t2, err := svc.CreateTaskInWeek(pid, w1ID, 2, false, &dto.WeekTaskCreateRequest{Title: "任务T2"})
	if err != nil {
		t.Fatalf("创建 T2 失败: %v", err)
	}
	if mustInt64(t, t2, "id") <= 0 {
		t.Fatalf("T2 创建结果缺少有效 ID: %#v", t2)
	}
	t3, err := svc.CreateTaskInWeek(pid, w2ID, 2, false, &dto.WeekTaskCreateRequest{Title: "任务T3"})
	if err != nil {
		t.Fatalf("创建 T3 失败: %v", err)
	}
	t3ID := mustInt64(t, t3, "id")
	if err := tx.Model(&models.TaskItem{}).Where("id = ?", t1ID).Update("status_code", "DOING").Error; err != nil {
		t.Fatalf("更新 T1 状态失败: %v", err)
	}
	if err := tx.Model(&models.TaskItem{}).Where("id = ?", t3ID).Update("status_code", "DONE").Error; err != nil {
		t.Fatalf("更新 T3 状态失败: %v", err)
	}

	listOf := func(r map[string]interface{}) []map[string]interface{} {
		return r["list"].([]map[string]interface{})
	}
	totalOf := func(r map[string]interface{}) int64 {
		switch v := r["total"].(type) {
		case int:
			return int64(v)
		case int64:
			return v
		default:
			t.Fatalf("total 类型异常: %T", r["total"])
			return -1
		}
	}

	// 正例：userId=3 → 仅 T1
	byUser, err := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{"userId": "3"})
	if err != nil {
		t.Fatalf("按负责人查询失败: %v", err)
	}
	if l := listOf(byUser); totalOf(byUser) != 1 || len(l) != 1 || mustStr(t, l[0], "title") != "任务T1" {
		t.Fatalf("userId=3 期望仅 T1 且 total=1，实际 %#v", l)
	}

	// 正例：statusCode=DOING → 仅 T1
	byStatus, err := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{"statusCode": "DOING"})
	if err != nil {
		t.Fatalf("按状态查询失败: %v", err)
	}
	if l := listOf(byStatus); totalOf(byStatus) != 1 || len(l) != 1 || mustStr(t, l[0], "title") != "任务T1" {
		t.Fatalf("statusCode=DOING 期望仅 T1 且 total=1，实际 %#v", l)
	}

	// 边界：weekPlanId=W2、userId=2、statusCode=DONE 三条件交集 → 仅 T3（T2 不在 W2 被排除）
	combined, err := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{
		"weekPlanId": strconv.FormatInt(w2ID, 10), "userId": "2", "statusCode": "DONE",
	})
	if err != nil {
		t.Fatalf("三条件组合查询失败: %v", err)
	}
	if l := listOf(combined); totalOf(combined) != 1 || len(l) != 1 || mustStr(t, l[0], "title") != "任务T3" {
		t.Fatalf("三条件交集期望仅 T3 且 total=1，实际 %#v", l)
	}

	// 边界：userId=2 过滤后第 1 页 size=1 → total=2、首条为未完成的 T2
	page1, err := svc.ListProjectTasks(pid, 1, 1, map[string]interface{}{"userId": "2"})
	if err != nil {
		t.Fatalf("过滤后分页查询失败: %v", err)
	}
	if l := listOf(page1); totalOf(page1) != 2 || len(l) != 1 || mustStr(t, l[0], "title") != "任务T2" {
		t.Fatalf("userId=2 首页期望 total=2 且首条 T2，实际 %#v", l)
	}

	// 反例：userId=abc 非法 → 忽略条件返回全部 3 条
	invalid, err := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{"userId": "abc"})
	if err != nil {
		t.Fatalf("非法 userId 查询失败: %v", err)
	}
	if l := listOf(invalid); totalOf(invalid) != 3 || len(l) != 3 {
		t.Fatalf("userId=abc 期望忽略条件返回 3 条，实际 %d 条", len(l))
	}

	// 反例：statusCode=FOO 未知状态 → 空列表
	unknown, err := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{"statusCode": "FOO"})
	if err != nil {
		t.Fatalf("未知状态查询失败: %v", err)
	}
	if l := listOf(unknown); totalOf(unknown) != 0 || len(l) != 0 {
		t.Fatalf("statusCode=FOO 期望空列表，实际 %d 条", len(l))
	}

	// 边界：userId=2 越界页 → total 仍为过滤后条数 2、列表为空
	beyond, err := svc.ListProjectTasks(pid, 3, 1, map[string]interface{}{"userId": "2"})
	if err != nil {
		t.Fatalf("越界页查询失败: %v", err)
	}
	if l := listOf(beyond); totalOf(beyond) != 2 || len(l) != 0 {
		t.Fatalf("越界页期望 total=2 且空列表，实际 total=%d 共 %d 条", totalOf(beyond), len(l))
	}

	// 反例：userId=999 无匹配 → 空列表
	noMatch, err := svc.ListProjectTasks(pid, 1, 10, map[string]interface{}{"userId": "999"})
	if err != nil {
		t.Fatalf("无匹配用户查询失败: %v", err)
	}
	if l := listOf(noMatch); totalOf(noMatch) != 0 || len(l) != 0 {
		t.Fatalf("userId=999 期望空列表，实际 %d 条", len(l))
	}
}

/* ============================== 项目阶段 ============================== */

// createTestStage 创建测试阶段（PM=operatorID）
func createTestStage(t *testing.T, svc *ProjectService, projectID, operatorID int64, name, start, end string) map[string]interface{} {
	t.Helper()
	res, err := svc.CreateStage(projectID, operatorID, false, &dto.StageCreateRequest{
		StageName: name, StartDate: start, EndDate: end,
	})
	if err != nil {
		t.Fatalf("创建测试阶段失败: %v", err)
	}
	return res
}

// startTestProject 将项目流转至 IN_PROGRESS（阶段维护的典型状态）
func startTestProject(t *testing.T, svc *ProjectService, projectID, operatorID int64) {
	t.Helper()
	if _, err := svc.ChangeProjectStatus(projectID, operatorID, false, &dto.ProjectStatusRequest{Action: "start"}); err != nil {
		t.Fatalf("开始项目失败: %v", err)
	}
}

// relDate 以今天为基准的相对日期（YYYY-MM-DD），避免用例随日历过期
func relDate(deltaDays int) string {
	return time.Now().AddDate(0, 0, deltaDays).Format("2006-01-02")
}

// TestStage_Create_Success 正例：PM 创建阶段成功（含中文名、落库与变更日志；对照分支：名称超长 400）
func TestStage_Create_Success(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)

	sres, err := svc.CreateStage(pid, 2, false, &dto.StageCreateRequest{
		StageName: "需求调研", StartDate: "2026-10-01", EndDate: "2026-10-10", Description: "调研业务需求",
	})
	if err != nil {
		t.Fatalf("创建阶段失败: %v", err)
	}
	if mustStr(t, sres, "stageName") != "需求调研" || mustStr(t, sres, "startDate") != "2026-10-01" ||
		mustStr(t, sres, "endDate") != "2026-10-10" || mustStr(t, sres, "description") != "调研业务需求" {
		t.Fatalf("返回字段不符: %#v", sres)
	}
	var cnt, logCnt int64
	tx.Model(&models.ProjectStage{}).Where("project_id = ? AND deleted = 0", pid).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("期望落库 1 行阶段，实际 %d", cnt)
	}
	tx.Model(&models.ProjectChangeLog{}).Where("project_id = ? AND action = ?", pid, "STAGE_ADD").Count(&logCnt)
	if logCnt != 1 {
		t.Fatalf("期望 1 条 STAGE_ADD 变更日志，实际 %d", logCnt)
	}

	// 对照分支（极值）：名称超 100 字符 → 400
	_, err = svc.CreateStage(pid, 2, false, &dto.StageCreateRequest{
		StageName: strings.Repeat("长", 101), StartDate: "2026-10-11", EndDate: "2026-10-20",
	})
	assertBizCode(t, err, 400, "阶段名称长度不能超过100个字符")
}

// TestStage_Create_DuplicateName_Rejected 反例：同项目重名拒绝且零副作用
func TestStage_Create_DuplicateName_Rejected(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)
	createTestStage(t, svc, pid, 2, "需求调研", "2026-10-01", "2026-10-10")

	_, err := svc.CreateStage(pid, 2, false, &dto.StageCreateRequest{
		StageName: "需求调研", StartDate: "2026-10-11", EndDate: "2026-10-20",
	})
	assertBizCode(t, err, 400, "阶段名称已存在")

	var cnt, logCnt int64
	tx.Model(&models.ProjectStage{}).Where("project_id = ? AND deleted = 0", pid).Count(&cnt)
	tx.Model(&models.ProjectChangeLog{}).Where("project_id = ? AND action = ?", pid, "STAGE_ADD").Count(&logCnt)
	if cnt != 1 || logCnt != 1 {
		t.Fatalf("重名拒绝后应保持 1 行阶段与 1 条 STAGE_ADD，实际 %d/%d", cnt, logCnt)
	}
}

// TestStage_Create_RecreateAfterSoftDelete_Allowed 幂等：软删后同名可重建
func TestStage_Create_RecreateAfterSoftDelete_Allowed(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)
	first := createTestStage(t, svc, pid, 2, "需求调研", "2026-10-01", "2026-10-10")
	firstID := mustInt64(t, first, "id")
	if err := svc.DeleteStage(pid, firstID, 2, false); err != nil {
		t.Fatalf("删除阶段失败: %v", err)
	}

	second := createTestStage(t, svc, pid, 2, "需求调研", "2026-11-01", "2026-11-10")
	secondID := mustInt64(t, second, "id")
	if secondID == firstID {
		t.Fatalf("重建应为新行，id 不应复用: %d", secondID)
	}
	var total, active int64
	tx.Model(&models.ProjectStage{}).Where("project_id = ?", pid).Count(&total)
	tx.Model(&models.ProjectStage{}).Where("project_id = ? AND deleted = 0", pid).Count(&active)
	if total != 2 || active != 1 {
		t.Fatalf("期望累计 2 行（其中 1 行软删）且有效仅一次出现，实际 %d/%d", total, active)
	}
}

// TestStage_Create_EndBeforeStart_Rejected 边界：结束早于开始拒绝；结束等于开始允许
func TestStage_Create_EndBeforeStart_Rejected(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)

	_, err := svc.CreateStage(pid, 2, false, &dto.StageCreateRequest{
		StageName: "倒置阶段", StartDate: "2026-10-01", EndDate: "2026-09-30",
	})
	assertBizCode(t, err, 400, "结束日期不能早于开始日期")
	var cnt int64
	tx.Model(&models.ProjectStage{}).Where("project_id = ?", pid).Count(&cnt)
	if cnt != 0 {
		t.Fatalf("倒置区间拒绝后不应落库，实际 %d 行", cnt)
	}

	// 边界：结束与开始相等（单日阶段）允许
	sres, err := svc.CreateStage(pid, 2, false, &dto.StageCreateRequest{
		StageName: "单日阶段", StartDate: "2026-10-01", EndDate: "2026-10-01",
	})
	if err != nil {
		t.Fatalf("单日阶段创建失败: %v", err)
	}
	if mustStr(t, sres, "endDate") != "2026-10-01" {
		t.Fatalf("单日阶段 endDate 不符: %v", sres["endDate"])
	}
	tx.Model(&models.ProjectStage{}).Where("project_id = ? AND deleted = 0", pid).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("单日阶段期望落库 1 行，实际 %d", cnt)
	}
}

// TestStage_Create_NonManager_Forbidden_AdminAllowed 权限：非 PM 403；ADMIN（非成员）放行
func TestStage_Create_NonManager_Forbidden_AdminAllowed(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)

	_, err := svc.CreateStage(pid, 3, false, &dto.StageCreateRequest{
		StageName: "越权阶段", StartDate: "2026-10-01", EndDate: "2026-10-10",
	})
	assertBizCode(t, err, 403, "仅项目经理可管理阶段")
	var cnt int64
	tx.Model(&models.ProjectStage{}).Where("project_id = ?", pid).Count(&cnt)
	if cnt != 0 {
		t.Fatalf("越权拒绝后不应落库，实际 %d 行", cnt)
	}

	// 对照：ADMIN 放行
	if _, err = svc.CreateStage(pid, 3, true, &dto.StageCreateRequest{
		StageName: "管理员阶段", StartDate: "2026-10-11", EndDate: "2026-10-20",
	}); err != nil {
		t.Fatalf("ADMIN 创建阶段应放行: %v", err)
	}
}

// TestStage_Create_TerminalProject_Rejected 异常：终态项目拒绝维护阶段
func TestStage_Create_TerminalProject_Rejected(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)
	if _, err := svc.ChangeProjectStatus(pid, 2, false, &dto.ProjectStatusRequest{Action: "close"}); err != nil {
		t.Fatalf("结项失败: %v", err)
	}

	_, err := svc.CreateStage(pid, 2, false, &dto.StageCreateRequest{
		StageName: "终态阶段", StartDate: "2026-10-01", EndDate: "2026-10-10",
	})
	assertBizCode(t, err, 400, "项目已进入终态，不可维护阶段")
	var cnt, logCnt int64
	tx.Model(&models.ProjectStage{}).Where("project_id = ?", pid).Count(&cnt)
	tx.Model(&models.ProjectChangeLog{}).Where("project_id = ? AND action = ?", pid, "STAGE_ADD").Count(&logCnt)
	if cnt != 0 || logCnt != 0 {
		t.Fatalf("终态拒绝后不应落库与记日志，实际 %d/%d", cnt, logCnt)
	}
}

// TestStage_Update_PartialMerge_And_SelfRename_Allowed 边界：部分更新合并校验与改名查重排除自身
func TestStage_Update_PartialMerge_And_SelfRename_Allowed(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	res := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)
	a := createTestStage(t, svc, pid, 2, "需求调研", "2026-10-01", "2026-10-10")
	aID := mustInt64(t, a, "id")
	createTestStage(t, svc, pid, 2, "开发联调", "2026-10-11", "2026-11-30")

	// 部分更新：仅改说明，名称与日期保持不变
	ures, err := svc.UpdateStage(pid, aID, 2, false, &dto.StageUpdateRequest{Description: strPtr("新说明")})
	if err != nil {
		t.Fatalf("部分更新失败: %v", err)
	}
	if mustStr(t, ures, "description") != "新说明" || mustStr(t, ures, "stageName") != "需求调研" ||
		mustStr(t, ures, "startDate") != "2026-10-01" || mustStr(t, ures, "endDate") != "2026-10-10" {
		t.Fatalf("部分更新未保持未传字段: %#v", ures)
	}

	// 改名等于自身原名：查重排除自身，放行
	if _, err = svc.UpdateStage(pid, aID, 2, false, &dto.StageUpdateRequest{StageName: strPtr("需求调研")}); err != nil {
		t.Fatalf("自身同名改名应放行: %v", err)
	}

	// 改名等于同项目他阶段名：400 且原名称保持不变
	_, err = svc.UpdateStage(pid, aID, 2, false, &dto.StageUpdateRequest{StageName: strPtr("开发联调")})
	assertBizCode(t, err, 400, "阶段名称已存在")
	var cnt int64
	tx.Model(&models.ProjectStage{}).Where("project_id = ? AND stage_name = ? AND deleted = 0", pid, "需求调研").Count(&cnt)
	if cnt != 1 {
		t.Fatalf("冲突拒绝后原名称应保持不变，实际命中「需求调研」%d 行", cnt)
	}

	// 部分更新合并校验：仅改 start 晚于现 end → 400
	_, err = svc.UpdateStage(pid, aID, 2, false, &dto.StageUpdateRequest{StartDate: strPtr("2026-10-20")})
	assertBizCode(t, err, 400, "结束日期不能早于开始日期")
}

// TestStage_Delete_SoftDelete_ThenInvisibleAnd404 反例：软删不可见、重复删除与跨项目 404
func TestStage_Delete_SoftDelete_ThenInvisibleAnd404(t *testing.T) {
	svc, tx := setupProjectServiceTx(t)
	resA := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pidA := mustInt64(t, resA, "id")
	startTestProject(t, svc, pidA, 2)
	stage := createTestStage(t, svc, pidA, 2, "需求调研", "2026-10-01", "2026-10-10")
	stageID := mustInt64(t, stage, "id")
	list, err := svc.ListStages(pidA)
	if err != nil || len(list) != 1 {
		t.Fatalf("删除前列表期望 1 条: err=%v len=%d", err, len(list))
	}

	// 跨项目隔离：用他项目删除该阶段 → 404
	resB := createTestProject(t, svc, 2, "2026-10-01", "2026-12-31")
	pidB := mustInt64(t, resB, "id")
	err = svc.DeleteStage(pidB, stageID, 2, false)
	assertBizCode(t, err, 404, "阶段不存在")

	// 软删成功：行保留 deleted=1，列表/详情不可见，日志 1 条
	if err = svc.DeleteStage(pidA, stageID, 2, false); err != nil {
		t.Fatalf("删除阶段失败: %v", err)
	}
	var rowCnt, deletedCnt, logCnt int64
	tx.Model(&models.ProjectStage{}).Where("id = ?", stageID).Count(&rowCnt)
	tx.Model(&models.ProjectStage{}).Where("id = ? AND deleted = ?", stageID, 1).Count(&deletedCnt)
	if rowCnt != 1 || deletedCnt != 1 {
		t.Fatalf("软删后应保留 1 行且 deleted=1，实际 %d/%d", rowCnt, deletedCnt)
	}
	list, _ = svc.ListStages(pidA)
	detail, err := svc.GetProjectDetail(pidA, 2, false)
	if err != nil {
		t.Fatalf("查询详情失败: %v", err)
	}
	stageViews, _ := detail["stages"].([]map[string]interface{})
	if len(list) != 0 || len(stageViews) != 0 {
		t.Fatalf("软删后列表/详情应不可见，实际 %d/%d", len(list), len(stageViews))
	}
	tx.Model(&models.ProjectChangeLog{}).Where("project_id = ? AND action = ?", pidA, "STAGE_DELETE").Count(&logCnt)
	if logCnt != 1 {
		t.Fatalf("期望 1 条 STAGE_DELETE 变更日志，实际 %d", logCnt)
	}

	// 重复删除 → 404
	err = svc.DeleteStage(pidA, stageID, 2, false)
	assertBizCode(t, err, 404, "阶段不存在")
}

// TestResolveCurrentStageItem_InclusiveBoundary_Overlap_None 边界：含两端、重叠取最晚开始、并列取 id 最大、无命中
func TestResolveCurrentStageItem_InclusiveBoundary_Overlap_None(t *testing.T) {
	mk := func(id int64, start, end string) models.ProjectStage {
		s, _ := time.Parse("2006-01-02", start)
		e, _ := time.Parse("2006-01-02", end)
		return models.ProjectStage{ID: id, StartDate: s, EndDate: e}
	}
	a := mk(1, "2026-06-01", "2026-06-20")
	b := mk(2, "2026-06-10", "2026-06-30")

	// 含两端：today 等于 start / 等于 end 均命中（单阶段切片隔离重叠）
	if hit := resolveCurrentStageItem([]models.ProjectStage{a}, "2026-06-01"); hit == nil || hit.ID != 1 {
		t.Fatalf("today=start 期望命中 A，实际 %v", hit)
	}
	if hit := resolveCurrentStageItem([]models.ProjectStage{a}, "2026-06-20"); hit == nil || hit.ID != 1 {
		t.Fatalf("today=end 期望命中 A，实际 %v", hit)
	}

	// 重叠：取 start_date 最晚（B）
	if hit := resolveCurrentStageItem([]models.ProjectStage{a, b}, "2026-06-15"); hit == nil || hit.ID != 2 {
		t.Fatalf("重叠期望命中 B（start 最晚），实际 %v", hit)
	}

	// 无命中与空切片
	if hit := resolveCurrentStageItem([]models.ProjectStage{a, b}, "2026-07-01"); hit != nil {
		t.Fatalf("区间外期望 nil，实际 id=%d", hit.ID)
	}
	if hit := resolveCurrentStageItem(nil, "2026-06-15"); hit != nil {
		t.Fatalf("空切片期望 nil，实际 id=%d", hit.ID)
	}

	// 并列 start：取 id 最大
	c1 := mk(3, "2026-06-01", "2026-06-10")
	c2 := mk(4, "2026-06-01", "2026-06-10")
	if hit := resolveCurrentStageItem([]models.ProjectStage{c1, c2}, "2026-06-05"); hit == nil || hit.ID != 4 {
		t.Fatalf("并列 start 期望取 id 最大(4)，实际 %v", hit)
	}
}

// TestProjectDetail_StagesSorted_CurrentStageId 边界：详情正序 + 命中/无命中/终态不计算
func TestProjectDetail_StagesSorted_CurrentStageId(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)

	// 场景一：乱序创建 3 个阶段（后置 → 前置 → 跨今天），详情按 start_date 正序
	res := createTestProject(t, svc, 2, relDate(-20), relDate(30))
	pid := mustInt64(t, res, "id")
	startTestProject(t, svc, pid, 2)
	createTestStage(t, svc, pid, 2, "后置阶段", relDate(10), relDate(20))
	createTestStage(t, svc, pid, 2, "前置阶段", relDate(-20), relDate(-10))
	hit := createTestStage(t, svc, pid, 2, "进行阶段", relDate(-5), relDate(5))
	hitID := mustInt64(t, hit, "id")

	detail, err := svc.GetProjectDetail(pid, 2, false)
	if err != nil {
		t.Fatalf("查询详情失败: %v", err)
	}
	stageViews, ok := detail["stages"].([]map[string]interface{})
	if !ok || len(stageViews) != 3 {
		t.Fatalf("详情 stages 期望 3 条，实际 %#v", detail["stages"])
	}
	wantOrder := []string{"前置阶段", "进行阶段", "后置阶段"}
	for i, want := range wantOrder {
		if stageViews[i]["stageName"] != want {
			t.Fatalf("stages 正序不符：第 %d 条期望 %s，实际 %v", i, want, stageViews[i]["stageName"])
		}
	}
	if got := detail["currentStageId"]; got == nil || got.(int64) != hitID {
		t.Fatalf("currentStageId 期望命中 %d，实际 %v", hitID, got)
	}

	// 场景二：阶段不跨今天 → currentStageId 为 nil
	resB := createTestProject(t, svc, 2, relDate(30), relDate(60))
	pidB := mustInt64(t, resB, "id")
	startTestProject(t, svc, pidB, 2)
	createTestStage(t, svc, pidB, 2, "未来阶段", relDate(30), relDate(40))
	detailB, err := svc.GetProjectDetail(pidB, 2, false)
	if err != nil {
		t.Fatalf("查询详情失败: %v", err)
	}
	if got := detailB["currentStageId"]; got != nil {
		t.Fatalf("无命中期望 nil，实际 %v", got)
	}

	// 场景三：终态项目即使阶段跨今天也不计算（阶段数据仍保留）
	resC := createTestProject(t, svc, 2, relDate(-20), relDate(30))
	pidC := mustInt64(t, resC, "id")
	startTestProject(t, svc, pidC, 2)
	createTestStage(t, svc, pidC, 2, "终止阶段", relDate(-5), relDate(5))
	if _, err := svc.ChangeProjectStatus(pidC, 2, false, &dto.ProjectStatusRequest{Action: "close"}); err != nil {
		t.Fatalf("结项失败: %v", err)
	}
	detailC, err := svc.GetProjectDetail(pidC, 2, false)
	if err != nil {
		t.Fatalf("查询详情失败: %v", err)
	}
	if got := detailC["currentStageId"]; got != nil {
		t.Fatalf("终态项目期望 currentStageId 为 nil，实际 %v", got)
	}
	if stageViewsC := detailC["stages"].([]map[string]interface{}); len(stageViewsC) != 1 {
		t.Fatalf("终态项目阶段数据应保留，实际 %d 条", len(stageViewsC))
	}
}

// TestListProjects_CurrentStage_AndTerminalEmpty 边界：清单 currentStage 命中/无阶段/终态为空串
func TestListProjects_CurrentStage_AndTerminalEmpty(t *testing.T) {
	svc, _ := setupProjectServiceTx(t)

	resA := createTestProject(t, svc, 2, relDate(-10), relDate(30))
	pidA := mustInt64(t, resA, "id")
	startTestProject(t, svc, pidA, 2)
	createTestStage(t, svc, pidA, 2, "开发阶段", relDate(-5), relDate(5))

	resB := createTestProject(t, svc, 2, relDate(-10), relDate(30))
	pidB := mustInt64(t, resB, "id")
	startTestProject(t, svc, pidB, 2)

	resC := createTestProject(t, svc, 2, relDate(-10), relDate(30))
	pidC := mustInt64(t, resC, "id")
	startTestProject(t, svc, pidC, 2)
	createTestStage(t, svc, pidC, 2, "终止阶段", relDate(-5), relDate(5))
	if _, err := svc.ChangeProjectStatus(pidC, 2, false, &dto.ProjectStatusRequest{Action: "close"}); err != nil {
		t.Fatalf("结项失败: %v", err)
	}

	result, err := svc.ListProjects(2, false, 1, 200, map[string]interface{}{"keyword": "TPRJ", "participant_id": "2"})
	if err != nil {
		t.Fatalf("查询列表失败: %v", err)
	}
	stageByID := map[int64]string{}
	for _, item := range result["list"].([]map[string]interface{}) {
		id, _ := item["id"].(int64)
		if v, ok := item["currentStage"].(string); ok {
			stageByID[id] = v
		}
	}
	if got := stageByID[pidA]; got != "开发阶段" {
		t.Fatalf("命中项目期望 currentStage=开发阶段，实际 %q", got)
	}
	if got, ok := stageByID[pidB]; !ok || got != "" {
		t.Fatalf("无阶段项目期望 currentStage 为空串，实际 %q（存在=%v）", got, ok)
	}
	if got := stageByID[pidC]; got != "" {
		t.Fatalf("终态项目期望 currentStage 为空串，实际 %q", got)
	}
}
