<template>
  <div class="prod-nodes">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>环境节点清单</span>
          <div>
            <el-button type="warning" :icon="Download" :loading="exporting" @click="handleExport">导出Excel</el-button>
            <el-button :icon="Document" @click="handleDownloadTemplate">下载模版</el-button>
            <el-button type="success" :icon="Upload" :loading="importing" @click="triggerImport">导入Excel</el-button>
            <el-button v-permission="'prodNode:add'" type="primary" :icon="Plus" @click="handleCreate">新建节点</el-button>
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
            <el-form-item label="环境类型">
              <el-select v-model="searchForm.envType" placeholder="全部" clearable filterable allow-create style="width: 100%">
                <el-option v-for="item in envTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="节点简称">
              <el-input v-model="searchForm.nodeName" placeholder="请输入节点简称" clearable @keyup.enter="handleSearch" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="节点IP">
              <el-input v-model="searchForm.nodeIP" placeholder="请输入节点IP" clearable @keyup.enter="handleSearch" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="用户名">
              <el-input v-model="searchForm.username" placeholder="请输入用户名" clearable @keyup.enter="handleSearch" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item>
              <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
              <el-button :icon="Refresh" @click="handleReset">重置</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" v-loading="loading" border stripe style="width: 100%">
        <el-table-column type="index" label="序号" width="55" align="center" />
        <el-table-column prop="bankType" label="银行类型" width="100" align="center" />
        <el-table-column prop="bankName" label="银行名称" min-width="120" />
        <el-table-column prop="envType" label="环境类型" min-width="120" />
        <el-table-column prop="nodeName" label="节点简称" min-width="100" />
        <el-table-column prop="nodeIp" label="节点IP" min-width="120" />
        <el-table-column prop="username" label="用户名" min-width="100" />
        <el-table-column prop="isMutualTrust" label="是否互信节点" min-width="110" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isMutualTrust === '是' || row.isMutualTrust === 'Y'" type="success">是</el-tag>
            <el-tag v-else type="info">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160" align="center">
          <template #default="{ row }">
            {{ formatDate(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'prodNode:view'" type="info" size="small" @click="handleView(row)">查看</el-button>
            <el-button v-permission="'prodNode:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'prodNode:delete'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.pageNum"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 20px; justify-content: flex-end;"
        @size-change="handleSizeChange"
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
import dayjs from 'dayjs'
import * as XLSX from 'xlsx'
import { getProdNodeList, deleteProdNode, importProdNodes } from '@/api/prodNode'
import { useDictStore } from '@/stores/dict'
import { useBankOptions } from '@/composables/useBankOptions'

const router = useRouter()
const dictStore = useDictStore()

// 银行类型 / 银行名称 下拉选项来自银行机构参数管理（不再使用数据字典关联）
const { bankTypeOptions, getBankNameOptions, loadBankParams } = useBankOptions()
const envTypeOptions = computed(() => dictStore.getDictByType('env_type').map(item => ({ label: item.dictLabel, value: item.dictValue })))

// 查询条件：银行名称联动（选了银行类型只展示该类型下的银行名称，未选则展示全部）
const searchBankNameOptions = computed(() => getBankNameOptions(searchForm.bankType))

// 查询条件银行类型变更时清空银行名称
const onSearchBankTypeChange = () => {
  searchForm.bankName = ''
}

const loading = ref(false)
const importing = ref(false)
const exporting = ref(false)
const tableData = ref([])
const fileInputRef = ref()

const searchForm = reactive({
  bankType: '',
  bankName: '',
  envType: '',
  nodeName: '',
  nodeIP: '',
  username: ''
})

const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

const formatDate = (d) => d ? dayjs(d).format('YYYY-MM-DD HH:mm:ss') : ''

const loadList = async () => {
  loading.value = true
  try {
    const res = await getProdNodeList({
      ...searchForm,
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize
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
    envType: '',
    nodeName: '',
    nodeIP: '',
    username: ''
  })
  pagination.pageNum = 1
  loadList()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.pageNum = 1
  loadList()
}

const handleView = (row) => {
  router.push(`/prod-nodes/${row.id}/detail`)
}

const handleCreate = () => {
  router.push('/prod-nodes/new')
}

const handleEdit = (row) => {
  router.push(`/prod-nodes/${row.id}/edit`)
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该节点？', '提示', {
      type: 'warning',
      confirmButtonText: '确认',
      cancelButtonText: '取消'
    })
    await deleteProdNode(row.id)
    ElMessage.success('删除成功')
    if (tableData.value.length === 1 && pagination.pageNum > 1) {
      pagination.pageNum--
    }
    loadList()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const triggerImport = () => {
  fileInputRef.value?.click()
}

// Excel 列名到字段名的映射（按列顺序）
const columnMap = ['bankType', 'bankName', 'envType', 'nodeName', 'nodeIp', 'username', 'isMutualTrust']

// Excel 表头（与 columnMap 顺序对应）
const excelHeaders = ['银行类型', '银行名称', '环境类型', '节点简称', '节点IP', '用户名', '是否互信节点']

// 导出 Excel（按当前查询条件，无条件则全量导出）
const handleExport = async () => {
  exporting.value = true
  try {
    const res = await getProdNodeList({
      ...searchForm,
      pageNum: 1,
      pageSize: 100000
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
    ws['!cols'] = excelHeaders.map(() => ({ wch: 18 }))
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, '生产环境节点清单')
    XLSX.writeFile(wb, `生产环境节点清单_${new Date().toLocaleDateString().replace(/\//g, '-')}.xlsx`)
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
  XLSX.utils.book_append_sheet(wb, ws, '生产环境节点清单')
  XLSX.writeFile(wb, '生产环境节点清单_导入模版.xlsx')
}

const handleFileChange = async (e) => {
  const file = e.target.files[0]
  if (!file) return

  importing.value = true
  try {
    const data = await file.arrayBuffer()
    const wb = XLSX.read(data)
    const ws = wb.Sheets[wb.SheetNames[0]]
    const rows = XLSX.utils.sheet_to_json(ws, { header: 1, defval: '' })

    if (rows.length < 2) {
      ElMessage.warning('Excel文件没有数据行')
      return
    }

    const nodes = []
    for (let i = 1; i < rows.length; i++) {
      const row = rows[i]
      if (!row[1] && !row[2] && !row[3] && !row[4]) continue
      nodes.push({
        bankType: String(row[0] || '').trim(),
        bankName: String(row[1] || '').trim(),
        envType: String(row[2] || '').trim(),
        nodeName: String(row[3] || '').trim(),
        nodeIp: String(row[4] || '').trim(),
        username: String(row[5] || '').trim(),
        isMutualTrust: String(row[6] || '').trim()
      })
    }

    if (nodes.length === 0) {
      ElMessage.warning('没有有效的数据行')
      return
    }

    const res = await importProdNodes(nodes)
    ElMessage.success(`成功导入 ${res.data.count} 条，跳过 ${res.data.skipped} 条`)
    pagination.pageNum = 1
    loadList()
  } catch (err) {
    ElMessage.error('导入失败：' + (err.message || '未知错误'))
  } finally {
    importing.value = false
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
.prod-nodes {
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}
</style>
