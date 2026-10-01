<template>
  <div class="role-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button v-permission="'system:role:add'" type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增角色
          </el-button>
        </div>
      </template>

      <!-- 表格 -->
      <el-table :data="tableData" stripe style="width: 100%">
        <el-table-column prop="roleName" label="角色名" width="150" />
        <el-table-column prop="roleCode" label="角色编码" width="150" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'system:role:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'system:role:permission'" type="warning" size="small" @click="handlePermission(row)">权限</el-button>
            <el-button v-permission="'system:role:delete'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
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
        <el-form-item label="角色名" prop="roleName">
          <el-input v-model="formData.roleName" placeholder="请输入角色名" />
        </el-form-item>
        <el-form-item label="角色编码" prop="roleCode">
          <el-input v-model="formData.roleCode" placeholder="请输入角色编码" :disabled="!!formData.id" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
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

    <!-- 权限分配对话框 -->
    <el-dialog
      v-model="permissionDialogVisible"
      title="分配权限"
      width="600px"
    >
      <el-tree
        ref="treeRef"
        :data="menuTreeData"
        :props="{ children: 'children', label: 'menuName' }"
        node-key="id"
        show-checkbox
        default-expand-all
      />
      <template #footer>
        <el-button @click="permissionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSavePermission">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getRoles, createRole, updateRole, deleteRole, getRoleMenus, saveRoleMenus } from '@/api/system'
import { getMenuTree } from '@/api/system'

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增角色')
const submitLoading = ref(false)
const formRef = ref(null)

const permissionDialogVisible = ref(false)
const currentRoleId = ref(null)
const treeRef = ref(null)
const menuTreeData = ref([])

const formData = reactive({
  id: null,
  roleName: '',
  roleCode: '',
  description: ''
})

const rules = {
  roleName: [{ required: true, message: '请输入角色名', trigger: 'blur' }],
  roleCode: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
}

const loadData = async () => {
  try {
    const res = await getRoles()
    tableData.value = res.data?.list || []
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  }
}

const handleAdd = () => {
  dialogTitle.value = '新增角色'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑角色'
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteRole(row.id)
      ElMessage.success('删除成功')
      loadData()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  })
}

const handlePermission = async (row) => {
  currentRoleId.value = row.id

  try {
    // 先显示对话框，让树组件先渲染
    permissionDialogVisible.value = true

    // 等待对话框和树组件渲染完成
    await nextTick()
    await nextTick() // 需要两次 nextTick 确保树完全渲染

    // 加载菜单树
    const menuRes = await getMenuTree()
    menuTreeData.value = menuRes.data || []

    // 再次等待树数据渲染
    await nextTick()

    // 加载该角色已分配的菜单（后端返回的是菜单 ID 数组）
    const roleMenuRes = await getRoleMenus(row.id)
    const roleMenuIds = roleMenuRes.data || []

    // 确保菜单 ID 是数字类型（与树节点 ID 类型一致）
    const checkedKeys = roleMenuIds.map(id => {
      if (typeof id === 'string') {
        return parseInt(id, 10)
      }
      return id
    })

    // 收集所有非叶子节点 ID（有子节点的父节点）
    // setCheckedKeys 会自动级联选中父节点的所有子节点，
    // 因此只设置叶子节点，父节点状态由子节点自动推算
    const parentIds = new Set()
    const collectParentIds = (nodes) => {
      for (const node of nodes) {
        if (node.children && node.children.length > 0) {
          parentIds.add(node.id)
          collectParentIds(node.children)
        }
      }
    }
    collectParentIds(menuTreeData.value)

    // 只保留叶子节点的 ID
    const leafCheckedKeys = checkedKeys.filter(id => !parentIds.has(id))

    // 设置选中的菜单项（仅叶子节点）
    if (treeRef.value) {
      treeRef.value.setCheckedKeys(leafCheckedKeys)
    }
  } catch (error) {
    console.error('加载权限失败:', error)
    ElMessage.error('加载权限失败')
  }
}

const handleSavePermission = async () => {
  if (!treeRef.value) return
  
  submitLoading.value = true
  try {
    const checkedKeys = treeRef.value.getCheckedKeys()
    const halfCheckedKeys = treeRef.value.getHalfCheckedKeys()
    // 合并半选节点（父节点）
    const menuIds = [...checkedKeys, ...halfCheckedKeys]
    
    await saveRoleMenus(currentRoleId.value, { menuIds })
    ElMessage.success('权限分配成功')
    permissionDialogVisible.value = false
  } catch (error) {
    console.error('保存权限失败:', error)
    ElMessage.error('保存权限失败')
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
          await updateRole(formData.id, formData)
          ElMessage.success('更新成功')
        } else {
          await createRole(formData)
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
  formData.roleName = ''
  formData.roleCode = ''
  formData.description = ''
  if (formRef.value) {
    formRef.value.clearValidate()
  }
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
.role-management {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
