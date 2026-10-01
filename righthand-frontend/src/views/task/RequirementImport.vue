<template>
  <div class="requirement-import">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>批量导入需求</span>
          <el-button type="success" @click="handleDownloadTemplate">
            <el-icon><Download /></el-icon>
            下载模板
          </el-button>
        </div>
      </template>

      <el-alert
        title="导入说明"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px;"
      >
        <div class="import-tips">
          <p>1. 请先下载 Excel 模板，按照模板格式填写需求数据</p>
          <p>2. 必填字段：需求标题、需求类型、优先级</p>
          <p>3. 优先级：HIGH（高）、MEDIUM（中）、LOW（低）</p>
          <p>4. 风险等级：HIGH（高）、MEDIUM（中）、LOW（低）</p>
          <p>5. 需求类型：BANK（行内需求）、INTERNAL（内部优化）</p>
          <p>6. 状态：PENDING_CLARIFY（待澄清）、PENDING_SCHEDULING（待排期）、DEVELOPING（开发中）、TESTING（测试中）、PENDING_RELEASE（待投产）、COMPLETED（已完成）</p>
          <p>7. 所属小组：LOAN_MIDDLE（贷中）、LOAN_PRE（贷前）、RISK（风险）</p>
          <p>8. 日期格式：yyyy-MM-dd HH:mm:ss（例：2024-01-15 10:30:00）</p>
        </div>
      </el-alert>

      <el-upload
        ref="uploadRef"
        drag
        :auto-upload="false"
        :on-change="handleFileChange"
        :limit="1"
        accept=".xlsx,.xls"
        style="width: 100%;"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            只能上传 xlsx/xls 格式的 Excel 文件
          </div>
        </template>
      </el-upload>

      <div v-if="file" class="upload-info" style="margin-top: 20px;">
        <el-alert
          :title="`已选择文件：${file.name}`"
          type="success"
          :closable="false"
          show-icon
        />
      </div>

      <div class="import-actions" style="margin-top: 20px; text-align: center;">
        <el-button 
          type="primary" 
          size="large" 
          @click="handleImport"
          :loading="importing"
          :disabled="!file"
        >
          开始导入
        </el-button>
        <el-button size="large" @click="$router.push('/task/requirements')">
          返回列表
        </el-button>
      </div>

      <!-- 导入结果 -->
      <el-dialog
        v-model="resultVisible"
        :title="importSuccess ? '导入成功' : '导入失败'"
        :type="importSuccess ? 'success' : 'error'"
        width="500px"
      >
        <div class="import-result">
          <el-result
            :icon="importSuccess ? 'success' : 'error'"
            :title="importSuccess ? '导入完成' : '导入失败'"
            :sub-title="importMessage"
          >
            <template #extra>
              <el-button type="primary" @click="resultVisible = false">确定</el-button>
              <el-button v-if="importSuccess" @click="$router.push('/task/requirements')">
                查看列表
              </el-button>
            </template>
          </el-result>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, UploadFilled } from '@element-plus/icons-vue'
import { batchImport } from '@/api/requirement'
import * as XLSX from 'xlsx'
import { useDictStore } from '@/stores/dict'

const dictStore = useDictStore()

const uploadRef = ref(null)
const file = ref(null)
const importing = ref(false)
const resultVisible = ref(false)
const importSuccess = ref(false)
const importMessage = ref('')

// Excel 列定义（与模版/导入解析共用）
const excelColumns = [
  { key: 'title', header: '需求名称' },
  { key: 'bankName', header: '所属银行' },
  { key: 'requirementNo', header: '需求编号' },
  { key: 'priority', header: '优先级' },
  { key: 'riskLevel', header: '风险等级' },
  { key: 'requirementType', header: '需求类型' },
  { key: 'mainSystem', header: '主R系统' },
  { key: 'developer', header: '开发主R|参与者' },
  { key: 'involvedSystems', header: '涉及系统' },
  { key: 'followerName', header: '需求跟进人' },
  { key: 'submitterName', header: '提交人' },
  { key: 'estimatedWorkload', header: '预计工作量(人天)' },
  { key: 'groupType', header: '所属小组' },
  { key: 'description', header: '需求描述' },
  { key: 'remark', header: '备注' }
]

// 处理文件选择
const handleFileChange = (uploadFile) => {
  file.value = uploadFile.raw
}

