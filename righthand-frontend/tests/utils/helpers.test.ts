import { describe, it, expect } from 'vitest'

describe('日期格式化工具', () => {
  it('应该正确格式化日期为YYYY-MM-DD', () => {
    const date = new Date('2024-01-15')
    const formatted = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
    
    expect(formatted).toBe('2024-01-15')
  })

  it('应该正确处理月份边界', () => {
    const date = new Date('2024-12-31')
    const formatted = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
    
    expect(formatted).toBe('2024-12-31')
  })

  it('应该正确处理月初', () => {
    const date = new Date('2024-01-01')
    const formatted = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
    
    expect(formatted).toBe('2024-01-01')
  })
})

describe('数组工具函数', () => {
  it('应该正确过滤空值', () => {
    const arr = [1, null, 2, undefined, 3, '', 4]
    const filtered = arr.filter(item => item !== null && item !== undefined && item !== '')
    
    expect(filtered).toEqual([1, 2, 3, 4])
  })

  it('应该正确去重', () => {
    const arr = [1, 2, 2, 3, 3, 4]
    const unique = [...new Set(arr)]
    
    expect(unique).toEqual([1, 2, 3, 4])
  })

  it('应该正确按属性排序', () => {
    const arr = [
      { name: 'Charlie', age: 30 },
      { name: 'Alice', age: 25 },
      { name: 'Bob', age: 35 }
    ]
    const sorted = [...arr].sort((a, b) => a.name.localeCompare(b.name))
    
    expect(sorted[0].name).toBe('Alice')
    expect(sorted[1].name).toBe('Bob')
    expect(sorted[2].name).toBe('Charlie')
  })
})

describe('字符串工具函数', () => {
  it('应该正确截断长字符串', () => {
    const str = '这是一个非常长的字符串,需要被截断'
    const truncated = str.length > 10 ? str.substring(0, 10) + '...' : str
    
    // 中文字符每个占一个长度,substring(0, 10)会截取前10个字符
    expect(truncated).toBe('这是一个非常长的字符...')
    expect(truncated.length).toBe(13) // 10个字符 + 3个省略号
  })

  it('应该正确处理短字符串', () => {
    const str = '短字符串'
    const truncated = str.length > 10 ? str.substring(0, 10) + '...' : str
    
    expect(truncated).toBe('短字符串')
  })

  it('应该正确转义HTML特殊字符', () => {
    const str = '<script>alert("xss")</script>'
    const escaped = str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;')
    
    expect(escaped).toContain('&lt;script&gt;')
    expect(escaped).not.toContain('<script>')
  })
})

describe('对象工具函数', () => {
  it('应该正确深拷贝对象', () => {
    const obj = { a: 1, b: { c: 2 } }
    const copy = JSON.parse(JSON.stringify(obj))
    
    expect(copy).toEqual(obj)
    expect(copy).not.toBe(obj)
    expect(copy.b).not.toBe(obj.b)
  })

  it('应该正确合并对象', () => {
    const obj1 = { a: 1, b: 2 }
    const obj2 = { b: 3, c: 4 }
    const merged = { ...obj1, ...obj2 }
    
    expect(merged).toEqual({ a: 1, b: 3, c: 4 })
  })

  it('应该正确检查对象是否为空', () => {
    const emptyObj = {}
    const nonEmptyObj = { a: 1 }
    
    expect(Object.keys(emptyObj).length).toBe(0)
    expect(Object.keys(nonEmptyObj).length).toBeGreaterThan(0)
  })
})

describe('数字工具函数', () => {
  it('应该正确格式化数字为两位小数', () => {
    const num = 3.14159
    const formatted = num.toFixed(2)
    
    expect(formatted).toBe('3.14')
  })

  it('应该正确处理整数', () => {
    const num = 5
    const formatted = num.toFixed(2)
    
    expect(formatted).toBe('5.00')
  })

  it('应该正确判断数字范围', () => {
    const num = 50
    const inRange = num >= 0 && num <= 100
    
    expect(inRange).toBe(true)
  })

  it('应该正确处理负数', () => {
    const num = -10
    const abs = Math.abs(num)
    
    expect(abs).toBe(10)
  })
})
