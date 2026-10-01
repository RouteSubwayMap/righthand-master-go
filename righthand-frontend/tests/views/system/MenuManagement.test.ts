import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import MenuManagement from '@/views/system/MenuManagement.vue'
import { getMenus, createMenu, updateMenu, deleteMenu } from '@/api/system'

vi.mock('@/api/system')

describe('MenuManagement.vue', () => {
  const mockMenus = [
    {
      id: 1,
      parentId: 0,
      menuName: '系统管理',
      menuType: 0,
      path: '/system',
      component: '',
      permission: '',
      icon: 'Setting',
      sortOrder: 1
    },
    {
      id: 2,
      parentId: 1,
      menuName: '用户管理',
      menuType: 1,
      path: '/system/users',
      component: 'system/UserManagement',
      permission: '',
      icon: 'User',
      sortOrder: 1
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    ;(getMenus as vi.Mock).mockResolvedValue({
      code: 200,
      data: mockMenus
    })
    ;(createMenu as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(updateMenu as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(deleteMenu as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
  })

  it('should render correctly', async () => {
    const wrapper = mount(MenuManagement)
    await wrapper.vm.$nextTick()
    
    expect(wrapper.find('.card-header span').text()).toBe('菜单管理')
    expect(wrapper.find('.el-button--primary').text()).toContain('新增菜单')
    expect(wrapper.find('.el-table').exists()).toBe(true)
  })

  it('should load menus on mount', async () => {
    mount(MenuManagement)
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(getMenus).toHaveBeenCalled()
  })

  it('should build tree structure correctly', async () => {
    const wrapper = mount(MenuManagement)
    await wrapper.vm.$nextTick()
    
    const buildTree = (wrapper.vm as any).buildTree
    const tree = buildTree(mockMenus, 0)
    
    expect(tree.length).toBe(1)
    expect(tree[0].children.length).toBe(1)
    expect(tree[0].menuName).toBe('系统管理')
    expect(tree[0].children[0].menuName).toBe('用户管理')
  })

  it('should call createMenu when handleSubmit is called for new menu', async () => {
    const wrapper = mount(MenuManagement)
    await wrapper.vm.$nextTick()
    
    const formData = (wrapper.vm as any).formData
    formData.menuName = '新菜单'
    formData.menuType = 1
    formData.path = '/newmenu'
    formData.component = 'NewMenu'
    
    const handleSubmit = (wrapper.vm as any).handleSubmit
    await handleSubmit()
    
    expect(createMenu).toHaveBeenCalled()
  })

  it('should call deleteMenu when handleDelete is called', async () => {
    const wrapper = mount(MenuManagement)
    await wrapper.vm.$nextTick()
    
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    
    const handleDelete = (wrapper.vm as any).handleDelete
    await handleDelete({ id: 1 })
    
    expect(deleteMenu).toHaveBeenCalledWith(1)
  })
})