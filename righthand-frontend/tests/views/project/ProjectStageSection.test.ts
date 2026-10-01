import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import ProjectStageSection from '@/views/project/components/ProjectStageSection.vue'
import { createProjectStage, updateProjectStage, deleteProjectStage } from '@/api/project'
import { ElMessageBox } from 'element-plus'

// 沙箱/CI 环境较慢，放大超时避免真实组件挂载（Element Plus）误报超时
vi.setConfig({ testTimeout: 90000, hookTimeout: 90000 })

vi.mock('@/api/project', () => ({
  createProjectStage: vi.fn(),
  updateProjectStage: vi.fn(),
  deleteProjectStage: vi.fn()
}))

vi.mock('element-plus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('element-plus')>()
  return {
    ...actual,
    ElMessage: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
    ElMessageBox: { confirm: vi.fn(), prompt: vi.fn() }
  }
})

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
    stageName: '开发阶段',
    startDate: '2026-01-21',
    endDate: '2026-02-10',
    description: ''
  }
]

function buildProject(overrides: Record<string, any> = {}) {
  return {
    id: 1,
    allowedActions: ['manageStage'],
    stages,
    currentStageId: 32,
    ...overrides
  }
}

async function mountSection(project: any = buildProject()) {
  const wrapper = mount(ProjectStageSection, { props: { project } })
  await flushPromises()
  await nextTick()
  return wrapper
}

function tableRows(wrapper: any) {
  return wrapper.findAll('.el-table__row')
}

