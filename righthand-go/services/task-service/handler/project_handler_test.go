package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"righthand-go/internal/config"
	"righthand-go/internal/database"
	"righthand-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 项目/消息 Handler 测试：真实 MySQL + 唯一 Key + 清理。
// 响应契约为 HTTP 200 + body {code,message,data}，断言 body code（对齐前端拦截器）。

func setupProjectHandlerTestDB(t *testing.T) *gorm.DB {
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

// setupProjectHandlerRouter 注册项目/消息路由并注入指定身份
func setupProjectHandlerRouter(t *testing.T, userID int64, isAdmin bool) *gin.Engine {
	t.Helper()
	setupProjectHandlerTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ph := NewProjectHandler()
	mh := NewMessageHandler()
	auth := func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("isAdmin", isAdmin)
		c.Next()
	}
	g := router.Group("/api", auth)
	{
		g.GET("/projects", ph.ListProjects)
		g.POST("/projects", ph.CreateProject)
		g.GET("/projects/:id", ph.GetProjectDetail)
		g.PUT("/projects/:id", ph.UpdateProject)
		g.DELETE("/projects/:id", ph.DeleteProject)
		g.PUT("/projects/:id/status", ph.ChangeProjectStatus)
		g.PUT("/projects/:id/risk", ph.ChangeProjectRisk)
		g.GET("/projects/:id/stages", ph.ListStages)
		g.POST("/projects/:id/stages", ph.CreateStage)
		g.PUT("/projects/:id/stages/:stageId", ph.UpdateStage)
		g.DELETE("/projects/:id/stages/:stageId", ph.DeleteStage)
		g.GET("/messages/unread-count", mh.UnreadCount)
		g.PUT("/messages/:id/read", mh.MarkRead)
	}
	return router
}

var handlerKeySeq int

func uniqueHandlerKey() string {
	handlerKeySeq++
	return fmt.Sprintf("HPRJ%d%d", time.Now().UnixNano()%100000000, handlerKeySeq)
}

// cleanupProjectRows 物理清理测试项目及其关联（避免污染开发库）
func cleanupProjectRows(t *testing.T, db *gorm.DB, projectID int64) {
	t.Helper()
	var weekIDs []int64
	db.Model(&models.ProjectWeeklyPlan{}).Where("project_id = ?", projectID).Pluck("id", &weekIDs)
	if len(weekIDs) > 0 {
		db.Exec("DELETE FROM task_week_plan_rel WHERE week_plan_id IN ?", weekIDs)
	}
	db.Exec("DELETE FROM project_weekly_plan WHERE project_id = ?", projectID)
	db.Exec("DELETE FROM project_member WHERE project_id = ?", projectID)
	db.Exec("DELETE FROM project_stage WHERE project_id = ?", projectID)
	db.Exec("DELETE FROM project_change_log WHERE project_id = ?", projectID)
	db.Exec("DELETE FROM project WHERE id = ?", projectID)
}

// doJSON 发送 JSON 请求并解析统一响应体
func doJSON(t *testing.T, router *gin.Engine, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("序列化请求体失败: %v", err)
		}
		buf = bytes.NewBuffer(raw)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s 期望 HTTP 200（统一响应契约），实际 %d: %s", method, path, w.Code, w.Body.String())
	}
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("响应体非 JSON: %s", w.Body.String())
	}
	return w.Code, out
}

func bodyCode(t *testing.T, out map[string]interface{}) int {
	t.Helper()
	code, ok := out["code"].(float64)
	if !ok {
		t.Fatalf("响应缺少 code 字段: %#v", out)
	}
	return int(code)
}

func bodyData(t *testing.T, out map[string]interface{}) map[string]interface{} {
	t.Helper()
	data, ok := out["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("响应缺少 data 对象: %#v", out)
	}
	return data
}

func bodyMessage(t *testing.T, out map[string]interface{}) string {
	t.Helper()
	msg, _ := out["message"].(string)
	return msg
}

