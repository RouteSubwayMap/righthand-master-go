<template>
  <div v-loading="pageLoading" class="project-detail-page">
    <template v-if="project">
      <!-- 头部信息卡 -->
      <el-card shadow="never" class="header-card">
        <div class="header-top">
          <div class="header-left">
            <el-button link @click="goBack">
              <el-icon><ArrowLeft /></el-icon>
              <span>返回</span>
            </el-button>
            <span class="project-name">{{ project.projectName }}</span>
            <el-tag size="small" effect="plain">{{ project.projectKey }}</el-tag>
            <el-tag :type="statusMeta(project.status).tagType">{{ statusMeta(project.status).label }}</el-tag>
            <el-tag :type="riskMeta(project.riskLevel).tagType" effect="plain">
              风险：{{ riskMeta(project.riskLevel).label }}
            </el-tag>
            <el-tag :type="deliveryMeta(project.deliveryStatus).tagType" effect="plain">
              {{ deliveryMeta(project.deliveryStatus).label }}
            </el-tag>
          </div>
          <div class="header-actions">
            <el-button
              v-for="act in sortedActions"
              :key="act"
              size="small"
              :type="actionButtonType(act)"
              :plain="act === 'edit' || act === 'editRisk' || act === 'manageMember'"
              @click="handleProjectAction(act)"
            >
              {{ actionLabel(act) }}
            </el-button>
          </div>
        </div>

        <!-- 进度概览 -->
        <div class="progress-overview">
          <div class="overview-item">
            <div class="overview-label">整体进度</div>
            <el-progress
              :percentage="project.progressPct || 0"
              :stroke-width="10"
              :color="progressColor(project.progressPct)"
              style="width: 220px"
            />
          </div>
          <div class="overview-item">
            <div class="overview-label">周计划完成</div>
            <div class="overview-value">{{ project.doneWeekCount || 0 }} / {{ project.weekCount || 0 }} 周</div>
          </div>
          <div class="overview-item">
            <div class="overview-label">项目周期</div>
            <div class="overview-value">{{ project.startDate }} ~ {{ project.endDate }}（{{ durationDays }} 天）</div>
          </div>
          <div class="overview-item">
            <div class="overview-label">延期情况</div>
            <div class="overview-value">
              <span v-if="overdueDays > 0" class="danger-text">已超期 {{ overdueDays }} 天</span>
              <span v-else>未超期</span>
              <span v-if="overdueWeekCount > 0" class="danger-text">（{{ overdueWeekCount }} 周已延期）</span>
            </div>
          </div>
          <div class="overview-item">
            <div class="overview-label">项目经理</div>
            <div class="overview-value">{{ project.managerName || '-' }}</div>
          </div>
        </div>

        <div v-if="project.riskNote" class="risk-note">
          <el-icon><WarningFilled /></el-icon>
          <span>风险说明：{{ project.riskNote }}</span>
        </div>

        <div v-if="project.description || project.planDesc" class="desc-block">
          <div v-if="project.description" class="desc-item">
            <span class="desc-label">项目描述：</span>{{ project.description }}
          </div>
          <div v-if="project.planDesc" class="desc-item">
            <span class="desc-label">整体计划：</span>{{ project.planDesc }}
          </div>
        </div>
      </el-card>

      <!-- Tabs -->
      <el-card shadow="never" class="tabs-card">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="项目整体计划" name="stages">
            <ProjectStageSection :project="project" @refresh="loadDetail" />
          </el-tab-pane>
          <el-tab-pane label="周计划" name="weeks">
            <WeeklyPlanSection :project="project" :users="users" @refresh="loadDetail" @add-task="handleAddTaskFromWeek" />
          </el-tab-pane>
          <el-tab-pane label="任务" name="tasks">
            <ProjectTaskTab
              :project="project"
              :users="users"
              :default-week-id="pendingTaskWeekId"
              @refresh="loadDetail"
              @consumed-default-week="pendingTaskWeekId = null"
            />
          </el-tab-pane>
          <el-tab-pane label="成员" name="members">
            <div class="member-section">
              <div class="member-toolbar" v-if="canManageMembers">
                <el-select v-model="newMemberId" filterable placeholder="选择要添加的成员" style="width: 220px" size="small">
                  <el-option v-for="u in addableUsers" :key="u.id" :label="userLabel(u)" :value="u.id" />
                </el-select>
                <el-button type="primary" size="small" :disabled="!newMemberId" :loading="addingMember" @click="handleAddMember">
                  添加成员
                </el-button>
              </div>
              <el-table :data="project.members || []" stripe>
                <el-table-column label="姓名" min-width="140">
                  <template #default="{ row }">{{ row.userName || userLabelFromId(row.userId) }}</template>
                </el-table-column>
                <el-table-column label="角色" width="120" align="center">
                  <template #default="{ row }">
                    <el-tag :type="row.memberRole === 'PM' ? 'primary' : 'info'" size="small" effect="plain">
                      {{ row.memberRole === 'PM' ? '项目经理' : '成员' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="加入时间" width="180">
                  <template #default="{ row }">{{ formatDateTime(row.createTime) }}</template>
                </el-table-column>
                <el-table-column label="操作" width="100">
                  <template #default="{ row }">
                    <el-button
                      v-if="canManageMembers && row.userId !== project.managerId"
                      link
                      type="danger"
                      size="small"
                      @click="handleRemoveMember(row)"
                    >
                      移除
                    </el-button>
                    <span v-else class="muted">-</span>
                  </template>
                </el-table-column>
                <template #empty>
                  <el-empty description="暂无成员" :image-size="60" />
                </template>
              </el-table>
              <div v-if="!canManageMembers" class="muted member-tip">仅项目经理/管理员可管理成员</div>
            </div>
          </el-tab-pane>
          <el-tab-pane label="变更历史" name="logs">
            <div class="logs-section">
              <el-empty v-if="!(project.logs || []).length" description="暂无变更记录" :image-size="60" />
              <el-timeline v-else>
                <el-timeline-item
                  v-for="log in project.logs"
                  :key="log.id"
                  :timestamp="formatDateTime(log.createTime)"
                  placement="top"
                  :type="logMeta(log.action).type"
                >
                  <div class="log-item">
                    <el-tag size="small" effect="plain" :type="logMeta(log.action).type">
                      {{ logMeta(log.action).label }}
                    </el-tag>
                    <span class="log-detail">{{ log.detail }}</span>
                    <span class="muted">操作人：{{ log.operatorName || userLabelFromId(log.operatorId) }}</span>
                  </div>
                </el-timeline-item>
              </el-timeline>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </template>

    <el-empty v-else-if="!pageLoading" description="项目不存在或已被删除">
      <el-button type="primary" @click="goBack">返回项目清单</el-button>
    </el-empty>

    <!-- 新建/编辑弹窗 -->
    <ProjectFormDialog v-model:visible="editDialogVisible" :project="project" :users="users" @saved="loadDetail" />

    <!-- 风险修改弹窗 -->
    <el-dialog v-model="riskDialogVisible" title="修改风险 / 交付状态" width="520px" :close-on-click-modal="false">
      <el-form ref="riskFormRef" :model="riskForm" :rules="riskRules" label-width="100px">
        <el-form-item label="风险等级" prop="riskLevel">
          <el-select v-model="riskForm.riskLevel" style="width: 100%">
            <el-option v-for="opt in RISK_LEVEL_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="交付状态" prop="deliveryStatus">
          <el-select v-model="riskForm.deliveryStatus" style="width: 100%">
            <el-option v-for="opt in DELIVERY_STATUS_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="风险说明" prop="riskNote">
          <el-input v-model="riskForm.riskNote" type="textarea" :rows="2" maxlength="500" placeholder="当前风险说明（可选）" />
        </el-form-item>
        <el-form-item label="变更原因" prop="reason">
          <el-input v-model="riskForm.reason" type="textarea" :rows="2" maxlength="300" placeholder="变更原因（必填，会记入变更历史）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="riskDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingRisk" @click="handleRiskSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, WarningFilled } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getProjectDetail, changeProjectStatus, changeProjectRisk, deleteProject, addProjectMember, removeProjectMember } from '@/api/project'
