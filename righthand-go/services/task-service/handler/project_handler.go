package handler

import (
	"errors"
	"strconv"

	"righthand-go/internal/utils"
	"righthand-go/services/task-service/dto"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 项目管理（项目 + 成员）
type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{
		projectService: service.NewProjectService(),
	}
}

// respondProjectError 业务错误 -> 响应码映射（BizError 携带语义码，其余 500）
func respondProjectError(c *gin.Context, err error, fallback string) {
	var bizErr *service.BizError
	if errors.As(err, &bizErr) {
		utils.ErrorWithCode(c, bizErr.Code, bizErr.Message)
		return
	}
	utils.ErrorWithCode(c, 500, fallback+": "+err.Error())
}

// ListProjects 项目列表（筛选/排序/分页）
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	filter := map[string]interface{}{}
	if v := c.Query("status"); v != "" {
		filter["status"] = v
	}
	if v := c.Query("riskLevel"); v != "" {
		filter["risk_level"] = v
	}
	if v := c.Query("deliveryStatus"); v != "" {
		filter["delivery_status"] = v
	}
	if v := c.Query("managerId"); v != "" {
		filter["manager_id"] = v
	}
	if v := c.Query("startDate"); v != "" {
		filter["start_date"] = v
	}
	if v := c.Query("endDate"); v != "" {
		filter["end_date"] = v
	}
	if v := c.Query("keyword"); v != "" {
		filter["keyword"] = v
	}
	if v := c.Query("participantId"); v != "" {
		filter["participant_id"] = v
	}
	if v := c.Query("sortBy"); v != "" {
		filter["sort_by"] = v
	}
	if v := c.Query("order"); v != "" {
		filter["order"] = v
	}

	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.ListProjects(userID, isAdmin, page, size, filter)
	if err != nil {
		respondProjectError(c, err, "查询项目列表失败")
		return
	}
	utils.Success(c, result)
}

// CreateProject 新建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req dto.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.CreateProject(userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "创建项目失败")
		return
	}
	utils.SuccessWithMessage(c, "创建成功", result)
}

// GetProjectDetail 项目详情（含周计划/成员/变更历史）
func (h *ProjectHandler) GetProjectDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.GetProjectDetail(id, userID, isAdmin)
	if err != nil {
		respondProjectError(c, err, "查询项目详情失败")
		return
	}
	utils.Success(c, result)
}

// UpdateProject 编辑项目
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	var req dto.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.UpdateProject(id, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "编辑项目失败")
		return
	}
	utils.SuccessWithMessage(c, "保存成功", result)
}

// DeleteProject 删除项目（仅 STARTING）
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	if err := h.projectService.DeleteProject(id, userID, isAdmin); err != nil {
		respondProjectError(c, err, "删除项目失败")
		return
	}
	utils.SuccessWithMessage(c, "删除成功", nil)
}

// ChangeProjectStatus 项目状态流转（start/pause/resume/close/terminate）
func (h *ProjectHandler) ChangeProjectStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	var req dto.ProjectStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.ChangeProjectStatus(id, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "状态流转失败")
		return
	}
	utils.Success(c, result)
}

// ChangeProjectRisk 修改风险等级/交付状态（必填原因）
func (h *ProjectHandler) ChangeProjectRisk(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	var req dto.ProjectRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.ChangeProjectRisk(id, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "修改风险信息失败")
		return
	}
	utils.Success(c, result)
}

/* ================================ 成员 ================================ */

// ListMembers 项目成员列表
func (h *ProjectHandler) ListMembers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	result, err := h.projectService.ListMembers(id)
	if err != nil {
		respondProjectError(c, err, "查询成员列表失败")
		return
	}
	utils.Success(c, result)
}

// AddMember 添加项目成员（幂等）
func (h *ProjectHandler) AddMember(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	var req dto.ProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, already, err := h.projectService.AddMember(id, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "添加成员失败")
		return
	}
	if already {
		utils.SuccessWithMessage(c, "成员已存在", result)
		return
	}
	utils.SuccessWithMessage(c, "添加成功", result)
}

// RemoveMember 移除项目成员
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	memberID, err := strconv.ParseInt(c.Param("memberId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的成员ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	if err := h.projectService.RemoveMember(id, memberID, userID, isAdmin); err != nil {
		respondProjectError(c, err, "移除成员失败")
		return
	}
	utils.SuccessWithMessage(c, "移除成功", nil)
}

/* ================================ 项目阶段 ================================ */

// ListStages 项目阶段列表（按开始时间正序）
func (h *ProjectHandler) ListStages(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	result, err := h.projectService.ListStages(id)
	if err != nil {
		respondProjectError(c, err, "查询阶段列表失败")
		return
	}
	utils.Success(c, result)
}

// CreateStage 新增项目阶段
func (h *ProjectHandler) CreateStage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	var req dto.StageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.CreateStage(id, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "新增阶段失败")
		return
	}
	utils.SuccessWithMessage(c, "创建成功", result)
}

// UpdateStage 编辑项目阶段
func (h *ProjectHandler) UpdateStage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	stageID, err := strconv.ParseInt(c.Param("stageId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的阶段ID")
		return
	}
	var req dto.StageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	result, err := h.projectService.UpdateStage(id, stageID, userID, isAdmin, &req)
	if err != nil {
		respondProjectError(c, err, "编辑阶段失败")
		return
	}
	utils.SuccessWithMessage(c, "保存成功", result)
}

// DeleteStage 删除项目阶段（软删除）
func (h *ProjectHandler) DeleteStage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的项目ID")
		return
	}
	stageID, err := strconv.ParseInt(c.Param("stageId"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的阶段ID")
		return
	}
	userID := c.GetInt64("userID")
	isAdmin := c.GetBool("isAdmin")

	if err := h.projectService.DeleteStage(id, stageID, userID, isAdmin); err != nil {
		respondProjectError(c, err, "删除阶段失败")
		return
	}
	utils.SuccessWithMessage(c, "删除成功", nil)
}
