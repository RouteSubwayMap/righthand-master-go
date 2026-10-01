<template>
  <el-dialog
    :model-value="visible"
    :title="isEdit ? '编辑项目' : '新建项目'"
    width="640px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:visible', $event)"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="项目名称" prop="projectName">
        <el-input v-model="form.projectName" maxlength="100" placeholder="请输入项目名称" />
      </el-form-item>
      <el-form-item label="项目Key" prop="projectKey">
        <el-input v-model="form.projectKey" maxlength="50" placeholder="唯一标识，如 NCBC" :disabled="isEdit" />
      </el-form-item>
      <el-form-item label="起止日期" prop="startDate">
        <div class="date-range">
          <el-date-picker
            v-model="form.startDate"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="开始日期"
            style="width: 100%"
          />
          <span class="date-sep">至</span>
          <el-date-picker
            v-model="form.endDate"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="结束日期"
            style="width: 100%"
          />
        </div>
      </el-form-item>
      <el-form-item label="项目经理" prop="managerId">
        <el-select v-model="form.managerId" filterable placeholder="请选择项目经理" style="width: 100%">
          <el-option v-for="u in users" :key="u.id" :label="userLabel(u)" :value="u.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="整体计划说明" prop="planDesc">
        <el-input
          v-model="form.planDesc"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="项目整体计划/里程碑说明（可选）"
        />
      </el-form-item>
      <el-form-item label="项目描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="2"
          maxlength="500"
          show-word-limit
          placeholder="项目背景/目标描述（可选）"
        />
      </el-form-item>
      <el-alert v-if="!isEdit" type="info" :closable="false" show-icon>
        新建后项目状态为「启动」，点「开始」后进入进行中；仅「启动」状态可删除。
      </el-alert>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { createProject, updateProject } from '@/api/project'

const props = defineProps({
  visible: { type: Boolean, default: false },
  project: { type: Object, default: null },
  users: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:visible', 'saved'])

const formRef = ref(null)
const saving = ref(false)

const isEdit = computed(() => !!props.project?.id)

const form = reactive({
  projectName: '',
  projectKey: '',
  startDate: '',
  endDate: '',
  managerId: null,
  planDesc: '',
  description: ''
})

const rules = {
  projectName: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  projectKey: [
    { required: true, message: '请输入项目Key', trigger: 'blur' },
    { pattern: /^[A-Za-z0-9_-]+$/, message: '仅支持字母、数字、下划线、中划线', trigger: 'blur' }
  ],
  startDate: [
    {
      validator: (rule, value, callback) => {
        if (!form.startDate || !form.endDate) return callback(new Error('请选择起止日期'))
        if (form.endDate < form.startDate) return callback(new Error('结束日期不能早于开始日期'))
        callback()
      },
      trigger: 'change'
    }
  ],
  managerId: [{ required: true, message: '请选择项目经理', trigger: 'change' }]
}

function userLabel(u) {
  return u.realName || u.nickname || u.username || `用户${u.id}`
}

watch(
  () => props.visible,
  (val) => {
    if (!val) return
    if (props.project) {
      form.projectName = props.project.projectName || ''
      form.projectKey = props.project.projectKey || ''
      form.startDate = props.project.startDate || ''
      form.endDate = props.project.endDate || ''
      form.managerId = props.project.managerId || null
      form.planDesc = props.project.planDesc || ''
      form.description = props.project.description || ''
    } else {
      form.projectName = ''
      form.projectKey = ''
      form.startDate = ''
      form.endDate = ''
      form.managerId = null
      form.planDesc = ''
      form.description = ''
    }
    formRef.value?.clearValidate()
  }
)

function handleClosed() {
  formRef.value?.resetFields()
}

async function handleSubmit() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    const payload = { ...form }
    if (isEdit.value) {
      await updateProject(props.project.id, payload)
      ElMessage.success('项目更新成功')
    } else {
      await createProject(payload)
      ElMessage.success('项目创建成功')
    }
    emit('update:visible', false)
    emit('saved')
  } catch {
    // 业务错误已由 request 拦截器统一提示
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.date-range {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.date-sep {
  color: #909399;
  flex-shrink: 0;
}
</style>
