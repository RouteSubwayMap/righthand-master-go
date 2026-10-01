package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"righthand-go/internal/utils"
	"righthand-go/internal/models"
	"righthand-go/services/auth-service/service"
)

// DictHandler 字典管理处理器
type DictHandler struct {
	dictService *service.DictService
}

// NewDictHandler 创建字典处理器实例
func NewDictHandler() *DictHandler {
	return &DictHandler{
		dictService: service.NewDictService(),
	}
}

// ListDicts 查询字典列表
func (h *DictHandler) ListDicts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	dictType := c.Query("dictType")

	dicts, total, err := h.dictService.ListDicts(page, pageSize, dictType)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, utils.PageResult{
		Total:    total,
		PageNum:  page,
		PageSize: pageSize,
		List:     dicts,
	})
}

// GetDictsByType 根据类型获取字典
func (h *DictHandler) GetDictsByType(c *gin.Context) {
	dictType := c.Param("type")

	dicts, err := h.dictService.GetDictsByType(dictType)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, dicts)
}

// GetDict 获取字典详情
func (h *DictHandler) GetDict(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的字典ID")
		return
	}

	dict, err := h.dictService.GetDictByID(id)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if dict == nil {
		utils.NotFound(c, "字典不存在")
		return
	}

	utils.Success(c, dict)
}

// CreateDict 创建字典
func (h *DictHandler) CreateDict(c *gin.Context) {
	var dict models.SysDict
	if err := c.ShouldBindJSON(&dict); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.dictService.CreateDict(&dict); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "创建成功", nil)
}

// UpdateDict 更新字典
func (h *DictHandler) UpdateDict(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的字典ID")
		return
	}

	var dict models.SysDict
	if err := c.ShouldBindJSON(&dict); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	dict.ID = id

	if err := h.dictService.UpdateDict(&dict); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteDict 删除字典
func (h *DictHandler) DeleteDict(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的字典ID")
		return
	}

	if err := h.dictService.DeleteDict(id); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}
