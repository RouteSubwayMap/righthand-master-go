package handler

import (
	"strconv"
	"time"

	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/task-service/dto"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		taskService: service.NewTaskService(),
	}
}

// ListTasks 获取事项列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	title := c.Query("title")           // 标题关键字
	itemType := c.Query("itemType")     // TASK/MEMO/REQUIREMENT
	statusCode := c.Query("statusCode") // TODO/DOING/DONE/CANCELLED
	priority := c.Query("priority")     // HIGH/MEDIUM/LOW
	category := c.Query("category")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	today := c.Query("today")
	overdue := c.Query("overdue")
	createUser := c.Query("createUser") // 创建人ID

	filter := make(map[string]interface{})
	if title != "" {
		filter["title"] = title
	}
	if itemType != "" {
		filter["item_type"] = itemType
	}
	if statusCode != "" {
		filter["status_code"] = statusCode
	}
	if priority != "" {
		filter["priority"] = priority
	}
	if category != "" {
		filter["category"] = category
	}
	if startDate != "" {
		filter["start_date"] = startDate
	}
	if endDate != "" {
		filter["end_date"] = endDate
	}
	if today == "true" {
		filter["today"] = true
	}
	if overdue == "true" {
		filter["overdue"] = true
	}
	if createUser != "" {
		filter["create_user"] = createUser
	}

	userID := c.GetInt64("userID")
	isAdmin, _ := c.Get("isAdmin")
	result, err := h.taskService.ListTasks(userID, isAdmin.(bool), page, size, filter)
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}

// GetTask 获取事项详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的事项ID")
		return
	}

	userID := c.GetInt64("userID")
	isAdmin, _ := c.Get("isAdmin")
	task, err := h.taskService.GetTask(id, userID, isAdmin.(bool))
	if err != nil {
		utils.ErrorWithCode(c, 404, "事项不存在")
		return
	}

	// 查询关联的需求信息
	reqID, reqNo, reqTitle, _ := h.taskService.GetTaskRequirement(id)

	utils.Success(c, gin.H{
		"id":                 task.ID,
		"userId":             task.UserID,
		"title":              task.Title,
		"description":        task.Description,
		"itemType":           task.ItemType,
		"category":           task.Category,
		"priority":           task.Priority,
		"tags":               task.Tags,
		"startTime":         task.StartTime,
		"deadline":           task.Deadline,
		"completedTime":      task.CompletedTime,
		"statusCode":         task.StatusCode,
		"progress":           task.Progress,
		"recurrenceType":     task.RecurrenceType,
		"recurrenceRule":     task.RecurrenceRule,
		"recurrenceEndType":  task.RecurrenceEndType,
		"recurrenceEndDate":  task.RecurrenceEndDate,
		"recurrenceCount":    task.RecurrenceCount,
		"parentRecurrenceId": task.ParentRecurrenceID,
		"createTime":         task.CreateTime,
		"createUser":         task.CreateUser,
		"updateTime":         task.UpdateTime,
		"updateUser":         task.UpdateUser,
		"requirementId":      reqID,
		"requirementNo":      reqNo,
		"requirementTitle":   reqTitle,
	})
}

// CustomTimeToTimePtr 将 CustomTime 指针转换为 time.Time 指针
func CustomTimeToTimePtr(ct *dto.CustomTime) *time.Time {
	if ct == nil {
		return nil
	}
	t := ct.Time
	return &t
}

