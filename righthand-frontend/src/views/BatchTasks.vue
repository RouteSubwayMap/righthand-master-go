<template>
  <div class="batch-tasks-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>批量任务清单</span>
          <div>
            <el-button type="warning" :icon="Download" :loading="exporting" @click="handleExport">导出Excel</el-button>
            <el-button :icon="Document" @click="handleDownloadTemplate">下载模版</el-button>
            <el-button type="success" :icon="Upload" :loading="importing" @click="triggerImport">导入Excel</el-button>
            <el-button v-permission="'batchTask:add'" type="primary" :icon="Plus" @click="handleCreate">新建任务</el-button>
          </div>
        </div>
      </template>

      <!-- 搜索栏（每行4个字段） -->
      <el-form :model="searchForm" label-width="80px" class="search-form">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-form-item label="银行类型">
              <el-select v-model="searchForm.bankType" placeholder="全部" clearable style="width: 100%" @change="onSearchBankTypeChange">
                <el-option v-for="item in bankTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="银行名称">
              <el-select v-model="searchForm.bankName" placeholder="全部" clearable filterable style="width: 100%">
                <el-option v-for="item in searchBankNameOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="所属系统">
              <el-select v-model="searchForm.systemName" placeholder="全部" clearable filterable style="width: 100%">
                <el-option v-for="item in systemNameOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="任务类型">
              <el-select v-model="searchForm.taskType" placeholder="全部" clearable style="width: 100%">
                <el-option v-for="item in batchTaskTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="6">
            <el-form-item label="任务名">
              <el-input v-model="searchForm.taskName" placeholder="请输入任务名" clearable @keyup.enter="handleSearch" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="任务描述">
              <el-input v-model="searchForm.taskDesc" placeholder="请输入任务描述" clearable @keyup.enter="handleSearch" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="上游依赖系统">
              <el-input v-model="searchForm.upstreamSystem" placeholder="请输入上游依赖系统" clearable @keyup.enter="handleSearch" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="下游依赖系统">
              <el-input v-model="searchForm.downstreamSystem" placeholder="请输入下游依赖系统" clearable @keyup.enter="handleSearch" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
              <el-button :icon="Refresh" @click="handleReset">重置</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <!-- 列表 -->
      <el-table :data="tableData" v-loading="loading" border stripe style="width: 100%">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="bankType" label="银行类型" width="90" align="center" />
        <el-table-column prop="bankName" label="银行名称" width="150" show-overflow-tooltip />
        <el-table-column prop="systemName" label="所属系统" width="100" align="center" />
        <el-table-column prop="taskType" label="任务类型" width="100" align="center" />
        <el-table-column prop="taskName" label="任务名" width="200" show-overflow-tooltip />
        <el-table-column prop="taskDesc" label="任务描述" min-width="150" show-overflow-tooltip />
        <el-table-column prop="execTime" label="执行时间" width="160" show-overflow-tooltip />
        <el-table-column prop="upstreamSystem" label="上游依赖系统" width="150" show-overflow-tooltip />
        <el-table-column prop="downstreamSystem" label="下游依赖系统" width="150" show-overflow-tooltip />
        <el-table-column prop="remark" label="备注" width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'batchTask:view'" type="info" size="small" @click="handleView(row)">查看</el-button>
            <el-button v-permission="'batchTask:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'batchTask:delete'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.pageNum"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
        @size-change="loadList"
        @current-change="loadList"
      />
    </el-card>

    <!-- 隐藏的文件选择 -->
    <input ref="fileInputRef" type="file" accept=".xlsx,.xls" style="display:none" @change="handleFileChange" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, Upload, Download, Document } from '@element-plus/icons-vue'
import { getBatchTaskList, deleteBatchTask, importBatchTasks } from '@/api/batchTask'
import { useDictStore } from '@/stores/dict'
import { useBankOptions } from '@/composables/useBankOptions'
import * as XLSX from 'xlsx'

const router = useRouter()
const dictStore = useDictStore()

// 银行类型 / 银行名称 下拉选项来自银行机构参数管理（不再使用数据字典关联）
const { bankTypeOptions, getBankNameOptions, loadBankParams } = useBankOptions()

// 查询条件：银行名称联动（选了银行类型只展示该类型下的银行名称，未选则展示全部）
const searchBankNameOptions = computed(() => getBankNameOptions(searchForm.bankType))

const systemNameOptions = computed(() => dictStore.getDictByType('system_name').map(item => ({ label: item.dictLabel, value: item.dictValue })))
const batchTaskTypeOptions = computed(() => dictStore.getDictByType('batch_task_type').map(item => ({ label: item.dictLabel, value: item.dictValue })))

const loading = ref(false)
const importing = ref(false)
const exporting = ref(false)
const tableData = ref([])
const fileInputRef = ref()

const searchForm = reactive({
  bankType: '',
  bankName: '',
  systemName: '',
  taskType: '',
  taskName: '',
  taskDesc: '',
  upstreamSystem: '',
  downstreamSystem: ''
})

