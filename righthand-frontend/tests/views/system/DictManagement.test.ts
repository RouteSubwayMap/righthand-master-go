import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import DictManagement from '@/views/system/DictManagement.vue'
import { getDicts, createDict, updateDict, deleteDict } from '@/api/system'

vi.mock('@/api/system')

describe('DictManagement.vue', () => {
  const mockDicts = [
    {
      id: 1,
      dictCode: 'MEMO_TYPE_WORK',
      dictName: '备忘类型 - 工作',
      dictType: 'memo_type',
      dictValue: '1',
      dictLabel: '工作',
      sortOrder: 1,
      status: 1,
      remark: ''
    },
    {
      id: 2,
      dictCode: 'MEMO_TYPE_LIFE',
      dictName: '备忘类型 - 生活',
      dictType: 'memo_type',
      dictValue: '2',
      dictLabel: '生活',
      sortOrder: 2,
      status: 1,
      remark: ''
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    ;(getDicts as vi.Mock).mockResolvedValue({
      code: 200,
      data: mockDicts
    })
    ;(createDict as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(updateDict as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(deleteDict as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
  })

  it('should render correctly', async () => {
    const wrapper = mount(DictManagement)
    await wrapper.vm.$nextTick()
    
    expect(wrapper.find('.card-header span').text()).toBe('字典管理')
    expect(wrapper.find('.el-button--primary').text()).toContain('新增字典')
    expect(wrapper.find('.el-table').exists()).toBe(true)
  })

  it('should load dicts on mount', async () => {
    mount(DictManagement)
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(getDicts).toHaveBeenCalled()
  })

  it('should filter dicts by type', async () => {
    const wrapper = mount(DictManagement)
    await wrapper.vm.$nextTick()
    
    const dictTypeOptions = (wrapper.vm as any).dictTypeOptions
    expect(dictTypeOptions.value).toContain('memo_type')
  })

  it('should call createDict when handleSubmit is called for new dict', async () => {
    const wrapper = mount(DictManagement)
    await wrapper.vm.$nextTick()
    
    const formData = (wrapper.vm as any).formData
    formData.dictCode = 'NEW_DICT'
    formData.dictName = '新字典'
    formData.dictType = 'new_type'
    formData.dictValue = '1'
    formData.dictLabel = '新标签'
    
    const handleSubmit = (wrapper.vm as any).handleSubmit
    await handleSubmit()
    
    expect(createDict).toHaveBeenCalled()
  })

  it('should call deleteDict when handleDelete is called', async () => {
    const wrapper = mount(DictManagement)
    await wrapper.vm.$nextTick()
    
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    
    const handleDelete = (wrapper.vm as any).handleDelete
    await handleDelete({ id: 1 })
    
    expect(deleteDict).toHaveBeenCalledWith(1)
  })

  it('should reset search form correctly', async () => {
    const wrapper = mount(DictManagement)
    await wrapper.vm.$nextTick()
    
    const searchForm = (wrapper.vm as any).searchForm
    searchForm.dictType = 'memo_type'
    
    const resetSearch = (wrapper.vm as any).resetSearch
    resetSearch()
    
    expect(searchForm.dictType).toBe('')
  })
})