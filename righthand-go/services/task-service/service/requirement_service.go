package service

import (
	"database/sql"
	"errors"
	"fmt"
	"righthand-go/internal/database"
	"righthand-go/internal/models"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RequirementService 需求管理服务
type RequirementService struct {
	db *gorm.DB // 测试时注入,否则使用全局DB
}

// NewRequirementService 创建需求服务实例
func NewRequirementService() *RequirementService {
	return &RequirementService{}
}

// getDB 获取数据库连接
func (s *RequirementService) getDB() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return database.GetDB()
}

// CheckRequirementNoExists 检查需求编号是否已存在
func (s *RequirementService) CheckRequirementNoExists(requirementNo string) (bool, error) {
	db := s.getDB()
	var count int64
	err := db.Model(&models.Requirement{}).Where("requirement_no = ? AND deleted = ?", requirementNo, 0).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateRequirement 创建需求
func (s *RequirementService) CreateRequirement(req *models.Requirement) error {
	db := s.getDB()

	// 校验需求编号是否已存在
	if req.RequirementNo != nil && strings.TrimSpace(*req.RequirementNo) != "" {
		exists, err := s.CheckRequirementNoExists(*req.RequirementNo)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("需求编号 %s 已存在", *req.RequirementNo)
		}
	}

	// 设置默认值
	if req.Priority == nil {
		priority := "MEDIUM"
		req.Priority = &priority
	}
	if req.RiskLevel == nil {
		riskLevel := "MEDIUM"
		req.RiskLevel = &riskLevel
	}
	if req.RequirementType == nil {
		reqType := "INTERNAL"
		req.RequirementType = &reqType
	}
	if req.Status == nil {
		status := "PENDING_CLARIFY"
		req.Status = &status
	}
	if req.Deleted == nil {
		deleted := 0
		req.Deleted = &deleted
	}

	return db.Create(req).Error
}

// GetRequirementByID 根据ID获取需求
// 需求为团队共享视图：所有登录用户均可查看任意需求详情（软删除数据除外）
func (s *RequirementService) GetRequirementByID(id int64) (*models.Requirement, error) {
	db := s.getDB()
	var requirement models.Requirement

	err := db.Where("id = ? AND deleted = ?", id, 0).First(&requirement).Error
	if err != nil {
		return nil, err
	}

	return &requirement, nil
}

// UpdateRequirement 更新需求
// 管理员可更新任意需求；普通用户（需求创建人或需求跟进人）可更新自己的或跟进的需求
func (s *RequirementService) UpdateRequirement(req *models.Requirement, isAdmin bool) error {
	db := s.getDB()

	// 检查需求是否存在(管理员可更新任意需求,普通用户仅可更新创建人或跟进人为自己的)
	var existing models.Requirement
	query := db.Where("id = ? AND deleted = ?", req.ID, 0)
	if !isAdmin {
		query = query.Where("user_id = ? OR follower_id = ?", req.UserID, req.UserID)
	}
	err := query.First(&existing).Error
	if err != nil {
		return err
	}

	// 使用 map 更新，确保 nil/零值字段（如清空的需求编号）也能正确写入数据库
	updates := map[string]interface{}{
		"title":              req.Title,
		"bank_name":          req.BankName,
		"requirement_no":     req.RequirementNo,
		"priority":           req.Priority,
		"risk_level":         req.RiskLevel,
		"requirement_type":   req.RequirementType,
		"status":             req.Status,
		"schedule_date":      req.ScheduleDate,
		"main_system":        req.MainSystem,
		"developer":          req.Developer,
		"involved_systems":   req.InvolvedSystems,
		"follower_id":        req.FollowerID,
		"follower_name":      req.FollowerName,
		"submitter_id":       req.SubmitterID,
		"submitter_name":     req.SubmitterName,
		"submit_time":        req.SubmitTime,
		"completion_date":    req.CompletionDate,
		"estimated_workload": req.EstimatedWorkload,
		"group_type":         req.GroupType,
		"description":        req.Description,
		"remark":             req.Remark,
		"update_user":        req.UpdateUser,
	}

	return db.Model(&models.Requirement{}).Where("id = ?", req.ID).Updates(updates).Error
}

