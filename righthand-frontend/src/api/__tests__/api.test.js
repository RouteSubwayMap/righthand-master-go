/**
 * API接口单元测试
 * 测试API函数的参数处理和URL构建
 */

import { describe, it, expect } from 'vitest'
import request from '@/utils/request'

// Mock axios
vi.mock('@/utils/request', () => ({
  default: vi.fn()
}))

describe('Task API Tests', () => {
  it('getTaskList should call correct endpoint', async () => {
    const mockResponse = { code: 200, data: { records: [], total: 0 } }
    request.mockResolvedValue(mockResponse)

    const { getTaskList } = await import('@/api/task')
    await getTaskList({ page: 1, size: 10 })

    expect(request).toHaveBeenCalledWith({
      url: '/api/tasks',
      method: 'get',
      params: { page: 1, size: 10 }
    })
  })

  it('getTaskDetail should call correct endpoint with id', async () => {
    const mockResponse = { code: 200, data: {} }
    request.mockResolvedValue(mockResponse)

    const { getTaskDetail } = await import('@/api/task')
    await getTaskDetail(123)

    expect(request).toHaveBeenCalledWith({
      url: '/api/tasks/123',
      method: 'get'
    })
  })

  it('createTask should call correct endpoint with data', async () => {
    const mockResponse = { code: 200, data: {} }
    request.mockResolvedValue(mockResponse)

    const { createTask } = await import('@/api/task')
    const taskData = { title: 'Test Task', itemType: 'TASK' }
    await createTask(taskData)

    expect(request).toHaveBeenCalledWith({
      url: '/api/tasks',
      method: 'post',
      data: taskData
    })
  })

  it('updateTask should call correct endpoint', async () => {
    const mockResponse = { code: 200, data: {} }
    request.mockResolvedValue(mockResponse)

    const { updateTask } = await import('@/api/task')
    const updateData = { title: 'Updated Task' }
    await updateTask(123, updateData)

    expect(request).toHaveBeenCalledWith({
      url: '/api/tasks/123',
      method: 'put',
      data: updateData
    })
  })

  it('deleteTask should call correct endpoint', async () => {
    const mockResponse = { code: 200, data: null }
    request.mockResolvedValue(mockResponse)

    const { deleteTask } = await import('@/api/task')
    await deleteTask(123)

    expect(request).toHaveBeenCalledWith({
      url: '/api/tasks/123',
      method: 'delete'
    })
  })

  it('updateTaskStatus should call correct endpoint', async () => {
    const mockResponse = { code: 200, data: null }
    request.mockResolvedValue(mockResponse)

    const { updateTaskStatus } = await import('@/api/task')
    await updateTaskStatus(123, { statusCode: 'DONE' })

    expect(request).toHaveBeenCalledWith({
      url: '/api/tasks/123/status',
      method: 'put',
      data: { statusCode: 'DONE' }
    })
  })

  it('getDashboardStats should call correct endpoint', async () => {
    const mockResponse = { code: 200, data: { todo: 0, doing: 0, done: 0, overdue: 0 } }
    request.mockResolvedValue(mockResponse)

    const { getDashboardStats } = await import('@/api/task')
    await getDashboardStats()

    expect(request).toHaveBeenCalledWith({
      url: '/api/dashboard/stats',
      method: 'get'
    })
  })
})

describe('Contact API Tests', () => {
  it('getContactList should call correct endpoint', async () => {
    const mockResponse = { code: 200, data: { records: [], total: 0 } }
    request.mockResolvedValue(mockResponse)

    const { getContactList } = await import('@/api/contact')
    await getContactList({ page: 1, size: 10 })

    expect(request).toHaveBeenCalledWith({
      url: '/api/contacts',
      method: 'get',
      params: { page: 1, size: 10 }
    })
  })

  it('createContact should call correct endpoint', async () => {
    const mockResponse = { code: 200, data: {} }
    request.mockResolvedValue(mockResponse)

    const { createContact } = await import('@/api/contact')
    const contactData = { name: '张三', phone: '13800000000' }
    await createContact(contactData)

    expect(request).toHaveBeenCalledWith({
      url: '/api/contacts',
      method: 'post',
      data: contactData
    })
  })
})
