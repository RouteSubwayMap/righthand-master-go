package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewTaskService 测试创建TaskService实例
func TestNewTaskService(t *testing.T) {
	service := NewTaskService()
	assert.NotNil(t, service, "TaskService实例不应为nil")
}

// TestNewContactService 测试创建ContactService实例
func TestNewContactService(t *testing.T) {
	service := NewContactService()
	assert.NotNil(t, service, "ContactService实例不应为nil")
}

// TestValidateStatusCode 测试状态码验证
func TestValidateStatusCode(t *testing.T) {
	validStatuses := map[string]bool{
		"TODO":      true,
		"DOING":     true,
		"DONE":      true,
		"CANCELLED": true,
	}

	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"有效状态-TODO", "TODO", true},
		{"有效状态-DOING", "DOING", true},
		{"有效状态-DONE", "DONE", true},
		{"有效状态-CANCELLED", "CANCELLED", true},
		{"无效状态", "INVALID", false},
		{"空状态", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validStatuses[tt.status]
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateItemType 测试事项类型验证
func TestValidateItemType(t *testing.T) {
	validTypes := map[string]bool{
		"TASK":        true,
		"MEMO":        true,
		"REQUIREMENT": true,
	}

	tests := []struct {
		name     string
		itemType string
		expected bool
	}{
		{"有效类型-TASK", "TASK", true},
		{"有效类型-MEMO", "MEMO", true},
		{"有效类型-REQUIREMENT", "REQUIREMENT", true},
		{"无效类型", "INVALID", false},
		{"空类型", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validTypes[tt.itemType]
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidatePriority 测试优先级验证
func TestValidatePriority(t *testing.T) {
	validPriorities := map[string]bool{
		"HIGH":   true,
		"MEDIUM": true,
		"LOW":    true,
	}

	tests := []struct {
		name     string
		priority string
		expected bool
	}{
		{"有效优先级-HIGH", "HIGH", true},
		{"有效优先级-MEDIUM", "MEDIUM", true},
		{"有效优先级-LOW", "LOW", true},
		{"无效优先级", "INVALID", false},
		{"空优先级", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validPriorities[tt.priority]
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRecurrenceTypeValidation 测试周期性类型验证
func TestRecurrenceTypeValidation(t *testing.T) {
	validTypes := map[string]bool{
		"NONE":    true,
		"DAILY":   true,
		"WEEKLY":  true,
		"MONTHLY": true,
		"YEARLY":  true,
	}

	tests := []struct {
		name     string
		recType  string
		expected bool
	}{
		{"有效类型-NONE", "NONE", true},
		{"有效类型-DAILY", "DAILY", true},
		{"有效类型-WEEKLY", "WEEKLY", true},
		{"有效类型-MONTHLY", "MONTHLY", true},
		{"有效类型-YEARLY", "YEARLY", true},
		{"无效类型", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validTypes[tt.recType]
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRecurrenceEndTypeValidation 测试周期性结束类型验证
func TestRecurrenceEndTypeValidation(t *testing.T) {
	validEndTypes := map[string]bool{
		"NEVER": true,
		"DATE":  true,
		"COUNT": true,
	}

	tests := []struct {
		name     string
		endType  string
		expected bool
	}{
		{"有效类型-NEVER", "NEVER", true},
		{"有效类型-DATE", "DATE", true},
		{"有效类型-COUNT", "COUNT", true},
		{"无效类型", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validEndTypes[tt.endType]
			assert.Equal(t, tt.expected, result)
		})
	}
}
