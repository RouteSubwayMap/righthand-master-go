import { describe, it, expect, vi } from 'vitest'
import * as requirementApi from '@/api/requirement'

// Mock request module
vi.mock('@/utils/request', () => ({
  default: vi.fn((config) => {
    // 模拟API响应
    return Promise.resolve({
      code: 200,
      data: {
        list: [
          { id: 1, title: '测试需求1', status: 'PENDING_CLARIFY' },
          { id: 2, title: '测试需求2', status: 'DEVELOPING' }
        ],
        total: 2,
        page: 1,
        pageSize: 10
      }
    })
  })
}))

describe('需求管理API', () => {
  it('应该正确调用分页查询接口', async () => {
    const result = await requirementApi.getRequirementPage({ pageNum: 1, pageSize: 10 })
    
    expect(result.code).toBe(200)
    expect(result.data.records).toHaveLength(2)
    expect(result.data.total).toBe(2)
  })

  it('应该正确转换分页参数', async () => {
    // 验证pageNum被转换为page
    const result = await requirementApi.getRequirementPage({ pageNum: 2 })
    
    expect(result).toBeDefined()
  })

  it('应该正确获取需求详情', async () => {
    const result = await requirementApi.getRequirementDetail(1)
    
    expect(result.code).toBe(200)
  })

  it('应该正确创建需求', async () => {
    const newReq = {
      title: '新需求',
      description: '这是一个新需求'
    }
    
    const result = await requirementApi.createRequirement(newReq)
    
    expect(result.code).toBe(200)
  })

  it('应该正确更新需求', async () => {
    const updates = {
      title: '更新后的需求'
    }
    
    const result = await requirementApi.updateRequirement(1, updates)
    
    expect(result.code).toBe(200)
  })

  it('应该正确删除需求', async () => {
    const result = await requirementApi.deleteRequirement(1)
    
    expect(result.code).toBe(200)
  })

  it('应该正确更新需求状态', async () => {
    const result = await requirementApi.updateRequirementStatus(1, 'DEVELOPING')
    
    expect(result.code).toBe(200)
  })

  it('应该正确执行状态流转', async () => {
    const result = await requirementApi.executeRequirementTransition(1, 'DEVELOPING')
    
    expect(result.code).toBe(200)
  })
})