func TestProjectHandler_CreateValidationAndDetail(t *testing.T) {
	router := setupProjectHandlerRouter(t, 2, false)
	db := database.GetDB()

	// 反例：缺必填 → code 400
	_, out := doJSON(t, router, "POST", "/api/projects", map[string]interface{}{"projectName": "缺字段"})
	if bodyCode(t, out) != 400 || !strings.Contains(bodyMessage(t, out), "必填") {
		t.Fatalf("缺必填期望 400/必填，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}

	// 正例：创建成功 → data 含 projectKey 与 allowedActions(含 start)
	key := uniqueHandlerKey()
	_, out = doJSON(t, router, "POST", "/api/projects", map[string]interface{}{
		"projectName": "Handler测试-" + key,
		"projectKey":  key,
		"managerId":   2,
		"startDate":   "2026-01-05",
		"endDate":     "2026-02-27",
	})
	if bodyCode(t, out) != 200 {
		t.Fatalf("创建项目失败: %s", bodyMessage(t, out))
	}
	data := bodyData(t, out)
	if data["projectKey"] != key {
		t.Fatalf("projectKey 不符: %v", data["projectKey"])
	}
	pid := int64(data["id"].(float64))
	t.Cleanup(func() { cleanupProjectRows(t, db, pid) })
	actions, _ := data["allowedActions"].([]interface{})
	if !containsAction(actions, "start") {
		t.Fatalf("创建后 allowedActions 应含 start，实际 %v", actions)
	}

	// 正例：详情 → 200，含 allowedActions
	_, out = doJSON(t, router, "GET", fmt.Sprintf("/api/projects/%d", pid), nil)
	if bodyCode(t, out) != 200 {
		t.Fatalf("详情失败: %s", bodyMessage(t, out))
	}
	if !containsAction(bodyData(t, out)["allowedActions"].([]interface{}), "delete") {
		t.Fatalf("STARTING 详情 allowedActions 应含 delete: %v", bodyData(t, out)["allowedActions"])
	}

	// 反例：无效 ID → 400；不存在 → 404
	_, out = doJSON(t, router, "GET", "/api/projects/abc", nil)
	if bodyCode(t, out) != 400 || !strings.Contains(bodyMessage(t, out), "无效") {
		t.Fatalf("无效 ID 期望 400/无效，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}
	_, out = doJSON(t, router, "GET", "/api/projects/999999999", nil)
	if bodyCode(t, out) != 404 {
		t.Fatalf("不存在项目期望 404，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}

	// 正例：删除 STARTING 项目 → 200/删除成功；删除后再查 404
	_, out = doJSON(t, router, "DELETE", fmt.Sprintf("/api/projects/%d", pid), nil)
	if bodyCode(t, out) != 200 || bodyMessage(t, out) != "删除成功" {
		t.Fatalf("删除期望 200/删除成功，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}
	_, out = doJSON(t, router, "GET", fmt.Sprintf("/api/projects/%d", pid), nil)
	if bodyCode(t, out) != 404 {
		t.Fatalf("删除后详情期望 404，实际 %d", bodyCode(t, out))
	}
}

func containsAction(actions []interface{}, target string) bool {
	for _, a := range actions {
		if s, ok := a.(string); ok && s == target {
			return true
		}
	}
	return false
}

func TestProjectHandler_PermissionAndManualProgressIgnored(t *testing.T) {
	// 以 user 2（PM/创建人）创建项目
	router := setupProjectHandlerRouter(t, 2, false)
	db := database.GetDB()
	key := uniqueHandlerKey()
	_, out := doJSON(t, router, "POST", "/api/projects", map[string]interface{}{
		"projectName": "Handler权限-" + key,
		"projectKey":  key,
		"managerId":   2,
		"startDate":   "2026-01-05",
		"endDate":     "2026-02-27",
	})
	pid := int64(bodyData(t, out)["id"].(float64))
	t.Cleanup(func() { cleanupProjectRows(t, db, pid) })

	// 正例：手动提交 progressPct 被忽略（DTO 无该字段，进度仅服务端计算）
	_, out = doJSON(t, router, "PUT", fmt.Sprintf("/api/projects/%d", pid), map[string]interface{}{
		"description": "尝试篡改进度", "progressPct": 55,
	})
	if bodyCode(t, out) != 200 {
		t.Fatalf("编辑失败: %s", bodyMessage(t, out))
	}
	if got := bodyData(t, out)["progressPct"]; got != float64(0) {
		t.Fatalf("progressPct 应被忽略保持 0，实际 %v", got)
	}

	// 反例：其他用户（user 3）视角 allowedActions 为空
	router3 := setupProjectHandlerRouter(t, 3, false)
	_, out = doJSON(t, router3, "GET", fmt.Sprintf("/api/projects/%d", pid), nil)
	actions, _ := bodyData(t, out)["allowedActions"].([]interface{})
	if len(actions) != 0 {
		t.Fatalf("非 PM 非创建人 allowedActions 应为空，实际 %v", actions)
	}

	// 反例：非 PM 删除 → 403
	_, out = doJSON(t, router3, "DELETE", fmt.Sprintf("/api/projects/%d", pid), nil)
	if bodyCode(t, out) != 403 {
		t.Fatalf("非 PM 删除期望 403，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}

	// 反例：非 PM 开始项目 → 403（路由层透传业务码）
	_, out = doJSON(t, router3, "PUT", fmt.Sprintf("/api/projects/%d/status", pid), map[string]interface{}{"action": "start"})
	if bodyCode(t, out) != 403 {
		t.Fatalf("非 PM 状态流转期望 403，实际 %d", bodyCode(t, out))
	}
}

func TestProjectHandler_MessageEndpoints(t *testing.T) {
	router := setupProjectHandlerRouter(t, 2, false)
	db := database.GetDB()

	// 基线未读数
	_, out := doJSON(t, router, "GET", "/api/messages/unread-count", nil)
	if bodyCode(t, out) != 200 {
		t.Fatalf("未读数查询失败: %s", bodyMessage(t, out))
	}
	baseCount := bodyData(t, out)["count"].(float64)

	// 反例：标记不存在的消息 → 404
	_, out = doJSON(t, router, "PUT", "/api/messages/999999999/read", nil)
	if bodyCode(t, out) != 404 || !strings.Contains(bodyMessage(t, out), "消息不存在") {
		t.Fatalf("不存在消息期望 404，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}

	// 正例：插入一条本人未读消息 → 未读数 +1 → 标记已读 → 回落
	msg := &models.Message{UserID: 2, Title: "单测消息", Content: "内容", Type: "PROJECT"}
	if err := db.Create(msg).Error; err != nil {
		t.Fatalf("插入测试消息失败: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM sys_message WHERE id = ?", msg.ID) })

	_, out = doJSON(t, router, "GET", "/api/messages/unread-count", nil)
	if got := bodyData(t, out)["count"].(float64); got != baseCount+1 {
		t.Fatalf("插入后未读数期望 %v，实际 %v", baseCount+1, got)
	}
	_, out = doJSON(t, router, "PUT", fmt.Sprintf("/api/messages/%d/read", msg.ID), nil)
	if bodyCode(t, out) != 200 {
		t.Fatalf("标记已读失败: %s", bodyMessage(t, out))
	}
	_, out = doJSON(t, router, "GET", "/api/messages/unread-count", nil)
	if got := bodyData(t, out)["count"].(float64); got != baseCount {
		t.Fatalf("标记已读后未读数期望 %v，实际 %v", baseCount, got)
	}
}

// TestStageHandler_InvalidId_BadRequest_RoutesExist 异常：阶段接口链路（创建/列表/编辑/软删/重复删除 404）+ 无效 ID 400
func TestStageHandler_InvalidId_BadRequest_RoutesExist(t *testing.T) {
	router := setupProjectHandlerRouter(t, 2, false)
	db := database.GetDB()

	// 链路前置：创建项目（STARTING 非终态，阶段维护放行）
	key := uniqueHandlerKey()
	_, out := doJSON(t, router, "POST", "/api/projects", map[string]interface{}{
		"projectName": "Handler阶段-" + key,
		"projectKey":  key,
		"managerId":   2,
		"startDate":   "2026-01-05",
		"endDate":     "2026-02-27",
	})
	if bodyCode(t, out) != 200 {
		t.Fatalf("创建项目失败: %s", bodyMessage(t, out))
	}
	pid := int64(bodyData(t, out)["id"].(float64))
	t.Cleanup(func() { cleanupProjectRows(t, db, pid) })

	// 正例：新增阶段 → 200 创建成功
	_, out = doJSON(t, router, "POST", fmt.Sprintf("/api/projects/%d/stages", pid), map[string]interface{}{
		"stageName": "需求调研", "startDate": "2026-01-05", "endDate": "2026-01-20",
	})
	if bodyCode(t, out) != 200 || bodyMessage(t, out) != "创建成功" {
		t.Fatalf("新增阶段期望 200/创建成功，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}
	stageID := int64(bodyData(t, out)["id"].(float64))

	// 正例：列表 → 200 且含新增阶段
	_, out = doJSON(t, router, "GET", fmt.Sprintf("/api/projects/%d/stages", pid), nil)
	if bodyCode(t, out) != 200 {
		t.Fatalf("阶段列表失败: %s", bodyMessage(t, out))
	}
	if list, ok := out["data"].([]interface{}); !ok || len(list) != 1 {
		t.Fatalf("阶段列表期望 1 条，实际 %#v", out["data"])
	}

	// 正例：编辑 → 200 保存成功
	_, out = doJSON(t, router, "PUT", fmt.Sprintf("/api/projects/%d/stages/%d", pid, stageID), map[string]interface{}{
		"stageName": "需求调研一期",
	})
	if bodyCode(t, out) != 200 || bodyMessage(t, out) != "保存成功" {
		t.Fatalf("编辑阶段期望 200/保存成功，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}

	// 反例：无效 ID → 400
	_, out = doJSON(t, router, "GET", "/api/projects/abc/stages", nil)
	if bodyCode(t, out) != 400 || !strings.Contains(bodyMessage(t, out), "无效的项目ID") {
		t.Fatalf("无效项目ID期望 400/无效的项目ID，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}
	_, out = doJSON(t, router, "PUT", fmt.Sprintf("/api/projects/%d/stages/abc", pid), map[string]interface{}{
		"description": "改",
	})
	if bodyCode(t, out) != 400 || !strings.Contains(bodyMessage(t, out), "无效的阶段ID") {
		t.Fatalf("无效阶段ID期望 400/无效的阶段ID，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}

	// 软删 → 200 删除成功；重复删除 → 404；行保留 deleted=1
	_, out = doJSON(t, router, "DELETE", fmt.Sprintf("/api/projects/%d/stages/%d", pid, stageID), nil)
	if bodyCode(t, out) != 200 || bodyMessage(t, out) != "删除成功" {
		t.Fatalf("删除阶段期望 200/删除成功，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}
	_, out = doJSON(t, router, "DELETE", fmt.Sprintf("/api/projects/%d/stages/%d", pid, stageID), nil)
	if bodyCode(t, out) != 404 {
		t.Fatalf("重复删除期望 404，实际 %d/%s", bodyCode(t, out), bodyMessage(t, out))
	}
	var deletedCnt int64
	db.Model(&models.ProjectStage{}).Where("id = ? AND deleted = 1", stageID).Count(&deletedCnt)
	if deletedCnt != 1 {
		t.Fatalf("软删后应保留 deleted=1 行，实际 %d", deletedCnt)
	}
}
