package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"righthand-go/services/task-service/dto"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BizError 业务错误（携带 HTTP 语义码，handler 统一转换响应）
type BizError struct {
	Code    int
	Message string
}

func (e *BizError) Error() string { return e.Message }

func bizErr(code int, msg string) error { return &BizError{Code: code, Message: msg} }

// 状态/标签映射（仅服务端生成文案用）
var (
	projectStatusLabels = map[string]string{
		"STARTING": "启动", "IN_PROGRESS": "进行中", "PAUSED": "已暂停", "CLOSED": "已结项", "TERMINATED": "已终止",
	}
	weekStatusLabels = map[string]string{
		"NOT_STARTED": "未开始", "IN_PROGRESS": "进行中", "DONE": "已完成", "SKIPPED": "已跳过",
	}
	riskLevelLabels      = map[string]string{"HIGH": "高", "MEDIUM": "中", "LOW": "低"}
	deliveryStatusLabels = map[string]string{
		"NORMAL": "正常", "AT_RISK": "有延期风险", "CANNOT_FINISH": "预计无法完成",
	}
	projectStatusActions = map[string]string{
		"start": "开始", "pause": "暂停", "resume": "恢复", "close": "结项", "terminate": "终止",
	}
)

// projectTransitions 项目状态机：action -> {From 集合, To}
var projectTransitions = map[string]struct {
	From []string
	To   string
}{
	"start":     {From: []string{"STARTING"}, To: "IN_PROGRESS"},
	"pause":     {From: []string{"IN_PROGRESS"}, To: "PAUSED"},
	"resume":    {From: []string{"PAUSED"}, To: "IN_PROGRESS"},
	"close":     {From: []string{"IN_PROGRESS"}, To: "CLOSED"},
	"terminate": {From: []string{"IN_PROGRESS", "PAUSED"}, To: "TERMINATED"},
}

// weekTransitions 周状态机：action -> {From 集合, To}
var weekTransitions = map[string]struct {
	From []string
	To   string
}{
	"start":  {From: []string{"NOT_STARTED"}, To: "IN_PROGRESS"},
	"done":   {From: []string{"IN_PROGRESS"}, To: "DONE"},
	"skip":   {From: []string{"NOT_STARTED", "IN_PROGRESS"}, To: "SKIPPED"},
	"reopen": {From: []string{"DONE", "SKIPPED"}, To: "IN_PROGRESS"},
}

var validWeekStatuses = []string{"NOT_STARTED", "IN_PROGRESS", "DONE", "SKIPPED"}

// ProjectService 项目服务
type ProjectService struct {
	db *gorm.DB // 测试时注入,否则使用全局DB
}

func NewProjectService() *ProjectService {
	return &ProjectService{}
}

// NewProjectServiceWithDB 使用指定 DB 构造（调度器/测试注入）
func NewProjectServiceWithDB(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// getDB 获取数据库连接
func (s *ProjectService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

/* ------------------------------ 通用工具 ------------------------------ */

func fmtDateStr(t time.Time) string { return t.Format("2006-01-02") }

func fmtDateTimeStr(t time.Time) string { return t.Format("2006-01-02 15:04:05") }

// fmtDateTimePtr nil -> nil，非 nil -> "2006-01-02 15:04:05"
func fmtDateTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return fmtDateTimeStr(*t)
}

func isTerminalStatus(status string) bool {
	return status == "CLOSED" || status == "TERMINATED"
}

func containsStr(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// canActAsPMOrAdmin ADMIN 或项目经理/创建人
func canActAsPMOrAdmin(p *models.Project, userID int64, isAdmin bool) bool {
	if isAdmin || p.ManagerID == userID {
		return true
	}
	return p.CreateUser != nil && *p.CreateUser == userID
}

// canActAsManagerOnly ADMIN 或项目经理（成员管理/新增周计划/解锁总结）
func canActAsManagerOnly(p *models.Project, userID int64, isAdmin bool) bool {
	return isAdmin || p.ManagerID == userID
}

/* ------------------------------ 查询辅助 ------------------------------ */

func (s *ProjectService) getProject(id int64) (*models.Project, error) {
	var p models.Project
	if err := s.getDB().Where("id = ? AND deleted = ?", id, 0).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizErr(404, "项目不存在")
		}
		return nil, err
	}
	return &p, nil
}

func (s *ProjectService) getWeek(projectID, weekID int64) (*models.ProjectWeeklyPlan, error) {
	var w models.ProjectWeeklyPlan
	if err := s.getDB().Where("id = ? AND project_id = ? AND deleted = ?", weekID, projectID, 0).First(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizErr(404, "周计划不存在")
		}
		return nil, err
	}
	return &w, nil
}

// userNames 批量查询用户姓名（去重、忽略非正数）
func (s *ProjectService) userNames(ids []int64) map[int64]string {
	result := make(map[int64]string)
	uniq := make([]int64, 0, len(ids))
	seen := make(map[int64]bool)
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			uniq = append(uniq, id)
		}
	}
	if len(uniq) == 0 {
		return result
	}
	var users []models.SysUser
	s.getDB().Model(&models.SysUser{}).Select("id", "real_name").Where("id IN ?", uniq).Find(&users)
	for _, u := range users {
		result[u.ID] = u.RealName
	}
	return result
}

func addProjectLog(db *gorm.DB, projectID int64, action, detail string, operatorID int64) error {
	return db.Create(&models.ProjectChangeLog{
		ProjectID:  projectID,
		Action:     action,
		Detail:     detail,
		OperatorID: operatorID,
	}).Error
}

// AddChangeLog 对外暴露的变更日志写入（调度器使用）
func (s *ProjectService) AddChangeLog(projectID int64, action, detail string, operatorID int64) error {
	return addProjectLog(s.getDB(), projectID, action, detail, operatorID)
}

/* ------------------------------ 权限计算 ------------------------------ */

// projectAllowedActions 按状态+当前用户计算项目可用操作
func projectAllowedActions(p *models.Project, userID int64, isAdmin bool) []string {
	actions := []string{}
	if isTerminalStatus(p.Status) {
		return actions
	}
	pmOrAdmin := canActAsPMOrAdmin(p, userID, isAdmin)
	managerOnly := canActAsManagerOnly(p, userID, isAdmin)
	switch p.Status {
	case "STARTING":
		if pmOrAdmin {
			actions = append(actions, "start", "edit", "delete", "editRisk")
		}
	case "IN_PROGRESS":
		if pmOrAdmin {
			actions = append(actions, "pause", "close", "terminate", "edit", "editRisk")
		}
		if managerOnly {
			actions = append(actions, "manageMember", "addWeek", "manageStage")
		}
	case "PAUSED":
		if pmOrAdmin {
			actions = append(actions, "resume", "terminate", "edit", "editRisk")
		}
		if managerOnly {
			actions = append(actions, "manageMember", "addWeek", "manageStage")
		}
	}
	return actions
}

// weekAllowedActions 按状态+当前用户计算周计划可用操作（总结已提交 -> 仅解锁）
func weekAllowedActions(p *models.Project, w *models.ProjectWeeklyPlan, userID int64, isAdmin bool) []string {
	actions := []string{}
	if isTerminalStatus(p.Status) {
		return actions
	}
	if w.SummaryAt != nil {
		if canActAsManagerOnly(p, userID, isAdmin) {
			actions = append(actions, "unlockSummary")
		}
		return actions
	}
	if userID == w.OwnerID || canActAsPMOrAdmin(p, userID, isAdmin) {
		actions = append(actions, "edit", "delete", "changeStatus", "submitSummary")
	}
	return actions
}

/* ------------------------------ 视图构建 ------------------------------ */

func (s *ProjectService) projectView(p *models.Project, userID int64, isAdmin bool, managerName string, memberCount int64) map[string]interface{} {
	view := map[string]interface{}{
		"id":             p.ID,
		"projectName":    p.ProjectName,
		"projectKey":     p.ProjectKey,
		"description":    p.Description,
		"planDesc":       p.PlanDesc,
		"managerId":      p.ManagerID,
		"managerName":    managerName,
		"startDate":      fmtDateStr(p.StartDate),
		"endDate":        fmtDateStr(p.EndDate),
		"status":         p.Status,
		"riskLevel":      p.RiskLevel,
		"deliveryStatus": p.DeliveryStatus,
		"riskNote":       p.RiskNote,
		"progressPct":    p.ProgressPct,
		"weekCount":      p.WeekCount,
		"doneWeekCount":  p.DoneWeekCount,
		"pausedAt":       fmtDateTimePtr(p.PausedAt),
		"closedAt":       fmtDateTimePtr(p.ClosedAt),
		"createUser":     p.CreateUser,
		"createTime":     fmtDateTimeStr(p.CreateTime),
		"updateTime":     fmtDateTimeStr(p.UpdateTime),
		"deleted":        p.Deleted,
		"memberCount":    memberCount,
		"allowedActions": projectAllowedActions(p, userID, isAdmin),
	}
	if p.CreateUser != nil {
		createNames := s.userNames([]int64{*p.CreateUser})
		view["createUserName"] = createNames[*p.CreateUser]
	}
	return view
}