// CreateTask 创建事项
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		Title             string          `json:"title" binding:"required"`
		Description       string          `json:"description"`
		ItemType          string          `json:"itemType" binding:"required,oneof=TASK MEMO REQUIREMENT"`
		Category          string          `json:"category"`
		Priority          string          `json:"priority" binding:"omitempty,oneof=HIGH MEDIUM LOW"`
		Tags              string          `json:"tags"`
		StartTime         *dto.CustomTime `json:"startTime"`
		Deadline          *dto.CustomTime `json:"deadline"`
		RecurrenceType    string          `json:"recurrenceType" binding:"omitempty,oneof=NONE DAILY WEEKLY MONTHLY YEARLY"`
		RecurrenceRule    string          `json:"recurrenceRule"`
		RecurrenceEndType string          `json:"recurrenceEndType" binding:"omitempty,oneof=NEVER DATE COUNT"`
		RecurrenceEndDate *dto.CustomTime `json:"recurrenceEndDate"`
		RecurrenceCount   int             `json:"recurrenceCount"`
		RequirementID     *int64          `json:"requirementId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Priority == "" {
		req.Priority = "MEDIUM"
	}
	if req.ItemType == "" {
		req.ItemType = "TASK"
	}
	if req.RecurrenceType == "" {
		req.RecurrenceType = "NONE"
	}

	userID := c.GetInt64("userID")

	task := &models.TaskItem{
		UserID:            userID,
		Title:             req.Title,
		Description:       req.Description,
		ItemType:          req.ItemType,
		Category:          req.Category,
		Priority:          req.Priority,
		Tags:              req.Tags,
		StartTime:         CustomTimeToTimePtr(req.StartTime),
		Deadline:          CustomTimeToTimePtr(req.Deadline),
		StatusCode:        "TODO",
		Progress:          0,
		RecurrenceType:    req.RecurrenceType,
		RecurrenceRule:    req.RecurrenceRule,
		RecurrenceEndType: req.RecurrenceEndType,
		RecurrenceEndDate: CustomTimeToTimePtr(req.RecurrenceEndDate),
		RecurrenceCount:   req.RecurrenceCount,
		CreateUser:        func() *int64 { return &userID }(),
	}

	if err := h.taskService.CreateTask(task); err != nil {
		utils.ErrorWithCode(c, 500, "创建失败: "+err.Error())
		return
	}

	// 若类型为需求且指定了关联需求，写入关联表
	if req.ItemType == "REQUIREMENT" && req.RequirementID != nil {
		_ = h.taskService.SetTaskRequirement(task.ID, *req.RequirementID, userID)
	}

	utils.SuccessWithMessage(c, "创建成功", task)
}

// UpdateTask 更新事项
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的事项ID")
		return
	}

	var req struct {
		Title             string          `json:"title"`
		Description       string          `json:"description"`
		ItemType          string          `json:"itemType"`
		Category          string          `json:"category"`
		Priority          string          `json:"priority"`
		Tags              string          `json:"tags"`
		StartTime         *dto.CustomTime `json:"startTime"`
		Deadline          *dto.CustomTime `json:"deadline"`
		RecurrenceType    string          `json:"recurrenceType"`
		RecurrenceRule    string          `json:"recurrenceRule"`
		RecurrenceEndType string          `json:"recurrenceEndType"`
		RecurrenceEndDate *dto.CustomTime `json:"recurrenceEndDate"`
		RecurrenceCount   int             `json:"recurrenceCount"`
		RequirementID     *int64          `json:"requirementId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ItemType != "" {
		updates["item_type"] = req.ItemType
	}
	updates["category"] = req.Category
	if req.Priority != "" {
		updates["priority"] = req.Priority
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.StartTime != nil {
		updates["start_time"] = CustomTimeToTimePtr(req.StartTime)
	}
	if req.Deadline != nil {
		updates["deadline"] = CustomTimeToTimePtr(req.Deadline)
	}
	if req.RecurrenceType != "" {
		updates["recurrence_type"] = req.RecurrenceType
	}
	if req.RecurrenceRule != "" {
		updates["recurrence_rule"] = req.RecurrenceRule
	}
	if req.RecurrenceEndType != "" {
		updates["recurrence_end_type"] = req.RecurrenceEndType
	}
	if req.RecurrenceEndDate != nil {
		updates["recurrence_end_date"] = CustomTimeToTimePtr(req.RecurrenceEndDate)
	}
	if req.RecurrenceCount > 0 {
		updates["recurrence_count"] = req.RecurrenceCount
	}

	updates["update_user"] = userID

	if err := h.taskService.UpdateTask(id, userID, updates); err != nil {
		utils.ErrorWithCode(c, 500, "更新失败: "+err.Error())
		return
	}

	// 处理需求关联：有 requirementId 则设置关联，无则清除关联
	if req.RequirementID != nil {
		_ = h.taskService.SetTaskRequirement(id, *req.RequirementID, userID)
	} else {
		_ = h.taskService.RemoveTaskRequirement(id)
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteTask 删除事项
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的事项ID")
		return
	}

	userID := c.GetInt64("userID")
	if err := h.taskService.DeleteTask(id, userID); err != nil {
		utils.ErrorWithCode(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// BatchDeleteTasks 批量删除事项
func (h *TaskHandler) BatchDeleteTasks(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "请求参数错误")
		return
	}

	if len(req.IDs) == 0 {
		utils.ErrorWithCode(c, 400, "请选择要删除的事项")
		return
	}

	userID := c.GetInt64("userID")
	if err := h.taskService.BatchDeleteTasks(req.IDs, userID); err != nil {
		utils.ErrorWithCode(c, 500, "批量删除失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "批量删除成功", nil)
}

// UpdateStatus 更新事项状态
func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的事项ID")
		return
	}

	var req struct {
		StatusCode string `json:"statusCode" binding:"required,oneof=TODO DOING DONE CANCELLED"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	updates := map[string]interface{}{
		"status_code": req.StatusCode,
		"update_user": userID,
	}

	// 如果状态变为DONE,记录完成时间
	if req.StatusCode == "DONE" {
		now := time.Now()
		updates["completed_time"] = &now
	}

	if err := h.taskService.UpdateTask(id, userID, updates); err != nil {
		utils.ErrorWithCode(c, 500, "更新状态失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "状态更新成功", nil)
}

// GetDashboardStats 获取仪表盘统计
func (h *TaskHandler) GetDashboardStats(c *gin.Context) {
	userID := c.GetInt64("userID")
	isAdmin, _ := c.Get("isAdmin")
	stats, err := h.taskService.GetDashboardStats(userID, isAdmin.(bool))
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询统计失败: "+err.Error())
		return
	}

	utils.Success(c, stats)
}

// TransferTask 转交事项
func (h *TaskHandler) TransferTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的事项ID")
		return
	}

	var req struct {
		TargetUserId int64 `json:"targetUserId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	if err := h.taskService.TransferTask(id, userID, req.TargetUserId); err != nil {
		utils.ErrorWithCode(c, 500, "转交失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "转交成功", nil)
}
