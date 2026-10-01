import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as apiLogin, getUserInfo as apiGetUserInfo } from '@/api/auth'
import { useDictStore } from './dict'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('righthand_token') || '')
  const savedUserInfo = JSON.parse(localStorage.getItem('righthand_user_info') || '{}')
  const userInfo = ref(savedUserInfo)
  const products = ref([]) // 用户可访问的产品列表
  const menus = ref([]) // 用户菜单
  const roles = ref(savedUserInfo.roles || []) // 用户角色（从localStorage恢复）
  const permissions = ref(new Set()) // 权限集合(用于快速查找)

  function setToken(newToken) {
    token.value = newToken
    localStorage.setItem('righthand_token', newToken)
  }

  function setUserInfo(info) {
    userInfo.value = info
    localStorage.setItem('righthand_user_info', JSON.stringify(info))
    // 如果 userInfo 中包含 roles，也需要同步到 roles
    if (info.roles && Array.isArray(info.roles)) {
      console.log('从 userInfo 中提取角色数据:', info.roles)
      setRoles(info.roles)
    }
  }

  function setProducts(productList) {
    products.value = productList
  }

  function setMenus(menuList) {
    menus.value = menuList
  }

  function setRoles(roleList) {
    roles.value = roleList
  }

  // 从菜单中提取所有权限标识
  function extractPermissions(menuList) {
    const perms = new Set()
    
    function traverse(menus) {
      menus.forEach(menu => {
        if (menu.permission) {
          perms.add(menu.permission)
        }
        if (menu.children && menu.children.length > 0) {
          traverse(menu.children)
        }
      })
    }
    
    traverse(menuList)
    return perms
  }

  // 设置菜单并自动提取权限
  function setMenusWithPermissions(menuList) {
    menus.value = menuList
    permissions.value = extractPermissions(menuList)
    // 持久化权限到localStorage
    localStorage.setItem('righthand_permissions', JSON.stringify([...permissions.value]))
  }

  // 登录方法
  async function login(loginForm) {
    const res = await apiLogin({
      username: loginForm.username,
      password: loginForm.password
    })
    if (res.code === 200) {
      setToken(res.data.token)
      return res
    }
    throw new Error(res.message || '登录失败')
  }

  // 获取用户信息
  async function getUserInfo() {
    const res = await apiGetUserInfo()
    console.log('获取用户信息响应:', res)
    if (res.code === 200) {
      setUserInfo(res.data)
      // 设置用户可访问的产品列表
      if (res.data.products && Array.isArray(res.data.products)) {
        setProducts(res.data.products)
      }
      // 设置用户角色 - 优先从 roles 字段读取，如果没有则尝试从 userInfo.roles 读取
      let roleData = res.data.roles
      if (!roleData || roleData.length === 0) {
        // 尝试从 userInfo 中读取（有些情况下 roles 可能在 userInfo 对象内）
        roleData = res.data.userInfo?.roles || []
      }
      console.log('用户角色数据:', roleData)
      if (roleData && Array.isArray(roleData)) {
        setRoles(roleData)
      } else {
        console.warn('没有获取到角色数据或角色数据格式不正确')
      }
      
      // 加载字典数据
      const dictStore = useDictStore()
      await dictStore.loadAllDicts()
      
      return res
    }
    throw new Error(res.message || '获取用户信息失败')
  }

  // 构建树形菜单结构
  function buildMenuTree(menus, parentId) {
    const tree = []
    const menuMap = {}
    
    // 先创建菜单映射
    menus.forEach(menu => {
      menuMap[menu.id] = { ...menu, children: [] }
    })
    
    // 构建树形结构
    menus.forEach(menu => {
      const parent = menuMap[menu.parentId]
      if (parent) {
        parent.children.push(menuMap[menu.id])
      } else if (menu.parentId === 0 || !menu.parentId) {
        tree.push(menuMap[menu.id])
      }
    })
    
    return tree
  }

  // 加载用户菜单
  async function loadUserMenus() {
    try {
      if (!token.value) {
        console.warn('未登录，无法加载菜单')
        return
      }
      
      const response = await fetch('/api/auth/menus', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const result = await response.json()
      if (result.code === 200 && result.data) {
        // 后端已经返回树形结构，直接使用
        setMenusWithPermissions(result.data)
      } else if (result.code === 401) {
        console.error('认证失败，请重新登录')
        logout()
      } else {
        console.warn('加载菜单失败:', result.message || '未知错误')
      }
    } catch (error) {
      console.error('加载菜单失败:', error)
    }
  }

  function logout() {
    token.value = ''
    userInfo.value = {}
    products.value = []
    menus.value = []
    roles.value = []
    permissions.value = new Set()
    localStorage.removeItem('righthand_token')
    localStorage.removeItem('righthand_user_info')
    localStorage.removeItem('righthand_permissions')
  }

  return {
    token,
    userInfo,
    products,
    menus,
    roles,
    permissions,
    setToken,
    setUserInfo,
    setProducts,
    setMenus,
    setMenusWithPermissions,
    setRoles,
    login,
    getUserInfo,
    loadUserMenus,
    logout
  }
})
