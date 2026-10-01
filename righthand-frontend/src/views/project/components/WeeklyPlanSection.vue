<template>
  <div class="weekly-plan-section">
    <div class="section-header">
      <div class="section-title">
        周计划时间轴
        <span class="muted">共 {{ weeks.length }} 周，已完成 {{ doneCount }} 周</span>
      </div>
      <el-button
        v-if="canAddWeek"
        type="primary"
        size="small"
        @click="openCreateDialog"
      >
        <el-icon><Plus /></el-icon>
        <span>新增周计划</span>
      </el-button>
    </div>

    <el-empty v-if="!weeks.length" description="暂无周计划，请点击「新增周计划」开始排期" :image-size="80" />

    <div v-else class="week-timeline">
      <div v-for="week in weeks" :key="week.id" class="week-item" :class="{ 'week-done': week.status === 'DONE' }">
        <div class="week-dot" :style="{ backgroundColor: weekMeta(week.status).color }"></div>
        <div class="week-card">
          <div class="week-card-header">
            <div class="week-title-wrap">
              <span class="week-no">第{{ week.weekNo }}周</span>
              <span class="week-title">{{ week.title }}</span>
              <el-tag :type="weekMeta(week.status).tagType" size="small">{{ weekMeta(week.status).label }}</el-tag>
              <el-tag v-if="week.overdue" type="danger" size="small" effect="dark">已延期</el-tag>
            </div>
            <div class="week-actions">
              <el-button
                v-if="canAddTaskToWeek(week)"
                link
                type="primary"
                size="small"
                @click="$emit('add-task', week.id)"
              >
                新建任务
              </el-button>
              <template v-for="act in week.allowedActions || []" :key="act">
                <el-dropdown
                  v-if="act === 'changeStatus'"
                  trigger="click"
                  @command="(cmd) => handleWeekStatus(week, cmd)"
                >
                  <el-button link type="primary" size="small">状态流转</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item v-for="a in statusActionsFor(week)" :key="a.value" :command="a.value">
                        {{ a.label }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                <el-button
                  v-else-if="act === 'submitSummary'"
                  link
                  type="success"
                  size="small"
                  @click="openSummaryDialog(week)"
                >
                  填写总结
                </el-button>
                <el-button
                  v-else-if="act === 'unlockSummary'"
                  link
                  type="warning"
                  size="small"
                  @click="handleUnlock(week)"
                >
                  解锁总结
                </el-button>
                <el-button v-else-if="act === 'edit'" link type="primary" size="small" @click="openEditDialog(week)">
                  编辑
                </el-button>
                <el-button v-else-if="act === 'delete'" link type="danger" size="small" @click="handleDelete(week)">
                  删除
                </el-button>
              </template>
            </div>
          </div>

          <div class="week-meta">
            <span><el-icon><Calendar /></el-icon> {{ week.startDate }} ~ {{ week.endDate }}</span>
            <span><el-icon><User /></el-icon> 负责人：{{ week.ownerName || '-' }}</span>
            <span><el-icon><Tickets /></el-icon> 任务：{{ week.taskCount || 0 }} 个</span>
          </div>

          <div v-if="week.planDesc" class="week-plan-desc">{{ week.planDesc }}</div>

          <!-- 已提交的总结（锁定态） -->
          <div v-if="weekSummary(week)" class="week-summary">
            <div class="summary-header">
              <span class="summary-title">周总结</span>
              <span class="muted">
                {{ week.summaryByName || '-' }} 提交于 {{ formatDateTime(week.summaryAt) }}
                <el-tag size="small" type="info" effect="plain">已锁定</el-tag>
              </span>
            </div>
            <div class="summary-grid">
              <div class="summary-item">
                <div class="summary-label">本周完成</div>
                <div class="summary-text">{{ weekSummary(week).finished || '-' }}</div>
              </div>
              <div class="summary-item">
                <div class="summary-label">下周计划</div>
                <div class="summary-text">{{ weekSummary(week).nextPlan || '-' }}</div>
              </div>
              <div class="summary-item">
                <div class="summary-label">风险与问题</div>
                <div class="summary-text">{{ weekSummary(week).risks || '-' }}</div>
              </div>
              <div class="summary-item">
                <div class="summary-label">所需支持</div>
                <div class="summary-text">{{ weekSummary(week).support || '-' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增/编辑周计划弹窗 -->
    <el-dialog
      v-model="weekDialogVisible"
      :title="editingWeek ? `编辑第${editingWeek.weekNo}周计划` : '新增周计划'"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form ref="weekFormRef" :model="weekForm" :rules="weekRules" label-width="90px">
        <el-form-item label="周主题" prop="title">
          <el-input v-model="weekForm.title" maxlength="100" placeholder="如：需求调研与接口设计" />
        </el-form-item>
        <el-form-item label="起止日期" prop="startDate">
          <div class="date-range">
            <el-date-picker v-model="weekForm.startDate" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" style="width: 100%" />
            <span class="date-sep">至</span>
            <el-date-picker v-model="weekForm.endDate" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" style="width: 100%" />
          </div>
        </el-form-item>
        <el-form-item label="负责人" prop="ownerId">
          <el-select v-model="weekForm.ownerId" filterable placeholder="请选择本周负责人" style="width: 100%">
            <el-option v-for="u in users" :key="u.id" :label="userLabel(u)" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="计划内容" prop="planDesc">
          <el-input v-model="weekForm.planDesc" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="本周计划完成的工作内容" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="weekForm.remark" maxlength="200" placeholder="备注（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="weekDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingWeek" @click="handleWeekSubmit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 周总结弹窗 -->
    <el-dialog v-model="summaryDialogVisible" :title="`填写第${summaryWeek?.weekNo}周总结`" width="600px" :close-on-click-modal="false">
      <el-alert type="info" :closable="false" show-icon class="summary-tip">
        提交后总结将被锁定，如需修改请联系项目经理/管理员解锁。
      </el-alert>
      <el-form ref="summaryFormRef" :model="summaryForm" :rules="summaryRules" label-width="100px">
        <el-form-item label="本周完成" prop="finished">
          <el-input v-model="summaryForm.finished" type="textarea" :rows="3" maxlength="1000" show-word-limit placeholder="本周实际完成的工作（必填）" />
        </el-form-item>
        <el-form-item label="下周计划" prop="nextPlan">
          <el-input v-model="summaryForm.nextPlan" type="textarea" :rows="3" maxlength="1000" show-word-limit placeholder="下周计划开展的工作（必填）" />
        </el-form-item>
        <el-form-item label="风险与问题" prop="risks">
          <el-input v-model="summaryForm.risks" type="textarea" :rows="2" maxlength="1000" show-word-limit placeholder="遇到的风险/阻塞（可选）" />
        </el-form-item>
        <el-form-item label="所需支持" prop="support">
          <el-input v-model="summaryForm.support" type="textarea" :rows="2" maxlength="1000" show-word-limit placeholder="需要项目组提供的支持（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="summaryDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingSummary" @click="handleSummarySubmit">提交并锁定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Calendar, User, Tickets } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { createWeekPlan, updateWeekPlan, deleteWeekPlan, changeWeekStatus, submitWeekSummary, unlockWeekSummary } from '@/api/project'
import { WEEK_STATUS, WEEK_STATUS_ACTION_LABELS, parseSummary } from '@/utils/projectEnums'

const props = defineProps({
  project: { type: Object, required: true },
  users: { type: Array, default: () => [] }
})

const emit = defineEmits(['refresh', 'add-task'])

const weeks = computed(() => props.project?.weeks || [])
const doneCount = computed(() => weeks.value.filter((w) => w.status === 'DONE').length)
const canAddWeek = computed(() => (props.project?.allowedActions || []).includes('addWeek'))

function userLabel(u) {
  return u.realName || u.nickname || u.username || `用户${u.id}`
}

function weekMeta(status) {
  return WEEK_STATUS[status] || { label: status || '-', tagType: 'info', color: '#909399' }
}

function formatDateTime(v) {
  if (!v) return '-'
  return dayjs(v).format('YYYY-MM-DD HH:mm')
}

function weekSummary(week) {
  if (!week.summary) return null
  return parseSummary(week.summary)
}

/** 周卡片「新建任务」入口：周未锁定且当前用户是该周负责人/PM（allowedActions 含 edit） */
function canAddTaskToWeek(week) {
  if (week.summaryAt) return false
  return (week.allowedActions || []).includes('edit')
}

/* ------------------------------- 状态流转 ------------------------------- */

const STATUS_ACTIONS_BY_STATUS = {
  NOT_STARTED: ['start', 'skip'],
  IN_PROGRESS: ['done', 'skip'],
  DONE: ['reopen'],
  SKIPPED: ['reopen']
}

function statusActionsFor(week) {
  return (STATUS_ACTIONS_BY_STATUS[week.status] || []).map((v) => ({
    value: v,
    label: WEEK_STATUS_ACTION_LABELS[v] || v
  }))
}

async function handleWeekStatus(week, action) {
  if (!week) return
  try {
    await changeWeekStatus(props.project.id, week.id, { action })
    ElMessage.success('状态已更新')
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  }
}

/* ------------------------------- 新增/编辑 ------------------------------- */

const weekDialogVisible = ref(false)
const savingWeek = ref(false)
const editingWeek = ref(null)
const weekFormRef = ref(null)

const weekForm = reactive({
  title: '',
  startDate: '',
  endDate: '',
  ownerId: null,
  planDesc: '',
  remark: ''
})

const weekRules = {
  title: [{ required: true, message: '请输入周主题', trigger: 'blur' }],
  startDate: [
    {
      validator: (rule, value, callback) => {
        if (!weekForm.startDate || !weekForm.endDate) return callback(new Error('请选择起止日期'))
        if (weekForm.endDate < weekForm.startDate) return callback(new Error('结束日期不能早于开始日期'))
        callback()
      },
      trigger: 'change'
    }
  ],
  ownerId: [{ required: true, message: '请选择负责人', trigger: 'change' }]
}

function openCreateDialog() {
  editingWeek.value = null
  // 默认起点：已有最后一周结束日的次日；否则项目开始日
  const last = weeks.value[weeks.value.length - 1]
  const start = last ? dayjs(last.endDate).add(1, 'day') : dayjs(props.project.startDate)
  weekForm.title = ''
  weekForm.startDate = start.format('YYYY-MM-DD')
  weekForm.endDate = start.add(4, 'day').format('YYYY-MM-DD')
  weekForm.ownerId = props.project.managerId || null
  weekForm.planDesc = ''
  weekForm.remark = ''
  weekDialogVisible.value = true
  nextTick(() => weekFormRef.value?.clearValidate())
}

function openEditDialog(week) {
  editingWeek.value = week
  weekForm.title = week.title || ''
  weekForm.startDate = week.startDate || ''
  weekForm.endDate = week.endDate || ''
  weekForm.ownerId = week.ownerId || null
  weekForm.planDesc = week.planDesc || ''
  weekForm.remark = week.remark || ''
  weekDialogVisible.value = true
  nextTick(() => weekFormRef.value?.clearValidate())
}

async function handleWeekSubmit() {
  try {
    await weekFormRef.value.validate()
  } catch {
    return
  }
  savingWeek.value = true
  try {
    const payload = { ...weekForm }
    if (editingWeek.value) {
      await updateWeekPlan(props.project.id, editingWeek.value.id, payload)
      ElMessage.success('周计划已更新')
    } else {
      await createWeekPlan(props.project.id, payload)
      ElMessage.success('周计划已创建')
    }
    weekDialogVisible.value = false
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  } finally {
    savingWeek.value = false
  }
}

async function handleDelete(week) {
  try {
    await ElMessageBox.confirm(`确认删除第${week.weekNo}周计划「${week.title}」？该周关联的任务将被移出。`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除'
    })
  } catch {
    return
  }
  try {
    await deleteWeekPlan(props.project.id, week.id)
    ElMessage.success('删除成功')
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  }
}

/* ------------------------------- 总结 ------------------------------- */

const summaryDialogVisible = ref(false)
const savingSummary = ref(false)
const summaryWeek = ref(null)
const summaryFormRef = ref(null)

const summaryForm = reactive({
  finished: '',
  nextPlan: '',
  risks: '',
  support: ''
})

const summaryRules = {
  finished: [{ required: true, message: '请填写本周完成内容', trigger: 'blur' }],
  nextPlan: [{ required: true, message: '请填写下周计划', trigger: 'blur' }]
}

function openSummaryDialog(week) {
  summaryWeek.value = week
  const exist = weekSummary(week)
  summaryForm.finished = exist?.finished || ''
  summaryForm.nextPlan = exist?.nextPlan || ''
  summaryForm.risks = exist?.risks || ''
  summaryForm.support = exist?.support || ''
  summaryDialogVisible.value = true
  nextTick(() => summaryFormRef.value?.clearValidate())
}

async function handleSummarySubmit() {
  try {
    await summaryFormRef.value.validate()
  } catch {
    return
  }
  savingSummary.value = true
  try {
    await submitWeekSummary(props.project.id, summaryWeek.value.id, { summary: { ...summaryForm } })
    ElMessage.success('总结已提交并锁定')
    summaryDialogVisible.value = false
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  } finally {
    savingSummary.value = false
  }
}

async function handleUnlock(week) {
  try {
    await ElMessageBox.confirm(`确认解锁第${week.weekNo}周的总结？解锁后该周计划恢复可编辑。`, '解锁确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    await unlockWeekSummary(props.project.id, week.id)
    ElMessage.success('已解锁')
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  }
}
</script>

<style scoped>
.weekly-plan-section {
  min-height: 200px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.muted {
  font-size: 12px;
  color: #909399;
  font-weight: 400;
  margin-left: 6px;
}

.week-timeline {
  position: relative;
  padding-left: 20px;
}

.week-timeline::before {
  content: '';
  position: absolute;
  left: 5px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: #e4e7ed;
}

.week-item {
  position: relative;
  margin-bottom: 16px;
}

.week-dot {
  position: absolute;
  left: -20px;
  top: 16px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 2px solid #fff;
  box-shadow: 0 0 0 1px #e4e7ed;
}

.week-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 12px 16px;
  background: #fff;
  transition: box-shadow 0.2s;
}

.week-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.week-done .week-card {
  background: #fafcfa;
}

.week-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
}

.week-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.week-no {
  font-weight: 600;
  color: #409eff;
}

.week-title {
  font-weight: 600;
  color: #303133;
}

.week-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.week-meta {
  display: flex;
  gap: 18px;
  flex-wrap: wrap;
  margin-top: 8px;
  font-size: 12px;
  color: #606266;
}

.week-meta .el-icon {
  vertical-align: -2px;
  margin-right: 2px;
}

.week-plan-desc {
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 8px 10px;
  white-space: pre-wrap;
}

.week-summary {
  margin-top: 10px;
  border-top: 1px dashed #e4e7ed;
  padding-top: 10px;
}

.summary-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.summary-title {
  font-weight: 600;
  font-size: 13px;
  color: #303133;
}

.summary-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 16px;
}

.summary-item {
  font-size: 12px;
}

.summary-label {
  color: #909399;
  margin-bottom: 2px;
}

.summary-text {
  color: #606266;
  white-space: pre-wrap;
}

.date-range {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.date-sep {
  color: #909399;
  flex-shrink: 0;
}

.summary-tip {
  margin-bottom: 12px;
}
</style>
