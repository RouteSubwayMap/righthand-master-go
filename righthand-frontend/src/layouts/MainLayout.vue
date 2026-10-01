<template>
  <div class="main-layout">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside width="200px">
        <div class="logo">左右手管家</div>
        
        <!-- 检查是否有菜单数据 -->
        <div v-if="!standaloneMenus.length && !directoryMenus.length" class="no-menus-tip">
          <el-icon :size="48" color="#909399"><InfoFilled /></el-icon>
          <p>暂无菜单权限</p>
          <span>请联系管理员分配菜单权限</span>
        </div>
        
        <el-menu
          v-else
          :default-active="$route.path"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
          router
          :unique-opened="false"
        >
          <!-- 独立菜单项（menuType=1 且 parent_id=0） -->
          <el-menu-item
            v-for="menu in standaloneMenus"
            :key="menu.id"
            :index="menu.path"
          >
            <el-icon>
              <component :is="icons[menu.icon] || icons['Document']" />
            </el-icon>
            <span>{{ menu.menuName }}</span>
          </el-menu-item>
          
          <!-- 目录菜单（menuType=0，包含子菜单） -->
          <el-sub-menu
            v-for="directory in directoryMenus"
            :key="directory.id"
            :index="String(directory.id)"
          >
            <template #title>
              <el-icon>
                <component :is="icons[directory.icon] || icons['Folder']" />
              </el-icon>
              <span>{{ directory.menuName }}</span>
            </template>
            <!-- 子菜单 -->
            <el-menu-item
              v-for="child in getChildMenus(directory.id)"
              :key="child.id"
              :index="getChildMenuPath(directory.path, child.path)"
            >
              <el-icon>
                <component :is="icons[child.icon] || icons['Document']" />
              </el-icon>
              <span>{{ child.menuName }}</span>
            </el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-aside>

      <!-- 主体内容 -->
      <el-container>
        <!-- 顶栏 -->
        <el-header>
          <div class="header-content">
            <div class="header-left">
              <div class="breadcrumb">
                <h2 class="page-title">{{ pageTitle }}</h2>
              </div>
            </div>
            <div class="user-info">
              <!-- 消息中心 -->
              <el-popover placement="bottom-end" :width="360" trigger="click" @show="loadRecentMessages">
                <template #reference>
                  <el-badge :value="unreadCount" :hidden="unreadCount === 0" type="danger" class="message-bell-badge">
                    <el-icon :size="20" class="message-bell-icon"><Bell /></el-icon>
                  </el-badge>
                </template>
                <div class="message-panel">
                  <div class="message-panel-header">
                    <span class="message-panel-title">消息通知</span>
                    <el-button link type="primary" size="small" :disabled="unreadCount === 0" @click="handleReadAll">
                      全部已读
                    </el-button>
                  </div>
                  <el-scrollbar max-height="320px">
                    <el-empty v-if="!recentMessages.length" description="暂无消息" :image-size="60" />
                    <div
                      v-for="msg in recentMessages"
                      :key="msg.id"
                      class="message-item"
                      @click="handleMessageClick(msg)"
                    >
                      <div class="message-item-title">
                        <span v-if="!msg.isRead" class="message-dot"></span>
                        {{ msg.title }}
                      </div>
                      <div class="message-item-content">{{ msg.content }}</div>
                      <div class="message-item-time">{{ formatMessageTime(msg.createTime) }}</div>
                    </div>
                  </el-scrollbar>
                </div>
              </el-popover>
              <el-dropdown @command="handleCommand">
                <div class="user-avatar-wrapper">
                  <el-avatar :size="36" :icon="UserFilled" class="user-avatar" />
                  <div class="user-details">
                    <span class="user-name">{{ userStore.userInfo.realName || userStore.userInfo.username }}</span>
                    <span class="user-role">{{ userRoleText }}</span>
                  </div>
                  <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </div>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="profile">
                      <el-icon><User /></el-icon>
                      个人信息
                    </el-dropdown-item>
                    <el-dropdown-item command="password">
                      <el-icon><Lock /></el-icon>
                      修改密码
                    </el-dropdown-item>
                    <el-dropdown-item divided command="logout">
                      <el-icon><SwitchButton /></el-icon>
                      退出登录
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-header>

        <!-- 内容区 -->
        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>

    <!-- 个人信息对话框 -->
    <el-dialog
      v-model="profileDialogVisible"
      title="个人信息"
      width="600px"
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户名">{{ userStore.userInfo.username }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ userStore.userInfo.realName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ userStore.userInfo.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ userStore.userInfo.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="角色">
          <template v-if="userStore.roles && userStore.roles.length > 0">
            <el-tag 
              v-for="role in userStore.roles" 
              :key="role.id || role.roleCode || index"
              size="small"
              style="margin-right: 5px; margin-bottom: 5px;"
            >
              {{ role.roleName || role.name || role.roleCode || '未知角色' }}
            </el-tag>
          </template>
          <span v-else style="color: #909399;">暂无角色</span>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="profileDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="修改密码"
      width="500px"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="100px"
      >
        <el-form-item label="当前密码" prop="oldPassword">
          <el-input 
            v-model="passwordForm.oldPassword" 
            type="password" 
            placeholder="请输入当前密码" 
            show-password 
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input 
            v-model="passwordForm.newPassword" 
            type="password" 
            placeholder="请输入新密码" 
            show-password 
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="passwordForm.confirmPassword" 
            type="password" 
            placeholder="请再次输入新密码" 
            show-password 
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { useRouter, useRoute } from 'vue-router'
import { computed, onMounted, onBeforeUnmount, watch, ref, reactive } from 'vue'
import { useUserStore } from '@/stores/user'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { UserFilled, ArrowDown, User, Lock, SwitchButton, InfoFilled, Grid, Notebook, Bell, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { changePassword } from '@/api/user'
import { getUnreadCount, getMessages, markMessageRead, markAllMessagesRead } from '@/api/message'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

// 对话框状态
const profileDialogVisible = ref(false)
const passwordDialogVisible = ref(false)
const passwordLoading = ref(false)
const passwordFormRef = ref(null)

// 未读消息数量
const unreadCount = ref(0)

// 未读数刷新定时器（60 秒轮询）
let unreadCountTimer = null

// 顶栏消息面板：最近 10 条
const recentMessages = ref([])

// 密码表单
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 密码验证规则
const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 位', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 获取用户菜单
const userMenus = computed(() => {
  return userStore.menus || []
})

// 当前产品标识
const currentProduct = computed(() => route.meta.product)

// 可访问的产品列表
const accessibleProducts = computed(() => {
  const products = [
    {
      id: 'auth',
      name: '用户权限管理',
      badge: '核心功能',
      badgeType: 'primary',
      icon: User,
      route: '/auth/users'
    },
    {
      id: 'memo',
      name: '备忘录管理',
      badge: '智能助手',
      badgeType: 'success',
      icon: Notebook,
      route: '/memo/dashboard'
    },
    {
      id: 'task',
      name: '任务管理',
      badge: '效率工具',
      badgeType: 'warning',
      icon: Grid,
      route: '/task/board'
    }
  ]
  // 根据用户权限过滤
  if (userStore.products.length > 0) {
    return products.filter(p => userStore.products.includes(p.id))
  }
  return products
})

// 当前产品图标
const currentProductIcon = computed(() => {
  const current = accessibleProducts.value.find(
    p => route.path.startsWith(`/${p.id}`)
  )
  return current?.icon || Grid
})

// 当前产品路由
const currentProductRoute = computed({
  get: () => {
    if (route.path.startsWith('/auth')) {
      return '/auth/users'
    } else if (route.path.startsWith('/memo')) {
      return '/memo/dashboard'
    } else if (route.path.startsWith('/task')) {
      return '/task/board'
    }
    return ''
  },
  set: (val) => val
})

// 独立菜单项（menuType=1 且 parent_id=0），按 sortOrder 排序，只显示当前产品的菜单
const standaloneMenus = computed(() => {
  let menus = (userStore.menus || []).filter(menu => menu.menuType === 1 && menu.parentId === 0)
  // 如果是特定产品，只显示该产品的菜单（根据 product_code 过滤）
  if (currentProduct.value) {
    menus = menus.filter(menu => menu.productCode === currentProduct.value)
  }
  return menus
})

// 目录菜单（menuType=0），只显示当前产品的目录
const directoryMenus = computed(() => {
  let menus = (userStore.menus || []).filter(menu => menu.menuType === 0)
  // 如果是特定产品，只显示该产品的目录（根据 product_code 过滤）
  if (currentProduct.value) {
    menus = menus.filter(menu => menu.productCode === currentProduct.value)
  }
  return [...menus].sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0))
})

