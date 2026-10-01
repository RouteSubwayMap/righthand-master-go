package handler

import (
	"strconv"

	"righthand-go/internal/utils"
	"righthand-go/services/task-service/dto"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

// ProjectWeekHandler 项目管理（周计划 + 总结 + 任务-周关联）
type ProjectWeekHandler struct {
	projectService *service.ProjectService
}

func NewProjectWeekHandler() *ProjectWeekHandler {
	return &ProjectWeekHandler{
		projectService: service.NewProjectService(),
	}
}

/* ================================ 周计划 ================================ */

// CreateWeek 新增周计划
func (h *ProjectWeekHandler) CreateWeek(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	var req dto.WeekCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.CreateWeek(id, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "新增周计划失败")
		return
	}
	utils.SuccessWithMessage(c, "创建成功", result)
}

// UpdateWeek 编辑周计划
func (h *ProjectWeekHandler) UpdateWeek(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	var req dto.WeekUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.UpdateWeek(id, weekID, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "编辑周计划失败")
		return
	}
	utils.SuccessWithMessage(c, "保存成功", result)
}

// DeleteWeek 删除周计划
func (h *ProjectWeekHandler) DeleteWeek(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	if err := h.projectService.DeleteWeek(id, weekID, userID, isAdmin); err != nil {
		respondProjectError(c, err, "删除周计划失败")
		return
	}
	utils.SuccessWithMessage(c, "删除成功", nil)
}

// ChangeWeekStatus 周状态流转
func (h *ProjectWeekHandler) ChangeWeekStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	var req dto.WeekStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.ChangeWeekStatus(id, weekID, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "周状态流转失败")
		return
	}
	utils.Success(c, result)
}

// SubmitWeekSummary 提交周总结（提交即锁定）
func (h *ProjectWeekHandler) SubmitWeekSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	var req dto.WeekSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.SubmitWeekSummary(id, weekID, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "提交总结失败")
		return
	}
	utils.SuccessWithMessage(c, "提交成功", result)
}

// UnlockWeekSummary 解锁周总结（PM/ADMIN）
func (h *ProjectWeekHandler) UnlockWeekSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.UnlockWeekSummary(id, weekID, userID, isAdmin)
	if err != nil {
		respondProjectError(c, err, "解锁总结失败")
		return
	}
	utils.SuccessWithMessage(c, "已解锁", result)
}

/* ============================ 任务-周关联 ============================ */

// ListProjectTasks 项目任务列表（支持 weekPlanId / userId / statusCode 过滤；按任务去重）
func (h *ProjectWeekHandler) ListProjectTasks(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	filters := map[string]interface{}{}
	if v := c.Query("weekPlanId"); v != "" {
		filters["weekPlanId"] = v
	}
	if v := c.Query("userId"); v != "" {
		filters["userId"] = v
	}
	if v := c.Query("statusCode"); v != "" {
		filters["statusCode"] = v
	}

	result, err := h.projectService.ListProjectTasks(id, page, size, filters)
	if err != nil {
		respondProjectError(c, err, "查询项目任务失败")
		return
	}
	utils.Success(c, result)
}

// CreateTaskInWeek 在指定周新建任务（创建事项 + 建关联）
func (h *ProjectWeekHandler) CreateTaskInWeek(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	var req dto.WeekTaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.CreateTaskInWeek(id, weekID, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "新建任务失败")
		return
	}
	utils.SuccessWithMessage(c, "创建成功", result)
}

// LinkTaskToWeek 将已有事项加入该周（幂等）
func (h *ProjectWeekHandler) LinkTaskToWeek(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的任务ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, already, err := h.projectService.LinkTaskToWeek(id, weekID, taskID, userID, isAdmin)
	if err != nil {
		respondProjectError(c, err, "关联任务失败")
		return
	}
	if already {
		utils.SuccessWithMessage(c, "任务已在该周", result)
		return
	}
	utils.SuccessWithMessage(c, "关联成功", result)
}

// UnlinkTaskFromWeek 将事项从该周移出
func (h *ProjectWeekHandler) UnlinkTaskFromWeek(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	weekID, err := strconv.ParseInt(c.Param("weekId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的周计划ID")
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的任务ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	if err := h.projectService.UnlinkTaskFromWeek(id, weekID, taskID, userID, isAdmin); err != nil {
		respondProjectError(c, err, "移出任务失败")
		return
	}
	utils.SuccessWithMessage(c, "已从该周移出", nil)
}

// MoveTaskToWeek 将任务移动到项目内其他周
func (h *ProjectWeekHandler) MoveTaskToWeek(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	taskID, err := strconv.ParseInt(c.Param("taskId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的任务ID")
		return
	}
	var req dto.TaskMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.MoveTaskToWeek(id, taskID, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "移动任务失败")
		return
	}
	utils.SuccessWithMessage(c, "移动成功", result)
}
