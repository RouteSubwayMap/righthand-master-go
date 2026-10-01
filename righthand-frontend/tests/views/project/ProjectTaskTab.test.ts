import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import ProjectTaskTab from '@/views/project/components/ProjectTaskTab.vue'
import TaskSelectDialog from '@/components/TaskSelectDialog.vue'
import { getProjectTasks, linkTaskToWeek } from '@/api/project'
import { ElMessage } from 'element-plus'

// 沙箱/CI 环境较慢，放大超时避免真实组件挂载（Element Plus）误报超时
vi.setConfig({ testTimeout: 90000, hookTimeout: 90000 })

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() })
}))

vi.mock('@/api/project', () => ({
  getProjectTasks: vi.fn(),
  createTaskInWeek: vi.fn(),
  linkTaskToWeek: vi.fn(),
  unlinkTaskFromWeek: vi.fn(),
  moveTaskToWeek: vi.fn()
}))

vi.mock('@/api/task', () => ({ getTaskList: vi.fn() }))

vi.mock('element-plus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('element-plus')>()
  return {
    ...actual,
    ElMessage: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
    ElMessageBox: { confirm: vi.fn() }
  }
})

const project = {
  id: 9,
  status: 'IN_PROGRESS',
  allowedActions: ['addWeek'],
  weeks: [
    { id: 21, weekNo: 1, title: '启动周', allowedActions: ['edit'] },
    { id: 22, weekNo: 2, title: '开发周', allowedActions: ['edit'] }
  ]
}

const users = [
  { id: 1, username: 'admin', realName: '管理员' },
  { id: 2, username: 'user', realName: '普通用户' }
]

async function mountTab() {
  const wrapper = mount(ProjectTaskTab, {
    props: { project, users }
  })
  await flushPromises()
  await nextTick()
  return wrapper
}

/** 打开关联弹窗 → 打开选择弹窗 → 选择一条事项（模拟子组件确认） */
async function openAndSelect(wrapper: any, row: Record<string, any>) {
  const vm = wrapper.vm as any
  vm.openLinkDialog()
  await nextTick()
  vm.taskSelectVisible = true
  await nextTick()
  wrapper.findComponent(TaskSelectDialog).vm.$emit('select', row)
  await nextTick()
  return vm
}

describe('ProjectTaskTab 关联已有事项（选择弹窗链路）', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(getProjectTasks as any).mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
    ;(linkTaskToWeek as any).mockResolvedValue({ code: 200, message: '关联成功', data: {} })
  })

  it('OpenLinkDialog_DefaultFirstWeek：打开关联弹窗可见并默认归属第一周', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any

    vm.openLinkDialog()
    await nextTick()

    expect(vm.linkDialogVisible).toBe(true)
    expect(vm.linkForm.weekPlanId).toBe(21)
    expect(vm.taskSelectVisible).toBe(false)
  })

  it('SelectTask_LinkSuccess_CloseDialog：选择事项回填显示并关联成功关闭（完整链路）', async () => {
    const wrapper = await mountTab()
    const vm = await openAndSelect(wrapper, { id: 12, title: '修复登录bug' })

    expect(vm.linkForm.taskId).toBe(12)
    expect(vm.linkForm.taskTitle).toBe('#12 修复登录bug')

    await vm.handleLink()
    await flushPromises()

    expect(linkTaskToWeek).toHaveBeenCalledWith(9, 21, 12)
    expect(vm.linkDialogVisible).toBe(false)
  })

  it('Link_ApiReject_DialogStaysOpen：关联请求失败时弹窗保持打开，不误关闭（异常）', async () => {
    ;(linkTaskToWeek as any).mockRejectedValueOnce(new Error('Error'))
    const wrapper = await mountTab()
    const vm = await openAndSelect(wrapper, { id: 12, title: '修复登录bug' })

    await vm.handleLink()
    await flushPromises()

    expect(linkTaskToWeek).toHaveBeenCalledTimes(1)
    expect(vm.linkDialogVisible).toBe(true)
  })

  it('Link_NoSelection_GuardBlocks：未选择事项直接提交被守卫拦截，不触发请求（反例）', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any

    vm.openLinkDialog()
    await nextTick()
    expect(vm.linkForm.taskId).toBeNull()

    await vm.handleLink()
    await flushPromises()

    expect(ElMessage.warning).toHaveBeenCalled()
    expect(linkTaskToWeek).not.toHaveBeenCalled()
    expect(vm.linkDialogVisible).toBe(true)
  })

  it('Reopen_ClearPreviousSelection：重新打开弹窗清空上一次的选择与回显（回归：无残留状态）', async () => {
    const wrapper = await mountTab()
    const vm = await openAndSelect(wrapper, { id: 12, title: '修复登录bug' })

    vm.openLinkDialog()
    await nextTick()

    expect(vm.linkForm.taskId).toBeNull()
    expect(vm.linkForm.taskTitle).toBe('')
  })

  it('OpenLinkDialog_WeekFilterFollows：周过滤已选第 2 周时打开弹窗，归属周跟随筛选周（边界）', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any

    vm.weekFilter = 22
    vm.openLinkDialog()
    await nextTick()

    expect(vm.linkForm.weekPlanId).toBe(22)
  })
})

