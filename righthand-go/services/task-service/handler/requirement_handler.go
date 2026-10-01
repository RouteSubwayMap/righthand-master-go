package handler

import (
	"database/sql"
	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/task-service/dto"
	"righthand-go/services/task-service/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RequirementHandler 需求管理处理器
type RequirementHandler struct {
	requirementService *service.RequirementService
	taskService        *service.TaskService
}

// NewRequirementHandler 创建需求处理器实例
func NewRequirementHandler() *RequirementHandler {
	return &RequirementHandler{
		requirementService: service.NewRequirementService(),
		taskService:        service.NewTaskService(),
	}
}

// CreateRequirement 创建需求
func (h *RequirementHandler) CreateRequirement(c *gin.Context) {
	var reqDTO dto.RequirementRequest

	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	// 转换DTO为模型
	req := dtoToModel(&reqDTO)
	req.UserID = userID.(int64)

	// 设置创建人(从上下文获取)
	if createUser, exists := c.Get("userID"); exists {
		uid := createUser.(int64)
		req.CreateUser = &uid
	}

	if err := h.requirementService.CreateRequirement(req); err != nil {
		utils.Error(c, "创建需求失败: "+err.Error())
		return
	}

	utils.Success(c, req)
}

// dtoToModel 将需求请求DTO转换为模型
func dtoToModel(reqDTO *dto.RequirementRequest) *models.Requirement {
	// 处理 Title（非指针类型）
	title := ""
	if reqDTO.Title != nil {
		title = *reqDTO.Title
	}

	req := &models.Requirement{
		Title:           title,
		BankName:        nilStringToPtr(reqDTO.BankName),
		RequirementNo:   nilStringToPtr(reqDTO.RequirementNo),
		Priority:        nilStringToPtr(reqDTO.Priority),
		RiskLevel:       nilStringToPtr(reqDTO.RiskLevel),
		RequirementType: nilStringToPtr(reqDTO.RequirementType),
		Status:          nilStringToPtr(reqDTO.Status),
		ScheduleDate:    nilStringToPtr(reqDTO.ScheduleDate),
		MainSystem:      nilStringToPtr(reqDTO.MainSystem),
		Developer:       nilStringToPtr(reqDTO.Developer),
		InvolvedSystems: nilStringToPtr(reqDTO.InvolvedSystems),
		FollowerID:      reqDTO.FollowerID,
		FollowerName:    nilStringToPtr(reqDTO.FollowerName),
		SubmitterID:     reqDTO.SubmitterID,
		SubmitterName:   nilStringToPtr(reqDTO.SubmitterName),
		SubmitTime:      dto.ParseDateTime(reqDTO.SubmitTime),
		CompletionDate:  dto.ParseDateTime(reqDTO.CompletionDate),
		GroupType:       nilStringToPtr(reqDTO.GroupType),
		Description:     nilStringToPtr(reqDTO.Description),
		Remark:          nilStringToPtr(reqDTO.Remark),
	}

	// 处理 estimatedWorkload
	if reqDTO.EstimatedWorkload != nil {
		req.EstimatedWorkload = sql.NullFloat64{
			Float64: *reqDTO.EstimatedWorkload,
			Valid:   true,
		}
	}

	return req
}

// nilStringToPtr 处理空字符串，转换为 nil 指针
func nilStringToPtr(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

// GetRequirement 获取需求详情
func (h *RequirementHandler) GetRequirement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的需求ID")
		return
	}

	if _, exists := c.Get("userID"); !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	requirement, err := h.requirementService.GetRequirementByID(id)
	if err != nil {
		utils.NotFound(c, "需求不存在")
		return
	}

	utils.Success(c, requirement)
}

// GetRequirementTasks 获取需求关联的事项列表
func (h *RequirementHandler) GetRequirementTasks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的需求ID")
		return
	}

	tasks, err := h.taskService.ListTasksByRequirementID(id)
	if err != nil {
		utils.Error(c, "查询关联事项失败")
		return
	}

	utils.Success(c, tasks)
}

