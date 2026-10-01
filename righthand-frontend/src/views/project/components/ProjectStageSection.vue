<template>
  <div class="project-stage-section">
    <div class="section-header">
      <div class="section-title">
        项目阶段计划
        <span class="muted">共 {{ stages.length }} 个阶段</span>
      </div>
      <el-button v-if="canManageStage" type="primary" size="small" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        <span>新增阶段</span>
      </el-button>
    </div>

    <el-empty v-if="!stages.length" description="暂无项目阶段，请点击「新增阶段」开始规划" :image-size="80" />

    <el-table v-else :data="stages" :row-class-name="rowClassName" stripe>
      <el-table-column label="序号" type="index" width="70" align="center" />
      <el-table-column label="阶段名称" min-width="180">
        <template #default="{ row }">
          <span class="stage-name">{{ row.stageName }}</span>
          <el-tag v-if="row.id === currentStageId" size="small" effect="dark" class="current-tag">当前阶段</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开始时间" width="130" align="center">
        <template #default="{ row }">{{ row.startDate || '-' }}</template>
      </el-table-column>
      <el-table-column label="结束时间" width="130" align="center">
        <template #default="{ row }">{{ row.endDate || '-' }}</template>
      </el-table-column>
      <el-table-column label="阶段说明" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ row.description || '-' }}</template>
      </el-table-column>
      <el-table-column v-if="canManageStage" label="操作" width="120" align="center">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEditDialog(row)">编辑</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无项目阶段" :image-size="60" />
      </template>
    </el-table>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="editingStage ? '编辑阶段' : '新增阶段'" width="520px" :close-on-click-modal="false">
      <el-form ref="stageFormRef" :model="stageForm" :rules="stageRules" label-width="90px">
        <el-form-item label="阶段名称" prop="stageName">
          <el-input v-model="stageForm.stageName" maxlength="100" show-word-limit placeholder="请输入阶段名称" />
        </el-form-item>
        <el-form-item label="开始时间" prop="startDate">
          <el-date-picker v-model="stageForm.startDate" type="date" value-format="YYYY-MM-DD" placeholder="选择开始日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="结束时间" prop="endDate">
          <el-date-picker v-model="stageForm.endDate" type="date" value-format="YYYY-MM-DD" placeholder="选择结束日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="阶段说明" prop="description">
          <el-input v-model="stageForm.description" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="阶段说明（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { createProjectStage, updateProjectStage, deleteProjectStage } from '@/api/project'

const props = defineProps({
  project: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['refresh'])

/* ------------------------------- 展示计算 ------------------------------- */

// 阶段列表由详情接口内嵌返回（已按开始时间正序），不额外请求
const stages = computed(() => props.project?.stages || [])

// 当前阶段 ID 由后端统一计算下发，前端不复制日期判定逻辑
const currentStageId = computed(() => props.project?.currentStageId ?? null)

// 操作按钮显隐完全由后端 allowedActions 决定（与其他卡片保持一致）
const canManageStage = computed(() => (props.project?.allowedActions || []).includes('manageStage'))

function rowClassName({ row }) {
  return row.id === currentStageId.value ? 'current-stage-row' : ''
}

/* ------------------------------- 新增 / 编辑 ------------------------------- */

const dialogVisible = ref(false)
const saving = ref(false)
const editingStage = ref(null)
const stageFormRef = ref(null)

const stageForm = reactive({
  stageName: '',
  startDate: '',
  endDate: '',
  description: ''
})

const stageRules = {
  stageName: [{ required: true, message: '请输入阶段名称', trigger: 'blur' }],
  startDate: [{ required: true, message: '请选择开始日期', trigger: 'change' }],
  endDate: [
    { required: true, message: '请选择结束日期', trigger: 'change' },
    {
      validator: (rule, value, callback) => {
        if (value && stageForm.startDate && value < stageForm.startDate) {
          callback(new Error('结束日期不能早于开始日期'))
          return
        }
        callback()
      },
      trigger: 'change'
    }
  ]
}

function openCreateDialog() {
  editingStage.value = null
  stageForm.stageName = ''
  stageForm.startDate = ''
  stageForm.endDate = ''
  stageForm.description = ''
  dialogVisible.value = true
  stageFormRef.value?.clearValidate()
}

function openEditDialog(stage) {
  editingStage.value = stage
  stageForm.stageName = stage.stageName || ''
  stageForm.startDate = stage.startDate || ''
  stageForm.endDate = stage.endDate || ''
  stageForm.description = stage.description || ''
  dialogVisible.value = true
  stageFormRef.value?.clearValidate()
}

async function handleSubmit() {
  try {
    await stageFormRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    const payload = { ...stageForm }
    if (editingStage.value) {
      await updateProjectStage(props.project.id, editingStage.value.id, payload)
      ElMessage.success('保存成功')
    } else {
      await createProjectStage(props.project.id, payload)
      ElMessage.success('新增成功')
    }
    dialogVisible.value = false
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

/* ------------------------------- 删除 ------------------------------- */

async function handleDelete(stage) {
  try {
    await ElMessageBox.confirm(`确认删除阶段「${stage.stageName}」？`, '删除确认', { type: 'warning', confirmButtonText: '删除' })
  } catch {
    return
  }
  try {
    await deleteProjectStage(props.project.id, stage.id)
    ElMessage.success('删除成功')
    emit('refresh')
  } catch {
    // 错误已由拦截器提示
  }
}
</script>

<style scoped>
.project-stage-section {
  min-height: 160px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.stage-name {
  margin-right: 8px;
}

.current-tag {
  vertical-align: middle;
}

.muted {
  color: #909399;
  font-size: 12px;
  font-weight: 400;
  margin-left: 6px;
}

/* 当前阶段行高亮：浅蓝背景 + 左侧主色边线（提升特异性覆盖 stripe / hover 行背景） */
.el-table :deep(.el-table__body .current-stage-row td.el-table__cell) {
  background-color: #ecf5ff;
}

.el-table :deep(.el-table__body .current-stage-row td.el-table__cell:first-child) {
  box-shadow: inset 3px 0 0 #409eff;
}
</style>
