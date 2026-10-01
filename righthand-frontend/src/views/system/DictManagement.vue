<template>
  <div class="dict-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>字典管理</span>
          <el-button v-permission="'system:dict:add'" type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增字典
          </el-button>
        </div>
      </template>

      <!-- 字典类型筛选 -->
      <el-form :inline="true" class="mb-4">
        <el-form-item label="字典类型">
          <el-select v-model="searchForm.dictType" placeholder="全部类型" clearable @change="loadData" style="width: 180px">
            <el-option
              v-for="item in dictTypeOptions"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" stripe style="width: 100%">
        <el-table-column prop="dictCode" label="字典编码" width="200" />
        <el-table-column prop="dictName" label="字典名称" width="250" />
        <el-table-column prop="dictType" label="字典类型" width="180" />
        <el-table-column prop="dictValue" label="字典值" />
        <el-table-column prop="dictLabel" label="字典标签" width="100" />
        <el-table-column prop="sortOrder" label="排序" width="80" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'system:dict:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button
              v-permission="'system:dict:delete'"
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页组件 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
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
        label-width="100px"
      >
        <el-form-item label="字典编码" prop="dictCode">
          <el-input v-model="formData.dictCode" placeholder="如：MEMO_TYPE_WORK" />
        </el-form-item>
        <el-form-item label="字典名称" prop="dictName">
          <el-input v-model="formData.dictName" placeholder="如：备忘类型 - 工作" />
        </el-form-item>
        <el-form-item label="字典类型" prop="dictType">
          <el-input v-model="formData.dictType" placeholder="如：memo_type" />
        </el-form-item>
        <el-form-item label="字典值" prop="dictValue">
          <el-input v-model="formData.dictValue" placeholder="如：1" />
        </el-form-item>
        <el-form-item label="字典标签" prop="dictLabel">
          <el-input v-model="formData.dictLabel" placeholder="如：工作" />
        </el-form-item>
        <el-form-item label="排序" prop="sortOrder">
          <el-input-number v-model="formData.sortOrder" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
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
import { Plus } from '@element-plus/icons-vue'
import { getDicts, createDict, updateDict, deleteDict } from '@/api/system'
import { isCancelError } from '@/utils/request'

const searchForm = reactive({
  dictType: ''
})

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增字典')
const submitLoading = ref(false)
const formRef = ref(null)

// 分页参数
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  id: null,
  dictCode: '',
  dictName: '',
  dictType: '',
  dictValue: '',
  dictLabel: '',
  sortOrder: 0,
  status: 1,
  remark: ''
})

const rules = {
  dictCode: [{ required: true, message: '请输入字典编码', trigger: 'blur' }],
  dictName: [{ required: true, message: '请输入字典名称', trigger: 'blur' }],
  dictType: [{ required: true, message: '请输入字典类型', trigger: 'blur' }],
  dictValue: [{ required: true, message: '请输入字典值', trigger: 'blur' }],
  dictLabel: [{ required: true, message: '请输入字典标签', trigger: 'blur' }],
}

// 获取所有字典类型（从字典数据中提取）
const dictTypeOptions = computed(() => {
  const types = new Set()
  tableData.value.forEach(item => {
    if (item.dictType) {
      types.add(item.dictType)
    }
  })
  return Array.from(types)
})

const loadData = async () => {
  try {
    const res = await getDicts({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      dictType: searchForm.dictType || undefined
    })
    tableData.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    if (!isCancelError(error)) {
      console.error('加载数据失败:', error)
      ElMessage.error('加载数据失败')
    }
  }
}

const resetSearch = () => {
  searchForm.dictType = ''
  pagination.pageNum = 1
  loadData()
}

// 每页条数改变
const handleSizeChange = (val) => {
  pagination.pageSize = val
  pagination.pageNum = 1
  loadData()
}

// 当前页改变
const handleCurrentChange = (val) => {
  pagination.pageNum = val
  loadData()
}

const handleAdd = () => {
  dialogTitle.value = '新增字典'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑字典'
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该字典项吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteDict(row.id)
      ElMessage.success('删除成功')
      loadData()
    } catch (error) {
      if (!isCancelError(error)) {
        console.error('删除失败:', error)
        ElMessage.error('删除失败')
      }
    }
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        // 只提交必要字段，排除时间字段等不必要的字段
        const submitData = {
          dictCode: formData.dictCode,
          dictName: formData.dictName,
          dictType: formData.dictType,
          dictValue: formData.dictValue,
          dictLabel: formData.dictLabel,
          sortOrder: formData.sortOrder,
          status: formData.status,
          remark: formData.remark
        }
        
        if (formData.id) {
          await updateDict(formData.id, submitData)
          ElMessage.success('更新成功')
        } else {
          await createDict(submitData)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadData()
      } catch (error) {
        if (!isCancelError(error)) {
          console.error('提交失败:', error)
          ElMessage.error(error.response?.data?.message || '操作失败')
        }
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const resetForm = () => {
  formData.id = null
  formData.dictCode = ''
  formData.dictName = ''
  formData.dictType = ''
  formData.dictValue = ''
  formData.dictLabel = ''
  formData.sortOrder = 0
  formData.status = 1
  formData.remark = ''
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dict-management {
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

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>
