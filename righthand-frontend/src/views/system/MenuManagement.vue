<template>
  <div class="menu-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>菜单管理</span>
          <el-button v-permission="'system:menu:add'" type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增菜单
          </el-button>
        </div>
      </template>

      <!-- 表格 -->
      <el-table 
        :data="tableData" 
        stripe 
        style="width: 100%"
        row-key="id"
        :tree-props="{ children: 'children' }"
      >
        <el-table-column prop="menuName" label="菜单名" width="200" />
        <el-table-column prop="menuType" label="类型" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.menuType === 0" type="warning">目录</el-tag>
            <el-tag v-else-if="row.menuType === 1" type="success">菜单</el-tag>
            <el-tag v-else-if="row.menuType === 2" type="info">按钮</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" />
        <el-table-column prop="component" label="组件路径" />
        <el-table-column prop="permission" label="权限标识" width="150" />
        <el-table-column prop="icon" label="图标" width="80">
          <template #default="{ row }">
            <el-icon v-if="row.icon"><component :is="row.icon" /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="sortOrder" label="排序" width="80" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'system:menu:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'system:menu:delete'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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
        <el-form-item label="父级菜单" prop="parentId">
          <el-tree-select
            v-model="formData.parentId"
            :data="menuTreeOptions"
            :props="{ children: 'children', label: 'menuName', value: 'id' }"
            check-strictly
            placeholder="选择父级菜单"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="菜单名" prop="menuName">
          <el-input v-model="formData.menuName" placeholder="请输入菜单名" />
        </el-form-item>
        <el-form-item label="菜单类型" prop="menuType">
          <el-radio-group v-model="formData.menuType">
            <el-radio :value="0">目录</el-radio>
            <el-radio :value="1">菜单</el-radio>
            <el-radio :value="2">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="路由路径" prop="path" v-if="formData.menuType !== 2">
          <el-input v-model="formData.path" placeholder="如：/system/user" />
        </el-form-item>
        <el-form-item label="组件路径" prop="component" v-if="formData.menuType === 1">
          <el-input v-model="formData.component" placeholder="如：system/User" />
        </el-form-item>
        <el-form-item label="权限标识" prop="permission" v-if="formData.menuType === 2">
          <el-input v-model="formData.permission" placeholder="如：system:user:add" />
        </el-form-item>
        <el-form-item label="图标" prop="icon" v-if="formData.menuType !== 2">
          <el-input v-model="formData.icon" placeholder="如图标：User" />
        </el-form-item>
        <el-form-item label="排序" prop="sortOrder">
          <el-input-number v-model="formData.sortOrder" :min="0" :max="999" />
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
import { getMenus, createMenu, updateMenu, deleteMenu } from '@/api/system'
import { isCancelError } from '@/utils/request'

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增菜单')
const submitLoading = ref(false)
const formRef = ref(null)

const formData = reactive({
  id: null,
  parentId: 0,
  menuName: '',
  menuType: 1,
  path: '',
  component: '',
  permission: '',
  icon: '',
  sortOrder: 0
})

const rules = {
  menuName: [{ required: true, message: '请输入菜单名', trigger: 'blur' }],
  menuType: [{ required: true, message: '请选择菜单类型', trigger: 'change' }],
}

// 将菜单列表转换为树形结构（用于选择父级）
const menuTreeOptions = computed(() => {
  const buildTree = (items, parentId) => {
    return items
      .filter(item => item.parentId === parentId)
      .map(item => ({
        ...item,
        children: buildTree(items, item.id)
      }))
  }
  return buildTree(tableData.value, 0)
})

const loadData = async () => {
  try {
    const res = await getMenus()
    // 构建树形结构
    const items = res.data || []
    tableData.value = buildTree(items, 0)
  } catch (error) {
    if (!isCancelError(error)) {
      console.error('加载数据失败:', error)
      ElMessage.error('加载数据失败')
    }
  }
}

const buildTree = (items, parentId) => {
  return items
    .filter(item => item.parentId === parentId)
    .map(item => ({
      ...item,
      children: buildTree(items, item.id)
    }))
}

const handleAdd = () => {
  dialogTitle.value = '新增菜单'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑菜单'
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该菜单吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteMenu(row.id)
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
        // 构建提交数据，排除不必要的字段
        const submitData = {
          parentId: formData.parentId,
          menuName: formData.menuName,
          menuType: formData.menuType,
          path: formData.path,
          component: formData.component,
          permission: formData.permission,
          icon: formData.icon,
          sortOrder: formData.sortOrder
        }
        
        if (formData.id) {
          await updateMenu(formData.id, submitData)
          ElMessage.success('更新成功')
        } else {
          await createMenu(submitData)
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
  formData.parentId = 0
  formData.menuName = ''
  formData.menuType = 1
  formData.path = ''
  formData.component = ''
  formData.permission = ''
  formData.icon = ''
  formData.sortOrder = 0
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.menu-management {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
