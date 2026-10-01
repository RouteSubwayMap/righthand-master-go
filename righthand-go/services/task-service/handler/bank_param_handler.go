package handler

import (
	"fmt"
	"strconv"

	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

type BankParamHandler struct {
	bankParamService *service.BankParamService
}

func NewBankParamHandler() *BankParamHandler {
	return &BankParamHandler{
		bankParamService: service.NewBankParamService(),
	}
}

// ListBankParams 获取银行机构参数列表
func (h *BankParamHandler) ListBankParams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	bankType := c.Query("bankType")
	bankName := c.Query("bankName")
	bankCode := c.Query("bankCode")

	result, err := h.bankParamService.ListBankParams(page, size, bankType, bankName, bankCode)
	if err != nil {
		utils.ErrorWithCode(c, 500, "获取列表失败: "+err.Error())
		return
	}
	utils.Success(c, result)
}

// GetBankParam 获取银行机构参数详情
func (h *BankParamHandler) GetBankParam(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的ID")
		return
	}

	param, err := h.bankParamService.GetBankParam(id)
	if err != nil {
		utils.ErrorWithCode(c, 500, "获取详情失败: "+err.Error())
		return
	}
	utils.Success(c, param)
}

// CreateBankParam 创建银行机构参数
func (h *BankParamHandler) CreateBankParam(c *gin.Context) {
	var req struct {
		BankType string `json:"bankType" binding:"required"`
		BankName string `json:"bankName" binding:"required"`
		BankCode string `json:"bankCode" binding:"required"`
		Reserve1 string `json:"reserve1"`
		Reserve2 string `json:"reserve2"`
		Reserve3 string `json:"reserve3"`
		Reserve4 string `json:"reserve4"`
		Reserve5 string `json:"reserve5"`
		Remark   string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	param := &models.BankParam{
		BankType:   req.BankType,
		BankName:   req.BankName,
		BankCode:   req.BankCode,
		Reserve1:   req.Reserve1,
		Reserve2:   req.Reserve2,
		Reserve3:   req.Reserve3,
		Reserve4:   req.Reserve4,
		Reserve5:   req.Reserve5,
		Remark:     req.Remark,
		CreateUser: &userID,
	}

	if err := h.bankParamService.CreateBankParam(param); err != nil {
		utils.ErrorWithCode(c, 500, "创建失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "创建成功", param)
}

// UpdateBankParam 更新银行机构参数
func (h *BankParamHandler) UpdateBankParam(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的ID")
		return
	}

	var req struct {
		BankType string `json:"bankType"`
		BankName string `json:"bankName"`
		BankCode string `json:"bankCode"`
		Reserve1 string `json:"reserve1"`
		Reserve2 string `json:"reserve2"`
		Reserve3 string `json:"reserve3"`
		Reserve4 string `json:"reserve4"`
		Reserve5 string `json:"reserve5"`
		Remark   string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	updates := map[string]interface{}{
		"bank_type":   req.BankType,
		"bank_name":   req.BankName,
		"bank_code":   req.BankCode,
		"reserve1":    req.Reserve1,
		"reserve2":    req.Reserve2,
		"reserve3":    req.Reserve3,
		"reserve4":    req.Reserve4,
		"reserve5":    req.Reserve5,
		"remark":      req.Remark,
		"update_user": userID,
	}

	if err := h.bankParamService.UpdateBankParam(id, updates); err != nil {
		utils.ErrorWithCode(c, 500, "更新失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteBankParam 删除银行机构参数
func (h *BankParamHandler) DeleteBankParam(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的ID")
		return
	}

	if err := h.bankParamService.DeleteBankParam(id); err != nil {
		utils.ErrorWithCode(c, 500, "删除失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "删除成功", nil)
}

// BatchImportParams 批量导入银行机构参数
func (h *BankParamHandler) BatchImportParams(c *gin.Context) {
	var req struct {
		Params []struct {
			BankType string `json:"bankType"`
			BankName string `json:"bankName"`
			BankCode string `json:"bankCode"`
			Reserve1 string `json:"reserve1"`
			Reserve2 string `json:"reserve2"`
			Reserve3 string `json:"reserve3"`
			Reserve4 string `json:"reserve4"`
			Reserve5 string `json:"reserve5"`
			Remark   string `json:"remark"`
		} `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	if len(req.Params) == 0 {
		utils.ErrorWithCode(c, 400, "导入数据不能为空")
		return
	}

	userID := c.GetInt64("userID")
	params := make([]models.BankParam, 0, len(req.Params))
	for _, p := range req.Params {
		// 跳过必填字段为空的行
		if p.BankType == "" || p.BankName == "" || p.BankCode == "" {
			continue
		}
		params = append(params, models.BankParam{
			BankType:   p.BankType,
			BankName:   p.BankName,
			BankCode:   p.BankCode,
			Reserve1:   p.Reserve1,
			Reserve2:   p.Reserve2,
			Reserve3:   p.Reserve3,
			Reserve4:   p.Reserve4,
			Reserve5:   p.Reserve5,
			Remark:     p.Remark,
			CreateUser: &userID,
		})
	}

	count, err := h.bankParamService.BatchImportParams(params)
	if err != nil {
		utils.ErrorWithCode(c, 500, "导入失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, fmt.Sprintf("成功导入 %d 条参数", count), map[string]interface{}{
		"count":   count,
		"skipped": len(req.Params) - count,
	})
}
