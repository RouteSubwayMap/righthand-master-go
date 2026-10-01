/**
 * 按钮权限自定义指令
 * v-permission="'system:user:add'"
 * v-permission="['system:user:add', 'system:user:edit']"
 */
import { hasPermission, hasAnyPermission } from '@/utils/permission'

export default {
  mounted(el, binding) {
    const { value } = binding
    
    if (value) {
      // 支持单个权限或权限数组
      const permissions = Array.isArray(value) ? value : [value]
      
      // 检查是否有权限
      const hasAuth = hasAnyPermission(permissions)
      
      // 如果没有权限，移除元素
      if (!hasAuth) {
        el.parentNode && el.parentNode.removeChild(el)
      }
    }
  }
}
