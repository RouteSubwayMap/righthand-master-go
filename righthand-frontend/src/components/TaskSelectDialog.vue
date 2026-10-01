<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="900px"
    append-to-body
    :close-on-click-modal="false"
  >
    <!-- 筛选区（与「事项管理」列表口径一致） -->
    <el-form :inline="true" :model="filters" @submit.prevent>
      <el-form-item label="标题">
        <el-input
          v-model="filters.title"
          placeholder="标题关键字"
          style="width: 180px"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
      </el-form-item>
      <el-form-item label="类型">
        <el-select v-model="filters.itemType" placeholder="全部" clearable style="width: 110px" @change="handleSearch">
          <el-option v-for="(meta, value) in TASK_ITEM_TYPE" :key="value" :label="meta.label" :value="value" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="filters.statusCode" placeholder="全部" clearable style="width: 110px" @change="handleSearch">
          <el-option v-for="(meta, value) in TASK_STATUS" :key="value" :label="meta.label" :value="value" />
        </el-select>
      </el-form-item>
      <el-form-item label="优先级">
        <el-select v-model="filters.priority" placeholder="全部" clearable style="width: 110px" @change="handleSearch">
          <el-option v-for="(meta, value) in TASK_PRIORITY" :key="value" :label="meta.label" :value="value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table
      v-loading="loading"
      :data="list"
      stripe
      highlight-current-row
      style="width: 100%"
      @row-click="handleRowClick"
      @row-dblclick="handleRowDblclick"
    >
      <el-table-column width="50" align="center">
        <template #default="{ row }">
          <el-radio v-model="selectedId" :value="row.id"><span /></el-radio>
        </template>
      </el-table-column>
      <el-table-column label="标题" min-width="220">
        <template #default="{ row }">
          <el-tooltip placement="top" :max-width="320" :content="row.description || '无描述'">
            <span>{{ row.title }}</span>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="itemTypeMeta(row.itemType).tagType" size="small" effect="plain">
            {{ itemTypeMeta(row.itemType).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="优先级" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="priorityMeta(row.priority).tagType" size="small" effect="plain">
            {{ priorityMeta(row.priority).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="statusMeta(row.statusCode).tagType" size="small">
            {{ statusMeta(row.statusCode).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="截止时间" width="120">
        <template #default="{ row }">{{ formatDate(row.deadline) }}</template>
      </el-table-column>
      <el-table-column label="进度" width="120">
        <template #default="{ row }">
          <el-progress :percentage="row.progress || 0" :stroke-width="8" />
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="没有符合条件的事项" :image-size="80" />
      </template>
    </el-table>

    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="pagination.pageNum"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        background
        @current-change="loadList"
        @size-change="handleSearch"
      />
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :disabled="!selectedId" @click="handleConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
/**
 * 事项选择弹窗（通用组件）
 *
 * 交互与「事项管理」列表页保持一致：标题/类型/状态/优先级筛选 + 分页表格。
 * 单选后点击「确定」（或双击行）向父组件抛出 select 事件并关闭。
 */
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { getTaskList } from '@/api/task'
import { TASK_STATUS, TASK_PRIORITY, TASK_ITEM_TYPE } from '@/utils/projectEnums'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '选择事项' },
  // 打开时的默认类型筛选（空串 = 全部类型）
  itemType: { type: String, default: '' },
  // 打开时的预选事项 ID（用于回显上次选择）
  selectedId: { type: Number, default: null }
})

const emit = defineEmits(['update:modelValue', 'select'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const loading = ref(false)
const list = ref([])
const selectedId = ref(null)

const filters = reactive({
  title: '',
  itemType: '',
  statusCode: '',
  priority: ''
})

const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

watch(
  () => props.modelValue,
  (v) => {
    if (v) openDialog()
  },
  { immediate: true }
)

function openDialog() {
  filters.title = ''
  filters.itemType = props.itemType || ''
  filters.statusCode = ''
  filters.priority = ''
  pagination.pageNum = 1
  selectedId.value = props.selectedId ?? null
  loadList()
}

async function loadList() {
  loading.value = true
  try {
    const params = {
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize
    }
    if (filters.title) params.title = filters.title
    if (filters.itemType) params.itemType = filters.itemType
    if (filters.statusCode) params.statusCode = filters.statusCode
    if (filters.priority) params.priority = filters.priority
    const res = await getTaskList(params)
    const data = res?.data || {}
    list.value = data.records || []
    pagination.total = data.total || 0
  } catch {
    // 错误已由请求拦截器提示；此处兜底空态，避免弹窗残留旧数据
    list.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.pageNum = 1
  loadList()
}

function handleReset() {
  filters.title = ''
  filters.itemType = props.itemType || ''
  filters.statusCode = ''
  filters.priority = ''
  pagination.pageNum = 1
  loadList()
}

function handleRowClick(row) {
  selectedId.value = row.id
}

function handleRowDblclick(row) {
  selectedId.value = row.id
  handleConfirm()
}

function handleConfirm() {
  const row = list.value.find((t) => t.id === selectedId.value)
  if (!row) {
    ElMessage.warning('请先选择一条事项')
    return
  }
  emit('select', row)
  visible.value = false
}

function itemTypeMeta(type) {
  return TASK_ITEM_TYPE[type] || { label: type || '-', tagType: 'info' }
}

function statusMeta(code) {
  return TASK_STATUS[code] || { label: code || '-', tagType: 'info' }
}

function priorityMeta(priority) {
  return TASK_PRIORITY[priority] || { label: priority || '-', tagType: 'info' }
}

function formatDate(v) {
  return v ? dayjs(v).format('YYYY-MM-DD') : '-'
}
</script>

<style scoped>
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
