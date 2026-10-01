<template>
  <div class="tasks-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h2>事项管理</h2>
      <el-button type="primary" @click="goToCreate">
        <el-icon><Plus /></el-icon>
        新建事项
      </el-button>
    </div>

    <!-- 过滤器 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="标题">
          <el-input v-model="filters.title" placeholder="请输入标题关键字" style="width: 200px" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.itemType" placeholder="全部" clearable style="width: 120px">
            <el-option label="任务" value="TASK" />
            <el-option label="备忘" value="MEMO" />
            <el-option label="需求" value="REQUIREMENT" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.statusCode" placeholder="全部" clearable style="width: 120px">
            <el-option label="待办" value="TODO" />
            <el-option label="进行中" value="DOING" />
            <el-option label="已完成" value="DONE" />
            <el-option label="已取消" value="CANCELLED" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="filters.priority" placeholder="全部" clearable style="width: 120px">
            <el-option label="高" value="HIGH" />
            <el-option label="中" value="MEDIUM" />
            <el-option label="低" value="LOW" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isAdmin" label="创建人">
          <el-select v-model="filters.createUser" placeholder="全部" clearable style="width: 150px">
            <el-option
              v-for="user in userList"
              :key="user.id"
              :label="user.realName || user.username"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Tab 切换视图 -->
    <el-tabs v-model="activeView" type="card" @tab-change="handleTabChange">
      <!-- 列表视图 -->
      <el-tab-pane label="列表" name="list">
        <el-table
          v-loading="loading"
          :data="taskList"
          stripe
          style="width: 100%"
        >
          <el-table-column prop="title" label="标题" min-width="200">
            <template #default="{ row }">
              <el-tooltip placement="top" :max-width="300">
                <template #content>
                  <span style="white-space: pre-wrap; word-break: break-word;">{{ row.description || '无描述' }}</span>
                </template>
                <span>{{ row.title }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="itemType" label="类型" width="100">
            <template #default="{ row }">
              <el-tooltip placement="top" :max-width="300">
                <template #content>
                  <span style="white-space: pre-wrap; word-break: break-word;">{{ row.description || '无描述' }}</span>
                </template>
                <el-tag :type="getItemTypeTag(row.itemType)">
                  {{ getItemTypeText(row.itemType) }}
                </el-tag>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="statusCode" label="状态" width="100">
            <template #default="{ row }">
              <el-tooltip placement="top" :max-width="300">
                <template #content>
                  <span style="white-space: pre-wrap; word-break: break-word;">{{ row.description || '无描述' }}</span>
                </template>
                <el-tag :type="getStatusType(row.statusCode)">
                  {{ getStatusText(row.statusCode) }}
                </el-tag>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="priority" label="优先级" width="100">
            <template #default="{ row }">
              <el-tooltip placement="top" :max-width="300">
                <template #content>
                  <span style="white-space: pre-wrap; word-break: break-word;">{{ row.description || '无描述' }}</span>
                </template>
                <el-tag :type="getPriorityType(row.priority)" size="small">
                  {{ getPriorityText(row.priority) }}
                </el-tag>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="deadline" label="截止时间" width="180">
            <template #default="{ row }">
              <el-tooltip placement="top" :max-width="300">
                <template #content>
                  <span style="white-space: pre-wrap; word-break: break-word;">{{ row.description || '无描述' }}</span>
                </template>
                <span>{{ formatTime(row.deadline) }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column prop="progress" label="进度" width="120">
            <template #default="{ row }">
              <el-tooltip placement="top" :max-width="300">
                <template #content>
                  <span style="white-space: pre-wrap; word-break: break-word;">{{ row.description || '无描述' }}</span>
                </template>
                <el-progress :percentage="row.progress" :stroke-width="8" />
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="420" fixed="right">
            <template #default="{ row }">
              <div style="display: flex; flex-wrap: nowrap; gap: 4px; overflow-x: auto;">
                <el-button type="info" size="small" @click="goToDetail(row)">查看</el-button>
                <template v-if="row.statusCode === 'DONE' || row.statusCode === 'CANCELLED'">
                  <el-button type="danger" size="small" @click="deleteTaskItem(row)">删除</el-button>
                  <el-button v-if="row.statusCode !== 'CANCELLED'" type="primary" size="small" @click="handleTransfer(row)">转交</el-button>
                </template>
                <template v-else>
                  <el-button type="primary" size="small" @click="goToEdit(row)">编辑</el-button>
                  <el-button v-if="row.statusCode === 'TODO'" type="warning" size="small" @click="updateStatus(row, 'DOING')">开始处理</el-button>
                  <el-button v-if="row.statusCode === 'DOING'" type="success" size="small" @click="updateStatus(row, 'DONE')">完成</el-button>
                  <el-button v-if="row.statusCode === 'TODO'" type="success" size="small" @click="updateStatus(row, 'DONE')">完成</el-button>
                  <el-button type="danger" size="small" @click="handleCancel(row)">取消</el-button>
                  <el-button type="primary" size="small" @click="handleTransfer(row)">转交</el-button>
                </template>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
          style="margin-top: 20px; justify-content: flex-end"
        />
      </el-tab-pane>

      <!-- 日历视图 -->
      <el-tab-pane label="日历" name="calendar">
        <el-calendar v-model="calendarDate">
          <template #date-cell="{ data }">
            <div class="calendar-cell" @click="handleDateClick(data)">
              <p>{{ data.day.split('-').slice(1).join('-') }}</p>
              <div class="calendar-events">
                <div
                  v-for="event in getDayEvents(data.day)"
                  :key="event.id"
                  class="calendar-event"
                  :class="`event-${event.statusCode}`"
                >
                  {{ event.title }}
                </div>
              </div>
            </div>
          </template>
        </el-calendar>
      </el-tab-pane>
    </el-tabs>

    <!-- 转交对话框 -->
    <el-dialog
      v-model="transferDialogVisible"
      title="转交事项"
      width="400px"
    >
      <el-form :model="transferForm" label-width="80px">
        <el-form-item label="选择用户" required>
          <el-select v-model="transferForm.targetUserId" placeholder="请选择转交用户" style="width: 100%">
            <el-option
              v-for="user in allUserList"
              :key="user.id"
              :label="user.realName || user.username"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="transferDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleTransferSubmit" :loading="transferring">
          确定转交
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getTaskList, deleteTask, updateTaskStatus, transferTask } from '@/api/task'
import { getUsers } from '@/api/system'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'

const router = useRouter()

// 是否使用 mock 数据（本地测试用）
const USE_MOCK_DATA = false

// Mock 数据
const mockTasks = [
  {
    id: 1,
    title: '完成需求文档评审',
    itemType: 'TASK',
    priority: 'HIGH',
    statusCode: 'TODO',
    deadline: new Date('2026-05-27 18:00:00').toISOString(),
    progress: 0,
    description: '完成产品需求文档的评审工作'
  },
  {
    id: 2,
    title: '开发用户管理模块',
    itemType: 'TASK',
    priority: 'MEDIUM',
    statusCode: 'DOING',
    deadline: new Date('2026-05-28 18:00:00').toISOString(),
    progress: 50,
    description: '实现用户管理功能'
  },
  {
    id: 3,
    title: '修复登录页面 bug',
    itemType: 'TASK',
    priority: 'HIGH',
    statusCode: 'DONE',
    deadline: new Date('2026-05-26 18:00:00').toISOString(),
    progress: 100,
    description: '修复登录页面的已知问题'
  },
  {
    id: 4,
    title: '编写单元测试',
    itemType: 'TASK',
    priority: 'LOW',
    statusCode: 'TODO',
    deadline: new Date('2026-05-29 18:00:00').toISOString(),
    progress: 0,
    description: '为核心模块编写单元测试'
  },
  {
    id: 5,
    title: '性能优化',
    itemType: 'TASK',
    priority: 'MEDIUM',
    statusCode: 'CANCELLED',
    deadline: new Date('2026-05-30 18:00:00').toISOString(),
    progress: 0,
    description: '系统性能优化工作'
  }
]

// 当前视图
const activeView = ref('list')

// 加载状态
const loading = ref(false)

// 用户store
const userStore = useUserStore()

// 判断是否为管理员
const isAdmin = computed(() => {
  return userStore.roles?.some(role => role.roleCode === 'ADMIN')
})

// 用户列表（用于创建人下拉框）
const userList = ref([])

// 过滤器
const filters = reactive({
  title: '',
  itemType: '',
  statusCode: '',
  priority: '',
  createUser: ''
})

// 分页
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 事项列表
const taskList = ref([])

// 日历日期
const calendarDate = ref(new Date())

// 对话框
const transferDialogVisible = ref(false)
const transferring = ref(false)

// 当前要转交的任务
const currentTransferTask = ref(null)

// 所有用户列表（用于转交）
const allUserList = ref([])

// 转交表单
const transferForm = reactive({
  targetUserId: null
})

// 加载数据
const loadData = async () => {
  if (USE_MOCK_DATA) {
    loading.value = false
    // 根据过滤器筛选 mock 数据
    let filtered = [...mockTasks]
    if (filters.title) {
      filtered = filtered.filter(t => t.title.includes(filters.title))
    }
    if (filters.itemType) {
      filtered = filtered.filter(t => t.itemType === filters.itemType)
    }
    if (filters.statusCode) {
      filtered = filtered.filter(t => t.statusCode === filters.statusCode)
    }
    if (filters.priority) {
      filtered = filtered.filter(t => t.priority === filters.priority)
    }
    
    // 分页
    const start = (pagination.pageNum - 1) * pagination.pageSize
    const end = start + pagination.pageSize
    taskList.value = filtered.slice(start, end)
    pagination.total = filtered.length
    return
  }
  
  loading.value = true
  try {
    const params = {
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      ...filters
    }
    const res = await getTaskList(params)
    if (res.code === 200) {
      taskList.value = res.data.records || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 加载用户列表
const loadUserList = async () => {
  try {
    const res = await getUsers({ pageNum: 1, pageSize: 100 })
    if (res.code === 200) {
      userList.value = res.data.list || res.data || []
      allUserList.value = res.data.list || res.data || []
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
  }
}

// 重置过滤器
const resetFilters = () => {
  filters.title = ''
  filters.itemType = ''
  filters.statusCode = ''
  filters.priority = ''
  filters.createUser = ''
  pagination.pageNum = 1
  loadData()
}

// Tab切换
const handleTabChange = (tab) => {
  console.log('切换到:', tab)
}

// 跳转到新建事项页面
const goToCreate = () => {
  router.push('/tasks/new')
}

// 跳转到编辑事项页面
const goToEdit = (row) => {
  router.push(`/tasks/${row.id}/edit`)
}

// 跳转到事项详情页面
const goToDetail = (row) => {
  router.push(`/tasks/${row.id}/detail`)
}

// 更新状态
const updateStatus = async (row, statusCode) => {
  if (USE_MOCK_DATA) {
    const task = mockTasks.find(t => t.id === row.id)
    if (task) {
      task.statusCode = statusCode
      if (statusCode === 'DONE') {
        task.progress = 100
      }
      ElMessage.success('状态更新成功')
      loadData()
    }
    return
  }
  
  try {
    await updateTaskStatus(row.id, { statusCode })
    ElMessage.success('状态更新成功')
    loadData()
  } catch (error) {
    ElMessage.error('状态更新失败')
  }
}

// 取消事项
const handleCancel = async (row) => {
  if (USE_MOCK_DATA) {
    const task = mockTasks.find(t => t.id === row.id)
    if (task) {
      task.statusCode = 'CANCELLED'
      ElMessage.success('已取消')
      loadData()
    }
    return
  }
  
  try {
    await updateTaskStatus(row.id, { statusCode: 'CANCELLED' })
    ElMessage.success('已取消')
    loadData()
  } catch (error) {
    ElMessage.error('取消失败')
  }
}

// 处理转交
const handleTransfer = (row) => {
  currentTransferTask.value = row
  transferForm.targetUserId = null
  transferDialogVisible.value = true
}

// 提交转交
const handleTransferSubmit = async () => {
  if (!transferForm.targetUserId) {
    ElMessage.warning('请选择转交用户')
    return
  }

  transferring.value = true
  try {
    await transferTask(currentTransferTask.value.id, {
      targetUserId: transferForm.targetUserId
    })
    ElMessage.success('转交成功')
    transferDialogVisible.value = false
    loadData()
  } catch (error) {
    ElMessage.error('转交失败')
  } finally {
    transferring.value = false
  }
}

// 删除事项
const deleteTaskItem = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该事项吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteTask(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 获取类型标签
const getItemTypeTag = (type) => {
  const map = { TASK: undefined, MEMO: 'success', REQUIREMENT: 'warning' }
  return map[type] || undefined
}

// 获取类型文本
const getItemTypeText = (type) => {
  const map = { TASK: '任务', MEMO: '备忘', REQUIREMENT: '需求' }
  return map[type] || type
}

// 获取状态类型
const getStatusType = (statusCode) => {
  const map = { TODO: undefined, DOING: 'warning', DONE: 'success', CANCELLED: 'info' }
  return map[statusCode] || undefined
}

// 获取状态文本
const getStatusText = (statusCode) => {
  const map = { TODO: '待办', DOING: '进行中', DONE: '已完成', CANCELLED: '已取消' }
  return map[statusCode] || statusCode
}

// 获取优先级类型
const getPriorityType = (priority) => {
  const map = { HIGH: 'danger', MEDIUM: 'warning', LOW: 'info' }
  return map[priority] || ''
}

// 获取优先级文本
const getPriorityText = (priority) => {
  const map = { HIGH: '高', MEDIUM: '中', LOW: '低' }
  return map[priority] || priority
}

// 格式化时间
const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
}

// 获取某天的事件
const getDayEvents = (day) => {
  return taskList.value.filter(item => {
    if (!item.deadline) return false
    return dayjs(item.deadline).format('YYYY-MM-DD') === day
  })
}

// 日期点击
const handleDateClick = (data) => {
  console.log('点击日期:', data.day)
}

onMounted(() => {
  loadData()
  loadUserList()
})
</script>

<style scoped lang="scss">
.tasks-container {
  padding: 20px;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h2 {
      margin: 0;
      font-size: 24px;
    }
  }

  .filter-card {
    margin-bottom: 20px;
  }

  .calendar-cell {
    height: 100%;
    padding: 5px;

    p {
      margin: 0;
      text-align: right;
      font-size: 12px;
      color: #909399;
    }

    .calendar-events {
      margin-top: 5px;

      .calendar-event {
        padding: 2px 5px;
        margin-bottom: 2px;
        font-size: 12px;
        border-radius: 3px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        background: #ecf5ff;
        color: #409eff;

        &.event-DOING {
          background: #fdf6ec;
          color: #e6a23c;
        }

        &.event-DONE {
          background: #f0f9ff;
          color: #67c23a;
        }

        &.event-CANCELLED {
          background: #f4f4f5;
          color: #909399;
        }
      }
    }
  }
}
</style>
