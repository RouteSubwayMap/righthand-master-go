package dto

import (
	"time"
)

// RequirementRequest 需求请求 DTO
type RequirementRequest struct {
	Title             *string  `json:"title"`
	BankName          *string  `json:"bankName"`
	RequirementNo     *string  `json:"requirementNo"`
	Priority          *string  `json:"priority"`
	RiskLevel         *string  `json:"riskLevel"`
	RequirementType   *string  `json:"requirementType"`
	Status            *string  `json:"status"`
	ScheduleDate      *string  `json:"scheduleDate"`
	MainSystem        *string  `json:"mainSystem"`
	Developer         *string  `json:"developer"`
	InvolvedSystems   *string  `json:"involvedSystems"`
	FollowerID        *int64   `json:"followerId"`
	FollowerName      *string  `json:"followerName"`
	SubmitterID       *int64   `json:"submitterId"`
	SubmitterName     *string  `json:"submitterName"`
	SubmitTime        *string  `json:"submitTime"`
	CompletionDate    *string  `json:"completionDate"`
	EstimatedWorkload *float64 `json:"estimatedWorkload"`
	GroupType         *string  `json:"groupType"`
	Description       *string  `json:"description"`
	Remark            *string  `json:"remark"`
}

// ParseDateTime 解析日期时间字符串，始终按北京时区(UTC+8)解析，
// 与 CustomTime.UnmarshalJSON 保持一致，避免服务器 OS 时区为 UTC 时入库多 8 小时。
func ParseDateTime(dateStr *string) *time.Time {
	if dateStr == nil || *dateStr == "" {
		return nil
	}

	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, *dateStr, beijingLoc); err == nil {
			return &t
		}
	}

	return nil
}