import { getAllUsers } from '@/api/user'
import WeeklyPlanSection from './components/WeeklyPlanSection.vue'
import ProjectTaskTab from './components/ProjectTaskTab.vue'
import ProjectFormDialog from './components/ProjectFormDialog.vue'
import ProjectStageSection from './components/ProjectStageSection.vue'
import {
  PROJECT_STATUS,
  RISK_LEVEL,
  DELIVERY_STATUS,
  CHANGE_LOG_ACTION_LABELS,
  RISK_LEVEL_OPTIONS,
  DELIVERY_STATUS_OPTIONS,
  actionLabel
} from '@/utils/projectEnums'

const route = useRoute()
const router = useRouter()

const projectId = computed(() => route.params.id)

const pageLoading = ref(false)
const project = ref(null)
const users = ref([])
// 默认页签：项目整体计划（点击清单「详情」默认展示）
const activeTab = ref('stages')

// 周卡片「新建任务」带入的任务归属周（传递给任务 Tab 消费后置空）
const pendingTaskWeekId = ref(null)

const editDialogVisible = ref(false)

/* ------------------------------- 展示工具 ------------------------------- */

function userLabel(u) {
  return u.realName || u.nickname || u.username || `用户${u.id}`
}

function userLabelFromId(id) {
  const u = users.value.find((x) => x.id === id)
  return u ? userLabel(u) : `用户${id}`
}

