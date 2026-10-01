import request from '@/utils/request'

// 获取今日代办
export function getTodayTasks(params) {
  return request({
    url: '/api/dashboard/today',
    method: 'get',
    params
  })
}

// 获取明日代办
export function getTomorrowTasks(params) {
  return request({
    url: '/api/dashboard/tomorrow',
    method: 'get',
    params
  })
}

// 获取未完成事项
export function getUnfinishedTasks(params) {
  return request({
    url: '/api/dashboard/unfinished',
    method: 'get',
    params
  })
}

// 获取本周代办
export function getThisWeekTasks(params) {
  return request({
    url: '/api/dashboard/thisWeek',
    method: 'get',
    params
  })
}

// 获取统计数据
export function getStatistics() {
  return request({
    url: '/api/dashboard/statistics',
    method: 'get'
  })
}
