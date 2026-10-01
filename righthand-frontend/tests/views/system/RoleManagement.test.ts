import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import RoleManagement from '@/views/system/RoleManagement.vue'
import { getRoles, createRole, updateRole, deleteRole, getMenus, getRoleMenus } from '@/api/system'

vi.mock('@/api/system')

describe('RoleManagement.vue', () => {
  const mockRoles = [
    {
      id: 1,
      roleName: '管理员',
      roleCode: 'admin',
      description: '系统管理员',
      status: 1
    },
    {
      id: 2,
      roleName: '普通用户',
      roleCode: 'user',
      description: '普通用户',
      status: 1
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    ;(getRoles as vi.Mock).mockResolvedValue({
      code: 200,
      data: {
        records: mockRoles,
        total: 2
      }
    })
    ;(createRole as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(updateRole as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(deleteRole as vi.Mock).mockResolvedValue({
      code: 200,
      data: {}
    })
    ;(getMenus as vi.Mock).mockResolvedValue({
      code: 200,
      data: []
    })
    ;(getRoleMenus as vi.Mock).mockResolvedValue({
      code: 200,
      data: []
    })
  })

  it('should render correctly', async () => {
    const wrapper = mount(RoleManagement)
    await wrapper.vm.$nextTick()
    
    expect(wrapper.find('.card-header span').text()).toBe('角色管理')
    expect(wrapper.find('.el-button--primary').text()).toContain('新增角色')
    expect(wrapper.find('.el-table').exists()).toBe(true)
  })

  it('should load roles on mount', async () => {
    mount(RoleManagement)
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(getRoles).toHaveBeenCalled()
  })

  it('should call createRole when handleSubmit is called for new role', async () => {
    const wrapper = mount(RoleManagement)
    await wrapper.vm.$nextTick()
    
    const formData = (wrapper.vm as any).formData
    formData.roleName = '新角色'
    formData.roleCode = 'newrole'
    formData.description = '新角色描述'
    formData.status = 1
    
    const handleSubmit = (wrapper.vm as any).handleSubmit
    await handleSubmit()
    
    expect(createRole).toHaveBeenCalled()
  })

  it('should call deleteRole when handleDelete is called', async () => {
    const wrapper = mount(RoleManagement)
    await wrapper.vm.$nextTick()
    
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    
    const handleDelete = (wrapper.vm as any).handleDelete
    await handleDelete({ id: 1 })
    
    expect(deleteRole).toHaveBeenCalledWith(1)
  })
})