// singleProjectView 单项目视图（自动查经理姓名与成员数）
func (s *ProjectService) singleProjectView(p *models.Project, userID int64, isAdmin bool) map[string]interface{} {
	names := s.userNames([]int64{p.ManagerID})
	var memberCount int64
	s.getDB().Model(&models.ProjectMember{}).Where("project_id = ? AND deleted = ?", p.ID, 0).Count(&memberCount)
	return s.projectView(p, userID, isAdmin, names[p.ManagerID], memberCount)
}

func (s *ProjectService) buildWeekView(p *models.Project, w *models.ProjectWeeklyPlan, userID int64, isAdmin bool, names map[int64]string, taskCount int64) map[string]interface{} {
	var summary interface{}
	if w.Summary != nil && *w.Summary != "" {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(*w.Summary), &obj); err == nil {
			summary = obj
		}
	}
	var summaryByName interface{}
	if w.SummaryBy != nil {
		if n, ok := names[*w.SummaryBy]; ok {
			summaryByName = n
		}
	}
	todayStr := time.Now().Format("2006-01-02")
	endStr := fmtDateStr(w.EndDate)
	return map[string]interface{}{
		"id":             w.ID,
		"projectId":      w.ProjectID,
		"weekNo":         w.WeekNo,
		"title":          w.Title,
		"planDesc":       w.PlanDesc,
		"startDate":      fmtDateStr(w.StartDate),
		"endDate":        endStr,
		"ownerId":        w.OwnerID,
		"ownerName":      names[w.OwnerID],
		"status":         w.Status,
		"remark":         w.Remark,
		"summary":        summary,
		"summaryBy":      w.SummaryBy,
		"summaryByName":  summaryByName,
		"summaryAt":      fmtDateTimePtr(w.SummaryAt),
		"sortOrder":      w.SortOrder,
		"overdue":        endStr < todayStr && (w.Status == "NOT_STARTED" || w.Status == "IN_PROGRESS"),
		"taskCount":      taskCount,
		"createTime":     fmtDateTimeStr(w.CreateTime),
		"updateTime":     fmtDateTimeStr(w.UpdateTime),
		"deleted":        w.Deleted,
		"allowedActions": weekAllowedActions(p, w, userID, isAdmin),
	}
}

// singleWeekView 单周视图（自动查姓名与任务数）
func (s *ProjectService) singleWeekView(p *models.Project, w *models.ProjectWeeklyPlan, userID int64, isAdmin bool) map[string]interface{} {
	ids := []int64{w.OwnerID}
	if w.SummaryBy != nil {
		ids = append(ids, *w.SummaryBy)
	}
	names := s.userNames(ids)
	var taskCount int64
	s.getDB().Model(&models.TaskWeekPlanRel{}).Where("week_plan_id = ?", w.ID).Count(&taskCount)
	return s.buildWeekView(p, w, userID, isAdmin, names, taskCount)
}

/* ------------------------------ 项目 ------------------------------ */

// ListProjects 项目列表（筛选/排序/分页，批量聚合经理姓名与成员数避免 N+1）
func (s *ProjectService) ListProjects(userID int64, isAdmin bool, page, size int, filter map[string]interface{}) (map[string]interface{}, error) {
	db := s.getDB()
	query := db.Model(&models.Project{}).Where("deleted = ?", 0)

	if v, ok := filter["status"].(string); ok && strings.TrimSpace(v) != "" {
		statuses := []string{}
		for _, item := range strings.Split(v, ",") {
			if item = strings.TrimSpace(item); item != "" {
				statuses = append(statuses, item)
			}
		}
		if len(statuses) > 0 {
			query = query.Where("status IN ?", statuses)
		}
	}
	if v, ok := filter["risk_level"].(string); ok && v != "" {
		query = query.Where("risk_level = ?", v)
	}
	if v, ok := filter["delivery_status"].(string); ok && v != "" {
		query = query.Where("delivery_status = ?", v)
	}
	if v, ok := filter["manager_id"].(string); ok && v != "" {
		if managerID, err := strconv.ParseInt(v, 10, 64); err == nil {
			query = query.Where("manager_id = ?", managerID)
		}
	}
	if v, ok := filter["keyword"].(string); ok && strings.TrimSpace(v) != "" {
		kw := "%" + strings.TrimSpace(v) + "%"
		query = query.Where("(project_name LIKE ? OR project_key LIKE ?)", kw, kw)
	}
	// 日期范围：与项目起止区间取交集
	if v, ok := filter["start_date"].(string); ok && v != "" {
		query = query.Where("end_date >= ?", v)
	}
	if v, ok := filter["end_date"].(string); ok && v != "" {
		query = query.Where("start_date <= ?", v)
	}
	// 参与人：所选用户为项目经理或项目成员
	if v, ok := filter["participant_id"].(string); ok && v != "" {
		if participantID, err := strconv.ParseInt(v, 10, 64); err == nil {
			query = query.Where("(manager_id = ? OR id IN (SELECT project_id FROM project_member WHERE user_id = ? AND deleted = 0))", participantID, participantID)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序：默认 end_date 升序，可切换
	orderColumn := "end_date"
	if v, ok := filter["sort_by"].(string); ok {
		if col, exists := map[string]string{"endDate": "end_date", "startDate": "start_date", "updateTime": "update_time"}[v]; exists {
			orderColumn = col
		}
	}
	orderDir := "ASC"
	if v, ok := filter["order"].(string); ok && strings.EqualFold(v, "desc") {
		orderDir = "DESC"
	}
	query = query.Order(orderColumn + " " + orderDir + ", id ASC")

	var projects []models.Project
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&projects).Error; err != nil {
		return nil, err
	}

	// 批量聚合经理姓名与成员数
	ids := make([]int64, 0, len(projects))
	managerIDs := make([]int64, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
		managerIDs = append(managerIDs, p.ManagerID)
	}
	names := s.userNames(managerIDs)
	memberCounts := map[int64]int64{}
	if len(ids) > 0 {
		type countRow struct {
			ProjectID int64
			Cnt       int64
		}
		var rows []countRow
		db.Model(&models.ProjectMember{}).
			Select("project_id, COUNT(*) AS cnt").
			Where("project_id IN ? AND deleted = ?", ids, 0).
			Group("project_id").Scan(&rows)
		for _, r := range rows {
			memberCounts[r.ProjectID] = r.Cnt
		}
	}

	// 批量查询各项目阶段并计算「当前阶段」（终态项目不计算）
	stagesByProject := map[int64][]models.ProjectStage{}
	if len(ids) > 0 {
		var stageRows []models.ProjectStage
		db.Where("project_id IN ? AND deleted = ?", ids, 0).Order("start_date ASC, id ASC").Find(&stageRows)
		for _, st := range stageRows {
			stagesByProject[st.ProjectID] = append(stagesByProject[st.ProjectID], st)
		}
	}
	todayStr := time.Now().Format("2006-01-02")

	list := make([]map[string]interface{}, 0, len(projects))
	for i := range projects {
		p := &projects[i]
		view := s.projectView(p, userID, isAdmin, names[p.ManagerID], memberCounts[p.ID])
		currentStage := ""
		if !isTerminalStatus(p.Status) {
			if hit := resolveCurrentStageItem(stagesByProject[p.ID], todayStr); hit != nil {
				currentStage = hit.StageName
			}
		}
		view["currentStage"] = currentStage
		list = append(list, view)
	}

	return map[string]interface{}{
		"total":    total,
		"pageNum":  page,
		"pageSize": size,
		"list":     list,
	}, nil
}

// GetProjectDetail 项目详情（含周计划/成员/变更历史）
func (s *ProjectService) GetProjectDetail(id, userID int64, isAdmin bool) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(id)
	if err != nil {
		return nil, err
	}

	// 周计划
	var weeks []models.ProjectWeeklyPlan
	if err := db.Where("project_id = ? AND deleted = ?", p.ID, 0).Order("week_no ASC, id ASC").Find(&weeks).Error; err != nil {
		return nil, err
	}
	weekIDs := make([]int64, 0, len(weeks))
	nameIDs := []int64{p.ManagerID}
	if p.CreateUser != nil {
		nameIDs = append(nameIDs, *p.CreateUser)
	}
	for _, w := range weeks {
		weekIDs = append(weekIDs, w.ID)
		nameIDs = append(nameIDs, w.OwnerID)
		if w.SummaryBy != nil {
			nameIDs = append(nameIDs, *w.SummaryBy)
		}
	}
	names := s.userNames(nameIDs)

	// 每周任务数
	taskCounts := map[int64]int64{}
	if len(weekIDs) > 0 {
		type countRow struct {
			WeekPlanID int64
			Cnt        int64
		}
		var rows []countRow
		db.Model(&models.TaskWeekPlanRel{}).
			Select("week_plan_id, COUNT(*) AS cnt").
			Where("week_plan_id IN ?", weekIDs).
			Group("week_plan_id").Scan(&rows)
		for _, r := range rows {
			taskCounts[r.WeekPlanID] = r.Cnt
		}
	}
	weekViews := make([]map[string]interface{}, 0, len(weeks))
	for i := range weeks {
		weekViews = append(weekViews, s.buildWeekView(p, &weeks[i], userID, isAdmin, names, taskCounts[weeks[i].ID]))
	}

	// 成员（含姓名）
	var members []models.ProjectMember
	if err := db.Where("project_id = ? AND deleted = ?", p.ID, 0).Order("id ASC").Find(&members).Error; err != nil {
		return nil, err
	}
	memberIDs := make([]int64, 0, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.UserID)
	}
	memberNames := s.userNames(memberIDs)
	memberViews := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		memberViews = append(memberViews, map[string]interface{}{
			"id":         m.ID,
			"projectId":  m.ProjectID,
			"userId":     m.UserID,
			"userName":   memberNames[m.UserID],
			"memberRole": m.MemberRole,
			"createTime": fmtDateTimeStr(m.CreateTime),
			"deleted":    m.Deleted,
		})
	}

	// 变更历史（倒序）
	var logs []models.ProjectChangeLog
	if err := db.Where("project_id = ?", p.ID).Order("create_time DESC, id DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	logIDs := make([]int64, 0, len(logs))
	for _, l := range logs {
		logIDs = append(logIDs, l.OperatorID)
	}
	logNames := s.userNames(logIDs)
	logViews := make([]map[string]interface{}, 0, len(logs))
	for _, l := range logs {
		operatorName := logNames[l.OperatorID]
		if l.OperatorID <= 0 {
			operatorName = "系统"
		}
		logViews = append(logViews, map[string]interface{}{
			"id":           l.ID,
			"projectId":    l.ProjectID,
			"action":       l.Action,
			"detail":       l.Detail,
			"operatorId":   l.OperatorID,
			"operatorName": operatorName,
			"createTime":   fmtDateTimeStr(l.CreateTime),
		})
	}

	// 项目阶段（按开始时间正序）
	var stages []models.ProjectStage
	if err := db.Where("project_id = ? AND deleted = ?", p.ID, 0).Order("start_date ASC, id ASC").Find(&stages).Error; err != nil {
		return nil, err
	}
	stageViews := make([]map[string]interface{}, 0, len(stages))
	for i := range stages {
		stageViews = append(stageViews, buildStageView(&stages[i]))
	}

	view := s.projectView(p, userID, isAdmin, names[p.ManagerID], int64(len(members)))
	view["weeks"] = weekViews
	view["members"] = memberViews
	view["logs"] = logViews
	view["stages"] = stageViews
	view["currentStageId"] = nil
	if !isTerminalStatus(p.Status) {
		if hit := resolveCurrentStageItem(stages, time.Now().Format("2006-01-02")); hit != nil {
			view["currentStageId"] = hit.ID
		}
	}
	return view, nil
}

