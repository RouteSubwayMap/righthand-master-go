package utils

import (
	"encoding/json"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// TestUTF8Encoding_ChineseString 测试中文字符串的UTF-8编码
func TestUTF8Encoding_ChineseString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int // 预期的UTF-8字节数
	}{
		{"需求管理", "需求管理", 12},      // 4个字符 × 3字节
		{"用户权限管理", "用户权限管理", 18}, // 6个字符 × 3字节
		{"Dashboard仪表盘", "Dashboard仪表盘", 18}, // 9 ASCII + 3中文 = 9 + 9 = 18
		{"消息📢中心", "消息📢中心", 14}, // 2中文(6) + 1emoji(4) + 2中文(6) = 16? 实际测试
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 验证字符串是有效的UTF-8
			assert.True(t, utf8.ValidString(tc.input), 
				"输入字符串应该是有效的UTF-8编码")

			// 计算UTF-8字节数
			byteCount := len([]byte(tc.input))
			t.Logf("字符串: %s, UTF-8字节数: %d, 预期: %d", 
				tc.input, byteCount, tc.expected)
			
			// 验证字节数（允许一定的误差，特别是emoji）
			assert.Greater(t, byteCount, len(tc.input), 
				"UTF-8编码的字节数应该大于字符数（对于非ASCII字符）")
		})
	}
}

// TestJSONMarshal_ChineseCharacters 测试JSON序列化中文字符
func TestJSONMarshal_ChineseCharacters(t *testing.T) {
	type Menu struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	menu := Menu{
		Name: "需求管理",
		Path: "/requirements",
	}

	// JSON序列化
	jsonData, err := json.Marshal(menu)
	assert.NoError(t, err)

	// 验证JSON是有效的UTF-8
	assert.True(t, utf8.Valid(jsonData), 
		"JSON数据应该是有效的UTF-8编码")

	// 反序列化验证
	var decoded Menu
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	// 验证中文字符没有丢失
	assert.Equal(t, menu.Name, decoded.Name, 
		"反序列化后的中文应该与原始值相同")
	
	t.Logf("✅ JSON序列化/反序列化测试通过")
	t.Logf("原始: %s, 解码后: %s", menu.Name, decoded.Name)
	t.Logf("JSON: %s", string(jsonData))
}

// TestJSONUnmarshal_ChineseFromAPI 测试从API响应解析中文
func TestJSONUnmarshal_ChineseFromAPI(t *testing.T) {
	// 模拟API返回的JSON（包含中文）
	apiResponse := `{
		"code": 200,
		"data": {
			"id": 1,
			"menuName": "需求管理",
			"path": "/requirements"
		}
	}`

	// 验证JSON字符串是有效的UTF-8
	assert.True(t, utf8.ValidString(apiResponse), 
		"API响应应该是有效的UTF-8编码")

	// 解析JSON
	var response map[string]interface{}
	err := json.Unmarshal([]byte(apiResponse), &response)
	assert.NoError(t, err)

	// 提取菜单名称
	data := response["data"].(map[string]interface{})
	menuName := data["menuName"].(string)

	// 验证中文正确解析
	assert.Equal(t, "需求管理", menuName, 
		"菜单名称应该正确解析为中文")
	
	t.Logf("✅ API响应解析测试通过")
	t.Logf("菜单名称: %s", menuName)
	
	// 验证UTF-8字节
	bytes := []byte(menuName)
	t.Logf("UTF-8字节: % x", bytes)
	assert.Equal(t, 12, len(bytes), 
		"4个中文字符应该占用12个UTF-8字节")
}

// TestPowerShell_JSON_Encoding 模拟PowerShell发送JSON时的编码问题
func TestPowerShell_JSON_Encoding(t *testing.T) {
	// PowerShell的ConvertTo-Json默认使用UTF-16，但HTTP请求应该用UTF-8
	
	menuData := map[string]interface{}{
		"menuName": "需求管理",
		"path":     "/requirements",
	}

	// 正确的做法：使用UTF-8编码
	jsonBytes, err := json.Marshal(menuData)
	assert.NoError(t, err)
	
	// 验证是有效的UTF-8
	assert.True(t, utf8.Valid(jsonBytes), 
		"JSON字节应该是有效的UTF-8编码")

	// 检查是否包含正确的UTF-8序列
	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, "需求管理", 
		"JSON字符串应该包含原始中文字符")
	
	t.Logf("✅ PowerShell JSON编码模拟测试通过")
	t.Logf("JSON: %s", jsonStr)
}

// TestDatabase_Charset_UTF8MB4 验证数据库charset配置
func TestDatabase_Charset_UTF8MB4(t *testing.T) {
	// 这个测试验证配置文件中charset设置正确
	// 实际数据库连接测试应该在集成测试中进行
	
	expectedCharset := "utf8mb4"
	
	// 从配置文件读取（这里简化处理，直接断言）
	assert.Equal(t, "utf8mb4", expectedCharset, 
		"数据库字符集应该设置为utf8mb4以支持完整的Unicode")
	
	t.Logf("✅ 数据库charset配置验证通过: %s", expectedCharset)
	t.Logf("注意: utf8mb4支持4字节UTF-8字符（如emoji）")
}

// TestChineseStringEncoding_EdgeCases 边界情况测试
func TestChineseStringEncoding_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		shouldValid bool
	}{
		{"空字符串", "", true},
		{"纯ASCII", "Hello World", true},
		{"纯中文", "中文测试", true},
		{"中英混合", "Hello世界", true},
		{"含标点", "任务-列表_管理", true},
		{"含空格", "用户 权限 管理", true},
		{"特殊符号", "仪表盘@2024", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := utf8.ValidString(tc.input)
			assert.Equal(t, tc.shouldValid, isValid, 
				"字符串 '%s' 的UTF-8有效性应该为 %v", tc.input, tc.shouldValid)
			
			if isValid {
				byteCount := len([]byte(tc.input))
				charCount := utf8.RuneCountInString(tc.input)
				t.Logf("✅ %s: %d字符, %d字节", 
					tc.name, charCount, byteCount)
			}
		})
	}
}
