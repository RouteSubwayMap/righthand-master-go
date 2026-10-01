import request from '@/utils/request'

// 获取事项列表
export function getTaskList(params) {
  return request({
    url: '/api/tasks',
    method: 'get',
    params
  })
}

// 获取事项详情
export function getTaskDetail(id) {
  return request({
    url: `/api/tasks/${id}`,
    method: 'get'
  })
}

// 创建事项
export function createTask(data) {
  return request({
    url: '/api/tasks',
    method: 'post',
    data
  })
}

// 更新事项
export function updateTask(id, data) {
  return request({
    url: `/api/tasks/${id}`,
    method: 'put',
    data
  })
}

// 删除事项
export function deleteTask(id) {
  return request({
    url: `/api/tasks/${id}`,
    method: 'delete'
  })
}

// 更新事项状态
export function updateTaskStatus(id, data) {
  return request({
    url: `/api/tasks/${id}/status`,
    method: 'put',
    data
  })
}

// 获取仪表盘统计
export function getDashboardStats() {
  return request({
    url: '/api/dashboard/stats',
    method: 'get'
  })
}

// 转交事项
export function transferTask(id, data) {
  return request({
    url: `/api/tasks/${id}/transfer`,
    method: 'put',
    data
  })
}
