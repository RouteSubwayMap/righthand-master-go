package handler

import (
	"fmt"
	"strconv"

	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

type BatchTaskHandler struct {
	batchTaskService *service.BatchTaskService
}

func NewBatchTaskHandler() *BatchTaskHandler {
	return &BatchTaskHandler{
		batchTaskService: service.NewBatchTaskService(),
	}
}

// ListBatchTasks 获取批量任务列表
func (h *BatchTaskHandler) ListBatchTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	bankType := c.Query("bankType")
	bankName := c.Query("bankName")
	systemName := c.Query("systemName")
	taskType := c.Query("taskType")
	taskName := c.Query("taskName")
	taskDesc := c.Query("taskDesc")
	upstreamSystem := c.Query("upstreamSystem")
	downstreamSystem := c.Query("downstreamSystem")

	result, err := h.batchTaskService.ListBatchTasks(page, size, bankType, bankName, systemName, taskType, taskName, taskDesc, upstreamSystem, downstreamSystem)
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询失败: "+err.Error())
		return
	}
	utils.Success(c, result)
}

// GetBatchTask 获取批量任务详情
func (h *BatchTaskHandler) GetBatchTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的任务ID")
		return
	}

	task, err := h.batchTaskService.GetBatchTask(id)
	if err != nil {
		utils.ErrorWithCode(c, 404, "任务不存在")
		return
	}
	utils.Success(c, task)
}

// CreateBatchTask 创建批量任务
func (h *BatchTaskHandler) CreateBatchTask(c *gin.Context) {
	var req struct {
		BankType         string `json:"bankType" binding:"required"`
		BankName         string `json:"bankName" binding:"required"`
		SystemName       string `json:"systemName" binding:"required"`
		TaskType         string `json:"taskType" binding:"required"`
		TaskName         string `json:"taskName" binding:"required"`
		TaskDesc         string `json:"taskDesc"`
		ExecTime         string `json:"execTime"`
		UpstreamSystem   string `json:"upstreamSystem"`
		UpstreamInfo     string `json:"upstreamInfo"`
		DownstreamSystem string `json:"downstreamSystem"`
		DownstreamInfo   string `json:"downstreamInfo"`
		Remark           string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	task := &models.BatchTask{
		BankType:         req.BankType,
		BankName:         req.BankName,
		SystemName:       req.SystemName,
		TaskType:         req.TaskType,
		TaskName:         req.TaskName,
		TaskDesc:         req.TaskDesc,
		ExecTime:         req.ExecTime,
		UpstreamSystem:   req.UpstreamSystem,
		UpstreamInfo:     req.UpstreamInfo,
		DownstreamSystem: req.DownstreamSystem,
		DownstreamInfo:   req.DownstreamInfo,
		Remark:           req.Remark,
		CreateUser:       &userID,
	}

	if err := h.batchTaskService.CreateBatchTask(task); err != nil {
		utils.ErrorWithCode(c, 500, "创建失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "创建成功", task)
}

// UpdateBatchTask 更新批量任务
func (h *BatchTaskHandler) UpdateBatchTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的任务ID")
		return
	}

	var req struct {
		BankType         string `json:"bankType"`
		BankName         string `json:"bankName"`
		SystemName       string `json:"systemName"`
		TaskType         string `json:"taskType"`
		TaskName         string `json:"taskName"`
		TaskDesc         string `json:"taskDesc"`
		ExecTime         string `json:"execTime"`
		UpstreamSystem   string `json:"upstreamSystem"`
		UpstreamInfo     string `json:"upstreamInfo"`
		DownstreamSystem string `json:"downstreamSystem"`
		DownstreamInfo   string `json:"downstreamInfo"`
		Remark           string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	updates := map[string]interface{}{
		"bank_type":         req.BankType,
		"bank_name":         req.BankName,
		"system_name":       req.SystemName,
		"task_type":         req.TaskType,
		"task_name":         req.TaskName,
		"task_desc":         req.TaskDesc,
		"exec_time":         req.ExecTime,
		"upstream_system":   req.UpstreamSystem,
		"upstream_info":     req.UpstreamInfo,
		"downstream_system": req.DownstreamSystem,
		"downstream_info":   req.DownstreamInfo,
		"remark":            req.Remark,
		"update_user":       userID,
	}

	if err := h.batchTaskService.UpdateBatchTask(id, updates); err != nil {
		utils.ErrorWithCode(c, 500, "更新失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteBatchTask 删除批量任务
func (h *BatchTaskHandler) DeleteBatchTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的任务ID")
		return
	}

	if err := h.batchTaskService.DeleteBatchTask(id); err != nil {
		utils.ErrorWithCode(c, 500, "删除失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "删除成功", nil)
}

// BatchImportTasks 批量导入任务
func (h *BatchTaskHandler) BatchImportTasks(c *gin.Context) {
	var req struct {
		Tasks []struct {
			BankType         string `json:"bankType"`
			BankName         string `json:"bankName"`
			SystemName       string `json:"systemName"`
			TaskType         string `json:"taskType"`
			TaskName         string `json:"taskName"`
			TaskDesc         string `json:"taskDesc"`
			ExecTime         string `json:"execTime"`
			UpstreamSystem   string `json:"upstreamSystem"`
			UpstreamInfo     string `json:"upstreamInfo"`
			DownstreamSystem string `json:"downstreamSystem"`
			DownstreamInfo   string `json:"downstreamInfo"`
			Remark           string `json:"remark"`
		} `json:"tasks"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	if len(req.Tasks) == 0 {
		utils.ErrorWithCode(c, 400, "导入数据不能为空")
		return
	}

	userID := c.GetInt64("userID")
	tasks := make([]models.BatchTask, 0, len(req.Tasks))
	for _, t := range req.Tasks {
		// 跳过必填字段为空的行
		if t.BankType == "" || t.BankName == "" || t.TaskName == "" {
			continue
		}
		tasks = append(tasks, models.BatchTask{
			BankType:         t.BankType,
			BankName:         t.BankName,
			SystemName:       t.SystemName,
			TaskType:         t.TaskType,
			TaskName:         t.TaskName,
			TaskDesc:         t.TaskDesc,
			ExecTime:         t.ExecTime,
			UpstreamSystem:   t.UpstreamSystem,
			UpstreamInfo:     t.UpstreamInfo,
			DownstreamSystem: t.DownstreamSystem,
			DownstreamInfo:   t.DownstreamInfo,
			Remark:           t.Remark,
			CreateUser:       &userID,
		})
	}

	count, err := h.batchTaskService.BatchImportTasks(tasks)
	if err != nil {
		utils.ErrorWithCode(c, 500, "导入失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, fmt.Sprintf("成功导入 %d 条任务", count), map[string]interface{}{
		"count":   count,
		"skipped": len(req.Tasks) - count,
	})
}
