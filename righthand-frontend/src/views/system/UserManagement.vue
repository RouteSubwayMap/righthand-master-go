<template>
  <div class="user-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button v-permission="'system:user:add'" type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增用户
          </el-button>
        </div>
      </template>

      <!-- 筛选条件 -->
      <div class="filter-bar">
        <el-form :inline="true" :model="queryParams">
          <el-form-item label="用户名">
            <el-input v-model="queryParams.username" placeholder="请输入用户名" clearable />
          </el-form-item>
          <el-form-item label="姓名">
            <el-input v-model="queryParams.realName" placeholder="请输入姓名" clearable />
          </el-form-item>
          <el-form-item label="角色">
            <el-select v-model="queryParams.roleId" placeholder="全部" clearable>
              <el-option
                v-for="role in allRoles"
                :key="role.id"
                :label="role.roleName"
                :value="role.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="queryParams.status" placeholder="全部" clearable>
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadData">查询</el-button>
            <el-button @click="resetQuery">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 表格 -->
      <el-table :data="tableData" stripe style="width: 100%">
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="realName" label="姓名" width="100" />
        <el-table-column prop="email" label="邮箱" width="180" />
        <el-table-column prop="phone" label="手机号" width="120" />
        <el-table-column label="角色" width="200">
          <template #default="{ row }">
            <el-tag 
              v-for="roleId in row.roleIds" 
              :key="roleId"
              size="small"
              style="margin-right: 5px; margin-bottom: 5px;"
            >
              {{ getRoleName(roleId) }}
            </el-tag>
            <span v-if="!row.roleIds || row.roleIds.length === 0" style="color: #999;">未分配角色</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="380" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'system:user:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'system:user:assign'" type="warning" size="small" @click="handleAssignRole(row)">分配角色</el-button>
            <el-button v-permission="'system:user:resetPwd'" type="info" size="small" @click="handleResetPassword(row)">重置密码</el-button>
            <el-button
              v-permission="'system:user:delete'"
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="pagination.total"
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
        <el-form-item label="用户名" prop="username">
          <el-input 
            id="username" 
            v-model="formData.username" 
            placeholder="请输入用户名" 
            :disabled="!!formData.id" 
          />
        </el-form-item>
        <el-form-item label="姓名" prop="realName">
          <el-input 
            id="realName" 
            v-model="formData.realName" 
            placeholder="请输入姓名" 
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input 
            id="email" 
            v-model="formData.email" 
            placeholder="请输入邮箱" 
          />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input 
            id="phone" 
            v-model="formData.phone" 
            placeholder="请输入手机号" 
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 分配角色对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      title="分配角色"
      width="500px"
    >
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox 
          v-for="role in allRoles" 
          :key="role.id" 
          :value="role.id"
          style="display: block; margin-bottom: 10px;"
        >
          {{ role.roleName }} ({{ role.roleCode }})
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSaveRole">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getUserPage, createUser, updateUser, deleteUser, getUserRoles, saveUserRoles, getAllRoles } from '@/api/system'
import { resetPassword } from '@/api/user'

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const submitLoading = ref(false)
const formRef = ref(null)

const roleDialogVisible = ref(false)
const currentUserId = ref(null)
const selectedRoleIds = ref([])
const allRoles = ref([])
const roleMap = ref(new Map()) // 角色 ID 到角色名称的映射

// 查询参数
const queryParams = reactive({
  username: '',
  realName: '',
  roleId: null,
  status: null
})

// 分页
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  id: null,
  username: '',
  realName: '',
  email: '',
  phone: '',
  status: 1
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  realName: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  email: [
    { type: 'email', message: '邮箱格式不正确', trigger: ['blur', 'change'] }
  ],
}

const loadData = async () => {
  try {
    // 加载所有角色，构建角色映射
    const rolesRes = await getAllRoles()
    const roles = rolesRes.data || []
    allRoles.value = roles
    roleMap.value = new Map(roles.map(role => [role.id, role.roleName]))
    
    // 分页查询用户
    const params = {
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      ...queryParams
    }
    const res = await getUserPage(params)
    if (res.code === 200) {
      const pageData = res.data
      tableData.value = pageData.list || pageData.records || []
      pagination.total = pageData.total || 0
      
      // 为每个用户加载角色 ID 列表
      for (const user of tableData.value) {
        try {
          const userRolesRes = await getUserRoles(user.id)
          const userRoles = userRolesRes.data || []
          // 从角色对象数组中提取角色 ID
          user.roleIds = userRoles.map(role => role.id)
        } catch (error) {
          console.error(`加载用户 ${user.username} 的角色失败:`, error)
          user.roleIds = []
        }
      }
      
      // 如果当前页没有数据且不是第一页，自动回到第一页
      if (tableData.value.length === 0 && pagination.pageNum > 1) {
        pagination.pageNum = 1
        await loadData()
      }
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增用户'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑用户'
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该用户吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteUser(row.id)
      ElMessage.success('删除成功')
      loadData()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  })
}

const handleResetPassword = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要将用户「${row.realName || row.username}」的密码重置为 88888888 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await resetPassword(row.id)
    ElMessage.success('重置密码成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重置密码失败:', error)
      ElMessage.error('重置密码失败')
    }
  }
}

const handleAssignRole = async (row) => {
  currentUserId.value = row.id
  
  try {
    // 先加载所有角色
    const rolesRes = await getAllRoles()
    allRoles.value = rolesRes.data || []
    
    // 加载该用户已分配的角色（后端返回的是角色对象数组，需要提取 ID）
    const userRolesRes = await getUserRoles(row.id)
    const userRoles = userRolesRes.data || []
    // 从角色对象数组中提取角色 ID
    selectedRoleIds.value = userRoles.map(role => role.id)
    
    // 确保角色数据加载完成后再显示对话框
    roleDialogVisible.value = true
  } catch (error) {
    console.error('加载角色失败:', error)
    ElMessage.error('加载角色失败')
  }
}

const handleSaveRole = async () => {
  submitLoading.value = true
  try {
    await saveUserRoles(currentUserId.value, { roleIds: selectedRoleIds.value })
    ElMessage.success('分配角色成功')
    roleDialogVisible.value = false
    await loadData() // 刷新列表
  } catch (error) {
    console.error('保存角色失败:', error)
    ElMessage.error('保存角色失败')
  } finally {
    submitLoading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        if (formData.id) {
          await updateUser(formData.id, formData)
          ElMessage.success('更新成功')
        } else {
          await createUser(formData)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadData()
      } catch (error) {
        console.error('提交失败:', error)
        ElMessage.error(error.response?.data?.message || '操作失败')
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const resetForm = () => {
  formData.id = null
  formData.username = ''
  formData.realName = ''
  formData.email = ''
  formData.phone = ''
  formData.status = 1
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 根据角色 ID 获取角色名称
const getRoleName = (roleId) => {
  return roleMap.value.get(roleId) || '未知角色'
}

// 重置查询
const resetQuery = () => {
  queryParams.username = ''
  queryParams.realName = ''
  queryParams.roleId = null
  queryParams.status = null
  pagination.pageNum = 1
  loadData()
}

// 处理每页条数变化
const handleSizeChange = () => {
  pagination.pageNum = 1
  loadData()
}

// 处理页码变化
const handleCurrentChange = () => {
  loadData()
}

onMounted(() => {
  loadData()
})

const formatDateTime = (date) => {
  if (!date) return '-'
  const d = new Date(date)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<style scoped>
.user-management {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-bar {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.pagination-bar {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