// DeleteRequirement 删除需求(逻辑删除)
func (s *RequirementService) DeleteRequirement(id int64, userID int64, isAdmin bool) error {
	db := s.getDB()

	deleted := 1
	query := db.Model(&models.Requirement{}).Where("id = ? AND deleted = ?", id, 0)
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}
	result := query.Update("deleted", deleted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到可删除的需求记录")
	}
	return nil
}

// ListRequirements 查询需求列表(支持分页和筛选)
func (s *RequirementService) ListRequirements(userID int64, isAdmin bool, page, pageSize int, status, priority, bankName, groupType, title, requirementNo, riskLevel, requirementType, followerName *string, createdBy *int64, completionDateStart, completionDateEnd *string) ([]models.Requirement, int64, error) {
	db := s.getDB()
	var requirements []models.Requirement
	var total int64

	query := db.Model(&models.Requirement{}).Where("deleted = ?", 0)

	// 状态筛选
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	// 优先级筛选
	if priority != nil && *priority != "" {
		query = query.Where("priority = ?", *priority)
	}

	// 银行筛选
	if bankName != nil && *bankName != "" {
		query = query.Where("bank_name LIKE ?", "%"+*bankName+"%")
	}

	// 小组类型筛选
	if groupType != nil && *groupType != "" {
		query = query.Where("group_type = ?", *groupType)
	}

	// 需求名称筛选
	if title != nil && *title != "" {
		query = query.Where("title LIKE ?", "%"+*title+"%")
	}

	// 需求编号筛选
	if requirementNo != nil && *requirementNo != "" {
		query = query.Where("requirement_no LIKE ?", "%"+*requirementNo+"%")
	}

	// 风险等级筛选
	if riskLevel != nil && *riskLevel != "" {
		query = query.Where("risk_level = ?", *riskLevel)
	}

	// 需求类型筛选
	if requirementType != nil && *requirementType != "" {
		query = query.Where("requirement_type = ?", *requirementType)
	}

	// 需求跟进人筛选
	if followerName != nil && *followerName != "" {
		query = query.Where("follower_name LIKE ?", "%"+*followerName+"%")
	}

	// 创建人筛选
	if createdBy != nil && *createdBy > 0 {
		query = query.Where("user_id = ?", *createdBy)
	}

	// 需求完成时间筛选
	if completionDateStart != nil && *completionDateStart != "" {
		query = query.Where("completion_date >= ?", *completionDateStart)
	}
	if completionDateEnd != nil && *completionDateEnd != "" {
		query = query.Where("completion_date <= ?", *completionDateEnd+" 23:59:59")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&requirements).Error
	if err != nil {
		return nil, 0, err
	}

	return requirements, total, nil
}

// UpdateStatus 更新需求状态
func (s *RequirementService) UpdateStatus(id int64, userID int64, newStatus string) error {
	db := s.getDB()

	// 验证状态值
	validStatuses := map[string]bool{
		"PENDING_CLARIFY":    true,
		"PENDING_SCHEDULING": true,
		"DEVELOPING":         true,
		"TESTING":            true,
		"COMPLETED":          true,
	}

	if !validStatuses[newStatus] {
		return gorm.ErrInvalidData
	}

	// 如果状态变为已完成,自动设置完成时间
	updates := map[string]interface{}{
		"status": newStatus,
	}

	if newStatus == "COMPLETED" {
		now := s.getDB().NowFunc()
		updates["completion_date"] = now
	}

	return db.Model(&models.Requirement{}).Where("id = ? AND deleted = ?", id, 0).Updates(updates).Error
}

