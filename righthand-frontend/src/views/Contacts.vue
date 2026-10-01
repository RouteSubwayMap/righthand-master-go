<template>
  <div class="contacts">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>通讯录</span>
          <div class="header-actions">
            <el-radio-group v-model="viewMode" size="small">
              <el-radio-button value="list">列表</el-radio-button>
              <el-radio-button value="board">看板</el-radio-button>
            </el-radio-group>
            <el-button v-permission="'contacts:add'" type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              新增联系人
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="mb-4">
        <el-form-item label="姓名">
          <el-input v-model="searchForm.name" placeholder="请输入姓名" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="职位">
          <el-input v-model="searchForm.position" placeholder="请输入职位" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="searchForm.contactGroup" placeholder="请选择分组" clearable style="width: 180px">
            <el-option
              v-for="item in groupOptions"
              :key="item.dictValue"
              :label="item.dictLabel"
              :value="item.dictValue"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 列表视图 -->
      <el-table 
        v-if="viewMode === 'list'"
        :data="tableData" 
        stripe 
        style="width: 100%"
      >
        <el-table-column prop="name" label="姓名" />
        <el-table-column prop="phone" label="电话" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="position" label="职位" />
        <el-table-column prop="contactGroup" label="分组">
          <template #default="{ row }">
            <el-tag>{{ dictStore.getDictLabel('contact_group', row.contactGroup) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="company" label="公司" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'contacts:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'contacts:delete'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 看板视图 -->
      <div v-else class="board-view">
        <el-scrollbar>
          <div class="board-columns">
            <div 
              v-for="group in groupBoardColumns" 
              :key="group.value" 
              class="board-column"
            >
              <div class="column-header" :style="getColumnHeaderStyle(group.color)">
                <span>{{ group.label }}</span>
                <el-tag size="small" type="info">{{ group.contactCount }}</el-tag>
              </div>
              <div class="column-content">
                <el-card 
                  v-for="contact in getContactsByGroup(group.value)" 
                  :key="contact.id"
                  class="contact-card"
                  :class="getCardClass(group.value)"
                  shadow="hover"
                  @click="handleEdit(contact)"
                >
                  <div class="contact-card-name">{{ contact.name }}</div>
                  <div class="contact-card-info">
                    <div v-if="contact.phone" class="info-item">
                      <el-icon><Phone /></el-icon>
                      <span>{{ contact.phone }}</span>
                    </div>
                    <div v-if="contact.email" class="info-item">
                      <el-icon><Message /></el-icon>
                      <span>{{ contact.email }}</span>
                    </div>
                    <div v-if="contact.company" class="info-item">
                      <el-icon><OfficeBuilding /></el-icon>
                      <span>{{ contact.company }}</span>
                    </div>
                    <div v-if="contact.position" class="info-item position-item">
                      <el-icon><User /></el-icon>
                      <span class="position-text">{{ formatPosition(contact.position) }}</span>
                    </div>
                  </div>
                </el-card>
              </div>
            </div>
            <!-- 如果没有联系人数据 -->
            <div v-if="groupBoardColumns.length === 0" class="empty-board">
              <el-icon size="64" color="#ccc"><Folder /></el-icon>
              <p>暂无联系人数据</p>
              <span>请联系管理员配置联系人群组或添加联系人</span>
            </div>
          </div>
        </el-scrollbar>
      </div>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="80px"
      >
        <el-form-item label="姓名" prop="name">
          <el-input v-model="formData.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="电话" prop="phone">
          <el-input v-model="formData.phone" placeholder="请输入电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="分组" prop="contactGroup">
          <el-select v-model="formData.contactGroup" placeholder="请选择分组" style="width: 100%">
            <el-option
              v-for="item in groupOptions"
              :key="item.dictValue"
              :label="item.dictLabel"
              :value="item.dictValue"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="公司" prop="company">
          <el-input v-model="formData.company" placeholder="请输入公司" />
        </el-form-item>
        <el-form-item label="职位" prop="position">
          <el-input v-model="formData.position" placeholder="请输入职位" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Phone, Message, OfficeBuilding, User, Folder } from '@element-plus/icons-vue'
import { getContactList, createContact, updateContact, deleteContact } from '@/api/contact'
import { useDictStore } from '@/stores/dict'

const dictStore = useDictStore()

// 视图模式：list | board，默认看板方式
const viewMode = ref('board')

// 根据分组生成看板列（只显示有数据的分组）
const groupBoardColumns = computed(() => {
  const groups = dictStore.getDictByType('contact_group')
  const colorMap = {
    'ACCOUNT_MANAGER': '#FF6B6B',
    'PRODUCT': '#4ECDC4',
    'R&D': '#45B7D1'
  }
  
  // 过滤出有数据的分组
  return groups
    .map(group => ({
      value: group.dictValue,
      label: group.dictLabel,
      color: colorMap[group.dictValue] || '#96CEB4',
      contactCount: tableData.value.filter(contact => contact.contactGroup === group.dictValue).length
    }))
    .filter(group => group.contactCount > 0) // 只保留有数据的分组
})

// 根据分组筛选联系人
const getContactsByGroup = (group) => {
  return tableData.value.filter(contact => contact.contactGroup === group)
}

// 格式化职位信息
const formatPosition = (position) => {
  if (!position) return ''
  return position.split(',').map(item => item.trim()).join('\n')
}

// 生成安全的卡片类名
const getCardClass = (groupValue) => {
  const safeValue = groupValue.replace(/[^a-zA-Z0-9]/g, '_')
  return `contact-card-${safeValue}`
}

// 生成列头部样式
const getColumnHeaderStyle = (color) => {
  const darkenColor = (hex, percent) => {
    const num = parseInt(hex.replace('#', ''), 16)
    const amt = Math.round(2.55 * percent)
    const R = Math.max((num >> 16) - amt, 0)
    const G = Math.max((num >> 8 & 0x00FF) - amt, 0)
    const B = Math.max((num & 0x0000FF) - amt, 0)
    return `#${(0x1000000 + R * 0x10000 + G * 0x100 + B).toString(16).slice(1)}`
  }
  
  const darkColor = darkenColor(color, 20)
  return {
    '--header-color': color,
    '--header-color-dark': darkColor
  }
}

const searchForm = reactive({
  name: '',
  position: '',
  contactGroup: ''
})

const tableData = ref([])
const groupOptions = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增联系人')
const submitLoading = ref(false)
const formRef = ref(null)

// 分页
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  id: null,
  name: '',
  phone: '',
  email: '',
  contactGroup: '',
  company: '',
  position: '',
  remark: ''
})