// CreateProject 创建项目（PM 自动加入成员表，事务）
func (s *ProjectService) CreateProject(userID int64, isAdmin bool, req *dto.ProjectCreateRequest) (map[string]interface{}, error) {
	db := s.getDB()
	if strings.TrimSpace(req.ProjectName) == "" || strings.TrimSpace(req.ProjectKey) == "" ||
		strings.TrimSpace(req.StartDate) == "" || strings.TrimSpace(req.EndDate) == "" || req.ManagerID <= 0 {
		return nil, bizErr(400, "项目名称、项目Key、起止日期、项目经理为必填项")
	}
	startDate, err := dto.ParseFlexibleDateTime(req.StartDate)
	if err != nil {
		return nil, bizErr(400, "开始日期格式错误")
	}
	endDate, err := dto.ParseFlexibleDateTime(req.EndDate)
	if err != nil {
		return nil, bizErr(400, "结束日期格式错误")
	}
	if endDate.Before(startDate) {
		return nil, bizErr(400, "结束日期不能早于开始日期")
	}

	var dupCount int64
	db.Model(&models.Project{}).Where("project_key = ? AND deleted = ?", req.ProjectKey, 0).Count(&dupCount)
	if dupCount > 0 {
		return nil, bizErr(400, fmt.Sprintf("项目Key「%s」已存在", req.ProjectKey))
	}

	p := &models.Project{
		ProjectName:    req.ProjectName,
		ProjectKey:     req.ProjectKey,
		Description:    req.Description,
		PlanDesc:       req.PlanDesc,
		ManagerID:      req.ManagerID,
		StartDate:      startDate,
		EndDate:        endDate,
		Status:         "STARTING",
		RiskLevel:      "MEDIUM",
		DeliveryStatus: "NORMAL",
		ProgressPct:    0,
		CreateUser:     &userID,
		UpdateUser:     &userID,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		member := &models.ProjectMember{
			ProjectID:  p.ID,
			UserID:     req.ManagerID,
			MemberRole: "PM",
			CreateUser: &userID,
			UpdateUser: &userID,
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "CREATE", fmt.Sprintf("创建项目「%s」", p.ProjectName), userID)
	})
	if err != nil {
		return nil, err
	}

	created, err := s.getProject(p.ID)
	if err != nil {
		return nil, err
	}
	return s.singleProjectView(created, userID, isAdmin), nil
}

// UpdateProject 编辑项目
func (s *ProjectService) UpdateProject(id, userID int64, isAdmin bool, req *dto.ProjectUpdateRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(id)
	if err != nil {
		return nil, err
	}
	if !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, bizErr(403, "无权限编辑该项目")
	}
	if isTerminalStatus(p.Status) {
		return nil, bizErr(400, "项目已结项/终止，不可编辑")
	}

	updates := map[string]interface{}{}
	if req.ProjectName != nil && strings.TrimSpace(*req.ProjectName) != "" {
		updates["project_name"] = *req.ProjectName
	}
	if req.ProjectKey != nil && strings.TrimSpace(*req.ProjectKey) != "" && *req.ProjectKey != p.ProjectKey {
		var dupCount int64
		db.Model(&models.Project{}).Where("project_key = ? AND deleted = ? AND id <> ?", *req.ProjectKey, 0, p.ID).Count(&dupCount)
		if dupCount > 0 {
			return nil, bizErr(400, fmt.Sprintf("项目Key「%s」已存在", *req.ProjectKey))
		}
		updates["project_key"] = *req.ProjectKey
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.PlanDesc != nil {
		updates["plan_desc"] = *req.PlanDesc
	}
	if req.ManagerID != nil && *req.ManagerID > 0 {
		updates["manager_id"] = *req.ManagerID
	}
	startDate := p.StartDate
	endDate := p.EndDate
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		if t, err := dto.ParseFlexibleDateTime(*req.StartDate); err == nil {
			startDate = t
			updates["start_date"] = t
		} else {
			return nil, bizErr(400, "开始日期格式错误")
		}
	}
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		if t, err := dto.ParseFlexibleDateTime(*req.EndDate); err == nil {
			endDate = t
			updates["end_date"] = t
		} else {
			return nil, bizErr(400, "结束日期格式错误")
		}
	}
	if endDate.Before(startDate) {
		return nil, bizErr(400, "结束日期不能早于开始日期")
	}
	if req.RiskLevel != nil && *req.RiskLevel != "" {
		updates["risk_level"] = *req.RiskLevel
	}
	if req.DeliveryStatus != nil && *req.DeliveryStatus != "" {
		updates["delivery_status"] = *req.DeliveryStatus
	}
	if req.RiskNote != nil {
		updates["risk_note"] = *req.RiskNote
	}

	if len(updates) > 0 {
		updates["update_user"] = userID
		if err := db.Model(&models.Project{}).Where("id = ?", p.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
		if err := addProjectLog(db, p.ID, "UPDATE", "编辑项目信息", userID); err != nil {
			return nil, err
		}
	}

	updated, err := s.getProject(p.ID)
	if err != nil {
		return nil, err
	}
	return s.singleProjectView(updated, userID, isAdmin), nil
}

