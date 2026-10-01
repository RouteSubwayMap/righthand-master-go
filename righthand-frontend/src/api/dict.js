import request from '@/utils/request'

/**
 * 字典管理 API
 */

// 分页查询字典列表
export function getDictPage(params) {
  return request({
    url: '/api/system/dicts',
    method: 'get',
    params
  })
}

// 根据字典类型获取字典列表
export function getDictsByType(dictType) {
  return request({
    url: `/api/system/dicts/type/${dictType}`,
    method: 'get'
  })
}

// 获取字典详情
export function getDictDetail(id) {
  return request({
    url: `/api/system/dicts/${id}`,
    method: 'get'
  })
}

// 创建字典
export function createDict(data) {
  return request({
    url: '/api/system/dicts',
    method: 'post',
    data
  })
}

// 更新字典
export function updateDict(id, data) {
  return request({
    url: `/api/system/dicts/${id}`,
    method: 'put',
    data
  })
}

// 删除字典
export function deleteDict(id) {
  return request({
    url: `/api/system/dicts/${id}`,
    method: 'delete'
  })
}

// 获取所有字典数据(用于缓存)
export function getAllDicts() {
  return request({
    url: '/api/system/dicts',
    method: 'get',
    params: { pageNum: 1, pageSize: 1000 }
  })
}
