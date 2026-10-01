/**
 * 按钮权限判断工具函数
 */
import { useUserStore } from '@/stores/user'

/**
 * 检查用户是否有指定权限
 * @param {string} permission - 权限标识，如 'system:user:add'
 * @returns {boolean} - 是否有权限
 */
export function hasPermission(permission) {
  if (!permission) return false
  
  const userStore = useUserStore()
  
  // 管理员拥有所有权限
  const roles = userStore.roles || []
  const isAdmin = roles.some(role => role.roleCode === 'ADMIN')
  if (isAdmin) return true
  
  // 优先使用缓存的权限集合(O(1)查找)
  if (userStore.permissions && userStore.permissions.has(permission)) {
    return true
  }
  
  // 降级方案：从菜单中查找(兼容旧数据)
  const menus = userStore.menus || []
  return menus.some(menu => menu.permission === permission)
}

/**
 * 检查用户是否有任何一个权限
 * @param {Array<string>} permissions - 权限标识数组
 * @returns {boolean} - 是否有任一权限
 */
export function hasAnyPermission(permissions) {
  if (!permissions || permissions.length === 0) return false
  return permissions.some(permission => hasPermission(permission))
}

/**
 * 检查用户是否同时拥有所有权限
 * @param {Array<string>} permissions - 权限标识数组
 * @returns {boolean} - 是否有所有权限
 */
export function hasAllPermissions(permissions) {
  if (!permissions || permissions.length === 0) return false
  return permissions.every(permission => hasPermission(permission))
}