const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 查询条件银行类型变更时清空银行名称
const onSearchBankTypeChange = () => {
  searchForm.bankName = ''
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await getBatchTaskList({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      ...searchForm
    })
    tableData.value = res.data.records || []
    pagination.total = res.data.total || 0
  } catch (e) {
    ElMessage.error('加载列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadList()
}

const handleReset = () => {
  Object.assign(searchForm, {
    bankType: '',
    bankName: '',
    systemName: '',
    taskType: '',
    taskName: '',
    taskDesc: '',
    upstreamSystem: '',
    downstreamSystem: ''
  })
  pagination.pageNum = 1
  loadList()
}

const handleView = (row) => {
  router.push(`/batch-tasks/${row.id}/detail`)
}

const handleCreate = () => {
  router.push('/batch-tasks/new')
}

const handleEdit = (row) => {
  router.push(`/batch-tasks/${row.id}/edit`)
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确认删除任务「${row.taskName}」吗？`, '提示', {
    type: 'warning'
  }).then(async () => {
    try {
      await deleteBatchTask(row.id)
      ElMessage.success('删除成功')
      loadList()
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

// Excel 列名到字段名的映射（按列顺序）
const columnMap = [
  'bankType', 'bankName', 'systemName', 'taskType', 'taskName',
  'taskDesc', 'execTime', 'upstreamSystem', 'upstreamInfo',
  'downstreamSystem', 'downstreamInfo', 'remark'
]

// Excel 表头（与 columnMap 顺序对应）
const excelHeaders = [
  '银行类型', '银行名称', '所属系统', '任务类型', '任务名',
  '任务描述', '执行时间', '上游依赖系统', '上游信息',
  '下游依赖系统', '下游信息', '备注'
]

const triggerImport = () => {
  fileInputRef.value.click()
}

// 导出 Excel（按当前查询条件，无条件则全量导出）
const handleExport = async () => {
  exporting.value = true
  try {
    const res = await getBatchTaskList({
      pageNum: 1,
      pageSize: 100000,
      ...searchForm
    })
    const records = res.data?.records || []
    if (records.length === 0) {
      ElMessage.warning('没有可导出的数据')
      return
    }

    // 组装数据：表头 + 数据行
    const aoa = [excelHeaders]
    records.forEach(row => {
      aoa.push(columnMap.map(field => row[field] ?? ''))
    })

    const ws = XLSX.utils.aoa_to_sheet(aoa)
    // 设置列宽
    ws['!cols'] = excelHeaders.map(() => ({ wch: 18 }))
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, '批量任务清单')
    XLSX.writeFile(wb, `批量任务清单_${new Date().toLocaleDateString().replace(/\//g, '-')}.xlsx`)
    ElMessage.success(`成功导出 ${records.length} 条数据`)
  } catch (err) {
    ElMessage.error('导出失败：' + (err.message || '未知错误'))
  } finally {
    exporting.value = false
  }
}

// 下载 Excel 导入模版
const handleDownloadTemplate = () => {
  const aoa = [excelHeaders]
  const ws = XLSX.utils.aoa_to_sheet(aoa)
  ws['!cols'] = excelHeaders.map(() => ({ wch: 18 }))
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, '批量任务清单')
  XLSX.writeFile(wb, '批量任务清单_导入模版.xlsx')
}

const handleFileChange = async (e) => {
  const file = e.target.files[0]
  if (!file) return

  importing.value = true
  try {
    const data = await file.arrayBuffer()
    const wb = XLSX.read(data, { type: 'array' })
    const ws = wb.Sheets[wb.SheetNames[0]]
    const rows = XLSX.utils.sheet_to_json(ws, { header: 1, defval: '' })

    if (rows.length < 2) {
      ElMessage.warning('Excel文件没有数据行')
      return
    }

    // 跳过表头行，从第2行开始
    const tasks = []
    for (let i = 1; i < rows.length; i++) {
      const row = rows[i]
      // 跳过全空行
      if (!row || row.every(v => v === '' || v == null)) continue

      const task = {}
      columnMap.forEach((field, idx) => {
        const val = row[idx]
        task[field] = val != null ? String(val).trim() : ''
      })
      tasks.push(task)
    }

    if (tasks.length === 0) {
      ElMessage.warning('未解析到有效数据')
      return
    }

    const res = await importBatchTasks(tasks)
    ElMessage.success(res.message || `成功导入 ${res.data?.count || 0} 条任务`)
    loadList()
  } catch (err) {
    const msg = err?.response?.data?.message || '导入失败'
    ElMessage.error(msg)
  } finally {
    importing.value = false
    // 清空 input，允许重复选择同一文件
    fileInputRef.value.value = ''
  }
}

onMounted(() => {
  dictStore.loadAllDicts()
  loadBankParams()
  loadList()
})
</script>

<style scoped>
.batch-tasks-container {
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 16px;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
