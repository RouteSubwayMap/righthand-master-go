import request from '@/utils/request'

/**
 * 需求管理 API (Go后端版本)
 */

// 分页查询需求列表
export function getRequirementPage(params) {
  // 保持原始参数格式，后端期望接收 pageNum 和 pageSize
  const apiParams = {
    ...params,
    pageNum: params.pageNum || 1,
    pageSize: params.pageSize || 10
  }
  
  return request({
    url: '/api/requirements',
    method: 'get',
    params: apiParams
  }).then(res => {
    // 处理响应为空的情况
    if (!res || typeof res !== 'object') {
      return {
        code: 500,
        message: '服务器响应异常',
        data: {
          records: [],
          total: 0,
          pageNum: 1,
          pageSize: 10
        }
      }
    }
    
    // 转换响应格式: list -> records
    if (res.code === 200 && res.data) {
      return {
        ...res,
        data: {
          records: res.data.list || res.data.records || [],
          total: res.data.total || 0,
          pageNum: res.data.pageNum || res.data.page || 1,
          pageSize: res.data.pageSize || 10
        }
      }
    }
    
    // 处理其他情况，返回默认空数据
    return {
      ...res,
      data: {
        records: [],
        total: 0,
        pageNum: 1,
        pageSize: 10
      }
    }
  }).catch(error => {
    // 处理网络错误或请求异常
    console.error('API请求失败:', error)
    return {
      code: 500,
      message: error.message || '网络请求失败',
      data: {
        records: [],
        total: 0,
        pageNum: 1,
        pageSize: 10
      }
    }
  })
}

// 获取需求详情
export function getRequirementDetail(id) {
  return request({
    url: `/api/requirements/${id}`,
    method: 'get'
  })
}

// 创建需求
export function createRequirement(data) {
  return request({
    url: '/api/requirements',
    method: 'post',
    data
  })
}

// 获取需求关联的事项列表
export function getRequirementTasks(id) {
  return request({
    url: `/api/requirements/${id}/tasks`,
    method: 'get'
  })
}

// 更新需求
export function updateRequirement(id, data) {
  return request({
    url: `/api/requirements/${id}`,
    method: 'put',
    data
  })
}

// 删除需求
export function deleteRequirement(id) {
  return request({
    url: `/api/requirements/${id}`,
    method: 'delete'
  })
}

// 更新需求状态
export function updateRequirementStatus(id, status) {
  return request({
    url: `/api/requirements/${id}/status`,
    method: 'put',
    data: { status }
  })
}

// 执行需求状态流转(适配器,兼容Java版的工作流transition接口)
export function executeRequirementTransition(id, transitionId, comment = '') {
  // Go后端直接使用updateRequirementStatus,transitionId映射到status
  // 这里需要根据transitionId转换为对应的status值
  // 由于Go后端简化了状态流转,我们直接调用updateRequirementStatus
  // transitionId 应该是状态值本身,如 'PENDING_CLARIFY', 'DEVELOPING' 等
  return updateRequirementStatus(id, transitionId)
}

// 批量导入需求
export function batchImport(requirements) {
  return request({
    url: '/api/requirements/import',
    method: 'post',
    data: { requirements }
  })
}

// 导出需求列表
export function exportRequirements(params) {
  return request({
    url: '/api/requirements/export',
    method: 'get',
    params
  }).then(res => {
    // Go后端目前返回JSON格式,前端可以转换为CSV或Excel
    // TODO: 后续可以实现真正的Excel导出
    if (res.code === 200 && res.data) {
      return {
        ...res,
        data: {
          requirements: res.data.data || [],
          count: res.data.count || 0
        }
      }
    }
    return res
  })
}
