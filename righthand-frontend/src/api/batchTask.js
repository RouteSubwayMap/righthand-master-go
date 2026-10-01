import request from '@/utils/request'

// 获取批量任务列表（分页）
export function getBatchTaskList(params) {
  return request({
    url: '/api/batch-tasks',
    method: 'get',
    params: {
      pageNum: params.pageNum,
      pageSize: params.pageSize,
      bankType: params.bankType || '',
      bankName: params.bankName || '',
      systemName: params.systemName || '',
      taskType: params.taskType || '',
      taskName: params.taskName || '',
      taskDesc: params.taskDesc || '',
      upstreamSystem: params.upstreamSystem || '',
      downstreamSystem: params.downstreamSystem || ''
    }
  })
}

// 获取批量任务详情
export function getBatchTaskDetail(id) {
  return request({
    url: `/api/batch-tasks/${id}`,
    method: 'get'
  })
}

// 创建批量任务
export function createBatchTask(data) {
  return request({
    url: '/api/batch-tasks',
    method: 'post',
    data
  })
}

// 更新批量任务
export function updateBatchTask(id, data) {
  return request({
    url: `/api/batch-tasks/${id}`,
    method: 'put',
    data
  })
}

// 删除批量任务
export function deleteBatchTask(id) {
  return request({
    url: `/api/batch-tasks/${id}`,
    method: 'delete'
  })
}

// 批量导入任务
export function importBatchTasks(tasks) {
  return request({
    url: '/api/batch-tasks/import',
    method: 'post',
    data: { tasks }
  })
}
