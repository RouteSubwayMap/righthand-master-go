<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card clickable" :class="{ active: activeList === 'today' }" @click="toggleList('today')">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #67C23A;">
              <el-icon :size="32"><Calendar /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.todayTodoCount }}</div>
              <div class="stat-label">今日代办</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="hover" class="stat-card clickable" :class="{ active: activeList === 'todo' }" @click="toggleList('todo')">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #409EFF;">
              <el-icon :size="32"><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.todoCount }}</div>
              <div class="stat-label">所有代办</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="hover" class="stat-card clickable" :class="{ active: activeList === 'doing' }" @click="toggleList('doing')">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #E6A23C;">
              <el-icon :size="32"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.doingCount }}</div>
              <div class="stat-label">进行中</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="hover" class="stat-card clickable" :class="{ active: activeList === 'overdue' }" @click="toggleList('overdue')">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #F56C6C;">
              <el-icon :size="32"><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.overdueCount }}</div>
              <div class="stat-label">已逾期</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 今日代办列表 -->
    <el-card v-show="activeList === 'today'" class="task-list-card mb-4">
      <template #header>
        <div class="card-header">
          <span style="color: #67C23A; font-weight: bold;">今日代办</span>
          <div class="filter-bar">
            <el-input v-model="todayTodoSearch.title" placeholder="标题查询" clearable style="width: 400px" @keyup.enter="searchTodayTodo" />
            <el-date-picker v-model="todayTodoSearch.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 300px" />
            <el-button type="primary" size="small" @click="searchTodayTodo">查询</el-button>
            <el-button size="small" @click="resetTodayTodoSearch">重置</el-button>
          </div>
        </div>
      </template>

      <el-table :data="todayTodoList" stripe style="width: 100%">
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
            <el-tag :type="getItemTypeTag(row.itemType)" size="small">
              {{ getItemTypeLabel(row.itemType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.priority === 'HIGH' ? 'danger' : row.priority === 'MEDIUM' ? 'warning' : 'info'"
              size="small"
            >
              {{ getPriorityLabel(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止时间" width="160">
          <template #default="{ row }">
            <span>{{ formatDateTime(row.deadline) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="420" fixed="right">
          <template #default="{ row }">
            <div style="display: flex; flex-wrap: nowrap; gap: 4px; overflow-x: auto;">
              <el-button type="info" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="warning" size="small" @click="handleStart(row)">开始处理</el-button>
              <el-button v-if="row.statusCode === 'DOING'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode !== 'DONE' && row.statusCode !== 'CANCELLED'" type="danger" size="small" @click="handleCancel(row)">取消</el-button>
              <el-button v-if="row.statusCode !== 'CANCELLED'" type="primary" size="small" @click="handleTransfer(row)">转交</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="todayTodoPagination.pageNum"
        v-model:page-size="todayTodoPagination.pageSize"
        :page-sizes="[5, 10, 20]"
        :total="todayTodoPagination.total"
        layout="total, sizes, prev, pager, next"
        @size-change="loadTodayTodoList"
        @current-change="loadTodayTodoList"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 所有代办列表 -->
    <el-card v-show="activeList === 'todo'" class="task-list-card mb-4">
      <template #header>
        <div class="card-header">
          <span style="color: #409EFF; font-weight: bold;">所有代办</span>
          <div class="filter-bar">
            <el-input v-model="todoSearch.title" placeholder="标题查询" clearable style="width: 270px" @keyup.enter="searchTodo" />
            <el-date-picker v-model="todoSearch.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 390px" />
            <el-button type="primary" size="small" @click="searchTodo">查询</el-button>
            <el-button size="small" @click="resetTodoSearch">重置</el-button>
          </div>
        </div>
      </template>

      <el-table :data="todoList" stripe style="width: 100%">
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
            <el-tag :type="getItemTypeTag(row.itemType)" size="small">
              {{ getItemTypeLabel(row.itemType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.priority === 'HIGH' ? 'danger' : row.priority === 'MEDIUM' ? 'warning' : 'info'"
              size="small"
            >
              {{ getPriorityLabel(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="statusCode" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.statusCode)" size="small">
              {{ getStatusLabel(row.statusCode) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止时间" width="160">
          <template #default="{ row }">
            <span>{{ formatDateTime(row.deadline) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="420" fixed="right">
          <template #default="{ row }">
            <div style="display: flex; flex-wrap: nowrap; gap: 4px; overflow-x: auto;">
              <el-button type="info" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="warning" size="small" @click="handleStart(row)">开始处理</el-button>
              <el-button v-if="row.statusCode === 'DOING'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode !== 'DONE' && row.statusCode !== 'CANCELLED'" type="danger" size="small" @click="handleCancel(row)">取消</el-button>
              <el-button v-if="row.statusCode !== 'CANCELLED'" type="primary" size="small" @click="handleTransfer(row)">转交</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="todoPagination.pageNum"
        v-model:page-size="todoPagination.pageSize"
        :page-sizes="[5, 10, 20]"
        :total="todoPagination.total"
        layout="total, sizes, prev, pager, next"
        @size-change="loadTodoList"
        @current-change="loadTodoList"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 进行中事项列表 -->
    <el-card v-show="activeList === 'doing'" class="task-list-card mb-4">
      <template #header>
        <div class="card-header">
          <span style="color: #E6A23C; font-weight: bold;">进行中事项</span>
          <div class="filter-bar">
            <el-input v-model="doingSearch.title" placeholder="标题查询" clearable style="width: 270px" @keyup.enter="searchDoing" />
            <el-date-picker v-model="doingSearch.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 390px" />
            <el-button type="primary" size="small" @click="searchDoing">查询</el-button>
            <el-button size="small" @click="resetDoingSearch">重置</el-button>
          </div>
        </div>
      </template>

      <el-table :data="doingList" stripe style="width: 100%">
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
            <el-tag :type="getItemTypeTag(row.itemType)" size="small">
              {{ getItemTypeLabel(row.itemType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.priority === 'HIGH' ? 'danger' : row.priority === 'MEDIUM' ? 'warning' : 'info'"
              size="small"
            >
              {{ getPriorityLabel(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="statusCode" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.statusCode)" size="small">
              {{ getStatusLabel(row.statusCode) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止时间" width="160">
          <template #default="{ row }">
            <span>{{ formatDateTime(row.deadline) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="420" fixed="right">
          <template #default="{ row }">
            <div style="display: flex; flex-wrap: nowrap; gap: 4px; overflow-x: auto;">
              <el-button type="info" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="warning" size="small" @click="handleStart(row)">开始处理</el-button>
              <el-button v-if="row.statusCode === 'DOING'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode !== 'DONE' && row.statusCode !== 'CANCELLED'" type="danger" size="small" @click="handleCancel(row)">取消</el-button>
              <el-button v-if="row.statusCode !== 'CANCELLED'" type="primary" size="small" @click="handleTransfer(row)">转交</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="doingPagination.pageNum"
        v-model:page-size="doingPagination.pageSize"
        :page-sizes="[5, 10, 20]"
        :total="doingPagination.total"
        layout="total, sizes, prev, pager, next"
        @size-change="loadDoingList"
        @current-change="loadDoingList"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 已逾期事项列表 -->
    <el-card v-show="activeList === 'overdue'" class="task-list-card mb-4">
      <template #header>
        <div class="card-header">
          <span style="color: #F56C6C; font-weight: bold;">已逾期事项</span>
          <div class="filter-bar">
            <el-input v-model="overdueSearch.title" placeholder="标题查询" clearable style="width: 270px" @keyup.enter="searchOverdue" />
            <el-date-picker v-model="overdueSearch.dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width: 390px" />
            <el-button type="primary" size="small" @click="searchOverdue">查询</el-button>
            <el-button size="small" @click="resetOverdueSearch">重置</el-button>
          </div>
        </div>
      </template>

      <el-table :data="overdueList" stripe style="width: 100%">
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
            <el-tag :type="getItemTypeTag(row.itemType)" size="small">
              {{ getItemTypeLabel(row.itemType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.priority === 'HIGH' ? 'danger' : row.priority === 'MEDIUM' ? 'warning' : 'info'"
              size="small"
            >
              {{ getPriorityLabel(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="statusCode" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.statusCode)" size="small">
              {{ getStatusLabel(row.statusCode) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="deadline" label="截止时间" width="160">
          <template #default="{ row }">
            <span>{{ formatDateTime(row.deadline) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="420" fixed="right">
          <template #default="{ row }">
            <div style="display: flex; flex-wrap: nowrap; gap: 4px; overflow-x: auto;">
              <el-button type="info" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="warning" size="small" @click="handleStart(row)">开始处理</el-button>
              <el-button v-if="row.statusCode === 'DOING'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode === 'TODO'" type="success" size="small" @click="handleComplete(row)">完成</el-button>
              <el-button v-if="row.statusCode !== 'DONE' && row.statusCode !== 'CANCELLED'" type="danger" size="small" @click="handleCancel(row)">取消</el-button>
              <el-button v-if="row.statusCode !== 'CANCELLED'" type="primary" size="small" @click="handleTransfer(row)">转交</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="overduePagination.pageNum"
        v-model:page-size="overduePagination.pageSize"
        :page-sizes="[5, 10, 20]"
        :total="overduePagination.total"
        layout="total, sizes, prev, pager, next"
        @size-change="loadOverdueList"
        @current-change="loadOverdueList"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

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
              v-for="user in userList"
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
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Calendar, Document, Timer, Warning } from '@element-plus/icons-vue'
import { getDashboardStats, getTaskList, updateTaskStatus, transferTask } from '@/api/task'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'

const router = useRouter()

const transferDialogVisible = ref(false)
const transferring = ref(false)

// 当前要转交的任务
const currentTransferTask = ref(null)

// 用户列表
const userList = ref([])

const userStore = useUserStore()

// 转交表单
const transferForm = reactive({
  targetUserId: null
})

// 获取用户列表
const loadUserList = async () => {
  try {
    const res = await fetch('/api/system/users/page?pageNum=1&pageSize=10000', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('righthand_token')}`,
        'Content-Type': 'application/json'
      }
    })
    const result = await res.json()
    if (result.code === 200 && result.data) {
      userList.value = result.data.list || result.data || []
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
  }
}

// 跳转到事项详情页
const handleView = (row) => {
  router.push({ path: `/tasks/${row.id}/detail`, query: { from: 'dashboard' } })
}

// 跳转到事项编辑页
const handleEdit = (row) => {
  router.push({ path: `/tasks/${row.id}/edit`, query: { from: 'dashboard' } })
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
    reloadVisibleLists()
  } catch (error) {
    ElMessage.error('转交失败')
  } finally {
    transferring.value = false
  }
}

// 是否使用 mock 数据（本地测试用）
const USE_MOCK_DATA = false

// Mock 数据
const mockTasks = [
  { id: 1, title: '完成需求文档评审', itemType: 'TASK', priority: 'HIGH', statusCode: 'TODO', deadline: new Date('2026-05-27 18:00:00').toISOString(), progress: 0 },
  { id: 2, title: '开发用户管理模块', itemType: 'TASK', priority: 'MEDIUM', statusCode: 'DOING', deadline: new Date('2026-05-28 18:00:00').toISOString(), progress: 50 },
  { id: 3, title: '修复登录页面 bug', itemType: 'TASK', priority: 'HIGH', statusCode: 'DONE', deadline: new Date('2026-05-26 18:00:00').toISOString(), progress: 100 },
  { id: 4, title: '编写单元测试', itemType: 'TASK', priority: 'LOW', statusCode: 'TODO', deadline: new Date('2026-05-29 18:00:00').toISOString(), progress: 0 },
  { id: 5, title: '性能优化', itemType: 'TASK', priority: 'MEDIUM', statusCode: 'CANCELLED', deadline: new Date('2026-05-30 18:00:00').toISOString(), progress: 0 }
]

// 统计数据
const stats = ref({
  todayTodoCount: 0,
  todoCount: 0,
  doingCount: 0,
  overdueCount: 0
})

// ===== 当前展示的列表（互斥，默认今日代办） =====
const activeList = ref('today')

const toggleList = (type) => {
  activeList.value = type
  if (type === 'today') loadTodayTodoList()
  else if (type === 'todo') loadTodoList()
  else if (type === 'doing') loadDoingList()
  else if (type === 'overdue') loadOverdueList()
}

// ===== 搜索条件 =====
const todayTodoSearch = reactive({ title: '', dateRange: null })
const todoSearch = reactive({ title: '', dateRange: null })
const doingSearch = reactive({ title: '', dateRange: null })
const overdueSearch = reactive({ title: '', dateRange: null })

// 构建查询参数
const buildSearchParams = (search, extra = {}) => {
  const params = { ...extra }
  if (search.title) params.title = search.title
  if (search.dateRange && search.dateRange.length === 2) {
    params.startDate = search.dateRange[0]
    params.endDate = search.dateRange[1]
  }
  return params
}

// ===== 今日代办 =====
const todayTodoList = ref([])
const todayTodoPagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })

const loadTodayTodoList = async () => {
  if (USE_MOCK_DATA) {
    todayTodoList.value = mockTasks.filter(t => t.statusCode === 'TODO')
    todayTodoPagination.total = todayTodoList.value.length
    return
  }

  try {
    const res = await getTaskList(buildSearchParams(todayTodoSearch, {
      today: true,
      pageNum: todayTodoPagination.pageNum,
      pageSize: todayTodoPagination.pageSize
    }))
    if (res.code === 200 && res.data) {
      todayTodoList.value = res.data.records || []
      todayTodoPagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('加载今日代办列表失败:', error)
  }
}

const searchTodayTodo = () => { todayTodoPagination.pageNum = 1; loadTodayTodoList() }
const resetTodayTodoSearch = () => {
  todayTodoSearch.title = ''
  todayTodoSearch.dateRange = null
  todayTodoPagination.pageNum = 1
  loadTodayTodoList()
}

// ===== 所有代办 =====
const todoList = ref([])
const todoPagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })

const loadTodoList = async () => {
  if (USE_MOCK_DATA) {
    todoList.value = mockTasks.filter(t => t.statusCode === 'TODO')
    todoPagination.total = todoList.value.length
    return
  }

  try {
    const res = await getTaskList(buildSearchParams(todoSearch, {
      statusCode: 'TODO',
      pageNum: todoPagination.pageNum,
      pageSize: todoPagination.pageSize
    }))
    if (res.code === 200 && res.data) {
      todoList.value = res.data.records || []
      todoPagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('加载待办列表失败:', error)
  }
}

const searchTodo = () => { todoPagination.pageNum = 1; loadTodoList() }
const resetTodoSearch = () => {
  todoSearch.title = ''
  todoSearch.dateRange = null
  todoPagination.pageNum = 1
  loadTodoList()
}

// ===== 进行中 =====
const doingList = ref([])
const doingPagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })

const loadDoingList = async () => {
  if (USE_MOCK_DATA) {
    doingList.value = mockTasks.filter(t => t.statusCode === 'DOING')
    doingPagination.total = doingList.value.length
    return
  }

  try {
    const res = await getTaskList(buildSearchParams(doingSearch, {
      statusCode: 'DOING',
      pageNum: doingPagination.pageNum,
      pageSize: doingPagination.pageSize
    }))
    if (res.code === 200 && res.data) {
      doingList.value = res.data.records || []
      doingPagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('加载中列表失败:', error)
  }
}

const searchDoing = () => { doingPagination.pageNum = 1; loadDoingList() }
const resetDoingSearch = () => {
  doingSearch.title = ''
  doingSearch.dateRange = null
  doingPagination.pageNum = 1
  loadDoingList()
}

// ===== 已逾期 =====
const overdueList = ref([])
const overduePagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })

const loadOverdueList = async () => {
  if (USE_MOCK_DATA) {
    overdueList.value = mockTasks.filter(t => t.statusCode !== 'DONE' && t.statusCode !== 'CANCELLED')
    overduePagination.total = overdueList.value.length
    return
  }

  try {
    const res = await getTaskList(buildSearchParams(overdueSearch, {
      overdue: true,
      pageNum: overduePagination.pageNum,
      pageSize: overduePagination.pageSize
    }))
    if (res.code === 200 && res.data) {
      overdueList.value = res.data.records || []
      overduePagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('加载已逾期列表失败:', error)
  }
}

const searchOverdue = () => { overduePagination.pageNum = 1; loadOverdueList() }
const resetOverdueSearch = () => {
  overdueSearch.title = ''
  overdueSearch.dateRange = null
  overduePagination.pageNum = 1
  loadOverdueList()
}

// ===== 加载统计数据 =====
const loadStats = async () => {
  try {
    const res = await getDashboardStats()
    if (res.code === 200 && res.data) {
      stats.value = {
        todayTodoCount: res.data.todayTodoCount || 0,
        todoCount: res.data.todoCount || 0,
        doingCount: res.data.doingCount || 0,
        overdueCount: res.data.overdueCount || 0
      }
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// ===== 辅助函数 =====
const getItemTypeLabel = (type) => {
  const map = { TASK: '任务', MEMO: '备忘', REQUIREMENT: '需求' }
  return map[type] || type
}

const getItemTypeTag = (type) => {
  const map = { TASK: undefined, MEMO: 'success', REQUIREMENT: 'warning' }
  return map[type] || undefined
}

const getPriorityLabel = (priority) => {
  const map = { HIGH: '高', MEDIUM: '中', LOW: '低' }
  return map[priority] || priority
}

const getStatusLabel = (status) => {
  const map = { TODO: '待办', DOING: '进行中', DONE: '已完成', CANCELLED: '已取消' }
  return map[status] || status
}

const getStatusType = (status) => {
  const map = { TODO: undefined, DOING: 'warning', DONE: 'success', CANCELLED: 'info' }
  return map[status] || undefined
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

// ===== 状态操作 =====
const reloadVisibleLists = () => {
  loadStats()
  // 刷新当前展示的列表
  const list = activeList.value
  if (list === 'today') loadTodayTodoList()
  else if (list === 'todo') loadTodoList()
  else if (list === 'doing') loadDoingList()
  else if (list === 'overdue') loadOverdueList()
}

// 开始处理事项
const handleStart = async (row) => {
  try {
    await updateTaskStatus(row.id, { statusCode: 'DOING' })
    ElMessage.success('已开始处理')
    reloadVisibleLists()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 完成事项
const handleComplete = async (row) => {
  try {
    await updateTaskStatus(row.id, { statusCode: 'DONE' })
    ElMessage.success('已完成')
    reloadVisibleLists()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 取消事项
const handleCancel = async (row) => {
  try {
    await updateTaskStatus(row.id, { statusCode: 'CANCELLED' })
    ElMessage.success('已取消')
    reloadVisibleLists()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(() => {
  loadStats()
  loadTodayTodoList()
  loadUserList()
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
}

.stat-card.clickable {
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card.clickable:hover {
  transform: translateY(-2px);
}

.stat-card.clickable.active {
  box-shadow: 0 0 0 2px var(--el-color-primary);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 15px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.stat-info {
  text-align: left;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.card-header {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.mb-4 {
  margin-bottom: 20px;
}
</style>
