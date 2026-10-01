import request from '@/utils/request'

/**
 * 获取所有用户列表（使用分页接口获取全部数据）
 */
export function getAllUsers() {
  return request({
    url: '/api/system/users/page',
    method: 'get',
    params: {
      pageNum: 1,
      pageSize: 10000
    }
  }).then(res => {
    // 将分页结果转换为数组格式，兼容原有接口
    if (res.code === 200 && res.data) {
      return {
        code: 200,
        data: res.data.list || res.data || []
      }
    }
    return res
  })
}

/**
 * 根据 ID 获取用户详情
 */
export function getUserById(id) {
  return request({
    url: `/api/system/users/${id}`,
    method: 'get'
  })
}

/**
 * 搜索用户
 */
export function searchUsers(keyword) {
  return request({
    url: `/api/system/users/search?keyword=${encodeURIComponent(keyword)}`,
    method: 'get'
  })
}

/**
 * 修改密码
 */
export function changePassword(data) {
  return request({
    url: '/api/system/users/change-password',
    method: 'post',
    data
  })
}

/**
 * 重置密码（管理员操作）
 */
export function resetPassword(id) {
  return request({
    url: `/api/system/users/${id}/reset-password`,
    method: 'put'
  })
}