describe('ProjectStageSection 项目阶段组件', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('阶段列表按后端正序渲染、标题统计正确；空数组展示空态（正例+对照）', async () => {
    const wrapper = await mountSection()

    const rows = tableRows(wrapper)
    expect(rows.length).toBe(2)
    expect(rows[0].text()).toContain('需求调研')
    expect(rows[1].text()).toContain('开发阶段')
    expect(wrapper.text()).toContain('共 2 个阶段')

    const empty = await mountSection(buildProject({ stages: [], currentStageId: null }))
    expect(empty.text()).toContain('暂无项目阶段')
    expect(tableRows(empty).length).toBe(0)
  })

  it('命中 currentStageId 的行高亮并展示「当前阶段」标签，其余行不受影响（正例）', async () => {
    const wrapper = await mountSection()

    const rows = tableRows(wrapper)
    expect(rows[1].classes()).toContain('current-stage-row')
    expect(rows[0].classes()).not.toContain('current-stage-row')

    const currentTags = wrapper.findAll('.el-tag').filter((t) => t.text() === '当前阶段')
    expect(currentTags.length).toBe(1)
  })

  it('currentStageId 缺省时所有行均不高亮、无标签（反例：无命中/终态项目）', async () => {
    const wrapper = await mountSection(buildProject({ currentStageId: null }))

    const rows = tableRows(wrapper)
    expect(rows.length).toBe(2)
    rows.forEach((row) => expect(row.classes()).not.toContain('current-stage-row'))
    expect(wrapper.findAll('.el-tag').filter((t) => t.text() === '当前阶段').length).toBe(0)
  })

  it('新增按钮与操作列仅由 allowedActions.manageStage 驱动（权限）', async () => {
    const withPerm = await mountSection(buildProject({ allowedActions: ['manageStage'] }))
    expect(withPerm.text()).toContain('新增阶段')
    expect(withPerm.text()).toContain('操作')
    expect(withPerm.findAll('button').some((b) => b.text() === '编辑')).toBe(true)
    expect(withPerm.findAll('button').some((b) => b.text() === '删除')).toBe(true)

    const withoutPerm = await mountSection(buildProject({ allowedActions: [] }))
    expect(withoutPerm.text()).not.toContain('新增阶段')
    expect(withoutPerm.text()).not.toContain('操作')
    expect(
      withoutPerm.findAll('button').some((b) => b.text() === '编辑' || b.text() === '删除')
    ).toBe(false)
  })

  it('新增与编辑提交成功后关闭弹窗、emit refresh 且入参正确（正例）', async () => {
    const wrapper = await mountSection()
    const vm = wrapper.vm as any
    ;(createProjectStage as any).mockResolvedValue({ code: 200 })
    ;(updateProjectStage as any).mockResolvedValue({ code: 200 })

    // 新增
    vm.openCreateDialog()
    await nextTick()
    await flushPromises()
    vm.stageForm.stageName = '联调阶段'
    vm.stageForm.startDate = '2026-10-01'
    vm.stageForm.endDate = '2026-10-10'
    await vm.handleSubmit()
    await flushPromises()

    expect(createProjectStage).toHaveBeenCalledTimes(1)
    expect((createProjectStage as any).mock.calls[0][0]).toBe(1)
    expect((createProjectStage as any).mock.calls[0][1]).toMatchObject({
      stageName: '联调阶段',
      startDate: '2026-10-01',
      endDate: '2026-10-10'
    })
    expect(vm.dialogVisible).toBe(false)
    expect(wrapper.emitted('refresh')?.length).toBe(1)

    // 编辑
    vm.openEditDialog(stages[0])
    await nextTick()
    await flushPromises()
    vm.stageForm.stageName = '需求调研（一期）'
    await vm.handleSubmit()
    await flushPromises()

    expect(updateProjectStage).toHaveBeenCalledTimes(1)
    expect((updateProjectStage as any).mock.calls[0][0]).toBe(1)
    expect((updateProjectStage as any).mock.calls[0][1]).toBe(31)
    expect((updateProjectStage as any).mock.calls[0][2]).toMatchObject({
      stageName: '需求调研（一期）'
    })
    expect(wrapper.emitted('refresh')?.length).toBe(2)
  })

  it('阶段名称留空时校验拦截：不发请求且弹窗保持打开（空值）', async () => {
    const wrapper = await mountSection()
    const vm = wrapper.vm as any
    ;(createProjectStage as any).mockResolvedValue({ code: 200 })

    vm.openCreateDialog()
    await nextTick()
    await flushPromises()
    vm.stageForm.startDate = '2026-10-01'
    vm.stageForm.endDate = '2026-10-10'
    await vm.handleSubmit()
    await flushPromises()

    expect(createProjectStage).not.toHaveBeenCalled()
    expect(vm.dialogVisible).toBe(true)
  })

  it('结束日期早于开始日期时校验拦截并提示（边界）', async () => {
    const wrapper = await mountSection()
    const vm = wrapper.vm as any
    ;(createProjectStage as any).mockResolvedValue({ code: 200 })

    vm.openCreateDialog()
    await nextTick()
    await flushPromises()
    vm.stageForm.stageName = '倒置阶段'
    vm.stageForm.startDate = '2026-06-10'
    vm.stageForm.endDate = '2026-06-09'
    await vm.handleSubmit()
    await flushPromises()
    await nextTick()

    expect(createProjectStage).not.toHaveBeenCalled()
    expect(vm.dialogVisible).toBe(true)

    // 提示文案由 endDate 自定义校验规则产生（DOM 错误提示在 happy-dom 下渲染时机不稳，直接断言规则输出）
    const endRule = vm.stageRules.endDate.find((r: any) => typeof r.validator === 'function')
    const message = await new Promise<string>((resolve) => {
      endRule.validator({}, '2026-06-09', (err: any) => resolve(err ? err.message : ''))
    })
    expect(message).toBe('结束日期不能早于开始日期')
  })

  it('接口异常（500）时弹窗保持打开、saving 复位且不刷新（异常）', async () => {
    const wrapper = await mountSection()
    const vm = wrapper.vm as any
    ;(createProjectStage as any).mockRejectedValueOnce(new Error('Request failed with status code 500'))

    vm.openCreateDialog()
    await nextTick()
    await flushPromises()
    vm.stageForm.stageName = '失败阶段'
    vm.stageForm.startDate = '2026-10-01'
    vm.stageForm.endDate = '2026-10-10'
    await vm.handleSubmit()
    await flushPromises()

    expect(vm.dialogVisible).toBe(true)
    expect(vm.saving).toBe(false)
    expect(wrapper.emitted('refresh')).toBeUndefined()
  })

  it('删除确认取消不发请求；确认后调用删除接口并刷新列表（反例+链路）', async () => {
    const wrapper = await mountSection()
    const vm = wrapper.vm as any

    // 取消：零请求
    ;(ElMessageBox.confirm as any).mockRejectedValueOnce('cancel')
    await vm.handleDelete(stages[0])
    await flushPromises()
    expect(deleteProjectStage).not.toHaveBeenCalled()
    expect(wrapper.emitted('refresh')).toBeUndefined()

    // 确认：调用接口 + 刷新
    ;(ElMessageBox.confirm as any).mockResolvedValueOnce('confirm')
    ;(deleteProjectStage as any).mockResolvedValue({ code: 200 })
    await vm.handleDelete(stages[0])
    await flushPromises()
    expect(deleteProjectStage).toHaveBeenCalledWith(1, 31)
    expect(wrapper.emitted('refresh')?.length).toBe(1)
  })
})