// 获取子菜单，从树形结构中获取（只返回页面菜单，排除按钮）
const getChildMenus = (parentId) => {
  const findMenu = (menus) => {
    for (let i = 0; i < menus.length; i++) {
      const menu = menus[i]
      if (String(menu.id) === String(parentId)) {
        // 过滤掉按钮类型（menuType=2），只返回目录和页面菜单
        const children = menu.children || []
        return children.filter(child => child.menuType !== 2)
      }
      if (menu.children && menu.children.length > 0) {
        const found = findMenu(menu.children)
        if (found && found.length > 0) {
          return found
        }
      }
    }
    return []
  }
  return findMenu(userStore.menus || [])
}

// 获取子菜单的完整路径
const getChildMenuPath = (parentPath, childPath) => {
  // 如果子路径已经是绝对路径，直接返回
  if (childPath.startsWith('/')) {
    return childPath
  }
  // 否则拼接父路径和子路径
  return `${parentPath}/${childPath}`
}

// 动态注册图标组件
const icons = ElementPlusIconsVue

// 当前产品标题
const currentProductTitle = computed(() => {
  if (route.meta.product === 'auth') {
    return '用户权限管理'
  } else if (route.meta.product === 'memo') {
    return '备忘录管理'
  } else if (route.meta.product === 'task') {
    return '任务管理'
  }
  return '左右手管家'
})

