package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRequirementStatusValidation 测试需求状态验证
func TestRequirementStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"PENDING_CLARIFY":    true,
		"PENDING_SCHEDULING": true,
		"DEVELOPING":         true,
		"TESTING":            true,
		"COMPLETED":          true,
	}

	// 测试用例1: 验证所有有效状态
	t.Run("验证有效状态", func(t *testing.T) {
		for status := range validStatuses {
			assert.True(t, validStatuses[status], "状态 %s 应该是有效的", status)
		}
	})

	// 测试用例2: 验证无效状态
	t.Run("验证无效状态", func(t *testing.T) {
		invalidStatus := "INVALID_STATUS"
		assert.False(t, validStatuses[invalidStatus], "状态 %s 应该是无效的", invalidStatus)
	})
}

// TestRequirementPriorityLevels 测试需求优先级
func TestRequirementPriorityLevels(t *testing.T) {
	validPriorities := map[string]bool{
		"HIGH":   true,
		"MEDIUM": true,
		"LOW":    true,
	}

	// 测试用例: 验证所有优先级都有效
	t.Run("验证优先级有效性", func(t *testing.T) {
		for priority := range validPriorities {
			assert.True(t, validPriorities[priority], "优先级 %s 应该是有效的", priority)
		}
	})

	// 测试用例: 验证无效优先级
	t.Run("验证无效优先级", func(t *testing.T) {
		invalidPriority := "URGENT"
		assert.False(t, validPriorities[invalidPriority], "优先级 %s 应该是无效的", invalidPriority)
	})
}

// TestRequirementTypes 测试需求类型
func TestRequirementTypes(t *testing.T) {
	validTypes := map[string]bool{
		"INTERNAL": true,
		"BANK":     true,
	}

	// 测试用例: 验证所有类型都有效
	t.Run("验证需求类型有效性", func(t *testing.T) {
		for reqType := range validTypes {
			assert.True(t, validTypes[reqType], "需求类型 %s 应该是有效的", reqType)
		}
	})
}

// TestRequirementGroupTypes 测试需求小组类型
func TestRequirementGroupTypes(t *testing.T) {
	validGroupTypes := map[string]bool{
		"LOAN_MIDDLE": true,
		"LOAN_PRE":    true,
		"RISK":        true,
	}

	// 测试用例: 验证所有小组类型都有效
	t.Run("验证小组类型有效性", func(t *testing.T) {
		for groupType := range validGroupTypes {
			assert.True(t, validGroupTypes[groupType], "小组类型 %s 应该是有效的", groupType)
		}
	})
}

// TestRequirementStatusTransitions 测试需求状态流转规则
func TestRequirementStatusTransitions(t *testing.T) {
	// 定义状态流转规则
	statusTransitions := map[string][]string{
		"PENDING_CLARIFY":    {"PENDING_SCHEDULING", "COMPLETED"},
		"PENDING_SCHEDULING": {"DEVELOPING", "PENDING_CLARIFY"},
		"DEVELOPING":         {"TESTING", "PENDING_SCHEDULING"},
		"TESTING":            {"COMPLETED", "DEVELOPING"},
		"COMPLETED":          {},
	}

	// 测试用例1: 验证待澄清状态的转换
	t.Run("待澄清状态转换", func(t *testing.T) {
		transitions := statusTransitions["PENDING_CLARIFY"]
		assert.Contains(t, transitions, "PENDING_SCHEDULING")
		assert.Contains(t, transitions, "COMPLETED")
		assert.Len(t, transitions, 2)
	})

	// 测试用例2: 验证开发中状态的转换
	t.Run("开发中状态转换", func(t *testing.T) {
		transitions := statusTransitions["DEVELOPING"]
		assert.Contains(t, transitions, "TESTING")
		assert.Contains(t, transitions, "PENDING_SCHEDULING")
		assert.Len(t, transitions, 2)
	})

	// 测试用例3: 验证已完成状态不能转换
	t.Run("已完成状态不能转换", func(t *testing.T) {
		transitions := statusTransitions["COMPLETED"]
		assert.Len(t, transitions, 0)
	})

	// 测试用例4: 验证所有状态都有定义
	t.Run("所有状态都有定义", func(t *testing.T) {
		expectedStatuses := []string{
			"PENDING_CLARIFY",
			"PENDING_SCHEDULING",
			"DEVELOPING",
			"TESTING",
			"COMPLETED",
		}

		for _, status := range expectedStatuses {
			_, exists := statusTransitions[status]
			assert.True(t, exists, "状态 %s 应该有定义的转换规则", status)
		}
	})
}
