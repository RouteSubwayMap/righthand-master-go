import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import TaskSelectDialog from '@/components/TaskSelectDialog.vue'
import { getTaskList } from '@/api/task'
import { ElMessage } from 'element-plus'

// 沙箱/CI 环境较慢，放大超时避免真实组件挂载（Element Plus）误报超时
vi.setConfig({ testTimeout: 90000, hookTimeout: 90000 })

vi.mock('@/api/task', () => ({ getTaskList: vi.fn() }))

vi.mock('element-plus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('element-plus')>()
  return {
    ...actual,
    ElMessage: { success: vi.fn(), error: vi.fn(), warning: vi.fn() }
  }
})

const recordA = {
  id: 12,
  title: '修复登录bug',
  description: '验证码失效问题',
  itemType: 'TASK',
  priority: 'HIGH',
  statusCode: 'TODO',
  deadline: '2026-05-22T12:00:00+08:00',
  progress: 0
}

const recordB = {
  id: 13,
  title: '编制周报',
  description: '',
  itemType: 'MEMO',
  priority: 'LOW',
  statusCode: 'DONE',
  deadline: null,
  progress: 100
}

function mountDialog(props: Record<string, any> = {}) {
  // 注意：弹窗内容经 Teleport 挂载到 body，单元测试聚焦状态与数据链路，
  // 表格渲染与真实点击由浏览器联调验证覆盖
  return mount(TaskSelectDialog, {
    props: { modelValue: true, ...props }
  })
}

describe('TaskSelectDialog 事项选择弹窗', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ;(getTaskList as any).mockResolvedValue({ code: 200, data: { records: [recordA, recordB], total: 2 } })
  })

  it('Open_DefaultPage_LoadFirstPage：打开即按默认分页加载并渲染记录与总数', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    const vm = wrapper.vm as any

    expect(getTaskList).toHaveBeenCalledWith({ pageNum: 1, pageSize: 10 })
    expect(vm.list.map((t: any) => t.title)).toEqual(['修复登录bug', '编制周报'])
    expect(vm.pagination.total).toBe(2)
  })

  it('Confirm_PresetRow_EmitsSelectAndCloses：预选事项回显，确定后抛出 select 并关闭', async () => {
    const wrapper = mountDialog({ selectedId: 12 })
    await flushPromises()
    const vm = wrapper.vm as any

    expect(vm.selectedId).toBe(12)

    vm.handleConfirm()
    await flushPromises()

    const emitted = wrapper.emitted('select') as any[]
    expect(emitted.length).toBe(1)
    expect(emitted[0][0].id).toBe(12)
    const modelEvents = wrapper.emitted('update:modelValue') as any[]
    expect(modelEvents[modelEvents.length - 1][0]).toBe(false)
  })

  it('Confirm_OtherRow_EmitsThatRow：勾选其他行后确定，抛出所选行', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    const vm = wrapper.vm as any

    vm.handleRowClick(recordB)
    await flushPromises()
    expect(vm.selectedId).toBe(13)

    vm.handleConfirm()
    await flushPromises()

    const emitted = wrapper.emitted('select') as any[]
    expect(emitted.length).toBe(1)
    expect(emitted[0][0].title).toBe('编制周报')
  })

  it('DblClick_Row_EmitsImmediately：双击行直接确认选择', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    const vm = wrapper.vm as any

    vm.handleRowDblclick(recordA)
    await flushPromises()

    const emitted = wrapper.emitted('select') as any[]
    expect(emitted.length).toBe(1)
    expect(emitted[0][0].id).toBe(12)
  })

  it('Search_Params_EmptyKeysOmittedAndPageReset：筛选参数有值携带、空值不携带且回到第 1 页', async () => {
    const wrapper = mountDialog({ itemType: 'TASK' })
    await flushPromises()

    expect(getTaskList).toHaveBeenCalledWith({ pageNum: 1, pageSize: 10, itemType: 'TASK' })

    const vm = wrapper.vm as any
    vm.pagination.pageNum = 3
    vm.filters.title = '登录'
    vm.filters.statusCode = 'TODO'
    vm.handleSearch()
    await flushPromises()

    expect(getTaskList).toHaveBeenLastCalledWith({
      pageNum: 1,
      pageSize: 10,
      itemType: 'TASK',
      title: '登录',
      statusCode: 'TODO'
    })
  })

  it('Confirm_NoSelection_WarnsStaysOpen：未选择事项时确定，提示且不抛出 select、不关闭（反例）', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    const vm = wrapper.vm as any

    vm.handleConfirm()
    await flushPromises()

    expect(ElMessage.warning).toHaveBeenCalled()
    expect(wrapper.emitted('select')).toBeFalsy()
    expect(wrapper.emitted('update:modelValue')).toBeFalsy()
  })

  it('Confirm_StaleSelectionNotInList_Warns：选中项不在当前列表时确定，提示且不抛出（反例：预选失效防御）', async () => {
    const wrapper = mountDialog({ selectedId: 999 })
    await flushPromises()
    const vm = wrapper.vm as any

    vm.handleConfirm()
    await flushPromises()

    expect(ElMessage.warning).toHaveBeenCalled()
    expect(wrapper.emitted('select')).toBeFalsy()
  })

  it('Load_ApiReject_ClearListAndResetLoading：接口失败时清空列表与总数并复位 loading，不残留旧数据（反例）', async () => {
    ;(getTaskList as any).mockRejectedValueOnce(new Error('Network Error'))

    const wrapper = mountDialog()
    await flushPromises()
    const vm = wrapper.vm as any

    expect(vm.list.length).toBe(0)
    expect(vm.pagination.total).toBe(0)
    expect(vm.loading).toBe(false)
  })

  it('Reset_RestorePresetTypeAndPageOne：重置恢复预置类型并回到第 1 页（边界）', async () => {
    const wrapper = mountDialog({ itemType: 'TASK' })
    await flushPromises()
    const vm = wrapper.vm as any

    vm.pagination.pageNum = 2
    vm.filters.title = '登录'
    vm.filters.itemType = 'MEMO'
    vm.filters.statusCode = 'DONE'
    vm.filters.priority = 'LOW'
    vm.handleReset()
    await flushPromises()

    expect(vm.filters.title).toBe('')
    expect(vm.filters.itemType).toBe('TASK')
    expect(vm.filters.statusCode).toBe('')
    expect(vm.filters.priority).toBe('')
    expect(getTaskList).toHaveBeenLastCalledWith({ pageNum: 1, pageSize: 10, itemType: 'TASK' })
  })

  it('PageChange_FetchWithNewPageNum：翻页请求携带新页码（边界）', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    const vm = wrapper.vm as any

    vm.pagination.pageNum = 2
    vm.loadList()
    await flushPromises()

    expect(getTaskList).toHaveBeenLastCalledWith({ pageNum: 2, pageSize: 10 })
  })

  it('PageSizeChange_BackToFirstPage：每页条数变化按新 pageSize 回到第 1 页重新加载（边界）', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    const vm = wrapper.vm as any

    vm.pagination.pageNum = 2
    vm.pagination.pageSize = 20
    vm.handleSearch() // el-pagination size-change 绑定 handleSearch
    await flushPromises()

    expect(getTaskList).toHaveBeenLastCalledWith({ pageNum: 1, pageSize: 20 })
  })

  it('Closed_NoRequest：未打开状态挂载不发起请求（反例：关闭态零请求）', async () => {
    const wrapper = mountDialog({ modelValue: false })
    await flushPromises()
    const vm = wrapper.vm as any

    expect(getTaskList).not.toHaveBeenCalled()
    expect(vm.list.length).toBe(0)
  })
})
