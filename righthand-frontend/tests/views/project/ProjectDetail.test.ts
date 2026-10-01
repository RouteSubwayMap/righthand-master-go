import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import ProjectDetail from '@/views/project/ProjectDetail.vue'
import WeeklyPlanSection from '@/views/project/components/WeeklyPlanSection.vue'
import { getProjectDetail } from '@/api/project'
import { getAllUsers } from '@/api/user'
import { getTaskList } from '@/api/task'

// 沙箱/CI 环境较慢，放大超时避免真实组件挂载（Element Plus）误报超时
vi.setConfig({ testTimeout: 90000, hookTimeout: 90000 })

const { pushMock } = vi.hoisted(() => ({ pushMock: vi.fn() }))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: '1' }, name: 'ProjectDetailPage' })
}))

vi.mock('@/api/project', () => ({
  getProjectDetail: vi.fn(),
  changeProjectStatus: vi.fn(),
  changeProjectRisk: vi.fn(),
  deleteProject: vi.fn(),
  addProjectMember: vi.fn(),
  removeProjectMember: vi.fn(),
  createProject: vi.fn(),
  updateProject: vi.fn(),
  createWeekPlan: vi.fn(),
  updateWeekPlan: vi.fn(),
  deleteWeekPlan: vi.fn(),
  changeWeekStatus: vi.fn(),
  submitWeekSummary: vi.fn(),
  unlockWeekSummary: vi.fn(),
  getProjectTasks: vi.fn(),
  createTaskInWeek: vi.fn(),
  linkTaskToWeek: vi.fn(),
  unlinkTaskFromWeek: vi.fn(),
  moveTaskToWeek: vi.fn(),
  createProjectStage: vi.fn(),
  updateProjectStage: vi.fn(),
  deleteProjectStage: vi.fn()
}))

vi.mock('@/api/user', () => ({ getAllUsers: vi.fn() }))

vi.mock('@/api/task', () => ({ getTaskList: vi.fn() }))

vi.mock('element-plus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('element-plus')>()
  return {
    ...actual,
    ElMessage: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
    ElMessageBox: { confirm: vi.fn(), prompt: vi.fn() }
  }
})

const weeks = [
  {
    id: 11,
    weekNo: 1,
    title: '需求调研',
    startDate: '2026-01-05',
    endDate: '2026-01-09',
    ownerId: 1,
    ownerName: '管理员',
    status: 'DONE',
    taskCount: 2,
    overdue: false,
    allowedActions: ['changeStatus', 'submitSummary']
  },
  {
    id: 12,
    weekNo: 2,
    title: '接口设计',
    startDate: '2026-01-12',
    endDate: '2026-01-16',
    ownerId: 2,
    ownerName: '普通用户',
    status: 'IN_PROGRESS',
    taskCount: 1,
    overdue: false,
    allowedActions: ['changeStatus', 'submitSummary', 'edit']
  },
  {
    id: 13,
    weekNo: 3,
    title: '核心开发',
    startDate: '2026-01-19',
    endDate: '2026-01-23',
    ownerId: 2,
    ownerName: '普通用户',
    status: 'IN_PROGRESS',
    taskCount: 0,
    overdue: true,
    allowedActions: ['changeStatus', 'submitSummary', 'edit']
  },
  {
    id: 14,
    weekNo: 4,
    title: '联调准备',
    startDate: '2026-01-26',
    endDate: '2026-01-30',
    ownerId: 1,
    ownerName: '管理员',
    status: 'NOT_STARTED',
    taskCount: 0,
    overdue: false,
    allowedActions: ['changeStatus'],
    summary: JSON.stringify({
      finished: '完成联调环境准备',
      nextPlan: '执行联调',
      risks: '环境不稳定',
      support: '需要运维支持'
    }),
    summaryByName: '管理员',
    summaryAt: '2026-01-30 18:00:00'
  }
]

const stages = [
  {
    id: 31,
    projectId: 1,
    stageName: '需求调研',
    startDate: '2026-01-05',
    endDate: '2026-01-20',
    description: '调研业务需求'
  },
  {
    id: 32,
    projectId: 1,
    stageName: '开发联调',
    startDate: '2026-01-21',
    endDate: '2026-02-10',
    description: ''
  },
  {
    id: 33,
    projectId: 1,
    stageName: '验收上线',
    startDate: '2026-02-11',
    endDate: '2026-02-20',
    description: '验收与上线'
  }
]

