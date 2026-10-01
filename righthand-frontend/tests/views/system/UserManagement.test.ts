import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import UserManagement from '@/views/system/UserManagement.vue'
import { getUserPage, getAllRoles, getUserRoles, createUser, updateUser } from '@/api/system'

// 组件真实调用的接口为分页版 getUserPage（/api/system/users/page），mock 必须与之一致
vi.mock('@/api/system')

// 矩阵：tests/cases/user-add-dialog-validation.md（6 场景与 6 个测试用例 1:1）
describe('UserManagement.vue', () => {
  const mockUsers = [
    { id: 1, username: 'admin', realName: '管理员', email: 'admin@example.com', phone: '13800000001', status: 1 },
    { id: 2, username: 'user', realName: '普通用户', email: 'user@example.com', phone: '13800000002', status: 1 }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    ;(getUserPage as vi.Mock).mockResolvedValue({
      code: 200,
      data: { list: mockUsers, total: 2 }
    })
    ;(getAllRoles as vi.Mock).mockResolvedValue({
      code: 200,
      data: [{ id: 1, roleName: '管理员', roleCode: 'ADMIN' }]
    })
    ;(getUserRoles as vi.Mock).mockResolvedValue({
      code: 200,
      data: [{ id: 1, roleName: '管理员' }]
    })
    ;(createUser as vi.Mock).mockResolvedValue({ code: 200, data: {} })
    ;(updateUser as vi.Mock).mockResolvedValue({ code: 200, data: {} })
  })

  // 挂载并等待 onMounted 的 loadData（含逐行 getUserRoles）完成。
  // 注意：不可使用 stubs: { teleport: true }，element-plus 2.13 的 ElTeleport 组件 name 即 "teleport"，
  // 会被该 stub 命中并吞掉弹窗插槽内容；且 appendToBody=false 时弹窗本就原位渲染，无需 stub。
  const mountAndLoad = async () => {
    const wrapper = mount(UserManagement)
    await flushPromises()
    await flushPromises()
    await wrapper.vm.$nextTick()
    return wrapper
  }

  // 按文本定位并点击按钮（黑盒交互，不访问组件内部状态）
  const clickButton = async (wrapper, text) => {
    const btn = wrapper.findAll('button').find(b => b.text().includes(text))
    expect(btn, `未找到按钮: ${text}`).toBeTruthy()
    await btn.trigger('click')
    await flushPromises()
    await wrapper.vm.$nextTick()
  }

  // 点击弹窗底部的「确定」按钮
  const clickDialogConfirm = async (wrapper) => {
    const ok = wrapper.findAll('.el-dialog__footer button').find(b => b.text().includes('确定'))
    expect(ok, '未找到弹窗确定按钮').toBeTruthy()
    await ok.trigger('click')
    await flushPromises()
    await flushPromises()
    await wrapper.vm.$nextTick()
  }

  it('AddDialog_NoPasswordField：新增弹窗无密码框，用户名与姓名输入框存在', async () => {
    const wrapper = await mountAndLoad()

    // 基础渲染回归：页头、按钮、表格
    expect(wrapper.find('.card-header span').text()).toBe('用户管理')
    expect(wrapper.find('.el-table').exists()).toBe(true)

    await clickButton(wrapper, '新增用户')

    expect(wrapper.find('.el-dialog__title').text()).toBe('新增用户')
    expect(wrapper.find('#password').exists()).toBe(false)
    expect(wrapper.find('#username').exists()).toBe(true)
    expect(wrapper.find('#realName').exists()).toBe(true)
    expect(createUser).not.toHaveBeenCalled()
  })

  it('Submit_NoPasswordInPayload：提交 payload 不含 password 键，弹窗关闭并刷新列表', async () => {
    const wrapper = await mountAndLoad()
    await clickButton(wrapper, '新增用户')

    await wrapper.find('#username').setValue('zhangsan')
    await wrapper.find('#realName').setValue('张三')
    await wrapper.find('#email').setValue('zhangsan@example.com')
    await wrapper.find('#phone').setValue('13800000003')
    await wrapper.vm.$nextTick()

    await clickDialogConfirm(wrapper)

    expect(createUser).toHaveBeenCalledTimes(1)
    const payload = createUser.mock.calls[0][0]
    expect(payload.username).toBe('zhangsan')
    expect(payload.realName).toBe('张三')
    expect(Object.prototype.hasOwnProperty.call(payload, 'password')).toBe(false)

    // 弹窗关闭（overlay 被 v-show 置为 display:none；happy-dom 下 isVisible() 对 display:none 不可靠，直接断言样式）
    expect(wrapper.find('.el-overlay').attributes('style')).toContain('display: none')
    expect(getUserPage).toHaveBeenCalledTimes(2)
  })

  it('Submit_MissingRealName_Blocked：姓名留空被拦截且弹窗保持打开', async () => {
    const wrapper = await mountAndLoad()
    await clickButton(wrapper, '新增用户')

    await wrapper.find('#username').setValue('zhangsan')
    await wrapper.vm.$nextTick()

    await clickDialogConfirm(wrapper)
    // 错误文案展示有 100ms 防抖（validateStateDebounced），等待后断言
    await new Promise(resolve => setTimeout(resolve, 200))
    await flushPromises()

    expect(createUser).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('请输入姓名')
    // 弹窗保持打开（overlay 未被隐藏）
    expect(wrapper.find('.el-overlay').attributes('style') ?? '').not.toContain('display: none')
  })

  it('Submit_MissingUsername_Blocked：用户名留空被拦截且弹窗保持打开', async () => {
    const wrapper = await mountAndLoad()
    await clickButton(wrapper, '新增用户')

    await wrapper.find('#realName').setValue('张三')
    await wrapper.vm.$nextTick()

    await clickDialogConfirm(wrapper)
    // 错误文案展示有 100ms 防抖（validateStateDebounced），等待后断言
    await new Promise(resolve => setTimeout(resolve, 200))
    await flushPromises()

    expect(createUser).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('请输入用户名')
    expect(wrapper.find('.el-overlay').attributes('style') ?? '').not.toContain('display: none')
  })

  it('EditDialog_NoPasswordField：编辑弹窗无密码框，用户名回填且禁用', async () => {
    const wrapper = await mountAndLoad()

    // 列表加载回归：2 条记录渲染
    expect(getUserPage).toHaveBeenCalledTimes(1)
    expect(wrapper.findAll('.el-table__row').length).toBe(2)

    await clickButton(wrapper, '编辑')

    expect(wrapper.find('.el-dialog__title').text()).toBe('编辑用户')
    expect(wrapper.find('#password').exists()).toBe(false)
    const usernameInput = wrapper.find('#username')
    expect(usernameInput.element.value).toBe('admin')
    expect(usernameInput.attributes('disabled')).toBeDefined()
  })

  it('Edit_Submit_NoPasswordInPayload：编辑提交 payload 不含 password 键', async () => {
    const wrapper = await mountAndLoad()
    await clickButton(wrapper, '编辑')

    await wrapper.find('#realName').setValue('管理员2')
    await wrapper.vm.$nextTick()

    await clickDialogConfirm(wrapper)

    expect(updateUser).toHaveBeenCalledTimes(1)
    expect(updateUser.mock.calls[0][0]).toBe(1)
    const payload = updateUser.mock.calls[0][1]
    expect(payload.realName).toBe('管理员2')
    expect(Object.prototype.hasOwnProperty.call(payload, 'password')).toBe(false)
    expect(wrapper.find('.el-overlay').attributes('style')).toContain('display: none')
    expect(getUserPage).toHaveBeenCalledTimes(2)
  })
})
