package dto

import (
	"fmt"
	"time"
)

// ParseFlexibleDateTime 支持日期/日期时间多格式解析（项目模块使用，无时区串按北京时区解析）
func ParseFlexibleDateTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("时间为空")
	}
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, beijingLoc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无效的时间格式：%s", s)
}

// ProjectCreateRequest 创建项目请求（必填校验由服务层完成，便于返回友好错误）
type ProjectCreateRequest struct {
	ProjectName string `json:"projectName"`
	ProjectKey  string `json:"projectKey"`
	Description string `json:"description"`
	PlanDesc    string `json:"planDesc"`
	ManagerID   int64  `json:"managerId"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}

// ProjectUpdateRequest 编辑项目请求（字段均可选，仅更新传入项）
type ProjectUpdateRequest struct {
	ProjectName    *string `json:"projectName"`
	ProjectKey     *string `json:"projectKey"`
	Description    *string `json:"description"`
	PlanDesc       *string `json:"planDesc"`
	ManagerID      *int64  `json:"managerId"`
	StartDate      *string `json:"startDate"`
	EndDate        *string `json:"endDate"`
	RiskLevel      *string `json:"riskLevel"`
	DeliveryStatus *string `json:"deliveryStatus"`
	RiskNote       *string `json:"riskNote"`
}

// ProjectStatusRequest 项目状态流转请求（action: start/pause/resume/close/terminate）
type ProjectStatusRequest struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// ProjectRiskRequest 项目风险/交付状态变更请求（必填原因）
type ProjectRiskRequest struct {
	RiskLevel      string  `json:"riskLevel"`
	DeliveryStatus string  `json:"deliveryStatus"`
	RiskNote       *string `json:"riskNote"`
	Reason         string  `json:"reason"`
}

// ProjectMemberRequest 添加项目成员请求
type ProjectMemberRequest struct {
	UserID     int64  `json:"userId"`
	MemberRole string `json:"memberRole"`
}

// StageCreateRequest 新增项目阶段请求（必填校验由服务层完成，对齐项目先例）
type StageCreateRequest struct {
	StageName   string `json:"stageName"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Description string `json:"description"`
}

// StageUpdateRequest 编辑项目阶段请求（字段均可选，仅更新传入项）
type StageUpdateRequest struct {
	StageName   *string `json:"stageName"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
	Description *string `json:"description"`
}

// WeekCreateRequest 新增周计划请求
type WeekCreateRequest struct {
	Title     string `json:"title"`
	PlanDesc  string `json:"planDesc"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	OwnerID   int64  `json:"ownerId"`
	Remark    string `json:"remark"`
}

// WeekUpdateRequest 编辑周计划请求（字段均可选）
type WeekUpdateRequest struct {
	Title     *string `json:"title"`
	PlanDesc  *string `json:"planDesc"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
	OwnerID   *int64  `json:"ownerId"`
	Remark    *string `json:"remark"`
}

// WeekStatusRequest 周状态流转请求（action: start/done/skip/reopen；也支持直接传 status）
type WeekStatusRequest struct {
	Action string `json:"action"`
	Status string `json:"status"`
}

// WeekSummaryContent 周总结结构化内容
type WeekSummaryContent struct {
	Finished string `json:"finished"`
	NextPlan string `json:"nextPlan"`
	Risks    string `json:"risks"`
	Support  string `json:"support"`
}

// WeekSummaryRequest 提交周总结请求（支持 { summary: {...} } 或平铺两种格式）
type WeekSummaryRequest struct {
	Summary  *WeekSummaryContent `json:"summary"`
	Finished string              `json:"finished"`
	NextPlan string              `json:"nextPlan"`
	Risks    string              `json:"risks"`
	Support  string              `json:"support"`
}

// WeekTaskCreateRequest 周内新建任务请求（创建事项 + 建关联）
type WeekTaskCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	UserID      *int64 `json:"userId"`
	Priority    string `json:"priority"`
	Deadline    string `json:"deadline"`
	Category    string `json:"category"`
}

// TaskMoveRequest 任务移动到其他周请求
type TaskMoveRequest struct {
	WeekPlanID int64 `json:"weekPlanId"`
}