// ExportRequirements 导出需求列表(返回所有数据)
func (s *RequirementService) ExportRequirements(userID int64, isAdmin bool, status, priority, bankName, groupType, title, requirementNo, riskLevel, requirementType, followerName *string, createdBy *int64, completionDateStart, completionDateEnd *string) ([]models.Requirement, error) {
	db := s.getDB()
	var requirements []models.Requirement

	query := db.Model(&models.Requirement{}).Where("deleted = ?", 0)

	// 状态筛选
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	// 优先级筛选
	if priority != nil && *priority != "" {
		query = query.Where("priority = ?", *priority)
	}

	// 银行筛选
	if bankName != nil && *bankName != "" {
		query = query.Where("bank_name LIKE ?", "%"+*bankName+"%")
	}

	// 小组类型筛选
	if groupType != nil && *groupType != "" {
		query = query.Where("group_type = ?", *groupType)
	}

	// 标题模糊查询
	if title != nil && *title != "" {
		query = query.Where("title LIKE ?", "%"+*title+"%")
	}

	// 需求编号模糊查询
	if requirementNo != nil && *requirementNo != "" {
		query = query.Where("requirement_no LIKE ?", "%"+*requirementNo+"%")
	}

	// 风险等级筛选
	if riskLevel != nil && *riskLevel != "" {
		query = query.Where("risk_level = ?", *riskLevel)
	}

	// 需求类型筛选
	if requirementType != nil && *requirementType != "" {
		query = query.Where("requirement_type = ?", *requirementType)
	}

	// 跟进人模糊查询
	if followerName != nil && *followerName != "" {
		query = query.Where("follower_name LIKE ?", "%"+*followerName+"%")
	}

	// 创建人筛选
	if createdBy != nil && *createdBy > 0 {
		query = query.Where("user_id = ?", *createdBy)
	}

	// 需求完成时间筛选
	if completionDateStart != nil && *completionDateStart != "" {
		query = query.Where("completion_date >= ?", *completionDateStart)
	}
	if completionDateEnd != nil && *completionDateEnd != "" {
		query = query.Where("completion_date <= ?", *completionDateEnd+" 23:59:59")
	}

	err := query.Order("create_time DESC").Find(&requirements).Error
	if err != nil {
		return nil, err
	}

	return requirements, nil
}

// ImportRequirements 批量导入需求(从CSV)
func (s *RequirementService) ImportRequirements(userID int64, csvContent []byte) (map[string]interface{}, error) {
	db := s.getDB()

	// 解析CSV内容
	lines := strings.Split(string(csvContent), "\n")
	if len(lines) < 2 {
		return nil, errors.New("CSV文件格式错误")
	}

	// 跳过表头
	dataLines := lines[1:]

	successCount := 0
	failCount := 0
	var errors []string

	for i, line := range dataLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析CSV行(简单分割,不处理引号内的逗号)
		fields := strings.Split(line, ",")
		if len(fields) < 15 {
			failCount++
			errors = append(errors, fmt.Sprintf("第%d行: 字段数量不足", i+2))
			continue
		}

		// 创建需求对象
		req := &models.Requirement{
			UserID:            userID,
			Title:             fields[0],
			BankName:          stringPtrOrNil(fields[1]),
			RequirementNo:     stringPtrOrNil(fields[2]),
			Priority:          stringPtrOrNil(fields[3]),
			RiskLevel:         stringPtrOrNil(fields[4]),
			RequirementType:   stringPtrOrNil(fields[5]),
			Status:            stringPtr("PENDING_CLARIFY"), // 默认状态
			MainSystem:        stringPtrOrNil(fields[6]),
			Developer:         stringPtrOrNil(fields[7]),
			InvolvedSystems:   stringPtrOrNil(fields[8]),
			FollowerName:      stringPtrOrNil(fields[9]),
			SubmitterName:     stringPtrOrNil(fields[10]),
			SubmitTime:        func() *time.Time { t := time.Now(); return &t }(),
			EstimatedWorkload: parseNullFloat64(fields[11]),
			GroupType:         stringPtrOrNil(fields[12]),
			Description:       stringPtrOrNil(fields[13]),
			Remark:            stringPtrOrNil(fields[14]),
			Deleted:           func() *int { d := 0; return &d }(),
			CreateUser:        &userID,
		}

		// 设置默认值
		if req.Priority == nil || *req.Priority == "" {
			priority := "MEDIUM"
			req.Priority = &priority
		}
		if req.RiskLevel == nil || *req.RiskLevel == "" {
			riskLevel := "MEDIUM"
			req.RiskLevel = &riskLevel
		}
		if req.RequirementType == nil || *req.RequirementType == "" {
			reqType := "INTERNAL"
			req.RequirementType = &reqType
		}

		// 保存到数据库
		if err := db.Create(req).Error; err != nil {
			failCount++
			errors = append(errors, fmt.Sprintf("第%d行: %s", i+2, err.Error()))
		} else {
			successCount++
		}
	}

	result := map[string]interface{}{
		"successCount": successCount,
		"failCount":    failCount,
		"totalCount":   len(dataLines),
		"errors":       errors,
	}

	return result, nil
}

