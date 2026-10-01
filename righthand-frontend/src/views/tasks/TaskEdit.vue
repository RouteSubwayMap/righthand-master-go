<template>
  <div class="task-edit">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑事项' : '新建事项' }}</span>
          <el-button @click="goBack" :icon="ArrowLeft" plain>返回列表</el-button>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
        label-position="right"
      >
        <!-- 基本信息 -->
        <el-divider content-position="left">基本信息</el-divider>

        <el-form-item label="标题" prop="title">
          <el-input v-model="formData.title" placeholder="请输入标题" maxlength="100" show-word-limit />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="类型" prop="itemType">
              <el-select v-model="formData.itemType" style="width: 100%">
                <el-option label="任务" value="TASK" />
                <el-option label="备忘" value="MEMO" />
                <el-option label="需求" value="REQUIREMENT" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="formData.priority" style="width: 100%">
                <el-option label="高" value="HIGH" />
                <el-option label="中" value="MEDIUM" />
                <el-option label="低" value="LOW" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="分类" prop="category">
              <el-select v-model="formData.category" placeholder="请选择分类" style="width: 100%" clearable>
                <el-option
                  v-for="item in categoryOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="截止时间" prop="deadline">
              <el-date-picker
                v-model="formData.deadline"
                type="datetime"
                placeholder="选择截止时间"
                format="YYYY-MM-DD HH:mm"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 详细描述 -->
        <el-divider content-position="left">详细描述</el-divider>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="5"
            placeholder="请输入事项的详细描述"
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>

        <!-- 周期性设置 -->
        <el-divider content-position="left">周期性设置</el-divider>
        <RecurrenceConfig v-model="formData.recurrenceConfig" />

        <!-- 需求关联（仅类型为需求时显示） -->
        <el-divider v-if="formData.itemType === 'REQUIREMENT'" content-position="left">需求关联</el-divider>
        <el-form-item v-if="formData.itemType === 'REQUIREMENT'" label="关联需求">
          <div v-if="formData.requirementId" style="display: flex; align-items: center; gap: 12px;">
            <el-link type="primary" :underline="false" @click="goToRequirementDetail">
              {{ formData.requirementNo || '查看' }} - {{ formData.requirementTitle || '' }}
            </el-link>
            <el-button type="danger" size="small" link @click="clearRequirement">清除关联</el-button>
            <el-button type="primary" size="small" link @click="openRequirementDialog">重新选择</el-button>
          </div>
          <el-link v-else type="primary" :underline="false" @click="openRequirementDialog">
            点击选择需求
          </el-link>
        </el-form-item>

        <!-- 操作按钮 -->
        <div class="form-actions">
          <el-button type="primary" @click="submitForm" :loading="submitting">
            {{ isEdit ? '保存修改' : '创建事项' }}
          </el-button>
          <el-button @click="goBack">取消</el-button>
        </div>
      </el-form>
    </el-card>

    <!-- 需求选择弹窗 -->
    <el-dialog
      v-model="requirementDialogVisible"
      title="选择需求"
      width="900px"
      :close-on-click-modal="false"
      append-to-body
    >
      <div style="margin-bottom: 12px; display: flex; align-items: center; gap: 8px; flex-wrap: wrap;">
        <el-input
          v-model="requirementSearchNo"
          placeholder="需求编号"
          clearable
          style="width: 180px"
          @keyup.enter="searchRequirement"
        />
        <el-input
          v-model="requirementSearchTitle"
          placeholder="需求标题"
          clearable
          style="width: 240px"
          @keyup.enter="searchRequirement"
        />
        <el-button size="small" type="primary" @click="searchRequirement">搜索</el-button>
        <el-button size="small" @click="resetRequirementSearch">重置</el-button>
      </div>
      <el-table
        :data="requirementList"
        v-loading="requirementLoading"
        max-height="400"
        border
        size="small"
        style="width: 100%"
      >
        <el-table-column prop="requirementNo" label="需求编号" width="140" />
        <el-table-column prop="title" label="需求标题" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getRequirementStatusTag(row.status)">
              {{ getRequirementStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.priority === 'HIGH' ? 'danger' : row.priority === 'MEDIUM' ? 'warning' : 'info'">
              {{ dictStore.getDictLabel('priority', row.priority) || row.priority }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="bankName" label="所属银行" width="120" show-overflow-tooltip />
        <el-table-column label="选择" width="70" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="handleRequirementSelect(row)">选择</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top: 12px; display: flex; justify-content: flex-end;">
        <el-pagination
          v-model:current-page="requirementPagination.pageNum"
          v-model:page-size="requirementPagination.pageSize"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          :total="requirementPagination.total"
          @size-change="loadRequirementList"
          @current-change="loadRequirementList"
        />
      </div>
    </el-dialog>

    <!-- 需求详情弹窗 -->
    <el-dialog
      v-model="requirementDetailVisible"
      title="需求详情"
      width="900px"
      append-to-body
    >
      <div v-if="currentRequirement" class="requirement-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="需求编号">{{ currentRequirement.requirementNo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="需求名称">{{ currentRequirement.title }}</el-descriptions-item>
          <el-descriptions-item label="所属银行">{{ currentRequirement.bankName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="优先级">
            <el-tag :type="currentRequirement.priority === 'HIGH' ? 'danger' : currentRequirement.priority === 'MEDIUM' ? 'warning' : 'info'" size="small">
              {{ currentRequirement.priority === 'HIGH' ? '高' : currentRequirement.priority === 'MEDIUM' ? '中' : '低' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <el-tag :type="currentRequirement.riskLevel === 'HIGH' ? 'danger' : currentRequirement.riskLevel === 'MEDIUM' ? 'warning' : 'info'" size="small">
              {{ currentRequirement.riskLevel === 'HIGH' ? '高' : currentRequirement.riskLevel === 'MEDIUM' ? '中' : '低' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="需求类型">
            <el-tag :type="currentRequirement.requirementType === 'BANK' ? 'primary' : 'success'" size="small">
              {{ currentRequirement.requirementType === 'BANK' ? '行内需求' : '内部优化' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="reqStatusTagType(currentRequirement.status)" size="small">
              {{ reqStatusLabel(currentRequirement.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="主 R 系统">{{ currentRequirement.mainSystem || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开发主 R|参与者">{{ currentRequirement.developer || '-' }}</el-descriptions-item>
          <el-descriptions-item label="涉及系统">{{ currentRequirement.involvedSystems || '-' }}</el-descriptions-item>
          <el-descriptions-item label="所属小组">{{ getGroupLabel(currentRequirement.groupType) }}</el-descriptions-item>
          <el-descriptions-item label="需求跟进人">{{ currentRequirement.followerName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="提交人">{{ currentRequirement.submitterName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="需求提出时间">{{ formatReqDateTime(currentRequirement.submitTime) }}</el-descriptions-item>
          <el-descriptions-item label="需求完成时间">{{ formatReqDateTime(currentRequirement.completionDate) }}</el-descriptions-item>
          <el-descriptions-item label="预计工作量（人天）">{{ getEstimatedWorkload(currentRequirement.estimatedWorkload) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">需求描述</el-divider>
        <div style="white-space: pre-wrap; line-height: 1.8; color: #606266;">
          {{ currentRequirement.description || '无' }}
        </div>

        <el-divider content-position="left">备注</el-divider>
        <div style="white-space: pre-wrap; line-height: 1.8; color: #606266;">
          {{ currentRequirement.remark || '-' }}
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import RecurrenceConfig from '@/components/RecurrenceConfig.vue'
import { createTask, updateTask, getTaskDetail } from '@/api/task'
import { getRequirementPage, getRequirementDetail } from '@/api/requirement'
import { useDictStore } from '@/stores/dict'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const dictStore = useDictStore()

// 需求状态标签
const getRequirementStatusLabel = (status) => {
  const map = {
    'PENDING_CLARIFY': '待澄清',
    'PENDING_SCHEDULING': '待排期',
    'DEVELOPING': '开发中',
    'TESTING': '测试中',
    'COMPLETED': '已完成'
  }
  return map[status] || status
}

const getRequirementStatusTag = (status) => {
  const map = {
    'PENDING_CLARIFY': 'info',
    'PENDING_SCHEDULING': 'warning',
    'DEVELOPING': 'primary',
    'TESTING': 'warning',
    'COMPLETED': 'success'
  }
  return map[status] || 'info'
}

// 表单引用
const formRef = ref(null)

// 页面加载与提交状态
const pageLoading = ref(false)
const submitting = ref(false)

// 是否为编辑模式（路由中带有 id 参数）
const taskId = computed(() => route.params.id)
const isEdit = computed(() => !!taskId.value)

// 表单校验规则
const rules = {
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' },
    { min: 2, max: 100, message: '标题长度为 2~100 个字符', trigger: 'blur' }
  ],
  itemType: [
    { required: true, message: '请选择类型', trigger: 'change' }
  ],
  priority: [
    { required: true, message: '请选择优先级', trigger: 'change' }
  ]
}

// 分类选项（从 memo_type 字典获取）
const categoryOptions = computed(() => {
  return dictStore.getDictByType('memo_type').map(item => ({
    label: item.dictLabel,
    value: item.dictValue
  }))
})

// 表单数据
const formData = reactive({
  id: null,
  title: '',
  description: '',
  itemType: 'TASK',
  priority: 'MEDIUM',
  category: '',
  deadline: null,
  tags: '',
  requirementId: null,
  requirementNo: '',
  requirementTitle: '',
  // 周期性任务配置
  recurrenceConfig: {
    recurrenceType: 'NONE',
    weekDays: [],
    monthDay: 1,
    endType: 'NEVER',
    endDate: null
  }
})

// ===== 需求关联 =====
const requirementDialogVisible = ref(false)
const requirementList = ref([])
const requirementLoading = ref(false)
const requirementSearchNo = ref('')
const requirementSearchTitle = ref('')
const requirementPagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 需求详情弹窗
const requirementDetailVisible = ref(false)
const currentRequirement = ref(null)
const detailLoading = ref(false)

// 打开需求选择弹窗
const openRequirementDialog = () => {
  requirementDialogVisible.value = true
  requirementPagination.pageNum = 1
  loadRequirementList()
}

// 加载需求列表
const loadRequirementList = async () => {
  requirementLoading.value = true
  try {
    const params = {
      pageNum: requirementPagination.pageNum,
      pageSize: requirementPagination.pageSize
    }
    if (requirementSearchNo.value) {
      params.requirementNo = requirementSearchNo.value
    }
    if (requirementSearchTitle.value) {
      params.title = requirementSearchTitle.value
    }
    const res = await getRequirementPage(params)
    if (res.code === 200) {
      requirementList.value = res.data.records || []
      requirementPagination.total = res.data.total || 0
    } else {
      requirementList.value = []
      requirementPagination.total = 0
    }
  } catch (error) {
    console.error('加载需求列表失败:', error)
    requirementList.value = []
    requirementPagination.total = 0
  } finally {
    requirementLoading.value = false
  }
}

// 搜索需求
const searchRequirement = () => {
  requirementPagination.pageNum = 1
  loadRequirementList()
}

// 重置搜索
const resetRequirementSearch = () => {
  requirementSearchNo.value = ''
  requirementSearchTitle.value = ''
  requirementPagination.pageNum = 1
  loadRequirementList()
}

// 选择需求
const handleRequirementSelect = (row) => {
  formData.requirementId = row.id
  formData.requirementNo = row.requirementNo || ''
  formData.requirementTitle = row.title || ''
  requirementDialogVisible.value = false
}

// 查看需求详情（弹窗）
const goToRequirementDetail = async () => {
  if (!formData.requirementId) return
  detailLoading.value = true
  requirementDetailVisible.value = true
  try {
    const res = await getRequirementDetail(formData.requirementId)
    if (res.code === 200) {
      currentRequirement.value = res.data
    } else {
      ElMessage.error('获取需求详情失败')
    }
  } catch (error) {
    console.error('获取需求详情失败:', error)
    ElMessage.error('获取需求详情失败')
  } finally {
    detailLoading.value = false
  }
}

// 需求状态标签辅助函数
const reqStatusLabel = (status) => {
  const map = {
    'PENDING_CLARIFY': '待澄清',
    'PENDING_SCHEDULING': '待排期',
    'DEVELOPING': '开发中',
    'TESTING': '测试中',
    'PENDING_RELEASE': '待投产',
    'COMPLETED': '已完成'
  }
  return map[status] || status || '-'
}

const reqStatusTagType = (status) => {
  const map = {
    'PENDING_CLARIFY': 'info',
    'PENDING_SCHEDULING': 'warning',
    'DEVELOPING': 'primary',
    'TESTING': 'warning',
    'PENDING_RELEASE': 'danger',
    'COMPLETED': 'success'
  }
  return map[status] || 'info'
}

const formatReqDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const d = dayjs(dateStr)
  return d.isValid() ? d.format('YYYY-MM-DD HH:mm:ss') : dateStr
}

// 所属小组标签（与需求管理页面一致）
const getGroupLabel = (groupType) => {
  const map = {
    'LOAN_MIDDLE': '贷中',
    'LOAN_PRE': '贷前',
    'RISK': '风险'
  }
  return map[groupType] || groupType || '-'
}

// 预计工作量显示值（与需求管理页面一致，处理后端 sql.NullFloat64 对象）
const getEstimatedWorkload = (workload) => {
  if (!workload) return '-'
  if (typeof workload === 'object' && workload.Float64 !== undefined) {
    return workload.Valid ? workload.Float64 : '-'
  }
  if (typeof workload === 'number') {
    return workload
  }
  const num = parseFloat(workload)
  return isNaN(num) ? '-' : num
}

// 清除需求关联
const clearRequirement = () => {
  formData.requirementId = null
  formData.requirementNo = ''
  formData.requirementTitle = ''
}

// 监听类型变化，切换为非需求时清除关联
watch(() => formData.itemType, (newType) => {
  if (newType !== 'REQUIREMENT') {
    formData.requirementId = null
    formData.requirementNo = ''
    formData.requirementTitle = ''
  }
})

// 重置表单
const resetForm = () => {
  Object.assign(formData, {
    id: null,
    title: '',
    description: '',
    itemType: 'TASK',
    priority: 'MEDIUM',
    category: '',
    deadline: null,
    tags: '',
    requirementId: null,
    requirementNo: '',
    requirementTitle: '',
    recurrenceConfig: {
      recurrenceType: 'NONE',
      weekDays: [],
      monthDay: 1,
      endType: 'NEVER',
      endDate: null
    }
  })
  formRef.value?.clearValidate()
}

// 加载事项详情（编辑模式）
const loadTaskDetail = async () => {
  pageLoading.value = true
  try {
    const res = await getTaskDetail(taskId.value)
    if (res.code === 200) {
      const row = res.data
      // 解析周期性配置
      let recurrenceRule = {}
      if (row.recurrenceRule && row.recurrenceRule !== '{}') {
        try {
          recurrenceRule = JSON.parse(row.recurrenceRule)
        } catch (e) {
          recurrenceRule = {}
        }
      }

      // 格式化结束日期为 YYYY-MM-DD 字符串格式（组件期望的格式）
      let endDateStr = null
      if (row.recurrenceEndDate) {
        endDateStr = dayjs(row.recurrenceEndDate).format('YYYY-MM-DD')
      }

      Object.assign(formData, {
        id: row.id,
        title: row.title,
        description: row.description,
        itemType: row.itemType,
        priority: row.priority,
        category: row.category,
        deadline: row.deadline ? dayjs(row.deadline).format('YYYY-MM-DD HH:mm:ss') : null,
        tags: row.tags,
        requirementId: row.requirementId ?? null,
        requirementNo: row.requirementNo ?? '',
        requirementTitle: row.requirementTitle ?? '',
        recurrenceConfig: {
          recurrenceType: row.recurrenceType || 'NONE',
          weekDays: recurrenceRule.weekDays || [],
          monthDay: recurrenceRule.monthDay ?? 1,
          endType: row.recurrenceEndType || 'NEVER',
          endDate: endDateStr
        }
      })

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

// 提交表单
const submitForm = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const { recurrenceConfig, requirementNo, requirementTitle, ...baseData } = formData
      const data = {
        ...baseData,
        deadline: formData.deadline ? dayjs(formData.deadline).format('YYYY-MM-DD HH:mm:ss') : null,
        // 类型非需求时清除需求关联
        requirementId: formData.itemType === 'REQUIREMENT' ? formData.requirementId : null,
        // 展开周期性配置字段
        recurrenceType: recurrenceConfig.recurrenceType,
        recurrenceRule: recurrenceConfig.recurrenceType !== 'NONE' ? JSON.stringify(recurrenceConfig) : '{}',
        recurrenceEndType: recurrenceConfig.endType,
        recurrenceEndDate: recurrenceConfig.endDate ? dayjs(recurrenceConfig.endDate).format('YYYY-MM-DD HH:mm:ss') : null,
        recurrenceCount: 0
      }

      if (isEdit.value) {
        await updateTask(formData.id, data)
        ElMessage.success('更新成功')
      } else {
        await createTask(data)
        ElMessage.success('创建成功')
      }
      // 提交成功后返回来源页
      const from = route.query.from
      router.push(from === 'dashboard' ? '/dashboard' : '/tasks')
    } catch (error) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    } finally {
      submitting.value = false
    }
  })
}

onMounted(async () => {
  // 加载字典数据
  dictStore.loadAllDicts()
  // 编辑模式下加载事项详情
  if (isEdit.value) {
    await loadTaskDetail()
  } else {
    resetForm()
  }
})
</script>

<style scoped lang="scss">
.task-edit {
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

  :deep(.el-form-item) {
    margin-bottom: 22px;
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
