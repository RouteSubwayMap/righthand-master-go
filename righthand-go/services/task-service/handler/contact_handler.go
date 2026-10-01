package handler

import (
	"strconv"

	"righthand-go/internal/models"
	"righthand-go/internal/utils"
	"righthand-go/services/task-service/service"

	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	contactService *service.ContactService
}

func NewContactHandler() *ContactHandler {
	return &ContactHandler{
		contactService: service.NewContactService(),
	}
}

// ListContacts 获取联系人列表
func (h *ContactHandler) ListContacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	group := c.Query("group")
	position := c.Query("position")

	userID := c.GetInt64("userID")
	result, err := h.contactService.ListContacts(userID, page, size, keyword, group, position)
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}

// GetContact 获取联系人详情
func (h *ContactHandler) GetContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的联系人ID")
		return
	}

	userID := c.GetInt64("userID")
	contact, err := h.contactService.GetContact(id, userID)
	if err != nil {
		utils.ErrorWithCode(c, 404, "联系人不存在")
		return
	}

	utils.Success(c, contact)
}

// CreateContact 创建联系人
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Gender       *int   `json:"gender"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		ContactGroup string `json:"contactGroup"`
		Company      string `json:"company"`
		Position     string `json:"position"`
		Remark       string `json:"remark"`
		Avatar       string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	contact := &models.Contact{
		UserID:       userID,
		Name:         req.Name,
		Gender:       req.Gender,
		Phone:        req.Phone,
		Email:        req.Email,
		ContactGroup: req.ContactGroup,
		Company:      req.Company,
		Position:     req.Position,
		Remark:       req.Remark,
		Avatar:       req.Avatar,
		CreateUser:   &userID,
	}

	if err := h.contactService.CreateContact(contact); err != nil {
		utils.ErrorWithCode(c, 500, "创建失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "创建成功", contact)
}

// UpdateContact 更新联系人
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的联系人ID")
		return
	}

	var req struct {
		Name         string `json:"name"`
		Gender       *int   `json:"gender"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		ContactGroup string `json:"contactGroup"`
		Company      string `json:"company"`
		Position     string `json:"position"`
		Remark       string `json:"remark"`
		Avatar       string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, 400, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("userID")
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.ContactGroup != "" {
		updates["contact_group"] = req.ContactGroup
	}
	if req.Company != "" {
		updates["company"] = req.Company
	}
	if req.Position != "" {
		updates["position"] = req.Position
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	updates["update_user"] = userID

	if err := h.contactService.UpdateContact(id, userID, updates); err != nil {
		utils.ErrorWithCode(c, 500, "更新失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteContact 删除联系人
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, 400, "无效的联系人ID")
		return
	}

	userID := c.GetInt64("userID")
	if err := h.contactService.DeleteContact(id, userID); err != nil {
		utils.ErrorWithCode(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GetContactGroups 获取联系人分组列表
func (h *ContactHandler) GetContactGroups(c *gin.Context) {
	userID := c.GetInt64("userID")
	groups, err := h.contactService.GetContactGroups(userID)
	if err != nil {
		utils.ErrorWithCode(c, 500, "查询分组失败: "+err.Error())
		return
	}

	utils.Success(c, groups)
}
