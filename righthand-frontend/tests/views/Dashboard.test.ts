import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Dashboard from '@/views/Dashboard.vue'
import { getDashboardStats, getTaskList, updateTaskStatus } from '@/api/task'

vi.mock('@/api/task')
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  // request.js 间接依赖 @/router，需要 createRouter/createWebHistory 基础桩
  createRouter: () => ({ push: vi.fn(), beforeEach: vi.fn() }),
  createWebHistory: () => ({})
}))

describe('Dashboard.vue', () => {
  const mockStats = {
    todoCount: 5,
    doingCount: 3,
    overdueCount: 2
  }

  const mockTasks = [
    {
      id: 1,
      title: '测试任务1',
      itemType: 'TASK',
      priority: 'HIGH',
      statusCode: 'TODO',
      deadline: '2024-01-15 18:00:00'
    },
    {
      id: 2,
      title: '测试任务2',
      itemType: 'MEMO',
      priority: 'MEDIUM',
      statusCode: 'DOING',
      deadline: '2024-01-16 18:00:00'
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    ;(getDashboardStats as vi.Mock).mockResolvedValue({
      code: 200,
      data: mockStats
    })
    ;(getTaskList as vi.Mock).mockResolvedValue({
      code: 200,
      data: {
        records: mockTasks,
        total: 2
      }
    })
    ;(updateTaskStatus as vi.Mock).mockResolvedValue({
      code: 200,
      data: null
    })
  })

  it('should render correctly', async () => {
    const wrapper = mount(Dashboard)
    await wrapper.vm.$nextTick()
    
    // 检查统计卡片是否存在
    expect(wrapper.find('.stats-row').exists()).toBe(true)
    expect(wrapper.find('.stat-card').exists()).toBe(true)
    
    // 检查三个统计卡片（待办、进行中、已逾期）
    const statCards = wrapper.findAll('.stat-card')
    expect(statCards.length).toBe(3)
    
    // 检查列表区域
    expect(wrapper.find('.task-list-card').exists()).toBe(true)
  })

  it('should load stats data on mount', async () => {
    mount(Dashboard)
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(getDashboardStats).toHaveBeenCalled()
  })

  it('should load task lists on mount', async () => {
    mount(Dashboard)
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(getTaskList).toHaveBeenCalled()
    // 应该调用3次：待办、进行中、已逾期
    expect(getTaskList).toHaveBeenCalledTimes(3)
  })

  it('should display correct stat values', async () => {
    const wrapper = mount(Dashboard)
    await wrapper.vm.$nextTick()
    
    // 获取统计数值元素
    const statValues = wrapper.findAll('.stat-value')
    expect(statValues[0].text()).toBe('5') // 待办
    expect(statValues[1].text()).toBe('3') // 进行中
    expect(statValues[2].text()).toBe('2') // 已逾期
  })

  it('should format date correctly', async () => {
    const wrapper = mount(Dashboard)
    await wrapper.vm.$nextTick()
    
    const formatDateTime = (wrapper.vm as any).formatDateTime
    const result = formatDateTime('2024-01-15T10:30:00')
    
    // 应该包含日期和时间
    expect(result).toContain('2024')
    expect(result).toContain('01')
  })

  it('should get correct item type label', async () => {
    const wrapper = mount(Dashboard)
    await wrapper.vm.$nextTick()
    
    const getItemTypeLabel = (wrapper.vm as any).getItemTypeLabel
    expect(getItemTypeLabel('TASK')).toBe('任务')
    expect(getItemTypeLabel('MEMO')).toBe('备忘')
    expect(getItemTypeLabel('REQUIREMENT')).toBe('需求')
    expect(getItemTypeLabel('UNKNOWN')).toBe('UNKNOWN')
  })

  it('should get correct priority label', async () => {
    const wrapper = mount(Dashboard)
    await wrapper.vm.$nextTick()
    
    const getPriorityLabel = (wrapper.vm as any).getPriorityLabel
    expect(getPriorityLabel('HIGH')).toBe('高')
    expect(getPriorityLabel('MEDIUM')).toBe('中')
    expect(getPriorityLabel('LOW')).toBe('低')
  })

  it('should get correct status label', async () => {
    const wrapper = mount(Dashboard)
    await wrapper.vm.$nextTick()
    
    const getStatusLabel = (wrapper.vm as any).getStatusLabel
    expect(getStatusLabel('TODO')).toBe('待办')
    expect(getStatusLabel('DOING')).toBe('进行中')
    expect(getStatusLabel('DONE')).toBe('已完成')
  })

  it('should call updateTaskStatus when handleComplete is called', async () => {
    const wrapper = mount(Dashboard)
    await wrapper.vm.$nextTick()
    
    const handleComplete = (wrapper.vm as any).handleComplete
    await handleComplete({ id: 1 })
    
    expect(updateTaskStatus).toHaveBeenCalledWith(1, { statusCode: 'DONE' })
  })
})