describe('ProjectTaskTab 负责人/状态筛选', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(getProjectTasks as any).mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
  })

  it('Filter_User_TriggersSearchFromFirstPage：设置负责人后回到第 1 页并携带 userId（正例）', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any
    vm.pageNum = 2

    vm.userFilter = 2
    vm.handleSearch()
    await flushPromises()

    expect(vm.pageNum).toBe(1)
    expect(getProjectTasks).toHaveBeenLastCalledWith(9, { pageNum: 1, pageSize: 10, userId: 2 })
  })

  it('Filter_Status_AppliesStatusCode：设置状态后请求携带 statusCode（正例）', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any

    vm.statusFilter = 'DONE'
    vm.handleSearch()
    await flushPromises()

    expect(getProjectTasks).toHaveBeenLastCalledWith(9, { pageNum: 1, pageSize: 10, statusCode: 'DONE' })
  })

  it('Filter_Combined_WeekUserStatus：周+负责人+状态三条件同时传递（边界）', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any

    vm.weekFilter = 22
    vm.userFilter = 2
    vm.statusFilter = 'DONE'
    await vm.loadTasks()
    await flushPromises()

    expect(getProjectTasks).toHaveBeenLastCalledWith(9, {
      pageNum: 1,
      pageSize: 10,
      weekPlanId: 22,
      userId: 2,
      statusCode: 'DONE'
    })
  })

  it('Filter_Cleared_OmitsFilterParams：清空三个筛选后请求不带过滤参数（反例）', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any

    vm.weekFilter = null
    vm.userFilter = null
    vm.statusFilter = null
    await vm.loadTasks()
    await flushPromises()

    expect(getProjectTasks).toHaveBeenLastCalledWith(9, { pageNum: 1, pageSize: 10 })
  })

  it('Filter_Reset_OnProjectSwitch：切换项目重置三个筛选并重新查询（反例）', async () => {
    const wrapper = await mountTab()
    const vm = wrapper.vm as any

    vm.weekFilter = 22
    vm.userFilter = 2
    vm.statusFilter = 'DONE'

    await wrapper.setProps({ project: { ...project, id: 10 } })
    await flushPromises()

    expect(vm.weekFilter).toBeNull()
    expect(vm.userFilter).toBeNull()
    expect(vm.statusFilter).toBeNull()
    expect(getProjectTasks).toHaveBeenLastCalledWith(10, { pageNum: 1, pageSize: 10 })
  })

  it('Filter_Toolbar_ThreeSelectsRendered：工具栏渲染周/负责人/状态 3 个下拉（边界）', async () => {
    const wrapper = await mountTab()
    const selects = wrapper.findAll('.toolbar-left .el-select')

    expect(selects.length).toBe(3)
    expect(selects[1].text()).toContain('全部负责人')
    expect(selects[2].text()).toContain('全部状态')
  })
})