// DeleteProject 删除项目（仅 STARTING，事务级联软删周计划/成员，物理删任务关联）
func (s *ProjectService) DeleteProject(id, userID int64, isAdmin bool) error {
	db := s.getDB()
	p, err := s.getProject(id)
	if err != nil {
		return err
	}
	if !canActAsPMOrAdmin(p, userID, isAdmin) {
		return bizErr(403, "无权限删除该项目")
	}
	if p.Status != "STARTING" {
		return bizErr(400, "仅「启动」状态的项目允许删除")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var weeks []models.ProjectWeeklyPlan
		if err := tx.Where("project_id = ? AND deleted = ?", p.ID, 0).Find(&weeks).Error; err != nil {
			return err
		}
		weekIDs := make([]int64, 0, len(weeks))
		for _, w := range weeks {
			weekIDs = append(weekIDs, w.ID)
		}
		if len(weekIDs) > 0 {
			if err := tx.Where("week_plan_id IN ?", weekIDs).Delete(&models.TaskWeekPlanRel{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.ProjectWeeklyPlan{}).Where("id IN ?", weekIDs).
				Updates(map[string]interface{}{"deleted": 1, "update_user": userID}).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.ProjectMember{}).Where("project_id = ? AND deleted = ?", p.ID, 0).
			Updates(map[string]interface{}{"deleted": 1, "update_user": userID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Project{}).Where("id = ?", p.ID).
			Updates(map[string]interface{}{"deleted": 1, "update_user": userID}).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "DELETE", fmt.Sprintf("删除项目「%s」", p.ProjectName), userID)
	})
}

// ChangeProjectStatus 项目状态流转（开始/暂停/恢复/结项/终止）
func (s *ProjectService) ChangeProjectStatus(id, userID int64, isAdmin bool, req *dto.ProjectStatusRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(id)
	if err != nil {
		return nil, err
	}
	if !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, bizErr(403, "无权限执行该操作")
	}
	trans, ok := projectTransitions[req.Action]
	if !ok {
		return nil, bizErr(400, "不支持的状态操作")
	}
	if isTerminalStatus(p.Status) {
		return nil, bizErr(400, "项目已进入终态，不可再流转")
	}
	if !containsStr(trans.From, p.Status) {
		return nil, bizErr(400, fmt.Sprintf("当前状态不允许执行「%s」操作", projectStatusActions[req.Action]))
	}
	if req.Action == "terminate" && strings.TrimSpace(req.Reason) == "" {
		return nil, bizErr(400, "终止项目必须填写原因")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      trans.To,
		"update_user": userID,
	}
	switch req.Action {
	case "pause":
		updates["paused_at"] = &now
	case "resume":
		updates["paused_at"] = nil
	case "close":
		updates["closed_at"] = &now
	case "terminate":
		updates["risk_note"] = req.Reason
		updates["closed_at"] = &now
	}
	if err := db.Model(&models.Project{}).Where("id = ?", p.ID).Updates(updates).Error; err != nil {
		return nil, err
	}

	detail := fmt.Sprintf("项目状态：%s → %s（%s", projectStatusLabels[p.Status], projectStatusLabels[trans.To], projectStatusActions[req.Action])
	if strings.TrimSpace(req.Reason) != "" {
		detail += "：" + req.Reason
	}
	detail += "）"
	if err := addProjectLog(db, p.ID, "STATUS", detail, userID); err != nil {
		return nil, err
	}

	updated, err := s.getProject(p.ID)
	if err != nil {
		return nil, err
	}
	return s.singleProjectView(updated, userID, isAdmin), nil
}

// ChangeProjectRisk 修改风险等级/交付状态（必填原因，写变更历史）
func (s *ProjectService) ChangeProjectRisk(id, userID int64, isAdmin bool, req *dto.ProjectRiskRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(id)
	if err != nil {
		return nil, err
	}
	if !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, bizErr(403, "无权限修改项目风险")
	}
	if isTerminalStatus(p.Status) {
		return nil, bizErr(400, "项目已进入终态，不可修改风险信息")
	}
	if strings.TrimSpace(req.Reason) == "" {
		return nil, bizErr(400, "修改风险等级/交付状态必须填写原因")
	}

	changes := []string{}
	updates := map[string]interface{}{}
	if req.RiskLevel != "" && req.RiskLevel != p.RiskLevel {
		changes = append(changes, fmt.Sprintf("风险等级：%s → %s", riskLevelLabels[p.RiskLevel], riskLevelLabels[req.RiskLevel]))
		updates["risk_level"] = req.RiskLevel
	}
	if req.DeliveryStatus != "" && req.DeliveryStatus != p.DeliveryStatus {
		changes = append(changes, fmt.Sprintf("交付状态：%s → %s", deliveryStatusLabels[p.DeliveryStatus], deliveryStatusLabels[req.DeliveryStatus]))
		updates["delivery_status"] = req.DeliveryStatus
	}
	if req.RiskNote != nil {
		updates["risk_note"] = *req.RiskNote
	}
	if len(changes) == 0 && req.RiskNote == nil {
		return nil, bizErr(400, "未检测到任何变更")
	}

	if len(updates) > 0 {
		updates["update_user"] = userID
		if err := db.Model(&models.Project{}).Where("id = ?", p.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	detail := strings.Join(changes, "；")
	if detail == "" {
		detail = "更新风险说明"
	}
	detail += fmt.Sprintf("（%s）", req.Reason)
	if err := addProjectLog(db, p.ID, "RISK_CHANGE", detail, userID); err != nil {
		return nil, err
	}

	updated, err := s.getProject(p.ID)
	if err != nil {
		return nil, err
	}
	return s.singleProjectView(updated, userID, isAdmin), nil
}

// RecalcProjectProgress 重算项目进度（仅服务端调用；SKIPPED 不计入分子分母；0 周时进度 0）
func (s *ProjectService) RecalcProjectProgress(projectID int64) error {
	db := s.getDB()
	var stats struct {
		Total int64
		Done  int64
	}
	err := db.Model(&models.ProjectWeeklyPlan{}).
		Select("COALESCE(SUM(CASE WHEN status <> ? THEN 1 ELSE 0 END), 0) AS total, COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS done", "SKIPPED", "DONE").
		Where("project_id = ? AND deleted = ?", projectID, 0).
		Scan(&stats).Error
	if err != nil {
		return err
	}
	pct := 0
	if stats.Total > 0 {
		pct = int(math.Round(float64(stats.Done) / float64(stats.Total) * 100))
	}
	return db.Model(&models.Project{}).Where("id = ?", projectID).Updates(map[string]interface{}{
		"progress_pct":    pct,
		"week_count":      stats.Total,
		"done_week_count": stats.Done,
	}).Error
}

/* ------------------------------ 成员管理 ------------------------------ */

func memberView(m models.ProjectMember, userName string) map[string]interface{} {
	return map[string]interface{}{
		"id":         m.ID,
		"projectId":  m.ProjectID,
		"userId":     m.UserID,
		"userName":   userName,
		"memberRole": m.MemberRole,
		"createTime": fmtDateTimeStr(m.CreateTime),
		"deleted":    m.Deleted,
	}
}

// ListMembers 项目成员列表
func (s *ProjectService) ListMembers(projectID int64) ([]map[string]interface{}, error) {
	db := s.getDB()
	if _, err := s.getProject(projectID); err != nil {
		return nil, err
	}
	var members []models.ProjectMember
	if err := db.Where("project_id = ? AND deleted = ?", projectID, 0).Order("id ASC").Find(&members).Error; err != nil {
		return nil, err
	}
	userIDs := make([]int64, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	names := s.userNames(userIDs)
	list := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		list = append(list, memberView(m, names[m.UserID]))
	}
	return list, nil
}

// AddMember 添加项目成员（幂等：重复添加返回已有记录 already=true）
func (s *ProjectService) AddMember(projectID, userID int64, isAdmin bool, req *dto.ProjectMemberRequest) (map[string]interface{}, bool, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, false, err
	}
	if !canActAsManagerOnly(p, userID, isAdmin) {
		return nil, false, bizErr(403, "仅项目经理可管理成员")
	}
	if isTerminalStatus(p.Status) {
		return nil, false, bizErr(400, "项目已进入终态，不可变更成员")
	}
	if req.UserID <= 0 {
		return nil, false, bizErr(400, "请选择要添加的成员")
	}

	var existed models.ProjectMember
	err = db.Where("project_id = ? AND user_id = ? AND deleted = ?", p.ID, req.UserID, 0).First(&existed).Error
	if err == nil {
		return memberView(existed, s.userNames([]int64{existed.UserID})[existed.UserID]), true, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	role := req.MemberRole
	if role == "" {
		role = "MEMBER"
	}
	m := &models.ProjectMember{
		ProjectID:  p.ID,
		UserID:     req.UserID,
		MemberRole: role,
		CreateUser: &userID,
		UpdateUser: &userID,
	}
	memberName := s.userNames([]int64{m.UserID})[m.UserID]
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "MEMBER_ADD", fmt.Sprintf("添加成员：%s", memberName), userID)
	}); err != nil {
		return nil, false, err
	}
	return memberView(*m, memberName), false, nil
}

