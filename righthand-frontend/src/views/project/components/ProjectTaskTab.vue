<template>
  <div class="project-task-tab">
    <div class="tab-toolbar">
      <div class="toolbar-left">
        <span class="toolbar-label">周过滤：</span>
        <el-select v-model="weekFilter" clearable placeholder="全部周" style="width: 220px" @change="handleSearch">
          <el-option label="全部（按任务去重）" :value="0" />
          <el-option
            v-for="w in weeks"
            :key="w.id"
            :label="`第${w.weekNo}周 ${w.title}`"
            :value="w.id"
          />
        </el-select>
        <span class="toolbar-label">负责人：</span>
        <el-select v-model="userFilter" clearable filterable placeholder="全部负责人" style="width: 150px" @change="handleSearch">
          <el-option v-for="u in users" :key="u.id" :label="userLabel(u)" :value="u.id" />
        </el-select>
        <span class="toolbar-label">状态：</span>
        <el-select v-model="statusFilter" clearable placeholder="全部状态" style="width: 120px" @change="handleSearch">
          <el-option v-for="(meta, value) in TASK_STATUS" :key="value" :label="meta.label" :value="value" />
        </el-select>
      </div>
      <div class="toolbar-right" v-if="canManageTasks">
        <el-button type="primary" size="small" @click="openCreateDialog()">
          <el-icon><Plus /></el-icon>
          <span>新建任务</span>
        </el-button>
        <el-button size="small" @click="openLinkDialog">
          <el-icon><Link /></el-icon>
          <span>关联已有事项</span>
        </el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column label="任务标题" min-width="240" show-overflow-tooltip>
        <template #default="{ row }">
          <el-link type="primary" :underline="false" @click="goTaskDetail(row)">{{ row.title }}</el-link>
        </template>
      </el-table-column>
      <el-table-column label="所属周" min-width="150">
        <template #default="{ row }">
          <el-tag v-for="w in row.weeks || []" :key="w.weekPlanId" size="small" class="week-tag" effect="plain">
            第{{ w.weekNo }}周
          </el-tag>
          <span v-if="!(row.weeks || []).length" class="muted">-</span>
        </template>
      </el-table-column>
      <el-table-column label="负责人" width="100">
        <template #default="{ row }">{{ row.userName || '-' }}</template>
      </el-table-column>
      <el-table-column label="优先级" width="80" align="center">
        <template #default="{ row }">
          <el-tag :type="priorityMeta(row.priority).tagType" size="small" effect="plain">
            {{ priorityMeta(row.priority).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="taskStatusMeta(row.statusCode).tagType" size="small">
            {{ taskStatusMeta(row.statusCode).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="截止时间" width="120">
        <template #default="{ row }">{{ formatDate(row.deadline) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-dropdown
            v-if="(row.weeks || []).length && canManageTasks"
            trigger="click"
            @command="(cmd) => handleUnlink(row, cmd)"
          >
            <el-button link type="warning" size="small">移出周</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="w in row.weeks" :key="w.weekPlanId" :command="w.weekPlanId">
                  从第{{ w.weekNo }}周移出
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown
            v-if="movableWeeks(row).length && canManageTasks"
            trigger="click"
            @command="(cmd) => handleMove(row, cmd)"
          >
            <el-button link type="primary" size="small">移动到其他周</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="w in movableWeeks(row)" :key="w.id" :command="w.id">
                  移入第{{ w.weekNo }}周 {{ w.title }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无任务" :image-size="80" />
      </template>
    </el-table>

    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="pageNum"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        background
        @current-change="loadTasks"
        @size-change="handleSearch"
      />
    </div>

    <!-- 新建任务弹窗 -->
    <el-dialog v-model="createDialogVisible" title="新建任务" width="560px" :close-on-click-modal="false">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="90px">
        <el-form-item label="归属周" prop="weekPlanId">
          <el-select v-model="createForm.weekPlanId" placeholder="请选择归属周" style="width: 100%">
            <el-option v-for="w in weeks" :key="w.id" :label="`第${w.weekNo}周 ${w.title}`" :value="w.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="任务标题" prop="title">
          <el-input v-model="createForm.title" maxlength="200" placeholder="请输入任务标题" />
        </el-form-item>
        <el-form-item label="负责人" prop="userId">
          <el-select v-model="createForm.userId" filterable placeholder="请选择负责人" style="width: 100%">
            <el-option v-for="u in users" :key="u.id" :label="userLabel(u)" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-select v-model="createForm.priority" style="width: 100%">
            <el-option label="高" value="HIGH" />
            <el-option label="中" value="MEDIUM" />
            <el-option label="低" value="LOW" />
          </el-select>
        </el-form-item>
        <el-form-item label="截止时间" prop="deadline">
          <el-date-picker
            v-model="createForm.deadline"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="截止日期（可选）"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="createForm.description" type="textarea" :rows="3" maxlength="1000" show-word-limit placeholder="任务描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 关联已有事项弹窗 -->
    <el-dialog v-model="linkDialogVisible" title="关联已有事项" width="640px" :close-on-click-modal="false">
      <el-form label-width="90px">
        <el-form-item label="归属周">
          <el-select v-model="linkForm.weekPlanId" placeholder="请选择归属周" style="width: 100%">
            <el-option v-for="w in weeks" :key="w.id" :label="`第${w.weekNo}周 ${w.title}`" :value="w.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择事项">
          <el-input
            :model-value="linkForm.taskTitle"
            readonly
            placeholder="请点击右侧按钮，从事项列表中筛选选择"
          >
            <template #append>
              <el-button @click="taskSelectVisible = true">选择事项</el-button>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      <el-alert type="info" :closable="false" show-icon>
        支持从「事项管理」列表中筛选并选择已有事项关联到本项目的周计划；重复关联同一周不会产生重复记录。
      </el-alert>
      <template #footer>
        <el-button @click="linkDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="linking" :disabled="!linkForm.taskId || !linkForm.weekPlanId" @click="handleLink">
          关联
        </el-button>
      </template>
    </el-dialog>

    <!-- 选择事项弹窗（与事项管理列表同款筛选/分页） -->
    <TaskSelectDialog v-model="taskSelectVisible" :selected-id="linkForm.taskId" @select="handleTaskSelected" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Link } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getProjectTasks, createTaskInWeek, linkTaskToWeek, unlinkTaskFromWeek, moveTaskToWeek } from '@/api/project'
import { TASK_STATUS, TASK_PRIORITY } from '@/utils/projectEnums'
import TaskSelectDialog from '@/components/TaskSelectDialog.vue'

const props = defineProps({
  project: { type: Object, required: true },
  users: { type: Array, default: () => [] },
  // 从周卡片「新建任务」带入的归属周（消费后由父组件置空）
  defaultWeekId: { type: Number, default: null }
})

const emit = defineEmits(['refresh', 'consumed-default-week'])

const router = useRouter()

const loading = ref(false)
const list = ref([])
const total = ref(0)
const pageNum = ref(1)
const pageSize = ref(10)
const weekFilter = ref(null)
const userFilter = ref(null)
const statusFilter = ref(null)

const weeks = computed(() => props.project?.weeks || [])

const canManageTasks = computed(() => {
  const status = props.project?.status
  if (status === 'CLOSED' || status === 'TERMINATED') return false
  const actions = props.project?.allowedActions || []
  if (actions.includes('addWeek')) return true // PM/ADMIN
  return weeks.value.some((w) => (w.allowedActions || []).includes('edit')) // 某周负责人
})

function userLabel(u) {
  return u.realName || u.nickname || u.username || `用户${u.id}`
}

function priorityMeta(priority) {
  return TASK_PRIORITY[priority] || { label: priority || '-', tagType: 'info' }
}

function taskStatusMeta(status) {
  return TASK_STATUS[status] || { label: status || '-', tagType: 'info' }
}

function formatDate(v) {
  if (!v) return '-'
  return dayjs(v).format('YYYY-MM-DD')
}

/** 目标周候选：排除任务当前已关联的周 */
function movableWeeks(row) {
  const current = new Set((row.weeks || []).map((w) => w.weekPlanId))
  return weeks.value.filter((w) => !current.has(w.id))
}

function handleSearch() {
  pageNum.value = 1
  loadTasks()
}

async function loadTasks() {
  loading.value = true
  try {
    const params = { pageNum: pageNum.value, pageSize: pageSize.value }
    if (weekFilter.value) params.weekPlanId = weekFilter.value
    if (userFilter.value) params.userId = userFilter.value
    if (statusFilter.value) params.statusCode = statusFilter.value
    const res = await getProjectTasks(props.project.id, params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

function goTaskDetail(row) {
  router.push(`/tasks/${row.id}/detail`)
}

/* ------------------------------- 新建任务 ------------------------------- */

const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref(null)

const createForm = reactive({
  weekPlanId: null,
  title: '',
  userId: null,
  priority: 'MEDIUM',
  deadline: '',
  description: ''
})

const createRules = {
  weekPlanId: [{ required: true, message: '请选择归属周', trigger: 'change' }],
  title: [{ required: true, message: '请输入任务标题', trigger: 'blur' }],
  userId: [{ required: true, message: '请选择负责人', trigger: 'change' }]
}

function openCreateDialog(presetWeekId) {
  createForm.weekPlanId = presetWeekId || weekFilter.value || weeks.value[0]?.id || null
  createForm.title = ''
  createForm.userId = null
  createForm.priority = 'MEDIUM'
  createForm.deadline = ''
  createForm.description = ''
  createDialogVisible.value = true
  createFormRef.value?.clearValidate()
}

async function handleCreate() {
  try {
    await createFormRef.value.validate()
  } catch {
    return
  }
  creating.value = true
  try {
    const { weekPlanId, ...payload } = createForm
    await createTaskInWeek(props.project.id, weekPlanId, payload)
    ElMessage.success('任务创建成功')
    createDialogVisible.value = false
    loadTasks()
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  } finally {
    creating.value = false
  }
}

/* ------------------------------- 关联已有 ------------------------------- */

const linkDialogVisible = ref(false)
const linking = ref(false)
const taskSelectVisible = ref(false)

const linkForm = reactive({
  weekPlanId: null,
  taskId: null,
  taskTitle: ''
})

function openLinkDialog() {
  linkForm.weekPlanId = weekFilter.value || weeks.value[0]?.id || null
  linkForm.taskId = null
  linkForm.taskTitle = ''
  linkDialogVisible.value = true
}

/** 选择弹窗确认后回填（弹窗自身负责关闭） */
function handleTaskSelected(row) {
  linkForm.taskId = row.id
  linkForm.taskTitle = `#${row.id} ${row.title}`
}

async function handleLink() {
  // 防御：与「关联」按钮 disabled 双保险，避免空参请求（对应用例 Link_NoSelection_GuardBlocks）
  if (!linkForm.taskId || !linkForm.weekPlanId) {
    ElMessage.warning('请先选择事项与归属周')
    return
  }
  linking.value = true
  try {
    await linkTaskToWeek(props.project.id, linkForm.weekPlanId, Number(linkForm.taskId))
    ElMessage.success('关联成功')
    linkDialogVisible.value = false
    loadTasks()
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  } finally {
    linking.value = false
  }
}

/* ------------------------------- 移出/移动 ------------------------------- */

async function handleUnlink(row, weekPlanId) {
  const week = weeks.value.find((w) => w.id === weekPlanId)
  try {
    await ElMessageBox.confirm(`确认将任务「${row.title}」从第${week?.weekNo}周移出？任务本身不会被删除。`, '移出确认', {
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    await unlinkTaskFromWeek(props.project.id, weekPlanId, row.id)
    ElMessage.success('已移出')
    loadTasks()
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  }
}

async function handleMove(row, weekPlanId) {
  const week = weeks.value.find((w) => w.id === weekPlanId)
  try {
    await moveTaskToWeek(props.project.id, row.id, { weekPlanId })
    ElMessage.success(`已移动到第${week?.weekNo}周`)
    loadTasks()
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  }
}

watch(
  () => props.project?.id,
  () => {
    pageNum.value = 1
    weekFilter.value = null
    userFilter.value = null
    statusFilter.value = null
    loadTasks()
  }
)

// 从周卡片「新建任务」：自动打开新建弹窗并预选归属周
watch(
  () => props.defaultWeekId,
  (val) => {
    if (!val) return
    openCreateDialog(val)
    emit('consumed-default-week')
  }
)

onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
.project-task-tab {
  min-height: 200px;
}

.tab-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 8px;
}

.toolbar-label {
  font-size: 13px;
  color: #606266;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.week-tag {
  margin-right: 4px;
}

.muted {
  color: #909399;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
