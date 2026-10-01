import { describe, it, expect, beforeAll } from 'vitest'
import axios from 'axios'

// 直接调用后端API(绕过Vite代理)
const API_BASE = 'http://localhost:9080'

// 存储认证token
let authToken: string

describe('系统管理API集成测试', () => {
  beforeAll(async () => {
    // 登录获取token
    const loginRes = await axios.post(`${API_BASE}/api/auth/login`, {
      username: 'admin',
      password: 'admin123'
    })
    
    authToken = loginRes.data.data.token
    console.log('✅ 登录成功,获取Token')
  })

  describe('用户管理API', () => {
    it('应该正确返回分页数据格式', async () => {
      const res = await axios.get(`${API_BASE}/api/system/users/page`, {
        params: {
          pageNum: 1,
          pageSize: 10
        },
        headers: {
          Authorization: `Bearer ${authToken}`
        }
      })
      
      expect(res.status).toBe(200)
      expect(res.data.data).toHaveProperty('list')  // 验证真实格式是list而不是records
      expect(res.data.data).toHaveProperty('total')
      expect(res.data.data).toHaveProperty('pageNum')
      expect(res.data.data).toHaveProperty('pageSize')
      expect(Array.isArray(res.data.data.list)).toBe(true)
      
      console.log(`✅ 用户分页API正常,总数: ${res.data.data.total}`)
    })
    
    it('应该正确处理空查询参数', async () => {
      const res = await axios.get(`${API_BASE}/api/system/users/page`, {
        params: {
          pageNum: 1,
          pageSize: 10,
          username: '',  // 空字符串
          realName: ''   // 空字符串
        },
        headers: {
          Authorization: `Bearer ${authToken}`
        }
      })
      
      // 不应该返回400错误
      expect(res.status).toBe(200)
      console.log('✅ 空参数处理正常')
    })
  })

  describe('角色管理API', () => {
    it('应该正确返回所有角色列表', async () => {
      const res = await axios.get(`${API_BASE}/api/system/roles/all`, {
        headers: {
          Authorization: `Bearer ${authToken}`
        }
      })
      
      expect(res.status).toBe(200)
      expect(Array.isArray(res.data.data)).toBe(true)  // 应该返回数组
      expect(res.data.data.length).toBeGreaterThan(0)
      
      console.log(`✅ 角色列表API正常,角色数: ${res.data.data.length}`)
    })
    
    it('应该正确返回分页角色列表', async () => {
      const res = await axios.get(`${API_BASE}/api/system/roles`, {
        params: {
          pageNum: 1,
          pageSize: 10
        },
        headers: {
          Authorization: `Bearer ${authToken}`
        }
      })
      
      expect(res.status).toBe(200)
      expect(res.data.data).toHaveProperty('list')
      expect(res.data.data).toHaveProperty('total')
      
      console.log(`✅ 角色分页API正常`)
    })
  })

  describe('菜单管理API', () => {
    it('应该正确返回菜单树', async () => {
      const res = await axios.get(`${API_BASE}/api/system/menus/tree`, {
        headers: {
          Authorization: `Bearer ${authToken}`
        }
      })
      
      expect(res.status).toBe(200)
      expect(Array.isArray(res.data.data)).toBe(true)
      
      console.log(`✅ 菜单树API正常`)
    })
  })

  describe('字典管理API', () => {
    it('应该正确返回字典列表', async () => {
      const res = await axios.get(`${API_BASE}/api/system/dicts`, {
        params: {
          page: 1,
          pageSize: 100
        },
        headers: {
          Authorization: `Bearer ${authToken}`
        }
      })
      
      expect(res.status).toBe(200)
      expect(res.data.data).toHaveProperty('list')
      
      console.log(`✅ 字典列表API正常`)
    })
  })

  describe('菜单管理API', () => {
    it('应该正确返回菜单树形结构', async () => {
      const res = await axios.get(`${API_BASE}/api/system/menus/tree`, {
        headers: {
          Authorization: `Bearer ${authToken}`
        }
      })
      
      expect(res.status).toBe(200)
      expect(Array.isArray(res.data.data)).toBe(true)
      
      // 验证树形结构
      const menus = res.data.data
      if (menus.length > 0) {
        const firstMenu = menus[0]
        expect(firstMenu).toHaveProperty('id')
        expect(firstMenu).toHaveProperty('menuName')
        // children可能不存在(当为空数组时会被omitempty忽略)
        
        console.log(`✅ 菜单树API正常,顶级菜单数: ${menus.length}`)
        
        // 检查有子菜单的菜单项
        const menuWithChildren = menus.find((m: any) => m.children && m.children.length > 0)
        if (menuWithChildren) {
          expect(menuWithChildren).toHaveProperty('children')
          expect(Array.isArray(menuWithChildren.children)).toBe(true)
          console.log(`   - "${menuWithChildren.menuName}" 有 ${menuWithChildren.children.length} 个子菜单`)
        } else {
          console.log(`   - 所有菜单都没有子菜单`)
        }
      }
    })
  })
})