// RemoveMember 移除项目成员（项目经理不可移除）
func (s *ProjectService) RemoveMember(projectID, memberID, userID int64, isAdmin bool) error {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return err
	}
	if !canActAsManagerOnly(p, userID, isAdmin) {
		return bizErr(403, "仅项目经理可管理成员")
	}
	if isTerminalStatus(p.Status) {
		return bizErr(400, "项目已进入终态，不可变更成员")
	}

	var m models.ProjectMember
	if err := db.Where("id = ? AND project_id = ? AND deleted = ?", memberID, p.ID, 0).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizErr(404, "成员不存在")
		}
		return err
	}
	if m.UserID == p.ManagerID {
		return bizErr(400, "项目经理不可移除")
	}
	memberName := s.userNames([]int64{m.UserID})[m.UserID]
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ProjectMember{}).Where("id = ?", m.ID).
			Updates(map[string]interface{}{"deleted": 1, "update_user": userID}).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "MEMBER_REMOVE", fmt.Sprintf("移除成员：%s", memberName), userID)
	})
}

/* ------------------------------ 项目阶段 ------------------------------ */

// stageNameMaxLen 阶段名称长度上限（对齐 DDL varchar(100)）
const stageNameMaxLen = 100

// buildStageView 阶段视图
func buildStageView(st *models.ProjectStage) map[string]interface{} {
	return map[string]interface{}{
		"id":          st.ID,
		"projectId":   st.ProjectID,
		"stageName":   st.StageName,
		"description": st.Description,
		"startDate":   fmtDateStr(st.StartDate),
		"endDate":     fmtDateStr(st.EndDate),
		"createTime":  fmtDateTimeStr(st.CreateTime),
		"updateTime":  fmtDateTimeStr(st.UpdateTime),
	}
}

// resolveCurrentStageItem 计算当前阶段（纯函数）：区间 [start_date, end_date] 含两端；
// 多个命中取 start_date 最晚（并列取 id 最大）；无命中返回 nil
func resolveCurrentStageItem(stages []models.ProjectStage, todayStr string) *models.ProjectStage {
	var hit *models.ProjectStage
	for i := range stages {
		st := &stages[i]
		startStr := fmtDateStr(st.StartDate)
		if startStr > todayStr || todayStr > fmtDateStr(st.EndDate) {
			continue
		}
		if hit == nil {
			hit = st
			continue
		}
		hitStart := fmtDateStr(hit.StartDate)
		if startStr > hitStart || (startStr == hitStart && st.ID > hit.ID) {
			hit = st
		}
	}
	return hit
}

// getStage 查询项目内阶段（deleted=0）
func (s *ProjectService) getStage(projectID, stageID int64) (*models.ProjectStage, error) {
	var st models.ProjectStage
	if err := s.getDB().Where("id = ? AND project_id = ? AND deleted = ?", stageID, projectID, 0).First(&st).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizErr(404, "阶段不存在")
		}
		return nil, err
	}
	return &st, nil
}

// ensureStageNameUnique 同一项目未删除范围内阶段名称唯一（excludeID>0 时排除自身）
func (s *ProjectService) ensureStageNameUnique(db *gorm.DB, projectID int64, stageName string, excludeID int64) error {
	query := db.Model(&models.ProjectStage{}).Where("project_id = ? AND stage_name = ? AND deleted = ?", projectID, stageName, 0)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizErr(400, "阶段名称已存在")
	}
	return nil
}

// canMaintainStage 阶段维护权限校验（对齐成员管理：ADMIN/项目经理，终态不可操作）
func canMaintainStage(p *models.Project, userID int64, isAdmin bool) error {
	if !canActAsManagerOnly(p, userID, isAdmin) {
		return bizErr(403, "仅项目经理可管理阶段")
	}
	if isTerminalStatus(p.Status) {
		return bizErr(400, "项目已进入终态，不可维护阶段")
	}
	return nil
}

// ListStages 项目阶段列表（按开始时间正序，团队共享视图对齐详情权限）
func (s *ProjectService) ListStages(projectID int64) ([]map[string]interface{}, error) {
	db := s.getDB()
	if _, err := s.getProject(projectID); err != nil {
		return nil, err
	}
	var stages []models.ProjectStage
	if err := db.Where("project_id = ? AND deleted = ?", projectID, 0).Order("start_date ASC, id ASC").Find(&stages).Error; err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(stages))
	for i := range stages {
		list = append(list, buildStageView(&stages[i]))
	}
	return list, nil
}

// CreateStage 新增项目阶段（校验顺序对齐 AddMember/CreateWeek）
func (s *ProjectService) CreateStage(projectID, userID int64, isAdmin bool, req *dto.StageCreateRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	if err := canMaintainStage(p, userID, isAdmin); err != nil {
		return nil, err
	}
	stageName := strings.TrimSpace(req.StageName)
	if stageName == "" {
		return nil, bizErr(400, "阶段名称为必填项")
	}
	if len([]rune(stageName)) > stageNameMaxLen {
		return nil, bizErr(400, "阶段名称长度不能超过100个字符")
	}
	startDate, err := dto.ParseFlexibleDateTime(req.StartDate)
	if err != nil {
		return nil, bizErr(400, "开始日期格式错误")
	}
	endDate, err := dto.ParseFlexibleDateTime(req.EndDate)
	if err != nil {
		return nil, bizErr(400, "结束日期格式错误")
	}
	if endDate.Before(startDate) {
		return nil, bizErr(400, "结束日期不能早于开始日期")
	}
	if err := s.ensureStageNameUnique(db, p.ID, stageName, 0); err != nil {
		return nil, err
	}

	st := &models.ProjectStage{
		ProjectID:   p.ID,
		StageName:   stageName,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		CreateUser:  &userID,
		UpdateUser:  &userID,
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(st).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "STAGE_ADD", fmt.Sprintf("新增阶段「%s」（%s ~ %s）", st.StageName, fmtDateStr(st.StartDate), fmtDateStr(st.EndDate)), userID)
	}); err != nil {
		return nil, err
	}
	return buildStageView(st), nil
}

