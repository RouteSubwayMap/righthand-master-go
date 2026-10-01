import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as projectApi from '@/api/project'

// 捕获请求配置（替代真实 axios 实例），验证各 API 的 URL / method / 参数拼装契约
const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }))

vi.mock('@/utils/request', () => ({ default: requestMock }))

describe('项目管理 API 契约', () => {
  beforeEach(() => {
    requestMock.mockReset()
    requestMock.mockResolvedValue({ code: 200, message: 'ok', data: {} })
  })

  describe('正例：请求拼装与响应透传', () => {
    it('列表查询使用 GET /api/projects 并透传分页筛选参数，响应原样透传', async () => {
      const resp = { code: 200, message: 'ok', data: { list: [], total: 0 } }
      requestMock.mockResolvedValueOnce(resp)
      const params = {
        pageNum: 1,
        pageSize: 10,
        status: 'IN_PROGRESS,PAUSED',
        sortBy: 'endDate',
        order: 'asc'
      }

      const result = await projectApi.getProjectList(params)

      expect(requestMock).toHaveBeenCalledWith({ url: '/api/projects', method: 'get', params })
      expect(result).toBe(resp)
    })

    it('新建项目使用 POST 并原样透传 data（中文项目名不丢失）', async () => {
      const data = {
        projectName: '新一代信贷核心系统对接',
        projectKey: 'NCBC',
        startDate: '2026-01-05',
        endDate: '2026-02-06',
        managerId: 1
      }

      await projectApi.createProject(data)

      expect(requestMock).toHaveBeenCalledWith({ url: '/api/projects', method: 'post', data })
    })

    it('详情/更新/删除共用 /api/projects/:id 路径，方法分别为 get/put/delete', async () => {
      await projectApi.getProjectDetail(5)
      await projectApi.updateProject(5, { planDesc: '整体计划调整' })
      await projectApi.deleteProject(5)

      expect(requestMock).toHaveBeenNthCalledWith(1, { url: '/api/projects/5', method: 'get' })
      expect(requestMock).toHaveBeenNthCalledWith(2, {
        url: '/api/projects/5',
        method: 'put',
        data: { planDesc: '整体计划调整' }
      })
      expect(requestMock).toHaveBeenNthCalledWith(3, { url: '/api/projects/5', method: 'delete' })
    })

    it('终止流转必带 action 与 reason', async () => {
      await projectApi.changeProjectStatus(5, { action: 'terminate', reason: '预算取消' })

      expect(requestMock).toHaveBeenCalledWith({
        url: '/api/projects/5/status',
        method: 'put',
        data: { action: 'terminate', reason: '预算取消' }
      })
    })
  })

  describe('反例：错误透传与契约负例', () => {
    it('request 层 reject 时 API 不吞错不包装，原样向上抛', async () => {
      const err = new Error('Error')
      requestMock.mockRejectedValueOnce(err)

      await expect(projectApi.getProjectList({ pageNum: 1 })).rejects.toBe(err)
    })

    it('提交周总结必须是嵌套 summary 对象，不得拍平', async () => {
      await projectApi.submitWeekSummary(5, 12, {
        summary: { finished: '完成A', nextPlan: 'B', risks: '', support: '' }
      })

      const cfg = requestMock.mock.calls[0][0]
      expect(cfg.url).toBe('/api/projects/5/weeks/12/summary')
      expect(cfg.method).toBe('post')
      expect(cfg.data.summary.finished).toBe('完成A')
      expect(cfg.data.finished).toBeUndefined()
    })

    it('解锁总结为 PUT 且不携带 body（避免空对象污染请求）', async () => {
      await projectApi.unlockWeekSummary(5, 12)

      expect(requestMock).toHaveBeenCalledWith({
        url: '/api/projects/5/weeks/12/summary/unlock',
        method: 'put'
      })
    })
  })

  describe('边界：参数缺省与 URL 组合', () => {
    it('列表查询不传参时不伪造默认分页参数', async () => {
      await projectApi.getProjectList()

      expect(requestMock).toHaveBeenCalledWith({ url: '/api/projects', method: 'get', params: undefined })
    })

    it('任务-周关联同 URL 以 POST/DELETE 区分加入与移出', async () => {
      await projectApi.linkTaskToWeek(5, 12, 88)
      await projectApi.unlinkTaskFromWeek(5, 12, 88)

      const [call1, call2] = requestMock.mock.calls.map((c) => c[0])
      expect(call1.url).toBe('/api/projects/5/weeks/12/tasks/88')
      expect(call2.url).toBe(call1.url)
      expect(call1.method).toBe('post')
      expect(call2.method).toBe('delete')
    })

    it('周计划接口项目 id 与 weekId 同时拼接正确', async () => {
      await projectApi.createWeekPlan(8, { title: '第1周' })
      await projectApi.changeWeekStatus(8, 14, { action: 'done' })
      await projectApi.updateWeekPlan(8, 14, { title: '第1周（修订）' })
      await projectApi.deleteWeekPlan(8, 14)

      expect(requestMock).toHaveBeenNthCalledWith(1, {
        url: '/api/projects/8/weeks',
        method: 'post',
        data: { title: '第1周' }
      })
      expect(requestMock).toHaveBeenNthCalledWith(2, {
        url: '/api/projects/8/weeks/14/status',
        method: 'put',
        data: { action: 'done' }
      })
      expect(requestMock).toHaveBeenNthCalledWith(3, {
        url: '/api/projects/8/weeks/14',
        method: 'put',
        data: { title: '第1周（修订）' }
      })
      expect(requestMock).toHaveBeenNthCalledWith(4, { url: '/api/projects/8/weeks/14', method: 'delete' })
    })

    it('任务查询缺省 weekPlanId 时不注入该字段（「全部」视图语义）', async () => {
      await projectApi.getProjectTasks(5, { pageNum: 1, pageSize: 20 })

      expect(requestMock).toHaveBeenCalledWith({
        url: '/api/projects/5/tasks',
        method: 'get',
        params: { pageNum: 1, pageSize: 20 }
      })
    })

    it('移动任务到其他周使用 PUT /tasks/:taskId/week 携带 weekPlanId', async () => {
      await projectApi.moveTaskToWeek(5, 88, { weekPlanId: 14 })

      expect(requestMock).toHaveBeenCalledWith({
        url: '/api/projects/5/tasks/88/week',
        method: 'put',
        data: { weekPlanId: 14 }
      })
    })
  })
})