// 下载模板（带数据字典下拉选项）
const handleDownloadTemplate = async () => {
  const ExcelJS = await import('exceljs')
  const workbook = new ExcelJS.Workbook()
  const ws = workbook.addWorksheet('需求导入模板')

  ws.addRow(excelColumns.map(c => c.header))
  ws.addRow(['示例需求', '上海银行', 'REQ-2024-001', 'HIGH', 'MEDIUM', 'INTERNAL', 'poseidon', '张三|R&D团队', '系统A;系统B', '李四', '王五', 5.0, 'LOAN_MIDDLE', '这是一个示例需求的描述', '这是备注信息'])
  ws.columns = excelColumns.map(() => ({ width: 18 }))

  const dictCols = [
    { colIndex: 1, dictType: 'bank_name' },
    { colIndex: 3, dictType: 'priority' },
    { colIndex: 4, dictType: 'risk_level' },
    { colIndex: 5, dictType: 'requirement_type' },
    { colIndex: 12, dictType: 'requirement_group' }
  ]

  const dictSheet = workbook.addWorksheet('字典数据', { state: 'hidden' })
  const dictRefs = {}
  dictCols.forEach((dc, i) => {
    const items = dictStore.getDictByType(dc.dictType)
    if (items && items.length > 0) {
      const colLetter = String.fromCharCode(65 + i)
      items.forEach((item, idx) => {
        dictSheet.getCell(`${colLetter}${idx + 1}`).value = item.dictValue
      })
      dictRefs[dc.colIndex] = `字典数据!$${colLetter}$1:$${colLetter}$${items.length}`
    }
  })

  dictCols.forEach(dc => {
    const ref = dictRefs[dc.colIndex]
    if (!ref) return
    const colLetter = String.fromCharCode(65 + dc.colIndex)
    for (let row = 2; row <= 1000; row++) {
      ws.getCell(`${colLetter}${row}`).dataValidation = {
        type: 'list',
        allowBlank: true,
        formulae: [ref]
      }
    }
  })

  const buffer = await workbook.xlsx.writeBuffer()
  const blob = new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = '需求导入模板.xlsx'
  link.click()
  URL.revokeObjectURL(url)
  ElMessage.success('模板下载成功')
}

onMounted(() => {
  dictStore.loadAllDicts()
})

// 执行导入
const handleImport = async () => {
  if (!file.value) {
    ElMessage.warning('请先选择文件')
    return
  }

  importing.value = true
  try {
    // 解析 Excel 文件
    const data = await file.value.arrayBuffer()
    const wb = XLSX.read(data)
    const ws = wb.Sheets[wb.SheetNames[0]]
    const rows = XLSX.utils.sheet_to_json(ws, { header: 1, defval: '' })

    if (rows.length < 2) {
      ElMessage.warning('Excel文件没有数据行')
      return
    }

    // 跳过表头，从第二行开始
    const requirements = []
    for (let i = 1; i < rows.length; i++) {
      const row = rows[i]
      if (!row[0]) continue
      requirements.push({
        title: String(row[0] || '').trim(),
        bankName: String(row[1] || '').trim(),
        requirementNo: String(row[2] || '').trim(),
        priority: String(row[3] || '').trim(),
        riskLevel: String(row[4] || '').trim(),
        requirementType: String(row[5] || '').trim(),
        mainSystem: String(row[6] || '').trim(),
        developer: String(row[7] || '').trim(),
        involvedSystems: String(row[8] || '').trim(),
        followerName: String(row[9] || '').trim(),
        submitterName: String(row[10] || '').trim(),
        estimatedWorkload: parseFloat(row[11]) || 0,
        groupType: String(row[12] || '').trim(),
        description: String(row[13] || '').trim(),
        remark: String(row[14] || '').trim()
      })
    }

    if (requirements.length === 0) {
      ElMessage.warning('没有有效的数据行')
      return
    }

    const res = await batchImport(requirements)
    if (res.code === 200) {
      importSuccess.value = true
      const result = res.data || {}
      importMessage.value = `成功: ${result.successCount || 0}条, 失败: ${result.failCount || 0}条`
      ElMessage.success(importMessage.value)

      if (uploadRef.value) {
        uploadRef.value.clearFiles()
      }
      file.value = null
    } else {
      importSuccess.value = false
      importMessage.value = res.message || '导入失败'
      ElMessage.error(res.message || '导入失败')
    }
  } catch (error) {
    console.error('导入失败:', error)
    importSuccess.value = false
    importMessage.value = error.message || '导入失败，请检查文件格式'
    ElMessage.error('导入失败，请检查文件格式')
  } finally {
    importing.value = false
    resultVisible.value = true
  }
}
</script>

<style scoped>
.requirement-import {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.import-tips {
  line-height: 1.8;
}

.import-tips p {
  margin: 5px 0;
  font-size: 14px;
}

.upload-info {
  margin-top: 15px;
}

.import-actions {
  margin-top: 20px;
  text-align: center;
}

.import-actions .el-button {
  min-width: 120px;
  margin: 0 10px;
}

.import-result {
  padding: 20px;
  text-align: center;
}
</style>
