import request from '@/utils/request'

// 获取字典列表
export function getDicts(params) {
  return request({
    url: '/api/system/dicts',
    method: 'get',
    params
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

// 获取用户列表（分页）
export function getUserPage(params) {
  return request({
    url: '/api/system/users/page',
    method: 'get',
    params
  })
}

// 获取用户列表（兼容旧接口，默认获取全部）
export function getUsers(params = {}) {
  // 如果传入了分页参数，使用分页接口
  if (params.pageNum || params.pageSize) {
    return request({
      url: '/api/system/users/page',
      method: 'get',
      params
    })
  }
  // 否则获取全部用户
  return request({
    url: '/api/system/users/page',
    method: 'get',
    params: {
      pageNum: 1,
      pageSize: 10000,
      ...params
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

// 创建用户
export function createUser(data) {
  return request({
    url: '/api/system/users',
    method: 'post',
    data
  })
}

// 更新用户
export function updateUser(id, data) {
  return request({
    url: `/api/system/users/${id}`,
    method: 'put',
    data
  })
}

// 删除用户
export function deleteUser(id) {
  return request({
    url: `/api/system/users/${id}`,
    method: 'delete'
  })
}

// 获取用户的角色列表
export function getUserRoles(userId) {
  return request({
    url: `/api/system/users/${userId}/roles`,
    method: 'get'
  })
}

// 保存或更新用户的角色
export function saveUserRoles(userId, data) {
  return request({
    url: `/api/system/users/${userId}/roles`,
    method: 'post',
    data
  })
}

// 获取所有角色列表
export function getAllRoles() {
  return request({
    url: '/api/system/roles/all',
    method: 'get'
  })
}

// ========== 角色管理 ==========

// 获取角色列表
export function getRoles(params) {
  return request({
    url: '/api/system/roles',
    method: 'get',
    params
  })
}

// 创建角色
export function createRole(data) {
  return request({
    url: '/api/system/roles',
    method: 'post',
    data
  })
}

// 更新角色
export function updateRole(id, data) {
  return request({
    url: `/api/system/roles/${id}`,
    method: 'put',
    data
  })
}

// 删除角色
export function deleteRole(id) {
  return request({
    url: `/api/system/roles/${id}`,
    method: 'delete'
  })
}

// 获取角色的菜单权限
export function getRoleMenus(roleId) {
  return request({
    url: `/api/system/roles/${roleId}/menus`,
    method: 'get'
  })
}

// 保存角色的菜单权限
export function saveRoleMenus(roleId, data) {
  return request({
    url: `/api/system/roles/${roleId}/menus`,
    method: 'post',
    data
  })
}

// ========== 菜单管理 ==========

// 获取菜单列表（树形）
export function getMenuTree() {
  return request({
    url: '/api/system/menus/tree',
    method: 'get'
  })
}

// 获取菜单列表
export function getMenus(params) {
  return request({
    url: '/api/system/menus',
    method: 'get',
    params
  })
}

// 创建菜单
export function createMenu(data) {
  return request({
    url: '/api/system/menus',
    method: 'post',
    data
  })
}

// 更新菜单
export function updateMenu(id, data) {
  return request({
    url: `/api/system/menus/${id}`,
    method: 'put',
    data
  })
}

// 删除菜单
export function deleteMenu(id) {
  return request({
    url: `/api/system/menus/${id}`,
    method: 'delete'
  })
}