// UpdateRequirement 更新需求
func (h *RequirementHandler) UpdateRequirement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的需求ID")
		return
	}

	var reqDTO dto.RequirementRequest
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 转换DTO为模型
	req := dtoToModel(&reqDTO)
	req.ID = id

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	req.UserID = userID.(int64)

	isAdmin, _ := c.Get("isAdmin")

	// 设置更新人(从上下文获取)
	if updateUser, exists := c.Get("userID"); exists {
		uid := updateUser.(int64)
		req.UpdateUser = &uid
	}

	if err := h.requirementService.UpdateRequirement(req, isAdmin.(bool)); err != nil {
		utils.Error(c, "更新需求失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteRequirement 删除需求
func (h *RequirementHandler) DeleteRequirement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的需求ID")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	role, _ := c.Get("role")
	isAdmin := role == "admin"

	if err := h.requirementService.DeleteRequirement(id, userID.(int64), isAdmin); err != nil {
		utils.Error(c, "删除需求失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// ListRequirements 查询需求列表
func (h *RequirementHandler) ListRequirements(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("pageNum", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取筛选参数
	status := c.Query("status")
	priority := c.Query("priority")
	bankName := c.Query("bankName")
	groupType := c.Query("groupType")
	title := c.Query("title")
	requirementNo := c.Query("requirementNo")
	riskLevel := c.Query("riskLevel")
	requirementType := c.Query("requirementType")
	followerName := c.Query("followerName")
	createdByStr := c.Query("createdBy")
	completionDateStart := c.Query("completionDateStart")
	completionDateEnd := c.Query("completionDateEnd")

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	isAdmin, _ := c.Get("isAdmin")

	var statusPtr, priorityPtr, bankNamePtr, groupTypePtr, titlePtr, requirementNoPtr, riskLevelPtr, requirementTypePtr, followerNamePtr *string
	if status != "" {
		statusPtr = &status
	}
	if priority != "" {
		priorityPtr = &priority
	}
	if bankName != "" {
		bankNamePtr = &bankName
	}
	if groupType != "" {
		groupTypePtr = &groupType
	}
	if title != "" {
		titlePtr = &title
	}
	if requirementNo != "" {
		requirementNoPtr = &requirementNo
	}
	if riskLevel != "" {
		riskLevelPtr = &riskLevel
	}
	if requirementType != "" {
		requirementTypePtr = &requirementType
	}
	if followerName != "" {
		followerNamePtr = &followerName
	}

	var createdByPtr *int64
	if createdByStr != "" {
		if createdBy, err := strconv.ParseInt(createdByStr, 10, 64); err == nil && createdBy > 0 {
			createdByPtr = &createdBy
		}
	}

	var completionDateStartPtr, completionDateEndPtr *string
	if completionDateStart != "" {
		completionDateStartPtr = &completionDateStart
	}
	if completionDateEnd != "" {
		completionDateEndPtr = &completionDateEnd
	}

	requirements, total, err := h.requirementService.ListRequirements(
		userID.(int64), isAdmin.(bool), page, pageSize, statusPtr, priorityPtr, bankNamePtr, groupTypePtr, titlePtr, requirementNoPtr, riskLevelPtr, requirementTypePtr, followerNamePtr,
		createdByPtr, completionDateStartPtr, completionDateEndPtr,
	)
	if err != nil {
		utils.Error(c, "查询需求列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":     requirements,
		"total":    total,
		"pageNum":  page,
		"pageSize": pageSize,
	})
}

// UpdateStatus 更新需求状态
func (h *RequirementHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的需求ID")
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	if err := h.requirementService.UpdateStatus(id, userID.(int64), body.Status); err != nil {
		utils.Error(c, "更新状态失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "状态更新成功", nil)
}

// ExportRequirements 导出需求列表
func (h *RequirementHandler) ExportRequirements(c *gin.Context) {
	// 获取筛选参数
	status := c.Query("status")
	priority := c.Query("priority")
	bankName := c.Query("bankName")
	groupType := c.Query("groupType")
	title := c.Query("title")
	requirementNo := c.Query("requirementNo")
	riskLevel := c.Query("riskLevel")
	requirementType := c.Query("requirementType")
	followerName := c.Query("followerName")
	createdByStr := c.Query("createdBy")
	completionDateStart := c.Query("completionDateStart")
	completionDateEnd := c.Query("completionDateEnd")

	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	isAdmin, _ := c.Get("isAdmin")

	var statusPtr, priorityPtr, bankNamePtr, groupTypePtr, titlePtr, requirementNoPtr, riskLevelPtr, requirementTypePtr, followerNamePtr *string
	if status != "" {
		statusPtr = &status
	}
	if priority != "" {
		priorityPtr = &priority
	}
	if bankName != "" {
		bankNamePtr = &bankName
	}
	if groupType != "" {
		groupTypePtr = &groupType
	}
	if title != "" {
		titlePtr = &title
	}
	if requirementNo != "" {
		requirementNoPtr = &requirementNo
	}
	if riskLevel != "" {
		riskLevelPtr = &riskLevel
	}
	if requirementType != "" {
		requirementTypePtr = &requirementType
	}
	if followerName != "" {
		followerNamePtr = &followerName
	}

	var createdByPtr *int64
	if createdByStr != "" {
		if createdBy, err := strconv.ParseInt(createdByStr, 10, 64); err == nil && createdBy > 0 {
			createdByPtr = &createdBy
		}
	}

	var completionDateStartPtr, completionDateEndPtr *string
	if completionDateStart != "" {
		completionDateStartPtr = &completionDateStart
	}
	if completionDateEnd != "" {
		completionDateEndPtr = &completionDateEnd
	}

	requirements, err := h.requirementService.ExportRequirements(
		userID.(int64), isAdmin.(bool),
		statusPtr, priorityPtr, bankNamePtr, groupTypePtr,
		titlePtr, requirementNoPtr, riskLevelPtr, requirementTypePtr, followerNamePtr,
		createdByPtr, completionDateStartPtr, completionDateEndPtr,
	)
	if err != nil {
		utils.Error(c, "导出数据失败: "+err.Error())
		return
	}

	// TODO: 实现Excel导出功能
	// 目前先返回JSON数据
	utils.Success(c, gin.H{
		"data":  requirements,
		"count": len(requirements),
	})
}

// DownloadTemplate 下载导入模板
func (h *RequirementHandler) DownloadTemplate(c *gin.Context) {
	// 生成CSV格式的模板文件
	template := "需求名称,所属银行,需求编号,优先级,风险等级,需求类型,主R系统,开发主R|参与者,涉及系统,需求跟进人,提交人,预计工作量(人天),所属小组,需求描述,备注\n"
	template += "示例需求,工商银行,REQ-2024-001,HIGH,MEDIUM,INTERNAL,R系统,张三|R&D团队,系统A;系统B,李四,王五,5.0,LOAN_MIDDLE,这是一个示例需求的描述,这是备注信息\n"

	// 设置响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=requirement_template.csv")
	c.Header("Content-Transfer-Encoding", "binary")

	// 添加BOM头以支持中文
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	c.Writer.WriteString(template)
}

// ImportRequirements 批量导入需求
func (h *RequirementHandler) ImportRequirements(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		Requirements []struct {
			Title             string  `json:"title"`
			BankName          string  `json:"bankName"`
			RequirementNo     string  `json:"requirementNo"`
			Priority          string  `json:"priority"`
			RiskLevel         string  `json:"riskLevel"`
			RequirementType   string  `json:"requirementType"`
			MainSystem        string  `json:"mainSystem"`
			Developer         string  `json:"developer"`
			InvolvedSystems   string  `json:"involvedSystems"`
			FollowerName      string  `json:"followerName"`
			SubmitterName     string  `json:"submitterName"`
			EstimatedWorkload float64 `json:"estimatedWorkload"`
			GroupType         string  `json:"groupType"`
			Description       string  `json:"description"`
			Remark            string  `json:"remark"`
		} `json:"requirements"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.Requirements) == 0 {
		utils.BadRequest(c, "导入数据不能为空")
		return
	}

	result, err := h.requirementService.ImportRequirementsFromJSON(userID.(int64), req.Requirements)
	if err != nil {
		utils.Error(c, "导入失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "导入完成", result)
}