// 页面标题
const pageTitle = computed(() => {
  return route.meta.title || '首页'
})

// 用户角色文本
const userRoleText = computed(() => {
  const roles = userStore.roles || []
  if (roles.length === 0) return '普通用户'
  return roles.map(r => r.roleName).join(', ')
})

onMounted(async () => {
  // 如果还没有加载菜单，则加载
  if (!userStore.menus || userStore.menus.length === 0) {
    await userStore.loadUserMenus()
  }
  // 加载未读消息数量并开启轮询
  await loadUnreadCount()
  unreadCountTimer = setInterval(loadUnreadCount, 60000)
})

// 监听路由变化，当切换到不同产品时，确保菜单已加载
watch(() => route.meta.product, async (newProduct, oldProduct) => {
  // 只在产品切换时重新加载菜单
  if (newProduct && newProduct !== oldProduct) {
    await userStore.loadUserMenus()
    // 产品切换时也刷新未读消息数量
    await loadUnreadCount()
  }
})


const handleCommand = async (command) => {
  if (command === 'logout') {
    await userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  } else if (command === 'profile') {
    profileDialogVisible.value = true
  } else if (command === 'password') {
    passwordDialogVisible.value = true
  }
}

// 修改密码
const handleChangePassword = async () => {
  if (!passwordFormRef.value) return
  
  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      passwordLoading.value = true
      try {
        const res = await changePassword({
          username: userStore.userInfo.username,
          oldPassword: passwordForm.oldPassword,
          newPassword: passwordForm.newPassword
        })
        
        if (res.code === 200) {
          ElMessage.success('密码修改成功，请重新登录')
          passwordDialogVisible.value = false
          // 清空表单
          passwordForm.oldPassword = ''
          passwordForm.newPassword = ''
          passwordForm.confirmPassword = ''
          // 退出登录
          await userStore.logout()
          router.push('/login')
        } else {
          ElMessage.error(res.msg || '密码修改失败')
        }
        
      } catch (error) {
        ElMessage.error(error.response?.data?.msg || '密码修改失败')
      } finally {
        passwordLoading.value = false
      }
    }
  })
}

