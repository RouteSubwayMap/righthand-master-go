import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import ProjectList from '@/views/project/ProjectList.vue'
import { getProjectList, deleteProject } from '@/api/project'
import { getAllUsers } from '@/api/user'
import { ElMessageBox } from 'element-plus'

// 沙箱/CI 环境较慢，放大超时避免真实组件挂载（Element Plus）误报超时
vi.setConfig({ testTimeout: 90000, hookTimeout: 90000 })

const { pushMock } = vi.hoisted(() => ({ pushMock: vi.fn() }))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {}, name: 'ProjectListPage' })
}))

vi.mock('@/api/project', () => ({
  getProjectList: vi.fn(),
  createProject: vi.fn(),
  updateProject: vi.fn(),
  deleteProject: vi.fn(),
  changeProjectStatus: vi.fn()
}))

vi.mock('@/api/user', () => ({ getAllUsers: vi.fn() }))

vi.mock('element-plus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('element-plus')>()
  return {
    ...actual,
    ElMessage: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
    ElMessageBox: { confirm: vi.fn(), prompt: vi.fn() }
  }
})

const inProgressRow = {
  id: 1,
  projectName: '新一代信贷核心系统对接',
  projectKey: 'NCBC',
  managerName: '管理员',
  status: 'IN_PROGRESS',
  riskLevel: 'MEDIUM',
  deliveryStatus: 'NORMAL',
  progressPct: 40,
  weekCount: 5,
  doneWeekCount: 2,
  startDate: '2026-01-05',
  endDate: '2026-02-06',
  updateTime: '2026-01-20 10:00:00',
  currentStage: '开发阶段',
  allowedActions: ['pause', 'close', 'edit', 'terminate', 'delete']
}

const startingRow = {
  id: 2,
  projectName: '中间业务平台二期',
  projectKey: 'DATA-MID-2',
  managerName: '普通用户',
  status: 'STARTING',
  riskLevel: 'LOW',
  deliveryStatus: 'NORMAL',
  progressPct: 0,
  weekCount: 0,
  doneWeekCount: 0,
  startDate: '2026-03-01',
  endDate: '2026-04-30',
  updateTime: '2026-01-18 09:00:00',
  currentStage: '',
  allowedActions: ['start', 'edit', 'delete']
}

async function mountList() {
  const wrapper = mount(ProjectList)
  await flushPromises()
  await nextTick()
  return wrapper
}

