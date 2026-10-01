package handler

import (
	"fmt"
	"strconv"

	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

type ProdNodeHandler struct {
	prodNodeService *service.ProdNodeService
}

func NewProdNodeHandler() *ProdNodeHandler {
	return &ProdNodeHandler{
		prodNodeService: service.NewProdNodeService(),
	}
}

// ListProdNodes 获取节点列表
func (h *ProdNodeHandler) ListProdNodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	bankType := c.Query("bankType")
	bankName := c.Query("bankName")
	envType := c.Query("envType")
	nodeName := c.Query("nodeName")
	nodeIP := c.Query("nodeIP")
	username := c.Query("username")

	result, err := h.prodNodeService.ListProdNodes(page, size, bankType, bankName, envType, nodeName, nodeIP, username)
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询失败: "+err.Error())
		return
	}
	utils.Success(c, result)
}

// GetProdNode 获取节点详情
func (h *ProdNodeHandler) GetProdNode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的节点ID")
		return
	}

	node, err := h.prodNodeService.GetProdNode(id)
	if err != nil {
		utils.ErrorWithCode(c, 404, "节点不存在")
		return
	}
	utils.Success(c, node)
}

// CreateProdNode 创建节点
func (h *ProdNodeHandler) CreateProdNode(c *gin.Context) {
	var req struct {
		BankType      string `json:"bankType"`
		BankName      string `json:"bankName" binding:"required"`
		EnvType       string `json:"envType" binding:"required"`
		NodeName      string `json:"nodeName" binding:"required"`
		NodeIP        string `json:"nodeIp" binding:"required"`
		Username      string `json:"username"`
		IsMutualTrust string `json:"isMutualTrust"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	node := &models.ProdNode{
		BankType:      req.BankType,
		BankName:      req.BankName,
		EnvType:       req.EnvType,
		NodeName:      req.NodeName,
		NodeIP:        req.NodeIP,
		Username:      req.Username,
		IsMutualTrust: req.IsMutualTrust,
		CreateUser:    &userID,
	}

	if err := h.prodNodeService.CreateProdNode(node); err != nil {
		utils.ErrorWithCode(c, 500, "创建失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "创建成功", node)
}

// UpdateProdNode 更新节点
func (h *ProdNodeHandler) UpdateProdNode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的节点ID")
		return
	}

	var req struct {
		BankType      string `json:"bankType"`
		BankName      string `json:"bankName"`
		EnvType       string `json:"envType"`
		NodeName      string `json:"nodeName"`
		NodeIP        string `json:"nodeIp"`
		Username      string `json:"username"`
		IsMutualTrust string `json:"isMutualTrust"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	updates := map[string]interface{}{
		"bank_type":       req.BankType,
		"bank_name":       req.BankName,
		"env_type":        req.EnvType,
		"node_name":       req.NodeName,
		"node_ip":         req.NodeIP,
		"username":        req.Username,
		"is_mutual_trust": req.IsMutualTrust,
		"update_user":     userID,
	}

	if err := h.prodNodeService.UpdateProdNode(id, updates); err != nil {
		utils.ErrorWithCode(c, 500, "更新失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteProdNode 删除节点
func (h *ProdNodeHandler) DeleteProdNode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的节点ID")
		return
	}

	if err := h.prodNodeService.DeleteProdNode(id); err != nil {
		utils.ErrorWithCode(c, 500, "删除失败: "+err.Error())
		return
	}
	utils.SuccessWithMessage(c, "删除成功", nil)
}

// BatchImportNodes 批量导入节点
func (h *ProdNodeHandler) BatchImportNodes(c *gin.Context) {
	var req struct {
		Nodes []struct {
			BankType      string `json:"bankType"`
			BankName      string `json:"bankName"`
			EnvType       string `json:"envType"`
			NodeName      string `json:"nodeName"`
			NodeIP        string `json:"nodeIp"`
			Username      string `json:"username"`
			IsMutualTrust string `json:"isMutualTrust"`
		} `json:"nodes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	if len(req.Nodes) == 0 {
		utils.ErrorWithCode(c, 400, "导入数据不能为空")
		return
	}

	userID := c.GetInt64("userID")
	nodes := make([]models.ProdNode, 0, len(req.Nodes))
	for _, n := range req.Nodes {
		if n.BankName == "" || n.EnvType == "" || n.NodeName == "" || n.NodeIP == "" {
			continue
		}
		nodes = append(nodes, models.ProdNode{
			BankType:      n.BankType,
			BankName:      n.BankName,
			EnvType:       n.EnvType,
			NodeName:      n.NodeName,
			NodeIP:        n.NodeIP,
			Username:      n.Username,
			IsMutualTrust: n.IsMutualTrust,
			CreateUser:    &userID,
		})
	}

	count, err := h.prodNodeService.BatchImportNodes(nodes)
	if err != nil {
		utils.ErrorWithCode(c, 500, "导入失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, fmt.Sprintf("成功导入 %d 条节点", count), map[string]interface{}{
		"count":   count,
		"skipped": len(req.Nodes) - count,
	})
}