// UpdateStage 编辑项目阶段（字段均可选；改名查重排除自身；结束不得早于开始）
func (s *ProjectService) UpdateStage(projectID, stageID, userID int64, isAdmin bool, req *dto.StageUpdateRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	if err := canMaintainStage(p, userID, isAdmin); err != nil {
		return nil, err
	}
	st, err := s.getStage(p.ID, stageID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	stageName := st.StageName
	if req.StageName != nil && strings.TrimSpace(*req.StageName) != "" {
		stageName = strings.TrimSpace(*req.StageName)
		if len([]rune(stageName)) > stageNameMaxLen {
			return nil, bizErr(400, "阶段名称长度不能超过100个字符")
		}
		updates["stage_name"] = stageName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	startDate := st.StartDate
	endDate := st.EndDate
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		t, err := dto.ParseFlexibleDateTime(*req.StartDate)
		if err != nil {
			return nil, bizErr(400, "开始日期格式错误")
		}
		startDate = t
		updates["start_date"] = t
	}
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		t, err := dto.ParseFlexibleDateTime(*req.EndDate)
		if err != nil {
			return nil, bizErr(400, "结束日期格式错误")
		}
		endDate = t
		updates["end_date"] = t
	}
	if endDate.Before(startDate) {
		return nil, bizErr(400, "结束日期不能早于开始日期")
	}
	if stageName != st.StageName {
		if err := s.ensureStageNameUnique(db, p.ID, stageName, st.ID); err != nil {
			return nil, err
		}
	}
	if len(updates) == 0 {
		return buildStageView(st), nil
	}
	updates["update_user"] = userID
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ProjectStage{}).Where("id = ?", st.ID).Updates(updates).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "STAGE_EDIT", fmt.Sprintf("编辑阶段「%s」（%s ~ %s）", stageName, fmtDateStr(startDate), fmtDateStr(endDate)), userID)
	}); err != nil {
		return nil, err
	}
	updated, err := s.getStage(p.ID, st.ID)
	if err != nil {
		return nil, err
	}
	return buildStageView(updated), nil
}

// DeleteStage 删除项目阶段（软删除）
func (s *ProjectService) DeleteStage(projectID, stageID, userID int64, isAdmin bool) error {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return err
	}
	if err := canMaintainStage(p, userID, isAdmin); err != nil {
		return err
	}
	st, err := s.getStage(p.ID, stageID)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ProjectStage{}).Where("id = ?", st.ID).
			Updates(map[string]interface{}{"deleted": 1, "update_user": userID}).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "STAGE_DELETE", fmt.Sprintf("删除阶段「%s」（%s ~ %s）", st.StageName, fmtDateStr(st.StartDate), fmtDateStr(st.EndDate)), userID)
	})
}

/* ------------------------------ 周计划 ------------------------------ */

// weekStatusActions 周状态操作中文名（错误文案用）
var weekStatusActions = map[string]string{
	"start": "开始", "done": "完成", "skip": "跳过", "reopen": "重新打开",
}

// CreateWeek 新增周计划（week_no 项目内自增；创建后重算项目进度）
func (s *ProjectService) CreateWeek(projectID, userID int64, isAdmin bool, req *dto.WeekCreateRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	if !canActAsManagerOnly(p, userID, isAdmin) {
		return nil, bizErr(403, "仅项目经理可新增周计划")
	}
	if isTerminalStatus(p.Status) {
		return nil, bizErr(400, "项目已进入终态，不可新增周计划")
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.StartDate) == "" ||
		strings.TrimSpace(req.EndDate) == "" || req.OwnerID <= 0 {
		return nil, bizErr(400, "周主题、起止日期、负责人为必填项")
	}
	startDate, err := dto.ParseFlexibleDateTime(req.StartDate)
	if err != nil {
		return nil, bizErr(400, "开始日期格式错误")
	}
	endDate, err := dto.ParseFlexibleDateTime(req.EndDate)
	if err != nil {
		return nil, bizErr(400, "结束日期格式错误")
	}
	if endDate.Before(startDate) {
		return nil, bizErr(400, "结束日期不能早于开始日期")
	}

	weekNo := 1
	var lastWeek models.ProjectWeeklyPlan
	if err := db.Where("project_id = ? AND deleted = ?", p.ID, 0).Order("week_no DESC").First(&lastWeek).Error; err == nil {
		weekNo = lastWeek.WeekNo + 1
	}

	w := &models.ProjectWeeklyPlan{
		ProjectID:  p.ID,
		WeekNo:     weekNo,
		Title:      req.Title,
		PlanDesc:   req.PlanDesc,
		StartDate:  startDate,
		EndDate:    endDate,
		OwnerID:    req.OwnerID,
		Status:     "NOT_STARTED",
		Remark:     req.Remark,
		SortOrder:  weekNo,
		CreateUser: &userID,
		UpdateUser: &userID,
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(w).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "WEEK_ADD", fmt.Sprintf("新增第%d周计划「%s」", weekNo, w.Title), userID)
	}); err != nil {
		return nil, err
	}
	if err := s.RecalcProjectProgress(p.ID); err != nil {
		return nil, err
	}
	return s.singleWeekView(p, w, userID, isAdmin), nil
}

// UpdateWeek 编辑周计划（PM/创建人/该周负责人；已提交总结锁定）
func (s *ProjectService) UpdateWeek(projectID, weekID, userID int64, isAdmin bool, req *dto.WeekUpdateRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return nil, err
	}
	if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, bizErr(403, "无权限编辑该周计划")
	}
	if w.SummaryAt != nil {
		return nil, bizErr(400, "该周总结已提交锁定，需先解锁才能编辑")
	}

	updates := map[string]interface{}{}
	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		updates["title"] = *req.Title
	}
	if req.PlanDesc != nil {
		updates["plan_desc"] = *req.PlanDesc
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	if req.OwnerID != nil && *req.OwnerID > 0 {
		updates["owner_id"] = *req.OwnerID
	}
	startDate := w.StartDate
	endDate := w.EndDate
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		t, err := dto.ParseFlexibleDateTime(*req.StartDate)
		if err != nil {
			return nil, bizErr(400, "开始日期格式错误")
		}
		startDate = t
		updates["start_date"] = t
	}
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		t, err := dto.ParseFlexibleDateTime(*req.EndDate)
		if err != nil {
			return nil, bizErr(400, "结束日期格式错误")
		}
		endDate = t
		updates["end_date"] = t
	}
	if endDate.Before(startDate) {
		return nil, bizErr(400, "结束日期不能早于开始日期")
	}

	if len(updates) > 0 {
		updates["update_user"] = userID
		if err := db.Model(&models.ProjectWeeklyPlan{}).Where("id = ?", w.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	refreshed, err := s.getWeek(p.ID, w.ID)
	if err != nil {
		return nil, err
	}
	if err := addProjectLog(db, p.ID, "WEEK_EDIT", fmt.Sprintf("编辑第%d周计划「%s」", refreshed.WeekNo, refreshed.Title), userID); err != nil {
		return nil, err
	}
	return s.singleWeekView(p, refreshed, userID, isAdmin), nil
}

// DeleteWeek 删除周计划（事务：清理该周任务关联 + 软删；删除后重算进度）
func (s *ProjectService) DeleteWeek(projectID, weekID, userID int64, isAdmin bool) error {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return err
	}
	if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
		return bizErr(403, "无权限删除该周计划")
	}
	if w.SummaryAt != nil {
		return bizErr(400, "该周总结已提交锁定，需先解锁才能删除")
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("week_plan_id = ?", w.ID).Delete(&models.TaskWeekPlanRel{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.ProjectWeeklyPlan{}).Where("id = ?", w.ID).
			Updates(map[string]interface{}{"deleted": 1, "update_user": userID}).Error; err != nil {
			return err
		}
		return addProjectLog(tx, p.ID, "WEEK_DELETE", fmt.Sprintf("删除第%d周计划「%s」", w.WeekNo, w.Title), userID)
	}); err != nil {
		return err
	}
	return s.RecalcProjectProgress(p.ID)
}

// ChangeWeekStatus 周状态流转（action 或直接 status；变更后重算项目进度）
func (s *ProjectService) ChangeWeekStatus(projectID, weekID, userID int64, isAdmin bool, req *dto.WeekStatusRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return nil, err
	}
	if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, bizErr(403, "无权限变更该周状态")
	}
	if w.SummaryAt != nil && req.Status == "NOT_STARTED" {
		return nil, bizErr(400, "该周总结已提交锁定")
	}

	var target string
	if req.Action != "" {
		trans, ok := weekTransitions[req.Action]
		if !ok {
			return nil, bizErr(400, "不支持的状态操作")
		}
		if !containsStr(trans.From, w.Status) {
			return nil, bizErr(400, fmt.Sprintf("当前状态不允许执行「%s」操作", weekStatusActions[req.Action]))
		}
		target = trans.To
	} else if req.Status != "" {
		if !containsStr(validWeekStatuses, req.Status) {
			return nil, bizErr(400, "非法的周状态")
		}
		target = req.Status
	} else {
		return nil, bizErr(400, "缺少目标状态")
	}

	if err := db.Model(&models.ProjectWeeklyPlan{}).Where("id = ?", w.ID).
		Updates(map[string]interface{}{"status": target, "update_user": userID}).Error; err != nil {
		return nil, err
	}
	if err := s.RecalcProjectProgress(p.ID); err != nil {
		return nil, err
	}
	if err := addProjectLog(db, p.ID, "WEEK_STATUS", fmt.Sprintf("第%d周状态变更为 %s", w.WeekNo, weekStatusLabels[target]), userID); err != nil {
		return nil, err
	}
	refreshed, err := s.getWeek(p.ID, w.ID)
	if err != nil {
		return nil, err
	}
	return s.singleWeekView(p, refreshed, userID, isAdmin), nil
}

