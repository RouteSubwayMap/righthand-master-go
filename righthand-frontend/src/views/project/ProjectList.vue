<template>
  <div class="project-list-page">
    <!-- 筛选区 -->
    <el-card shadow="never" class="filter-card">
      <el-form :inline="true" :model="query" class="filter-form" @submit.prevent>
        <el-form-item label="状态">
          <el-select
            v-model="query.statusList"
            multiple
            collapse-tags
            collapse-tags-tooltip
            clearable
            placeholder="全部状态"
            style="width: 200px"
            @change="handleSearch"
          >
            <el-option v-for="opt in PROJECT_STATUS_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="风险">
          <el-select v-model="query.riskLevel" clearable placeholder="全部风险" style="width: 120px" @change="handleSearch">
            <el-option v-for="opt in RISK_LEVEL_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="项目经理">
          <el-select v-model="query.managerId" clearable filterable placeholder="全部经理" style="width: 150px" @change="handleSearch">
            <el-option v-for="u in users" :key="u.id" :label="userLabel(u)" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="参与人">
          <el-select v-model="query.participantId" clearable filterable placeholder="全部参与人" style="width: 150px" @change="handleSearch">
            <el-option v-for="u in users" :key="u.id" :label="userLabel(u)" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            value-format="YYYY-MM-DD"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 240px"
            @change="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="query.keyword"
            placeholder="搜索名称/Key"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
        </el-form-item>
        <el-form-item label="排序">
          <el-select v-model="query.sort" style="width: 170px" @change="handleSearch">
            <el-option label="结束日期（升序）" value="endDate_asc" />
            <el-option label="结束日期（降序）" value="endDate_desc" />
            <el-option label="最近更新优先" value="updateTime_desc" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            <span>查询</span>
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表区 -->
    <el-card shadow="never" class="table-card">
      <div class="table-header">
        <div class="table-title">
          项目清单
          <span class="total-count">共 {{ total }} 个</span>
        </div>
        <el-button v-permission="'project:add'" type="success" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          <span>新建项目</span>
        </el-button>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column label="项目名称" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="goDetail(row)">{{ row.projectName }}</el-link>
            <div class="project-key">{{ row.projectKey }}</div>
          </template>
        </el-table-column>
        <el-table-column label="起止 / 工期" width="230">
          <template #default="{ row }">
            <div>{{ row.startDate }} ~ {{ row.endDate }}</div>
            <div class="muted">共 {{ durationDays(row.startDate, row.endDate) }} 天</div>
          </template>
        </el-table-column>
        <el-table-column label="项目经理" width="110">
          <template #default="{ row }">{{ row.managerName || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="statusMeta(row.status).tagType" size="small">{{ statusMeta(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="风险" width="90" align="center">
          <template #default="{ row }">
            <span class="risk-dot" :style="{ backgroundColor: riskMeta(row.riskLevel).color }"></span>
            <span>{{ riskMeta(row.riskLevel).label }}</span>
          </template>
        </el-table-column>
        <el-table-column label="交付状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="deliveryMeta(row.deliveryStatus).tagType" size="small" effect="plain">
              {{ deliveryMeta(row.deliveryStatus).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="进度" width="170">
          <template #default="{ row }">
            <el-progress
              :percentage="row.progressPct || 0"
              :stroke-width="8"
              :color="progressColor(row.progressPct)"
            />
            <div class="muted">{{ row.doneWeekCount || 0 }}/{{ row.weekCount || 0 }} 周已完成</div>
          </template>
        </el-table-column>
        <el-table-column label="当前阶段" width="130" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.currentStage" size="small" effect="plain">{{ row.currentStage }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="最近更新" width="160">
          <template #default="{ row }">{{ formatDateTime(row.updateTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="goDetail(row)">详情</el-button>
            <el-button
              v-for="act in primaryActions(row)"
              :key="act"
              size="small"
              :type="actionButtonType(act)"
              @click="handleProjectAction(row, act)"
            >
              {{ actionLabel(act) }}
            </el-button>
            <el-dropdown v-if="moreActions(row).length" trigger="click" @command="(cmd) => handleProjectAction(row, cmd)">
              <el-button size="small">
                更多<el-icon><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-for="act in moreActions(row)" :key="act" :command="act" :style="act === 'delete' || act === 'terminate' ? { color: '#f56c6c' } : {}">
                    {{ actionLabel(act) }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无项目" :image-size="80" />
        </template>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="query.pageNum"
          v-model:page-size="query.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadList"
          @size-change="handleSearch"
        />
      </div>
    </el-card>

    <!-- 新建/编辑弹窗 -->
    <ProjectFormDialog v-model:visible="dialogVisible" :project="editingProject" :users="users" @saved="loadList" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, ArrowDown } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getProjectList, deleteProject, changeProjectStatus } from '@/api/project'
import { getAllUsers } from '@/api/user'
import ProjectFormDialog from './components/ProjectFormDialog.vue'
import {
  PROJECT_STATUS,
  RISK_LEVEL,
  DELIVERY_STATUS,
  PROJECT_STATUS_OPTIONS,
  RISK_LEVEL_OPTIONS,
  actionLabel
} from '@/utils/projectEnums'

const router = useRouter()

const loading = ref(false)
const list = ref([])
const total = ref(0)
const users = ref([])
const dateRange = ref([])

const query = reactive({
  pageNum: 1,
  pageSize: 10,
  statusList: [],
  riskLevel: '',
  managerId: null,
  participantId: null,
  keyword: '',
  // 默认 end_date 升序，可切换（endDate_asc / endDate_desc / updateTime_desc）
  sort: 'endDate_asc'
})

const dialogVisible = ref(false)
const editingProject = ref(null)

const ACTION_ORDER = ['start', 'resume', 'pause', 'close', 'edit', 'terminate', 'delete']

function userLabel(u) {
  return u.realName || u.nickname || u.username || `用户${u.id}`
}

function statusMeta(status) {
  return PROJECT_STATUS[status] || { label: status || '-', tagType: 'info' }
}

function riskMeta(level) {
  return RISK_LEVEL[level] || { label: level || '-', color: '#909399' }
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

function durationDays(start, end) {
  if (!start || !end) return 0
  return dayjs(end).diff(dayjs(start), 'day') + 1
}

/** 按固定优先级排序的操作按钮 */
function sortedActions(row) {
  const actions = row.allowedActions || []
  return ACTION_ORDER.filter((a) => actions.includes(a))
}

function primaryActions(row) {
  const actions = sortedActions(row)
  return actions.length > 2 ? actions.slice(0, 2) : actions
}

function moreActions(row) {
  const actions = sortedActions(row)
  return actions.length > 2 ? actions.slice(2) : []
}

function actionButtonType(act) {
  if (act === 'delete' || act === 'terminate') return 'danger'
  if (act === 'edit') return 'success'
  if (['start', 'resume', 'pause', 'close'].includes(act)) return 'warning'
  return 'primary'
}

function handleSearch() {
  query.pageNum = 1
  loadList()
}

function handleReset() {
  query.statusList = []
  query.riskLevel = ''
  query.managerId = null
  query.participantId = null
  query.keyword = ''
  query.sort = 'endDate_asc'
  dateRange.value = []
  query.pageNum = 1
  loadList()
}

async function loadList() {
  loading.value = true
  try {
    const [sortBy, order] = query.sort.split('_')
    const params = {
      pageNum: query.pageNum,
      pageSize: query.pageSize,
      sortBy,
      order
    }
    if (query.statusList.length) params.status = query.statusList.join(',')
    if (query.riskLevel) params.riskLevel = query.riskLevel
    if (query.managerId) params.managerId = query.managerId
    if (query.keyword) params.keyword = query.keyword
    if (query.participantId) params.participantId = query.participantId
    if (dateRange.value?.length === 2) {
      params.startDate = dateRange.value[0]
      params.endDate = dateRange.value[1]
    }
    const res = await getProjectList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
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

function goDetail(row) {
  router.push(`/projects/${row.id}/detail`)
}

function openCreateDialog() {
  editingProject.value = null
  dialogVisible.value = true
}

const STATUS_CONFIRM_LABELS = { start: '开始', pause: '暂停', resume: '恢复', close: '结项' }

async function handleProjectAction(row, act) {
  try {
    if (act === 'edit') {
      editingProject.value = row
      dialogVisible.value = true
      return
    }
    if (act === 'delete') {
      await ElMessageBox.confirm(
        `确认删除项目「${row.projectName}」？删除后不可恢复（仅「启动」状态可删除）。`,
        '删除确认',
        { type: 'warning', confirmButtonText: '删除', confirmButtonClass: 'el-button--danger' }
      )
      await deleteProject(row.id)
      ElMessage.success('删除成功')
      loadList()
      return
    }
    if (act === 'terminate') {
      const { value } = await ElMessageBox.prompt(
        `请输入终止项目「${row.projectName}」的原因（必填）：`,
        '终止项目',
        {
          type: 'warning',
          inputType: 'textarea',
          inputValidator: (v) => (v && v.trim() ? true : '终止原因不能为空'),
          confirmButtonText: '终止'
        }
      )
      await changeProjectStatus(row.id, { action: 'terminate', reason: value.trim() })
      ElMessage.success('项目已终止')
      loadList()
      return
    }
    // start / pause / resume / close
    await ElMessageBox.confirm(`确认${STATUS_CONFIRM_LABELS[act] || act}项目「${row.projectName}」？`, '操作确认', {
      type: 'warning'
    })
    await changeProjectStatus(row.id, { action: act })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) {
    if (e === 'cancel' || e === 'close' || e?.message === 'cancel') return
    // 业务错误已由拦截器提示
  }
}

onMounted(() => {
  loadUsers()
  loadList()
})
</script>

<style scoped>
.project-list-page {
  padding: 16px;
}

.filter-card {
  margin-bottom: 12px;
}

.filter-form {
  margin-bottom: -18px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.table-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.total-count {
  font-size: 12px;
  font-weight: 400;
  color: #909399;
  margin-left: 8px;
}

.project-key {
  font-size: 12px;
  color: #909399;
}

.muted {
  font-size: 12px;
  color: #909399;
}

.risk-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