// ImportRequirementsFromJSON 从JSON数组导入需求
func (s *RequirementService) ImportRequirementsFromJSON(userID int64, items []struct {
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
}) (map[string]interface{}, error) {
	db := s.getDB()

	successCount := 0
	failCount := 0
	var errors []string

	now := time.Now()
	deleted := 0

	for i, item := range items {
		if strings.TrimSpace(item.Title) == "" {
			failCount++
			errors = append(errors, fmt.Sprintf("第%d行: 需求名称不能为空", i+2))
			continue
		}

		// 校验需求编号是否已存在
		trimmedNo := strings.TrimSpace(item.RequirementNo)
		if trimmedNo != "" {
			exists, err := s.CheckRequirementNoExists(trimmedNo)
			if err != nil {
				failCount++
				errors = append(errors, fmt.Sprintf("第%d行: 校验需求编号失败: %s", i+2, err.Error()))
				continue
			}
			if exists {
				failCount++
				errors = append(errors, fmt.Sprintf("第%d行: 需求编号 %s 已存在", i+2, trimmedNo))
				continue
			}
		}

		req := &models.Requirement{
			UserID:            userID,
			Title:             item.Title,
			BankName:          stringPtrOrNil(item.BankName),
			RequirementNo:     stringPtrOrNil(item.RequirementNo),
			Priority:          stringPtrOrNil(item.Priority),
			RiskLevel:         stringPtrOrNil(item.RiskLevel),
			RequirementType:   stringPtrOrNil(item.RequirementType),
			Status:            stringPtr("PENDING_CLARIFY"),
			MainSystem:        stringPtrOrNil(item.MainSystem),
			Developer:         stringPtrOrNil(item.Developer),
			InvolvedSystems:   stringPtrOrNil(item.InvolvedSystems),
			FollowerName:      stringPtrOrNil(item.FollowerName),
			SubmitterName:     stringPtrOrNil(item.SubmitterName),
			SubmitTime:        &now,
			EstimatedWorkload: sql.NullFloat64{Float64: item.EstimatedWorkload, Valid: item.EstimatedWorkload > 0},
			GroupType:         stringPtrOrNil(item.GroupType),
			Description:       stringPtrOrNil(item.Description),
			Remark:            stringPtrOrNil(item.Remark),
			Deleted:           &deleted,
			CreateUser:        &userID,
		}

		// 设置默认值
		if req.Priority == nil || *req.Priority == "" {
			priority := "MEDIUM"
			req.Priority = &priority
		}
		if req.RiskLevel == nil || *req.RiskLevel == "" {
			riskLevel := "MEDIUM"
			req.RiskLevel = &riskLevel
		}
		if req.RequirementType == nil || *req.RequirementType == "" {
			reqType := "INTERNAL"
			req.RequirementType = &reqType
		}

		if err := db.Create(req).Error; err != nil {
			failCount++
			errors = append(errors, fmt.Sprintf("第%d行: %s", i+2, err.Error()))
		} else {
			successCount++
		}
	}

	result := map[string]interface{}{
		"successCount": successCount,
		"failCount":    failCount,
		"totalCount":   len(items),
		"errors":       errors,
	}

	return result, nil
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func stringPtrOrNil(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func parseNullFloat64(s string) sql.NullFloat64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullFloat64{Valid: false}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}