// SubmitWeekSummary 提交周总结（提交即锁定：summary_by/summary_at）
func (s *ProjectService) SubmitWeekSummary(projectID, weekID, userID int64, isAdmin bool, req *dto.WeekSummaryRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return nil, err
	}
	if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, bizErr(403, "仅周负责人可提交总结")
	}
	if w.SummaryAt != nil {
		return nil, bizErr(400, "该周总结已提交，请勿重复提交")
	}

	content := dto.WeekSummaryContent{}
	if req.Summary != nil {
		content = *req.Summary
	} else {
		content = dto.WeekSummaryContent{
			Finished: req.Finished,
			NextPlan: req.NextPlan,
			Risks:    req.Risks,
			Support:  req.Support,
		}
	}
	if strings.TrimSpace(content.Finished) == "" {
		return nil, bizErr(400, "「本周完成」为必填项")
	}
	if strings.TrimSpace(content.NextPlan) == "" {
		return nil, bizErr(400, "「下周计划」为必填项")
	}

	raw, err := json.Marshal(map[string]string{
		"finished": content.Finished,
		"nextPlan": content.NextPlan,
		"risks":    content.Risks,
		"support":  content.Support,
	})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	summaryStr := string(raw)
	if err := db.Model(&models.ProjectWeeklyPlan{}).Where("id = ?", w.ID).Updates(map[string]interface{}{
		"summary":     summaryStr,
		"summary_by":  userID,
		"summary_at":  &now,
		"update_user": userID,
	}).Error; err != nil {
		return nil, err
	}
	if err := addProjectLog(db, p.ID, "SUMMARY_SUBMIT", fmt.Sprintf("提交第%d周总结", w.WeekNo), userID); err != nil {
		return nil, err
	}
	refreshed, err := s.getWeek(p.ID, w.ID)
	if err != nil {
		return nil, err
	}
	return s.singleWeekView(p, refreshed, userID, isAdmin), nil
}

// UnlockWeekSummary 解锁周总结（PM/ADMIN；保留总结内容，仅清空锁定标记）
func (s *ProjectService) UnlockWeekSummary(projectID, weekID, userID int64, isAdmin bool) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return nil, err
	}
	if !canActAsManagerOnly(p, userID, isAdmin) {
		return nil, bizErr(403, "仅项目经理/管理员可解锁总结")
	}
	if w.SummaryAt == nil {
		return nil, bizErr(400, "该周总结未锁定，无需解锁")
	}

	if err := db.Model(&models.ProjectWeeklyPlan{}).Where("id = ?", w.ID).
		Updates(map[string]interface{}{"summary_at": nil, "summary_by": nil, "update_user": userID}).Error; err != nil {
		return nil, err
	}
	if err := addProjectLog(db, p.ID, "SUMMARY_UNLOCK", fmt.Sprintf("解锁第%d周总结", w.WeekNo), userID); err != nil {
		return nil, err
	}
	refreshed, err := s.getWeek(p.ID, w.ID)
	if err != nil {
		return nil, err
	}
	return s.singleWeekView(p, refreshed, userID, isAdmin), nil
}

/* ------------------------------ 任务-周关联 ------------------------------ */

// projectTaskEntry 任务与其周关联集合（列表聚合用）
type projectTaskEntry struct {
	task  models.TaskItem
	links []models.TaskWeekPlanRel
}

// taskWeekView 构建任务视图（含所属周 weekNos/weeks，与前端契约一致）
func taskWeekView(t *models.TaskItem, links []models.TaskWeekPlanRel, weekByID map[int64]*models.ProjectWeeklyPlan, names map[int64]string) map[string]interface{} {
	weeks := make([]map[string]interface{}, 0, len(links))
	for _, l := range links {
		w, ok := weekByID[l.WeekPlanID]
		if !ok {
			continue
		}
		weeks = append(weeks, map[string]interface{}{
			"weekPlanId": w.ID,
			"weekNo":     w.WeekNo,
			"title":      w.Title,
		})
	}
	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i]["weekNo"].(int) < weeks[j]["weekNo"].(int)
	})
	weekNos := make([]int, 0, len(weeks))
	for _, w := range weeks {
		weekNos = append(weekNos, w["weekNo"].(int))
	}
	var deadline interface{}
	if t.Deadline != nil {
		deadline = fmtDateTimeStr(*t.Deadline)
	}
	return map[string]interface{}{
		"id":          t.ID,
		"userId":      t.UserID,
		"userName":    names[t.UserID],
		"title":       t.Title,
		"itemType":    t.ItemType,
		"priority":    t.Priority,
		"statusCode":  t.StatusCode,
		"progress":    t.Progress,
		"deadline":    deadline,
		"description": t.Description,
		"weekNos":     weekNos,
		"weeks":       weeks,
	}
}

// ListProjectTasks 项目任务列表（按任务去重；可按 weekPlanId/userId/statusCode 过滤；未完成优先+截止时间升序）
func (s *ProjectService) ListProjectTasks(projectID int64, page, size int, filter map[string]interface{}) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}

	var weeks []models.ProjectWeeklyPlan
	if err := db.Where("project_id = ? AND deleted = ?", p.ID, 0).Find(&weeks).Error; err != nil {
		return nil, err
	}
	weekByID := make(map[int64]*models.ProjectWeeklyPlan, len(weeks))
	weekIDs := make([]int64, 0, len(weeks))
	for i := range weeks {
		weekByID[weeks[i].ID] = &weeks[i]
		weekIDs = append(weekIDs, weeks[i].ID)
	}

	empty := map[string]interface{}{"total": int64(0), "pageNum": page, "pageSize": size, "list": []map[string]interface{}{}}
	if len(weekIDs) == 0 {
		return empty, nil
	}

	weekPlanID := int64(0)
	if v, ok := filter["weekPlanId"].(string); ok && v != "" {
		weekPlanID, _ = strconv.ParseInt(v, 10, 64)
	}
	var links []models.TaskWeekPlanRel
	linkQuery := db.Where("week_plan_id IN ?", weekIDs)
	if weekPlanID > 0 {
		if _, ok := weekByID[weekPlanID]; !ok {
			return empty, nil
		}
		linkQuery = db.Where("week_plan_id = ?", weekPlanID)
	}
	if err := linkQuery.Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return empty, nil
	}

	// 按任务聚合（一个任务可关联多周，按任务去重）
	linksByTask := map[int64][]models.TaskWeekPlanRel{}
	taskIDs := make([]int64, 0, len(links))
	for _, l := range links {
		if _, exists := linksByTask[l.TaskID]; !exists {
			taskIDs = append(taskIDs, l.TaskID)
		}
		linksByTask[l.TaskID] = append(linksByTask[l.TaskID], l)
	}
	// 负责人/状态过滤（userId 非法或缺失时忽略该条件；未知状态码等值匹配自然返回空列表）
	taskQuery := db.Where("id IN ? AND deleted = ?", taskIDs, 0)
	if v, ok := filter["userId"].(string); ok {
		if uid, uerr := strconv.ParseInt(strings.TrimSpace(v), 10, 64); uerr == nil && uid > 0 {
			taskQuery = taskQuery.Where("user_id = ?", uid)
		}
	}
	if v, ok := filter["statusCode"].(string); ok && strings.TrimSpace(v) != "" {
		taskQuery = taskQuery.Where("status_code = ?", strings.TrimSpace(v))
	}
	var tasks []models.TaskItem
	if err := taskQuery.Find(&tasks).Error; err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		userIDs = append(userIDs, t.UserID)
	}
	names := s.userNames(userIDs)

	entries := make([]projectTaskEntry, 0, len(tasks))
	for i := range tasks {
		if ls, ok := linksByTask[tasks[i].ID]; ok {
			entries = append(entries, projectTaskEntry{task: tasks[i], links: ls})
		}
	}

	// 排序：未完成优先；截止时间升序，空值最后
	sort.Slice(entries, func(i, j int) bool {
		di, dj := 0, 0
		if entries[i].task.StatusCode == "DONE" {
			di = 1
		}
		if entries[j].task.StatusCode == "DONE" {
			dj = 1
		}
		if di != dj {
			return di < dj
		}
		dlI, dlJ := entries[i].task.Deadline, entries[j].task.Deadline
		if dlI == nil && dlJ == nil {
			return false
		}
		if dlI == nil {
			return false
		}
		if dlJ == nil {
			return true
		}
		return dlI.Before(*dlJ)
	})

	total := len(entries)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	list := make([]map[string]interface{}, 0, end-start)
	for i := start; i < end; i++ {
		list = append(list, taskWeekView(&entries[i].task, entries[i].links, weekByID, names))
	}
	return map[string]interface{}{"total": total, "pageNum": page, "pageSize": size, "list": list}, nil
}

