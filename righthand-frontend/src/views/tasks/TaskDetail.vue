<template>
  <div class="task-detail">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>事项详情</span>
          <el-button @click="goBack" :icon="ArrowLeft" plain>返回列表</el-button>
        </div>
      </template>

      <template v-if="!pageLoading && detailData">
        <!-- 基本信息 -->
        <el-divider content-position="left">基本信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题" :span="2">
            {{ detailData.title || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="getItemTypeTag(detailData.itemType)" size="small">
              {{ getItemTypeText(detailData.itemType) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="优先级">
            <el-tag :type="getPriorityType(detailData.priority)" size="small">
              {{ getPriorityText(detailData.priority) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="分类">
            {{ dictStore.getDictLabel('memo_type', detailData.category) || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(detailData.statusCode)" size="small">
              {{ getStatusText(detailData.statusCode) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="截止时间" :span="2">
            {{ formatTime(detailData.deadline) }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 详细描述 -->
        <el-divider content-position="left">详细描述</el-divider>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="描述">
            <span class="description-text">{{ detailData.description || '暂无描述' }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 周期性设置 -->
        <el-divider content-position="left">周期性设置</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="重复类型">
            {{ recurrenceTypeText }}
          </el-descriptions-item>
          <el-descriptions-item label="结束条件">
            {{ endTypeText }}
          </el-descriptions-item>
          <el-descriptions-item v-if="isDateEnd" label="结束日期" :span="2">
            {{ formatTime(detailData.recurrenceEndDate, 'YYYY-MM-DD') }}
          </el-descriptions-item>
          <el-descriptions-item v-if="recurrenceSummary" label="重复说明" :span="2">
            {{ recurrenceSummary }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 需求关联（仅类型为需求时显示） -->
        <template v-if="detailData.itemType === 'REQUIREMENT'">
          <el-divider content-position="left">需求关联</el-divider>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="关联需求">
              <template v-if="detailData.requirementId">
                <el-link type="primary" :underline="false" @click="showRequirementDetail">
                  [{{ detailData.requirementNo }}] {{ detailData.requirementTitle }}
                </el-link>
              </template>
              <span v-else>未关联需求</span>
            </el-descriptions-item>
          </el-descriptions>
        </template>

        <!-- 操作按钮 -->
        <div class="form-actions">
          <el-button type="primary" @click="goEdit">编辑事项</el-button>
          <el-button @click="goBack">关闭</el-button>
        </div>
      </template>
    </el-card>

    <!-- 需求详情弹窗 -->
    <el-dialog
      v-model="reqDialogVisible"
      title="需求详情"
      width="700px"
      :destroy-on-close="true"
    >
      <el-descriptions v-loading="reqLoading" :column="2" border>
        <el-descriptions-item label="需求编号">{{ reqDetail.requirementNo || '—' }}</el-descriptions-item>
        <el-descriptions-item label="需求标题" :span="2">{{ reqDetail.title || '—' }}</el-descriptions-item>
        <el-descriptions-item label="所属小组">{{ getGroupLabel(reqDetail.group) }}</el-descriptions-item>
        <el-descriptions-item label="需求状态">{{ reqDetail.status || '—' }}</el-descriptions-item>
        <el-descriptions-item label="提出人">{{ reqDetail.proposer || '—' }}</el-descriptions-item>
        <el-descriptions-item label="预计工作量">{{ getEstimatedWorkload(reqDetail.estimatedWorkload) }}</el-descriptions-item>
        <el-descriptions-item label="计划完成日期" :span="2">{{ reqDetail.scheduleDate || '—' }}</el-descriptions-item>
        <el-descriptions-item label="需求描述" :span="2">{{ reqDetail.description || '—' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="reqDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getTaskDetail } from '@/api/task'
import { getRequirementDetail } from '@/api/requirement'
import { useDictStore } from '@/stores/dict'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const dictStore = useDictStore()

// 页面加载状态
const pageLoading = ref(false)

// 详情数据
const detailData = ref(null)

// 获取事项ID
const taskId = computed(() => route.params.id)

// 周期性类型映射
const recurrenceTypeMap = {
  'NONE': '不重复',
  'DAILY': '每天',
  'WEEKLY': '每周',
  'MONTHLY': '每月'
}
const endTypeMap = {
  'NEVER': '永不结束',
  'DATE': '到指定日期',
  'COUNT': '重复次数'
}

// 周期性相关计算属性
const recurrenceTypeText = computed(() => {
  return recurrenceTypeMap[detailData.value?.recurrenceType] || '不重复'
})

const endTypeText = computed(() => {
  return endTypeMap[detailData.value?.recurrenceEndType] || '永不结束'
})

const isDateEnd = computed(() => {
  return detailData.value?.recurrenceEndType === 'DATE'
})

// 重复说明
const recurrenceSummary = computed(() => {
  if (!detailData.value) return ''
  const data = detailData.value
  const type = data.recurrenceType
  if (!type || type === 'NONE') return ''

  let desc = ''
  // 解析 recurrenceRule 获取周几/几号
  let rule = {}
  if (data.recurrenceRule && data.recurrenceRule !== '{}') {
    try { rule = JSON.parse(data.recurrenceRule) } catch (e) { rule = {} }
  }

  switch (type) {
    case 'DAILY':
      desc = '每天重复'
      break
    case 'WEEKLY': {
      const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
      const days = (rule.weekDays || []).map(d => dayNames[d]).join('、')
      desc = `每周${days || ''}重复`
      break
    }
    case 'MONTHLY':
      desc = rule.monthDay === 0 ? '每月月末重复' : `每月${rule.monthDay}号重复`
      break
  }

  if (data.recurrenceEndType === 'DATE' && data.recurrenceEndDate) {
    desc += `，到 ${dayjs(data.recurrenceEndDate).format('YYYY-MM-DD')} 结束`
  } else {
    desc += '，永不结束'
  }
  return desc
})

// 加载事项详情
const loadTaskDetail = async () => {
  pageLoading.value = true
  try {
    const res = await getTaskDetail(taskId.value)
    if (res.code === 200) {
      detailData.value = res.data
    } else {
      ElMessage.error('加载事项详情失败')
      goBack()
    }
  } catch (error) {
    ElMessage.error('加载事项详情失败')
    goBack()
  } finally {
    pageLoading.value = false
  }
}

// 返回来源页
const goBack = () => {
  const from = route.query.from
  router.push(from === 'dashboard' ? '/dashboard' : '/tasks')
}

// 跳转到编辑页
const goEdit = () => {
  const from = route.query.from
  router.push({
    path: `/tasks/${taskId.value}/edit`,
    query: from ? { from } : undefined
  })
}

// 获取类型标签类型
const getItemTypeTag = (type) => {
  const map = { TASK: undefined, MEMO: 'success', REQUIREMENT: 'warning' }
  return map[type] || undefined
}

// 获取类型文本
const getItemTypeText = (type) => {
  const map = { TASK: '任务', MEMO: '备忘', REQUIREMENT: '需求' }
  return map[type] || type || '-'
}

// 获取状态标签类型
const getStatusType = (statusCode) => {
  const map = { TODO: undefined, DOING: 'warning', DONE: 'success', CANCELLED: 'info' }
  return map[statusCode] || undefined
}

// 获取状态文本
const getStatusText = (statusCode) => {
  const map = { TODO: '待办', DOING: '进行中', DONE: '已完成', CANCELLED: '已取消' }
  return map[statusCode] || statusCode || '-'
}

// 获取优先级标签类型
const getPriorityType = (priority) => {
  const map = { HIGH: 'danger', MEDIUM: 'warning', LOW: 'info' }
  return map[priority] || ''
}

// 获取优先级文本
const getPriorityText = (priority) => {
  const map = { HIGH: '高', MEDIUM: '中', LOW: '低' }
  return map[priority] || priority || '-'
}

// 格式化时间
const formatTime = (time, fmt = 'YYYY-MM-DD HH:mm:ss') => {
  return time ? dayjs(time).format(fmt) : '-'
}

// 需求详情弹窗
const reqDialogVisible = ref(false)
const reqLoading = ref(false)
const reqDetail = ref({})

const showRequirementDetail = async () => {
  if (!detailData.value?.requirementId) return
  reqDialogVisible.value = true
  reqLoading.value = true
  try {
    const res = await getRequirementDetail(detailData.value.requirementId)
    reqDetail.value = res.data || {}
  } catch (error) {
    ElMessage.error('加载需求详情失败')
  } finally {
    reqLoading.value = false
  }
}

const getGroupLabel = (group) => {
  const map = { LOAN_MIDDLE: '贷中', LOAN_PRE: '贷前', RISK: '风险' }
  return map[group] || group || '—'
}

const getEstimatedWorkload = (val) => {
  if (val == null) return '—'
  const num = typeof val === 'object' ? val.Float64 : parseFloat(val)
  if (isNaN(num)) return '—'
  return `${num} 天`
}

onMounted(async () => {
  await dictStore.loadAllDicts()
  if (taskId.value) {
    loadTaskDetail()
  } else {
    goBack()
  }
})
</script>

<style scoped lang="scss">
.task-detail {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 16px;
    font-weight: 600;
  }

  :deep(.el-divider__text) {
    font-size: 14px;
    font-weight: 600;
    color: #303133;
  }

  :deep(.el-descriptions) {
    margin-bottom: 8px;
  }

  .description-text {
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.6;
  }

  .form-actions {
    margin-top: 32px;
    padding-top: 20px;
    border-top: 1px solid #ebeef5;
    text-align: center;

    .el-button {
      min-width: 120px;
      margin: 0 8px;
    }
  }
}
</style>
