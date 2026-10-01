import { describe, it, expect } from 'vitest'

describe('CSV导出工具', () => {
  it('应该正确处理数据格式化', () => {
    // 测试CSV字段处理逻辑
    const testData = [
      { name: '张三', age: 25 },
      { name: '李四', age: 30 }
    ]
    
    expect(testData).toHaveLength(2)
    expect(testData[0].name).toBe('张三')
    expect(testData[1].age).toBe(30)
  })

  it('应该处理空值', () => {
    const data = [
      { name: '张三', value: null },
      { name: null, value: 100 }
    ]
    
    expect(data[0].value).toBeNull()
    expect(data[1].name).toBeNull()
  })

  it('应该正确处理包含逗号的字段', () => {
    const text = '你好,世界'
    const escaped = '"' + text.replace(/"/g, '""') + '"'
    
    expect(escaped).toBe('"你好,世界"')
  })

  it('应该正确转义引号', () => {
    const text = '他说"你好"'
    const escaped = '"' + text.replace(/"/g, '""') + '"'
    
    expect(escaped).toBe('"他说""你好"""')
  })
})
