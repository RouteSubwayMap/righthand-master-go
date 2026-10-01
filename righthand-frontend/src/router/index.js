import { createRouter, createWebHistory } from 'vue-router'
import { cancelAllRequests } from '@/utils/request'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录 - 左右手管家系统' }
  },
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '左右手管家系统', requiresAuth: true },
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'HomeDashboard',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '仪表盘 - 左右手管家系统', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'DashboardIndex',
        component: () => import('@/views/Dashboard.vue')
      }
    ]
  },
  {
    path: '/auth',
    name: 'AuthSystem',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { product: 'auth', title: '用户权限管理', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'AuthIndex',
        redirect: '/auth/users'
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: () => import('@/views/system/UserManagement.vue')
      },
      {
        path: 'roles',
        name: 'RoleManagement',
        component: () => import('@/views/system/RoleManagement.vue')
      },
      {
        path: 'menus',
        name: 'MenuManagement',
        component: () => import('@/views/system/MenuManagement.vue')
      },
      {
        path: 'dicts',
        name: 'DictManagement',
        component: () => import('@/views/system/DictManagement.vue')
      }
    ]
  },
  {
    path: '/tasks',
    name: 'TasksSystem',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '事项管理 - 左右手管家系统', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'TasksIndex',
        component: () => import('@/views/tasks/Tasks.vue')
      },
      {
        path: 'new',
        name: 'TaskCreate',
        component: () => import('@/views/tasks/TaskEdit.vue'),
        meta: { title: '新建事项 - 左右手管家系统' }
      },
      {
        path: ':id/edit',
        name: 'TaskEdit',
        component: () => import('@/views/tasks/TaskEdit.vue'),
        meta: { title: '编辑事项 - 左右手管家系统' }
      },
      {
        path: ':id/detail',
        name: 'TaskDetail',
        component: () => import('@/views/tasks/TaskDetail.vue'),
        meta: { title: '事项详情 - 左右手管家系统' }
      }
    ]
  },
  {
    path: '/requirements',
    name: 'RequirementsSystem',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '需求管理 - 左右手管家系统', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'RequirementListPage',
        component: () => import('@/views/task/RequirementList.vue')
      },
      {
        path: 'import',
        name: 'RequirementImportPage',
        component: () => import('@/views/task/RequirementImport.vue')
      }
    ]
  },
  {
    path: '/projects',
    name: 'ProjectsSystem',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '项目管理 - 左右手管家系统', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'ProjectListPage',
        component: () => import('@/views/project/ProjectList.vue')
      },
      {
        path: ':id/detail',
        name: 'ProjectDetailPage',
        component: () => import('@/views/project/ProjectDetail.vue'),
        meta: { title: '项目详情 - 左右手管家系统' }
      }
    ]
  },
  {
    path: '/contacts',
    name: 'ContactsSystem',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '联系人管理 - 左右手管家系统', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'ContactsIndex',
        component: () => import('@/views/Contacts.vue')
      }
    ]
  },
  {
    path: '/batch-tasks',
    name: 'BatchTasksSystem',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '批量任务清单 - 左右手管家系统', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'BatchTasksIndex',
        component: () => import('@/views/BatchTasks.vue')
      },
      {
        path: 'new',
        name: 'BatchTaskNew',
        component: () => import('@/views/batchtasks/BatchTaskEdit.vue')
      },
      {
        path: ':id/edit',
        name: 'BatchTaskEdit',
        component: () => import('@/views/batchtasks/BatchTaskEdit.vue')
      },
      {
        path: ':id/detail',
        name: 'BatchTaskDetail',
        component: () => import('@/views/batchtasks/BatchTaskDetail.vue')
      }
    ]
  },
  {
    path: '/prod-nodes',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'ProdNodes',
        component: () => import('@/views/ProdNodes.vue')
      },
      {
        path: 'new',
        name: 'ProdNodeCreate',
        component: () => import('@/views/prodnodes/ProdNodeEdit.vue')
      },
      {
        path: ':id/edit',
        name: 'ProdNodeEdit',
        component: () => import('@/views/prodnodes/ProdNodeEdit.vue')
      },
      {
        path: ':id/detail',
        name: 'ProdNodeDetail',
        component: () => import('@/views/prodnodes/ProdNodeDetail.vue')
      }
    ]
  },
  {
    path: '/bank-params',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'BankParams',
        component: () => import('@/views/BankParams.vue')
      },
      {
        path: 'new',
        name: 'BankParamCreate',
        component: () => import('@/views/BankParamEdit.vue')
      },
      {
        path: ':id/edit',
        name: 'BankParamEdit',
        component: () => import('@/views/BankParamEdit.vue')
      },
      {
        path: ':id/detail',
        name: 'BankParamDetail',
        component: () => import('@/views/bankparams/BankParamDetail.vue')
      }
    ]
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由切换前，立即取消上一个页面的所有待处理请求
router.beforeEach(async (to, from, next) => {
  const token = localStorage.getItem('righthand_token')
  
  // 在路由切换前取消所有待处理请求，防止组件卸载后仍有异步操作
  cancelAllRequests()
  
  // 1. 认证检查
  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }
  
  // 2. 已登录用户访问登录页，重定向到首页
  if (to.path === '/login' && token) {
    next('/')
    return
  }
  
  // 3. 如果需要认证且有token，确保用户信息已加载
  if (to.meta.requiresAuth && token) {
    const userStore = useUserStore()
    
    // 如果菜单为空，说明用户信息未加载，先加载
    if (userStore.menus.length === 0) {
      try {
        await userStore.loadUserMenus()
      } catch (error) {
        console.error('加载用户菜单失败:', error)
        // 加载失败可能是token过期，清除token并重定向到登录页
        userStore.logout()
        next('/login')
        return
      }
    }
    
    // 4. 权限检查（如果路由配置了requiredPermission）
    if (to.meta.requiredPermission) {
      const { hasPermission } = await import('@/utils/permission')
      if (!hasPermission(to.meta.requiredPermission)) {
        console.warn(`无权访问页面: ${to.path}, 需要权限: ${to.meta.requiredPermission}`)
        next('/dashboard') // 重定向到仪表盘或403页面
        return
      }
    }
  }
  
  next()
  
  // 设置页面标题
  if (to.meta.title) {
    document.title = to.meta.title
  }
})



export default router
