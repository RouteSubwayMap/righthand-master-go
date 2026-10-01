import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Contacts from '@/views/Contacts.vue'
import { getContactList, createContact, updateContact, deleteContact } from '@/api/contact'

vi.mock('@/api/contact')
vi.mock('@/stores/dict', () => ({
  useDictStore: () => ({
    getDictByType: vi.fn().mockReturnValue([
      { dictValue: 'ACCOUNT_MANAGER', dictLabel: '客户经理' },
      { dictValue: 'PRODUCT', dictLabel: '产品' },
      { dictValue: 'R&D', dictLabel: '研发' }
    ]),
    getDictLabel: vi.fn().mockImplementation((type, value) => {
      const map: Record<string, string> = {
        'ACCOUNT_MANAGER': '客户经理',
        'PRODUCT': '产品',
        'R&D': '研发'
      }
      return map[value] || value
    })
  })
}))

describe('Contacts.vue', () => {
  const mockContacts = [
    {
      id: 1,
      name: '张三',
      phone: '13800000001',
      email: 'zhangsan@example.com',
      contactGroup: 'ACCOUNT_MANAGER',
      company: '测试公司',
      position: '经理',
      remark: '备注'
    },
    {
      id: 2,
      name: '李四',
      phone: '13800000002',
      email: 'lisi@example.com',
      contactGroup: 'PRODUCT',
      company: '测试公司',
      position: '产品经理',
      remark: ''
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    ;(getContactList as vi.Mock).mockResolvedValue({
      code: 200,
      data: {
        records: mockContacts,
        total: 2
      }
    })
    ;(createContact as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(updateContact as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(deleteContact as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
  })

  it('should render correctly', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    // 检查标题
    expect(wrapper.find('.card-header span').text()).toBe('通讯录')
    
    // 检查搜索表单
    expect(wrapper.find('form').exists()).toBe(true)
    
    // 检查视图切换
    expect(wrapper.find('.el-radio-group').exists()).toBe(true)
    
    // 默认是看板视图
    expect(wrapper.find('.board-view').exists()).toBe(true)
  })

  it('should load contacts data on mount', async () => {
    mount(Contacts)
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(getContactList).toHaveBeenCalled()
  })

  it('should switch between list and board view', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    // 默认是看板视图
    expect(wrapper.find('.board-view').exists()).toBe(true)
    expect(wrapper.find('.el-table').exists()).toBe(false)
    
    // 切换到列表视图
    const listRadio = wrapper.find('input[value="list"]')
    await listRadio.setChecked()
    
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.el-table').exists()).toBe(true)
    expect(wrapper.find('.board-view').exists()).toBe(false)
  })

  it('should filter contacts by group', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    const getContactsByGroup = (wrapper.vm as any).getContactsByGroup
    const result = getContactsByGroup('ACCOUNT_MANAGER')
    
    expect(result.length).toBe(1)
    expect(result[0].name).toBe('张三')
  })

  it('should format position correctly', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    const formatPosition = (wrapper.vm as any).formatPosition
    const result = formatPosition('职位1,职位2,职位3')
    
    expect(result).toBe('职位1\n职位2\n职位3')
  })

  it('should generate safe card class', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    const getCardClass = (wrapper.vm as any).getCardClass
    expect(getCardClass('ACCOUNT_MANAGER')).toBe('contact-card-ACCOUNT_MANAGER')
    expect(getCardClass('R&D')).toBe('contact-card-R_D')
  })

  it('should call createContact when handleSubmit is called for new contact', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    const handleSubmit = (wrapper.vm as any).handleSubmit
    const formData = (wrapper.vm as any).formData
    
    formData.name = '新联系人'
    formData.phone = '13800000003'
    formData.email = 'new@example.com'
    formData.contactGroup = 'PRODUCT'
    
    await handleSubmit()
    
    expect(createContact).toHaveBeenCalledWith({
      name: '新联系人',
      phone: '13800000003',
      email: 'new@example.com',
      contactGroup: 'PRODUCT',
      company: '',
      position: '',
      remark: ''
    })
  })

  it('should call updateContact when handleSubmit is called for existing contact', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    const formData = (wrapper.vm as any).formData
    formData.id = 1
    formData.name = '更新姓名'
    formData.phone = '13800000001'
    formData.email = 'updated@example.com'
    
    const handleSubmit = (wrapper.vm as any).handleSubmit
    await handleSubmit()
    
    expect(updateContact).toHaveBeenCalledWith(1, {
      id: 1,
      name: '更新姓名',
      phone: '13800000001',
      email: 'updated@example.com',
      contactGroup: '',
      company: '',
      position: '',
      remark: ''
    })
  })

  it('should call deleteContact when handleDelete is called', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    const handleDelete = (wrapper.vm as any).handleDelete
    
    // Mock ElMessageBox.confirm
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    
    await handleDelete({ id: 1 })
    
    expect(deleteContact).toHaveBeenCalledWith(1)
  })

  it('should reset search form correctly', async () => {
    const wrapper = mount(Contacts)
    await wrapper.vm.$nextTick()
    
    const searchForm = (wrapper.vm as any).searchForm
    searchForm.name = '测试'
    searchForm.position = '经理'
    searchForm.contactGroup = 'PRODUCT'
    
    const resetSearch = (wrapper.vm as any).resetSearch
    resetSearch()
    
    expect(searchForm.name).toBe('')
    expect(searchForm.position).toBe('')
    expect(searchForm.contactGroup).toBe('')
  })
})