const rules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }]
}

const loadData = async () => {
  try {
    const res = await getContactList({
      ...searchForm,
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize
    })
    console.log('联系人 API 响应:', res)
    if (res.data) {
      tableData.value = res.data.records || res.data.list || []
      pagination.total = res.data.total || 0
      console.log('联系人数据:', tableData.value, '总数:', pagination.total)
    }
  } catch (error) {
    console.error('加载联系人失败:', error)
  }
}

const resetSearch = () => {
  searchForm.name = ''
  searchForm.position = ''
  searchForm.contactGroup = ''
  loadData()
}

const handleAdd = () => {
  dialogTitle.value = '新增联系人'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑联系人'
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该联系人吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteContact(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      // 静默失败
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        const submitData = {
          name: formData.name,
          phone: formData.phone,
          email: formData.email,
          contactGroup: formData.contactGroup,
          company: formData.company,
          position: formData.position,
          remark: formData.remark
        }
        
        if (formData.id) {
          submitData.id = formData.id
          await updateContact(formData.id, submitData)
          ElMessage.success('更新成功')
        } else {
          await createContact(submitData)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadData()
      } catch (error) {
        const errorMsg = error.response?.data?.message || error.message || '提交失败'
        ElMessage.error(errorMsg)
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const resetForm = () => {
  formData.id = null
  formData.name = ''
  formData.phone = ''
  formData.email = ''
  formData.contactGroup = ''
  formData.company = ''
  formData.position = ''
  formData.remark = ''
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

const loadDictData = async () => {
  // 先检查字典数据是否已加载
  const existingDicts = dictStore.getDictByType('contact_group')
  if (!existingDicts || existingDicts.length === 0) {
    // 如果字典数据未加载，先加载所有字典
    await dictStore.loadAllDicts()
  }
  groupOptions.value = dictStore.getDictByType('contact_group')
}

onMounted(async () => {
  await loadDictData()
  loadData()
})

const handleSizeChange = (val) => {
  pagination.pageSize = val
  pagination.pageNum = 1
  loadData()
}

const handleCurrentChange = (val) => {
  pagination.pageNum = val
  loadData()
}
</script>

<style scoped>
.contacts {
  padding: 0;
}

.mb-4 {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.board-view {
  min-height: 500px;
  padding: 16px;
}

.board-columns {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.board-column {
  width: 100%;
  background-color: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
}

.column-header {
  padding: 16px 20px;
  color: white;
  font-weight: 600;
  font-size: 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  background-image: linear-gradient(135deg, var(--header-color) 0%, var(--header-color-dark) 100%);
  transition: all 0.3s ease;
}

.column-header:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.column-content {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  max-height: 600px;
  overflow-y: auto;
  padding-right: 8px;
}

.contact-card {
  cursor: pointer;
  transition: all 0.3s;
}

.contact-card :deep(.el-card__body) {
  border: 2px solid #e4e7ed;
  border-radius: 4px;
}

.contact-card-ACCOUNT_MANAGER :deep(.el-card__body) {
  border-color: #FF6B6B;
}

.contact-card-ACCOUNT_MANAGER:hover :deep(.el-card__body) {
  border-color: #ff5252;
  box-shadow: 0 4px 12px rgba(255, 107, 107, 0.3);
}

.contact-card-PRODUCT :deep(.el-card__body) {
  border-color: #4ECDC4;
}

.contact-card-PRODUCT:hover :deep(.el-card__body) {
  border-color: #3dbdb5;
  box-shadow: 0 4px 12px rgba(78, 205, 196, 0.3);
}

.contact-card-R_D :deep(.el-card__body) {
  border-color: #45B7D1;
}

.contact-card-R_D:hover :deep(.el-card__body) {
  border-color: #3aa7c1;
  box-shadow: 0 4px 12px rgba(69, 183, 209, 0.3);
}

.contact-card:hover {
  transform: translateY(-2px);
}

.contact-card-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
  text-align: center;
  padding-bottom: 10px;
  border-bottom: 2px solid #f0f0f0;
  transition: color 0.3s;
}

.contact-card-ACCOUNT_MANAGER .contact-card-name {
  color: #FF6B6B;
}

.contact-card-PRODUCT .contact-card-name {
  color: #4ECDC4;
}

.contact-card-R_D .contact-card-name {
  color: #45B7D1;
}

.contact-card-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
}

.info-item .el-icon {
  color: #909399;
  font-size: 14px;
  margin-top: 2px;
  flex-shrink: 0;
}

.position-item {
  align-items: flex-start;
}

.position-text {
  white-space: pre-line;
  line-height: 1.6;
  flex: 1;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 空看板状态样式 */
.empty-board {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: #909399;
}

.empty-board p {
  margin: 15px 0 8px;
  font-size: 16px;
  font-weight: 500;
}

.empty-board span {
  font-size: 14px;
  color: #c0c4cc;
}
</style>