const detailData: Record<string, any> = {
  id: 1,
  projectName: '新一代信贷核心系统对接',
  projectKey: 'NCBC',
  description: '信贷核心系统重构',
  planDesc: '分五周推进',
  managerId: 1,
  managerName: '管理员',
  startDate: '2026-01-05',
  endDate: '2026-02-06',
  status: 'IN_PROGRESS',
  riskLevel: 'MEDIUM',
  deliveryStatus: 'NORMAL',
  riskNote: '',
  progressPct: 40,
  weekCount: 5,
  doneWeekCount: 2,
  allowedActions: ['pause', 'close', 'edit', 'editRisk', 'manageMember', 'addWeek', 'terminate', 'delete'],
  members: [
    { id: 1, userId: 1, userName: '管理员', memberRole: 'PM', createTime: '2026-01-05 10:00:00' },
    { id: 2, userId: 2, userName: '普通用户', memberRole: 'MEMBER', createTime: '2026-01-06 10:00:00' }
  ],
  logs: [
    { id: 1, action: 'CREATE', detail: '创建项目', operatorName: '管理员', createTime: '2026-01-05 10:00:00' },
    { id: 2, action: 'STATUS', detail: '启动 → 进行中', operatorName: '管理员', createTime: '2026-01-06 09:00:00' }
  ],
  weeks,
  stages,
  currentStageId: 32
}

async function mountDetail() {
  const wrapper = mount(ProjectDetail)
  await flushPromises()
  await nextTick()
  return wrapper
}

describe('ProjectDetail 项目详情页', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(getProjectDetail as any).mockResolvedValue({ code: 200, data: detailData })
    ;(getAllUsers as any).mockResolvedValue({
      code: 200,
      data: [
        { id: 1, username: 'admin', realName: '管理员' },
        { id: 2, username: 'user', realName: '普通用户' }
      ]
    })
    ;(getTaskList as any).mockResolvedValue({ code: 200, data: { list: [], total: 0 } })
  })

  it('详情加载：按路由 id 请求并渲染名称/状态/进度概览', async () => {
    const wrapper = await mountDetail()

    expect(getProjectDetail).toHaveBeenCalledWith('1')
    const text = wrapper.text()
    expect(text).toContain('新一代信贷核心系统对接')
    expect(text).toContain('NCBC')
    expect(text).toContain('进行中')
    expect(text).toContain('2 / 5 周')
    expect(text).toContain('40%')
  })

  it('详情默认激活「项目整体计划」页签且位于「周计划」之前，渲染阶段行并高亮当前阶段（正例）', async () => {
    const wrapper = await mountDetail()

    const tabLabels = wrapper.findAll('.el-tabs__item').map((t) => t.text())
    expect(tabLabels[0]).toBe('项目整体计划')
    expect(tabLabels[1]).toBe('周计划')

    // 默认页签（用户反馈：点击清单「详情」默认展示「项目整体计划」）
    expect(wrapper.find('.el-tabs__item.is-active').text()).toBe('项目整体计划')

    const text = wrapper.text()
    expect(text).toContain('开发联调')
    expect(text).toContain('验收上线')
    expect(text).toContain('共 3 个阶段')

    // 当前阶段（currentStageId=32 命中「开发联调」）高亮并展示标签
    expect(wrapper.findAll('.current-stage-row').length).toBe(1)
    const currentTags = wrapper.findAll('.el-tag').filter((t) => t.text() === '当前阶段')
    expect(currentTags.length).toBe(1)
  })

  it('头部操作按钮按 allowedActions 优先级完整渲染', async () => {
    const wrapper = await mountDetail()

    const headerBtnTexts = wrapper.findAll('.header-actions button').map((b) => b.text())
    expect(headerBtnTexts).toEqual([
      '暂停',
      '结项',
      '编辑',
      '修改风险',
      '成员管理',
      '新增周计划',
      '终止',
      '删除'
    ])
  })

  it('周计划 Tab 渲染：周卡片、已延期标签与总结锁定态', async () => {
    const wrapper = await mountDetail()

    expect(wrapper.findAll('.week-item').length).toBe(4)
    const text = wrapper.text()
    expect(text).toContain('第1周')
    expect(text).toContain('需求调研')

    const overdueTags = wrapper.findAll('.el-tag').filter((t) => t.text() === '已延期')
    expect(overdueTags.length).toBe(1)

    // 第4周已提交总结：锁定态 + 结构化内容 + 提交人
    const lockedTags = wrapper.findAll('.el-tag').filter((t) => t.text() === '已锁定')
    expect(lockedTags.length).toBe(1)
    expect(text).toContain('完成联调环境准备')
    expect(text).toContain('需要运维支持')
    expect(text).toContain('管理员 提交于')
  })

  it('详情接口失败时展示空态而非报错页', async () => {
    ;(getProjectDetail as any).mockRejectedValueOnce(new Error('Error'))

    const wrapper = await mountDetail()

    expect(wrapper.text()).toContain('项目不存在或已被删除')
  })

  it('allowedActions 未声明 manageMember/terminate 时不渲染对应入口', async () => {
    ;(getProjectDetail as any).mockResolvedValueOnce({
      code: 200,
      data: { ...detailData, allowedActions: ['edit'] }
    })

    const wrapper = await mountDetail()
    const vm = wrapper.vm as any

    expect(wrapper.findAll('.header-actions button').map((b) => b.text())).toEqual(['编辑'])
    expect(vm.canManageMembers).toBe(false)
  })

  it('周级提交总结按钮仅按周 allowedActions 渲染；未声明解锁则不出现解锁入口', async () => {
    const wrapper = await mountDetail()

    const submitBtns = wrapper.findAll('button').filter((b) => b.text() === '填写总结')
    expect(submitBtns.length).toBe(3)

    const unlockBtns = wrapper.findAll('button').filter((b) => b.text() === '解锁总结')
    expect(unlockBtns.length).toBe(0)
  })

  it('终态项目不计算超期天数（边界）', async () => {
    ;(getProjectDetail as any).mockResolvedValueOnce({
      code: 200,
      data: { ...detailData, status: 'CLOSED' }
    })

    const wrapper = await mountDetail()
    const vm = wrapper.vm as any

    expect(vm.overdueDays).toBe(0)
    expect(wrapper.text()).toContain('未超期')
  })

  it('未知枚举值降级展示，不抛异常（边界）', async () => {
    const wrapper = await mountDetail()
    const vm = wrapper.vm as any

    expect(vm.statusMeta('WHATEVER')).toEqual({ label: 'WHATEVER', tagType: 'info' })
    expect(vm.statusMeta(null)).toEqual({ label: '-', tagType: 'info' })
    expect(vm.riskMeta('WHATEVER')).toEqual({ label: 'WHATEVER', tagType: 'info' })
    expect(vm.deliveryMeta(undefined)).toEqual({ label: '-', tagType: 'info' })
  })
})