describe('ProjectList 项目清单页', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(getProjectList as any).mockResolvedValue({
      code: 200,
      data: { list: [inProgressRow, startingRow], total: 2, pageNum: 1, pageSize: 10 }
    })
    ;(getAllUsers as any).mockResolvedValue({
      code: 200,
      data: [
        { id: 1, username: 'admin', realName: '管理员' },
        { id: 2, username: 'user', realName: '普通用户' }
      ]
    })
  })

  it('挂载后按默认排序加载列表：endDate_asc 拆分为 sortBy/order', async () => {
    await mountList()

    expect(getProjectList).toHaveBeenCalledTimes(1)
    expect((getProjectList as any).mock.calls[0][0]).toMatchObject({
      pageNum: 1,
      pageSize: 10,
      sortBy: 'endDate',
      order: 'asc'
    })
  })

  it('列表渲染项目名称/Key/状态标签/进度/总数', async () => {
    const wrapper = await mountList()
    const text = wrapper.text()

    expect(text).toContain('新一代信贷核心系统对接')
    expect(text).toContain('NCBC')
    expect(text).toContain('进行中')
    expect(text).toContain('启动')
    expect(text).toContain('2/5 周已完成')
    expect(text).toContain('共 2 个')
  })

  it('「当前阶段」列位于「进度」与「最近更新」之间，按行渲染且空值显示 -（正例）', async () => {
    const wrapper = await mountList()

    const headers = wrapper.findAll('.el-table__header th').map((th) => th.text())
    const iProgress = headers.indexOf('进度')
    const iStage = headers.indexOf('当前阶段')
    const iUpdate = headers.indexOf('最近更新')
    expect(iProgress).toBeGreaterThan(-1)
    expect(iStage).toBe(iProgress + 1)
    expect(iUpdate).toBe(iStage + 1)

    const rows = wrapper.findAll('.el-table__row')
    expect(rows[0].findAll('.el-tag').map((t) => t.text())).toContain('开发阶段')
    expect(rows[1].findAll('.el-tag').map((t) => t.text())).not.toContain('开发阶段')
    expect(rows[1].text()).toContain('-')
  })

  it('操作按钮按 allowedActions 渲染：前 2 个为主按钮，剩余进「更多」', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any

    // 排序优先级：pause→close→edit→terminate→delete（更早的进主按钮位）
    expect(vm.primaryActions(inProgressRow)).toEqual(['pause', 'close'])
    expect(vm.moreActions(inProgressRow)).toEqual(['edit', 'terminate', 'delete'])
    expect(vm.primaryActions(startingRow)).toEqual(['start', 'edit'])
    expect(vm.moreActions(startingRow)).toEqual(['delete'])

    // DOM：主按钮文案可见；「删除」只出现在更多下拉（未展开不渲染）
    const buttonTexts = wrapper.findAll('button').map((b) => b.text())
    expect(buttonTexts).toContain('暂停')
    expect(buttonTexts).toContain('结项')
    expect(buttonTexts).toContain('开始')
    expect(buttonTexts.filter((t) => t === '编辑').length).toBe(1)
    expect(buttonTexts.filter((t) => t === '删除').length).toBe(0)
    expect(buttonTexts.filter((t) => t.includes('更多')).length).toBe(2)
  })

  it('删除确认通过后调用接口并刷新列表', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any
    ;(ElMessageBox.confirm as any).mockResolvedValue('confirm')
    ;(deleteProject as any).mockResolvedValue({ code: 200 })
    const callsBefore = (getProjectList as any).mock.calls.length

    await vm.handleProjectAction(startingRow, 'delete')
    await flushPromises()

    expect(deleteProject).toHaveBeenCalledWith(2)
    expect((getProjectList as any).mock.calls.length).toBe(callsBefore + 1)
  })

  it('删除确认被取消时不调用删除接口（取消保护）', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any
    ;(ElMessageBox.confirm as any).mockRejectedValue('cancel')

    await vm.handleProjectAction(startingRow, 'delete')

    expect(deleteProject).not.toHaveBeenCalled()
  })

  it('allowedActions 为空或缺失时不渲染任何操作按钮（防御）', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any

    expect(vm.primaryActions({ allowedActions: [] })).toEqual([])
    expect(vm.moreActions({ allowedActions: [] })).toEqual([])
    expect(vm.sortedActions({})).toEqual([])
  })

  it('allowedActions 中的未知动作被过滤，不渲染后端未声明的操作', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any

    expect(vm.sortedActions({ allowedActions: ['hack', 'pause', 'unknown', 'delete'] })).toEqual([
      'pause',
      'delete'
    ])
  })

  it('接口失败时 loading 复位且不写入脏数据（不悬挂）', async () => {
    ;(getProjectList as any).mockRejectedValueOnce(new Error('Error'))

    const wrapper = await mountList()
    const vm = wrapper.vm as any

    expect(vm.loading).toBe(false)
    expect(vm.list).toEqual([])
    expect(vm.total).toBe(0)
  })

  it('筛选参数组装：状态多选逗号拼接、参与人单选、日期范围两段才生效', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any

    vm.query.statusList = ['IN_PROGRESS', 'PAUSED']
    vm.query.participantId = 2
    vm.dateRange = ['2026-01-01', '2026-02-01']
    vm.handleSearch()
    await flushPromises()

    const params = (getProjectList as any).mock.calls.at(-1)[0]
    expect(params.status).toBe('IN_PROGRESS,PAUSED')
    expect(params.participantId).toBe(2)
    expect(params.startDate).toBe('2026-01-01')
    expect(params.endDate).toBe('2026-02-01')
    expect(params.pageNum).toBe(1)
  })

  it('未选择参与人/项目经理时不携带对应参数（防御）', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any

    vm.query.participantId = null
    vm.query.managerId = null
    vm.handleSearch()
    await flushPromises()

    const params = (getProjectList as any).mock.calls.at(-1)[0]
    expect(params.participantId).toBeUndefined()
    expect(params.managerId).toBeUndefined()
  })

  it('日期范围仅选一段时不携带 startDate/endDate（边界）', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any

    vm.dateRange = ['2026-01-01']
    vm.handleSearch()
    await flushPromises()

    const params = (getProjectList as any).mock.calls.at(-1)[0]
    expect(params.startDate).toBeUndefined()
    expect(params.endDate).toBeUndefined()
  })

  it('工期天数按自然日闭区间计算，缺参返回 0（边界）', async () => {
    const wrapper = await mountList()
    const vm = wrapper.vm as any

    expect(vm.durationDays('2026-01-30', '2026-02-02')).toBe(4)
    expect(vm.durationDays('', '2026-02-02')).toBe(0)
    expect(vm.durationDays('2026-01-01', '')).toBe(0)
  })
})