// 切换产品
const switchProduct = (routePath) => {
  if (routePath) {
    router.push(routePath)
  }
}

// 加载未读消息数量
const loadUnreadCount = async () => {
  try {
    const res = await getUnreadCount()
    if (res.code === 200) {
      const count = res.data?.count ?? res.data ?? 0
      unreadCount.value = typeof count === 'number' ? count : parseInt(count) || 0
    }
  } catch (error) {
    // 静默失败
  }
}

// 加载最近 10 条消息（打开铃铛面板时）
const loadRecentMessages = async () => {
  try {
    const res = await getMessages({ pageNum: 1, pageSize: 10 })
    if (res.code === 200) {
      recentMessages.value = res.data?.list || []
    }
  } catch (error) {
    // 静默失败
  }
}

// 点击消息：标记已读；项目类消息跳转项目详情
const handleMessageClick = async (msg) => {
  try {
    if (!msg.isRead) {
      await markMessageRead(msg.id)
      msg.isRead = 1
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  } catch (error) {
    // 静默失败
  }
  if (msg.type === 'PROJECT' && msg.businessId) {
    router.push(`/projects/${msg.businessId}/detail`)
  }
}

// 全部已读
const handleReadAll = async () => {
  try {
    await markAllMessagesRead()
    recentMessages.value = recentMessages.value.map((m) => ({ ...m, isRead: 1 }))
    unreadCount.value = 0
    ElMessage.success('已全部标记为已读')
  } catch (error) {
    // 错误已由拦截器提示
  }
}

// 消息时间展示
const formatMessageTime = (v) => {
  if (!v) return ''
  return dayjs(v).format('MM-DD HH:mm')
}

// 组件卸载时清理定时器
onBeforeUnmount(() => {
  if (unreadCountTimer) {
    clearInterval(unreadCountTimer)
    unreadCountTimer = null
  }
})

// 提供刷新未读数方法给子组件使用
defineExpose({
  refreshUnreadCount: loadUnreadCount
})
</script>

<style scoped>
.main-layout {
  width: 100%;
  height: 100vh;
}

.el-container {
  height: 100%;
}

.el-aside {
  background-color: #304156;
  color: #fff;
}

.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  font-size: 20px;
  font-weight: bold;
  color: #fff;
  background-color: #2b3a4b;
}

.no-menus-tip {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
  color: #909399;
}

.no-menus-tip p {
  margin: 15px 0 8px;
  font-size: 14px;
  font-weight: 500;
}

.no-menus-tip span {
  font-size: 12px;
  color: #c0c4cc;
}

.el-header {
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0,21,41,.08);
  padding: 0 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.product-selector {
  width: 220px;
}

.product-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.product-name {
  flex: 1;
  font-size: 14px;
}

.product-badge {
  margin-left: auto;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.message-bell-badge {
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  transition: background-color 0.3s;
}

.message-bell-badge:hover {
  background-color: #f5f7fa;
}

.message-bell-icon {
  color: #606266;
  vertical-align: middle;
}

.message-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f2f5;
  margin-bottom: 4px;
}

.message-panel-title {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
}

.message-item {
  padding: 10px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.message-item:hover {
  background-color: #f5f7fa;
}

.message-item-title {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 6px;
}

.message-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f56c6c;
  flex-shrink: 0;
}

.message-item-content {
  font-size: 12px;
  color: #606266;
  margin-top: 4px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.message-item-time {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.user-avatar-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 8px;
  transition: background-color 0.3s;
}

.user-avatar-wrapper:hover {
  background-color: #f5f7fa;
}

.user-avatar {
  background-color: #409EFF;
  color: #fff;
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.user-role {
  font-size: 12px;
  color: #909399;
}

.el-icon--right {
  font-size: 14px;
  color: #909399;
}

.el-main {
  background-color: #f0f2f5;
  padding: 20px;
}
</style>