function statusMeta(status) {
  return PROJECT_STATUS[status] || { label: status || '-', tagType: 'info' }
}

function riskMeta(level) {
  return RISK_LEVEL[level] || { label: level || '-', tagType: 'info' }
}

function deliveryMeta(status) {
  return DELIVERY_STATUS[status] || { label: status || '-', tagType: 'info' }
}

function progressColor(pct) {
  if (pct >= 100) return '#67c23a'
  if (pct >= 50) return '#409eff'
  return '#e6a23c'
}

function formatDateTime(v) {
  if (!v) return '-'
  return dayjs(v).format('YYYY-MM-DD HH:mm')
}

const LOG_ACTION_TYPES = {
  CREATE: 'primary',
  STATUS: 'primary',
  RISK_CHANGE: 'warning',
  MEMBER_ADD: 'success',
  MEMBER_REMOVE: 'danger',
  SUMMARY_UNLOCK: 'warning',
  STAGE_ADD: 'primary',
  STAGE_EDIT: 'info',
  STAGE_DELETE: 'danger',
  UPDATE: 'info'
}

function logMeta(action) {
  return {
    label: CHANGE_LOG_ACTION_LABELS[action] || action || '-',
    type: LOG_ACTION_TYPES[action] || 'info'
  }
}

const durationDays = computed(() => {
  if (!project.value?.startDate || !project.value?.endDate) return 0
  return dayjs(project.value.endDate).diff(dayjs(project.value.startDate), 'day') + 1
})

const overdueDays = computed(() => {
  const p = project.value
  if (!p || p.status === 'CLOSED' || p.status === 'TERMINATED') return 0
  const today = dayjs().startOf('day')
  const end = dayjs(p.endDate).startOf('day')
  return end.isBefore(today) ? today.diff(end, 'day') : 0
})

const overdueWeekCount = computed(() => (project.value?.weeks || []).filter((w) => w.overdue).length)

/* ------------------------------- 操作按钮 ------------------------------- */

const ACTION_ORDER = ['start', 'resume', 'pause', 'close', 'edit', 'editRisk', 'manageMember', 'addWeek', 'terminate', 'delete']

const sortedActions = computed(() => {
  const actions = project.value?.allowedActions || []
  return ACTION_ORDER.filter((a) => actions.includes(a))
})

function actionButtonType(act) {
  if (act === 'delete' || act === 'terminate') return 'danger'
  if (act === 'close') return 'success'
  if (act === 'start' || act === 'resume') return 'primary'
  return 'default'
}

const STATUS_CONFIRM_LABELS = { start: '开始', pause: '暂停', resume: '恢复', close: '结项' }

async function handleProjectAction(act) {
  const p = project.value
  try {
    if (act === 'edit') {
      editDialogVisible.value = true
      return
    }
    if (act === 'editRisk') {
      openRiskDialog()
      return
    }
    if (act === 'manageMember') {
      activeTab.value = 'members'
      return
    }
    if (act === 'addWeek') {
      activeTab.value = 'weeks'
      return
    }
    if (act === 'delete') {
      await ElMessageBox.confirm(
        `确认删除项目「${p.projectName}」？删除后不可恢复（仅「启动」状态可删除）。`,
        '删除确认',
        { type: 'warning', confirmButtonText: '删除' }
      )
      await deleteProject(p.id)
      ElMessage.success('删除成功')
      goBack()
      return
    }
    if (act === 'terminate') {
      const { value } = await ElMessageBox.prompt(
        `请输入终止项目「${p.projectName}」的原因（必填）：`,
        '终止项目',
        {
          type: 'warning',
          inputType: 'textarea',
          inputValidator: (v) => (v && v.trim() ? true : '终止原因不能为空'),
          confirmButtonText: '终止'
        }
      )
      await changeProjectStatus(p.id, { action: 'terminate', reason: value.trim() })
      ElMessage.success('项目已终止')
      loadDetail()
      return
    }
    await ElMessageBox.confirm(`确认${STATUS_CONFIRM_LABELS[act] || act}项目「${p.projectName}」？`, '操作确认', {
      type: 'warning'
    })
    await changeProjectStatus(p.id, { action: act })
    ElMessage.success('操作成功')
    loadDetail()
  } catch (e) {
    if (e === 'cancel' || e === 'close') return
    // 业务错误已由拦截器提示
  }
}

/* ------------------------------- 风险修改 ------------------------------- */

const riskDialogVisible = ref(false)
const savingRisk = ref(false)
const riskFormRef = ref(null)

