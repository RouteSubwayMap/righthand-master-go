<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>银行机构信息管理</span>
          <div>
            <el-button type="warning" :icon="Download" :loading="exporting" @click="handleExport">导出Excel</el-button>
            <el-button :icon="Document" @click="handleDownloadTemplate">下载模版</el-button>
            <el-button type="success" :icon="Upload" :loading="importing" @click="triggerImport">导入Excel</el-button>
            <el-button v-permission="'bankParam:add'" type="primary" :icon="Plus" @click="handleCreate">新建参数</el-button>
          </div>
        </div>
      </template>

      <!-- 查询条件 -->
      <el-form :model="searchForm" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="银行类型">
              <el-select v-model="searchForm.bankType" placeholder="全部" clearable style="width: 100%" @change="onSearchBankTypeChange">
                <el-option v-for="item in bankTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="银行名称">
              <el-select v-model="searchForm.bankName" placeholder="全部" clearable filterable style="width: 100%">
                <el-option v-for="item in searchBankNameOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="银行机构号">
              <el-input v-model="searchForm.bankCode" placeholder="请输入银行机构号" clearable @keyup.enter="handleSearch" />
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
        <el-table-column prop="bankType" label="银行类型" width="100" align="center" />
        <el-table-column prop="bankName" label="银行名称" min-width="160" show-overflow-tooltip />
        <el-table-column prop="bankCode" label="银行机构号" min-width="150" align="center" />
        <el-table-column prop="createTime" label="创建时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.createTime) }}</template>
        </el-table-column>
        <el-table-column prop="updateTime" label="修改时间" width="170" align="center">
          <template #default="{ row }">{{ formatTime(row.updateTime) }}</template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'bankParam:view'" type="info" size="small" @click="handleView(row)">查看</el-button>
            <el-button v-permission="'bankParam:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'bankParam:delete'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </el-card>

    <!-- 隐藏的文件输入 -->
    <input ref="fileInputRef" type="file" accept=".xlsx,.xls" style="display: none" @change="handleFileChange" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, Upload, Download, Document } from '@element-plus/icons-vue'
import {
  getBankParamList, deleteBankParam, importBankParams
} from '@/api/bankParam'
import { useBankOptions } from '@/composables/useBankOptions'
import * as XLSX from 'xlsx'
import dayjs from 'dayjs'

const router = useRouter()

// 银行类型 / 银行名称 下拉选项来自银行机构参数管理（不再使用数据字典关联）
const { bankTypeOptions, getBankNameOptions } = useBankOptions()

const loading = ref(false)
const importing = ref(false)
const exporting = ref(false)
const tableData = ref([])
const fileInputRef = ref()

const searchForm = reactive({
  bankType: '',
  bankName: '',
  bankCode: ''
})

const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 查询条件：银行名称联动（选了银行类型只展示该类型下的银行名称，未选则展示全部）
const searchBankNameOptions = computed(() => getBankNameOptions(searchForm.bankType))

// 查询条件银行类型变更时清空银行名称
const onSearchBankTypeChange = () => {
  searchForm.bankName = ''
}

// 时间格式化
const formatTime = (time) => {
  if (!time) return ''
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// Excel 列名到字段名的映射（按列顺序）
const columnMap = ['bankType', 'bankName', 'bankCode', 'reserve1', 'reserve2', 'reserve3', 'reserve4', 'reserve5', 'remark']
const excelHeaders = ['银行类型', '银行名称', '银行机构号', '备用字段1', '备用字段2', '备用字段3', '备用字段4', '备用字段5', '备注']

const loadList = async () => {
  loading.value = true
  try {
    const res = await getBankParamList({
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
  Object.assign(searchForm, { bankType: '', bankName: '', bankCode: '' })
  pagination.pageNum = 1
  loadList()
}

const handleCreate = () => {
  router.push('/bank-params/new')
}

const handleView = (row) => {
  router.push(`/bank-params/${row.id}/detail`)
}

const handleEdit = (row) => {
  router.push(`/bank-params/${row.id}/edit`)
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确认删除「${row.bankName}」的银行参数吗？`, '提示', {
    type: 'warning'
  }).then(async () => {
    try {
      await deleteBankParam(row.id)
      ElMessage.success('删除成功')
      loadList()
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

// 导入 Excel
const triggerImport = () => {
  fileInputRef.value.click()
}

const handleFileChange = async (e) => {
  const file = e.target.files[0]
  if (!file) return
  importing.value = true
  try {
    const data = await file.arrayBuffer()
    const wb = XLSX.read(data)
    const ws = wb.Sheets[wb.SheetNames[0]]
    const rows = XLSX.utils.sheet_to_json(ws, { header: 1 })
    if (rows.length < 2) {
      ElMessage.warning('文件没有数据行')
      return
    }
    // 跳过表头，从第二行开始
    const params = []
    for (let i = 1; i < rows.length; i++) {
      const row = rows[i]
      if (!row || row.length === 0) continue
      const param = {}
      columnMap.forEach((field, idx) => {
        param[field] = row[idx] != null ? String(row[idx]).trim() : ''
      })
      params.push(param)
    }
    const res = await importBankParams(params)
    ElMessage.success(res.message || `成功导入 ${res.data?.count || 0} 条数据`)
    loadList()
  } catch (err) {
    ElMessage.error('导入失败：' + (err.message || '未知错误'))
  } finally {
    importing.value = false
    e.target.value = ''
  }
}

// 导出 Excel
const handleExport = async () => {
  exporting.value = true
  try {
    const res = await getBankParamList({
      pageNum: 1,
      pageSize: 100000,
      ...searchForm
    })
    const records = res.data?.records || []
    if (records.length === 0) {
      ElMessage.warning('没有可导出的数据')
      return
    }
    const aoa = [excelHeaders]
    records.forEach(row => {
      aoa.push(columnMap.map(field => row[field] ?? ''))
    })
    const ws = XLSX.utils.aoa_to_sheet(aoa)
    ws['!cols'] = excelHeaders.map(() => ({ wch: 18 }))
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, '银行机构参数')
    XLSX.writeFile(wb, `银行机构参数_${new Date().toLocaleDateString().replace(/\//g, '-')}.xlsx`)
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
  XLSX.utils.book_append_sheet(wb, ws, '银行机构参数')
  XLSX.writeFile(wb, '银行机构参数_导入模版.xlsx')
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
.page-container {
  padding: 16px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