// CreateTaskInWeek 在指定周新建任务（事务：创建事项 + 建周关联）
func (s *ProjectService) CreateTaskInWeek(projectID, weekID, userID int64, isAdmin bool, req *dto.WeekTaskCreateRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return nil, err
	}
	if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, bizErr(403, "仅周负责人或项目经理可在该周新建任务")
	}
	if w.SummaryAt != nil {
		return nil, bizErr(400, "该周总结已提交锁定，不可新增任务")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, bizErr(400, "任务标题为必填项")
	}

	assignee := w.OwnerID
	if req.UserID != nil && *req.UserID > 0 {
		assignee = *req.UserID
	}
	priority := req.Priority
	if priority == "" {
		priority = "MEDIUM"
	}
	task := &models.TaskItem{
		UserID:            assignee,
		Title:             req.Title,
		Description:       req.Description,
		ItemType:          "TASK",
		Category:          req.Category,
		Priority:          priority,
		StatusCode:        "TODO",
		Progress:          0,
		RecurrenceType:    "NONE",
		RecurrenceRule:    "{}", // MySQL JSON 列不接受空字符串
		RecurrenceEndType: "NEVER",
		RecurrenceCount:   10,
		CreateUser:        &userID,
		UpdateUser:        &userID,
	}
	if strings.TrimSpace(req.Deadline) != "" {
		t, err := dto.ParseFlexibleDateTime(req.Deadline)
		if err != nil {
			return nil, bizErr(400, "截止时间格式错误")
		}
		task.Deadline = &t
	}
	link := &models.TaskWeekPlanRel{WeekPlanID: w.ID, CreateUser: &userID}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		link.TaskID = task.ID
		return tx.Create(link).Error
	}); err != nil {
		return nil, err
	}
	names := s.userNames([]int64{task.UserID})
	return taskWeekView(task, []models.TaskWeekPlanRel{*link}, map[int64]*models.ProjectWeeklyPlan{w.ID: w}, names), nil
}

// LinkTaskToWeek 将已有事项加入该周（幂等：已关联时 already=true）
func (s *ProjectService) LinkTaskToWeek(projectID, weekID, taskID, userID int64, isAdmin bool) (map[string]interface{}, bool, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, false, err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return nil, false, err
	}
	if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
		return nil, false, bizErr(403, "仅周负责人或项目经理可关联任务")
	}
	if w.SummaryAt != nil {
		return nil, false, bizErr(400, "该周总结已提交锁定，不可关联任务")
	}

	var task models.TaskItem
	if err := db.Where("id = ? AND deleted = ?", taskID, 0).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, bizErr(404, "事项不存在")
		}
		return nil, false, err
	}
	names := s.userNames([]int64{task.UserID})
	weekByID := map[int64]*models.ProjectWeeklyPlan{w.ID: w}

	var existing []models.TaskWeekPlanRel
	if err := db.Where("task_id = ? AND week_plan_id = ?", task.ID, w.ID).Find(&existing).Error; err != nil {
		return nil, false, err
	}
	if len(existing) > 0 {
		return taskWeekView(&task, existing, weekByID, names), true, nil
	}
	link := &models.TaskWeekPlanRel{TaskID: task.ID, WeekPlanID: w.ID, CreateUser: &userID}
	if err := db.Create(link).Error; err != nil {
		return nil, false, err
	}
	return taskWeekView(&task, []models.TaskWeekPlanRel{*link}, weekByID, names), false, nil
}

// UnlinkTaskFromWeek 将事项从该周移出
func (s *ProjectService) UnlinkTaskFromWeek(projectID, weekID, taskID, userID int64, isAdmin bool) error {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return err
	}
	w, err := s.getWeek(p.ID, weekID)
	if err != nil {
		return err
	}
	if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
		return bizErr(403, "仅周负责人或项目经理可移出任务")
	}
	if w.SummaryAt != nil {
		return bizErr(400, "该周总结已提交锁定，不可移出任务")
	}
	result := db.Where("task_id = ? AND week_plan_id = ?", taskID, w.ID).Delete(&models.TaskWeekPlanRel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bizErr(404, "该任务不在此周计划中")
	}
	return nil
}

// MoveTaskToWeek 将任务移动到项目内其他周（事务：移出本项目全部关联 + 移入目标周）
func (s *ProjectService) MoveTaskToWeek(projectID, taskID, userID int64, isAdmin bool, req *dto.TaskMoveRequest) (map[string]interface{}, error) {
	db := s.getDB()
	p, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	var target models.ProjectWeeklyPlan
	if err := db.Where("id = ? AND project_id = ? AND deleted = ?", req.WeekPlanID, p.ID, 0).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizErr(404, "目标周计划不存在")
		}
		return nil, err
	}

	var weeks []models.ProjectWeeklyPlan
	if err := db.Where("project_id = ? AND deleted = ?", p.ID, 0).Find(&weeks).Error; err != nil {
		return nil, err
	}
	weekByID := make(map[int64]*models.ProjectWeeklyPlan, len(weeks))
	weekIDs := make([]int64, 0, len(weeks))
	for i := range weeks {
		weekByID[weeks[i].ID] = &weeks[i]
		weekIDs = append(weekIDs, weeks[i].ID)
	}

	var curLinks []models.TaskWeekPlanRel
	if len(weekIDs) > 0 {
		if err := db.Where("task_id = ? AND week_plan_id IN ?", taskID, weekIDs).Find(&curLinks).Error; err != nil {
			return nil, err
		}
	}
	if len(curLinks) == 0 {
		return nil, bizErr(404, "该任务不在本项目任一周期中")
	}
	// 校验任务当前所在周的锁定态与权限
	for _, l := range curLinks {
		w, ok := weekByID[l.WeekPlanID]
		if !ok {
			continue
		}
		if w.SummaryAt != nil {
			return nil, bizErr(400, fmt.Sprintf("第%d周总结已锁定，不可移动任务", w.WeekNo))
		}
		if userID != w.OwnerID && !canActAsPMOrAdmin(p, userID, isAdmin) {
			return nil, bizErr(403, "无权限移动该任务")
		}
	}
	if target.SummaryAt != nil {
		return nil, bizErr(400, "目标周总结已锁定，不可移入")
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ? AND week_plan_id IN ?", taskID, weekIDs).Delete(&models.TaskWeekPlanRel{}).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&models.TaskWeekPlanRel{}).Where("task_id = ? AND week_plan_id = ?", taskID, target.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return tx.Create(&models.TaskWeekPlanRel{TaskID: taskID, WeekPlanID: target.ID, CreateUser: &userID}).Error
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var task models.TaskItem
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
		// 事项已删除时返回空视图（对齐前端契约 task ? view : null）
		return nil, nil
	}
	// 返回任务全部周关联视图
	var allLinks []models.TaskWeekPlanRel
	if err := db.Where("task_id = ?", taskID).Find(&allLinks).Error; err != nil {
		return nil, err
	}
	allWeekIDs := make([]int64, 0, len(allLinks))
	for _, l := range allLinks {
		allWeekIDs = append(allWeekIDs, l.WeekPlanID)
	}
	allWeekByID := map[int64]*models.ProjectWeeklyPlan{}
	if len(allWeekIDs) > 0 {
		var allWeeks []models.ProjectWeeklyPlan
		if err := db.Where("id IN ? AND deleted = ?", allWeekIDs, 0).Find(&allWeeks).Error; err != nil {
			return nil, err
		}
		for i := range allWeeks {
			allWeekByID[allWeeks[i].ID] = &allWeeks[i]
		}
	}
	names := s.userNames([]int64{task.UserID})
	return taskWeekView(&task, allLinks, allWeekByID, names), nil
}