const riskForm = reactive({
  riskLevel: 'MEDIUM',
  deliveryStatus: 'NORMAL',
  riskNote: '',
  reason: ''
})

const riskRules = {
  riskLevel: [{ required: true, message: '请选择风险等级', trigger: 'change' }],
  deliveryStatus: [{ required: true, message: '请选择交付状态', trigger: 'change' }],
  reason: [
    { required: true, message: '请填写变更原因', trigger: 'blur' },
    { min: 2, message: '变更原因至少 2 个字符', trigger: 'blur' }
  ]
}

function openRiskDialog() {
  riskForm.riskLevel = project.value.riskLevel || 'MEDIUM'
  riskForm.deliveryStatus = project.value.deliveryStatus || 'NORMAL'
  riskForm.riskNote = project.value.riskNote || ''
  riskForm.reason = ''
  riskDialogVisible.value = true
  riskFormRef.value?.clearValidate()
}

async function handleRiskSubmit() {
  try {
    await riskFormRef.value.validate()
  } catch {
    return
  }
  savingRisk.value = true
  try {
    await changeProjectRisk(project.value.id, { ...riskForm })
    ElMessage.success('风险信息已更新')
    riskDialogVisible.value = false
    loadDetail()
  } catch {
    // 错误已由拦截器提示
  } finally {
    savingRisk.value = false
  }
}

/* ------------------------------- 成员管理 ------------------------------- */

const newMemberId = ref(null)
const addingMember = ref(false)

const canManageMembers = computed(() => (project.value?.allowedActions || []).includes('manageMember'))

const addableUsers = computed(() => {
  const existed = new Set((project.value?.members || []).map((m) => m.userId))
  return users.value.filter((u) => !existed.has(u.id))
})

async function handleAddMember() {
  addingMember.value = true
  try {
    await addProjectMember(project.value.id, { userId: newMemberId.value, memberRole: 'MEMBER' })
    ElMessage.success('成员已添加')
    newMemberId.value = null
    loadDetail()
  } catch {
    // 错误已由拦截器提示
  } finally {
    addingMember.value = false
  }
}

async function handleRemoveMember(member) {
  const name = member.userName || userLabelFromId(member.userId)
  try {
    await ElMessageBox.confirm(`确认将「${name}」移出项目成员？`, '移除确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    await removeProjectMember(project.value.id, member.id)
    ElMessage.success('已移除')
    loadDetail()
  } catch {
    // 错误已由拦截器提示
  }
}

/* ------------------------------- 数据加载 ------------------------------- */

async function loadDetail() {
  if (!projectId.value) return
  pageLoading.value = true
  try {
    const res = await getProjectDetail(projectId.value)
    project.value = res.data
  } catch {
    project.value = null
  } finally {
    pageLoading.value = false
  }
}

async function loadUsers() {
  try {
    const res = await getAllUsers()
    users.value = res.data || []
  } catch {
    users.value = []
  }
}

function goBack() {
  router.push('/projects')
}

// 周卡片「新建任务」：切到任务 Tab 并带入归属周
function handleAddTaskFromWeek(weekId) {
  activeTab.value = 'tasks'
  pendingTaskWeekId.value = weekId
}

watch(
  () => route.params.id,
  () => {
    if (route.name === 'ProjectDetailPage') {
      // 切换项目时同样回到默认页签「项目整体计划」
      activeTab.value = 'stages'
      loadDetail()
    }
  }
)

onMounted(() => {
  loadUsers()
  loadDetail()
})
</script>

<style scoped>
.project-detail-page {
  padding: 16px;
  min-height: 400px;
}

.header-card {
  margin-bottom: 12px;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.project-name {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.progress-overview {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f2f5;
}

.overview-item {
  min-width: 140px;
}

.overview-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}

.overview-value {
  font-size: 13px;
  color: #303133;
}

.danger-text {
  color: #f56c6c;
}

.risk-note {
  margin-top: 12px;
  padding: 8px 12px;
  background: #fdf6ec;
  border-radius: 4px;
  color: #e6a23c;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.desc-block {
  margin-top: 12px;
  font-size: 13px;
  color: #606266;
}

.desc-item {
  margin-bottom: 4px;
  white-space: pre-wrap;
}

.desc-label {
  color: #909399;
}

.tabs-card {
  min-height: 300px;
}

.member-section {
  min-height: 160px;
}

.member-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.member-tip {
  margin-top: 8px;
  font-size: 12px;
}

.logs-section {
  padding: 4px 0;
}

.log-item {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.log-detail {
  color: #606266;
  font-size: 13px;
}

.muted {
  color: #909399;
  font-size: 12px;
}
</style>
