<template>
  <div class="batch-task-edit">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑批量任务' : '新建批量任务' }}</span>
          <el-button @click="goBack" :icon="ArrowLeft" plain>返回列表</el-button>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="120px"
        label-position="right"
      >
        <!-- 基本信息 -->
        <el-divider content-position="left">基本信息</el-divider>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="银行类型" prop="bankType">
              <el-select v-model="formData.bankType" placeholder="请选择" filterable style="width: 100%" @change="onBankTypeChange">
                <el-option v-for="item in bankTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="银行名称" prop="bankName">
              <el-select v-model="formData.bankName" placeholder="请选择" filterable allow-create style="width: 100%">
                <el-option v-for="item in formBankNameOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="所属系统" prop="systemName">
              <el-select v-model="formData.systemName" placeholder="请选择" filterable allow-create style="width: 100%">
                <el-option v-for="item in systemNameOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="任务类型" prop="taskType">
              <el-select v-model="formData.taskType" placeholder="请选择" style="width: 100%">
                <el-option v-for="item in batchTaskTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="执行时间" prop="execTime">
              <el-input v-model="formData.execTime" placeholder="cron或时间点，如22:00" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="任务名" prop="taskName">
              <el-input v-model="formData.taskName" placeholder="Job标识" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="任务描述" prop="taskDesc">
          <el-input v-model="formData.taskDesc" type="textarea" :rows="2" placeholder="请输入任务描述" />
        </el-form-item>

        <!-- 依赖信息 -->
        <el-divider content-position="left">依赖信息</el-divider>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="上游依赖系统" prop="upstreamSystem">
              <el-input v-model="formData.upstreamSystem" placeholder="多个用 | 分隔" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下游依赖系统" prop="downstreamSystem">
              <el-input v-model="formData.downstreamSystem" placeholder="多个用 | 分隔" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="上游依赖信息" prop="upstreamInfo">
          <el-input v-model="formData.upstreamInfo" type="textarea" :rows="2" placeholder="依赖条件/文件说明" />
        </el-form-item>
        <el-form-item label="下游依赖信息" prop="downstreamInfo">
          <el-input v-model="formData.downstreamInfo" type="textarea" :rows="2" placeholder="输出文件说明" />
        </el-form-item>

        <!-- 其他 -->
        <el-divider content-position="left">其他</el-divider>

        <el-form-item label="备注" prop="remark">
          <el-input v-model="formData.remark" type="textarea" :rows="2" placeholder="备注信息" />
        </el-form-item>

        <!-- 操作按钮 -->
        <el-form-item>
          <el-button @click="goBack">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { createBatchTask, updateBatchTask, getBatchTaskDetail } from '@/api/batchTask'
import { useDictStore } from '@/stores/dict'
import { useBankOptions } from '@/composables/useBankOptions'

const route = useRoute()
const router = useRouter()
const dictStore = useDictStore()

// 银行类型 / 银行名称 下拉选项来自银行机构参数管理（不再使用数据字典关联）
const { bankTypeOptions, getBankNameOptions, loadBankParams } = useBankOptions()
const systemNameOptions = computed(() => dictStore.getDictByType('system_name').map(item => ({ label: item.dictLabel, value: item.dictValue })))
const batchTaskTypeOptions = computed(() => dictStore.getDictByType('batch_task_type').map(item => ({ label: item.dictLabel, value: item.dictValue })))

// 银行名称联动：选了银行类型只展示该类型下的银行名称，未选则展示全部
const formBankNameOptions = computed(() => getBankNameOptions(formData.bankType))

// 银行类型变更时清空银行名称
const onBankTypeChange = () => {
  formData.bankName = ''
}

const pageLoading = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const formRef = ref()

const formData = reactive({
  id: null,
  bankType: '',
  bankName: '',
  systemName: '',
  taskType: '',
  taskName: '',
  taskDesc: '',
  execTime: '',
  upstreamSystem: '',
  upstreamInfo: '',
  downstreamSystem: '',
  downstreamInfo: '',
  remark: ''
})

const rules = {
  bankType: [{ required: true, message: '请选择银行类型', trigger: 'change' }],
  bankName: [{ required: true, message: '请输入银行名称', trigger: 'blur' }],
  systemName: [{ required: true, message: '请选择所属系统', trigger: 'change' }],
  taskType: [{ required: true, message: '请选择任务类型', trigger: 'change' }],
  taskName: [{ required: true, message: '请输入任务名', trigger: 'blur' }]
}

const goBack = () => {
  router.push('/batch-tasks')
}

const loadDetail = async (id) => {
  pageLoading.value = true
  try {
    const res = await getBatchTaskDetail(id)
    const data = res.data
    Object.assign(formData, {
      id: data.id,
      bankType: data.bankType || '',
      bankName: data.bankName || '',
      systemName: data.systemName || '',
      taskType: data.taskType || '',
      taskName: data.taskName || '',
      taskDesc: data.taskDesc || '',
      execTime: data.execTime || '',
      upstreamSystem: data.upstreamSystem || '',
      upstreamInfo: data.upstreamInfo || '',
      downstreamSystem: data.downstreamSystem || '',
      downstreamInfo: data.downstreamInfo || '',
      remark: data.remark || ''
    })
  } catch (e) {
    ElMessage.error('加载数据失败')
    goBack()
  } finally {
    pageLoading.value = false
  }
}

const handleSubmit = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updateBatchTask(formData.id, formData)
        ElMessage.success('更新成功')
      } else {
        await createBatchTask(formData)
        ElMessage.success('创建成功')
      }
      goBack()
    } catch (e) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    } finally {
      submitting.value = false
    }
  })
}

onMounted(() => {
  dictStore.loadAllDicts()
  loadBankParams()
  const id = route.params.id
  if (id) {
    isEdit.value = true
    loadDetail(id)
  }
})
</script>

<style scoped>
.batch-task-edit {
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
}

:deep(.el-divider__text) {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

:deep(.el-form-item) {
  margin-bottom: 22px;
}
</style>