describe('WeeklyPlanSection 周计划组件', () => {
  const buildProject = (overrides: Record<string, any> = {}) => ({
    id: 1,
    startDate: '2026-01-05',
    endDate: '2026-02-06',
    managerId: 1,
    allowedActions: ['addWeek'],
    weeks,
    ...overrides
  })

  function mountSection(project: any) {
    return mount(WeeklyPlanSection, {
      props: {
        project,
        users: [{ id: 1, username: 'admin', realName: '管理员' }]
      }
    })
  }

  it('渲染周完成统计与全部周卡片', () => {
    const wrapper = mountSection(buildProject())

    expect(wrapper.text()).toContain('共 4 周，已完成 1 周')
    expect(wrapper.findAll('.week-item').length).toBe(4)
  })

  it('未授予 addWeek 时不渲染「新增周计划」按钮（反例：无权限无入口）', () => {
    const withPerm = mountSection(buildProject({ allowedActions: ['addWeek'] }))
    const withoutPerm = mountSection(buildProject({ allowedActions: [] }))

    expect(withPerm.text()).toContain('新增周计划')
    expect(withoutPerm.text()).not.toContain('新增周计划')
  })

  it('解锁总结按钮仅在周 allowedActions 声明 unlockSummary 时渲染（边界）', () => {
    const project = buildProject({
      weeks: [
        {
          id: 21,
          weekNo: 1,
          title: '联调执行',
          startDate: '2026-01-05',
          endDate: '2026-01-09',
          status: 'DONE',
          allowedActions: ['unlockSummary']
        }
      ]
    })
    const wrapper = mountSection(project)

    const unlockBtns = wrapper.findAll('button').filter((b) => b.text() === '解锁总结')
    expect(unlockBtns.length).toBe(1)
  })

  it('周卡片「新建任务」受总结锁定与 edit 权限双重约束（边界）', () => {
    const wrapper = mountSection(buildProject())
    const vm = wrapper.vm as any

    expect(vm.canAddTaskToWeek({ summaryAt: null, allowedActions: ['edit'] })).toBe(true)
    expect(vm.canAddTaskToWeek({ summaryAt: '2026-01-30 18:00:00', allowedActions: ['edit'] })).toBe(false)
    expect(vm.canAddTaskToWeek({ summaryAt: null, allowedActions: ['changeStatus'] })).toBe(false)
  })

  it('周状态流转动作映射：DONE/SKIPPED 仅 reopen，未知状态为空（边界）', () => {
    const wrapper = mountSection(buildProject())
    const vm = wrapper.vm as any

    expect(vm.statusActionsFor({ status: 'NOT_STARTED' }).map((a: any) => a.value)).toEqual(['start', 'skip'])
    expect(vm.statusActionsFor({ status: 'DONE' }).map((a: any) => a.value)).toEqual(['reopen'])
    expect(vm.statusActionsFor({ status: 'SKIPPED' }).map((a: any) => a.value)).toEqual(['reopen'])
    expect(vm.statusActionsFor({ status: 'BOGUS' })).toEqual([])
  })

  it('无周计划时展示空态提示（反例：空数据不渲染时间轴）', () => {
    const wrapper = mountSection(buildProject({ weeks: [] }))

    expect(wrapper.text()).toContain('暂无周计划')
    expect(wrapper.findAll('.week-item').length).toBe(0)
  })
})
