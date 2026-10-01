<template>
  <div class="prod-node-edit">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑节点' : '新建节点' }}</span>
          <el-button @click="goBack" :icon="ArrowLeft" plain>返回列表</el-button>
        </div>
      </template>

      <el-form ref="formRef" :model="formData" :rules="rules" label-width="100px">
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
            <el-form-item label="环境类型" prop="envType">
              <el-select v-model="formData.envType" placeholder="请选择" filterable allow-create style="width: 100%">
                <el-option v-for="item in envTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="节点简称" prop="nodeName">
              <el-input v-model="formData.nodeName" placeholder="请输入节点简称" maxlength="100" show-word-limit />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="节点IP" prop="nodeIp">
              <el-input v-model="formData.nodeIp" placeholder="请输入节点IP" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="用户名" prop="username">
              <el-input v-model="formData.username" placeholder="请输入用户名" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="是否互信" prop="isMutualTrust">
              <el-select v-model="formData.isMutualTrust" placeholder="请选择" clearable style="width: 100%">
                <el-option label="是" value="是" />
                <el-option label="否" value="否" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 操作按钮 -->
        <div class="form-actions">
          <el-button @click="goBack">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="submitForm">{{ isEdit ? '保存' : '创建' }}</el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getProdNodeDetail, createProdNode, updateProdNode } from '@/api/prodNode'
import { useDictStore } from '@/stores/dict'
import { useBankOptions } from '@/composables/useBankOptions'

const route = useRoute()
const router = useRouter()
const dictStore = useDictStore()

const formRef = ref()
const pageLoading = ref(false)
const submitting = ref(false)

const isEdit = computed(() => !!route.params.id)

// 银行类型 / 银行名称 下拉选项来自银行机构参数管理（不再使用数据字典关联）
const { bankTypeOptions, getBankNameOptions, loadBankParams } = useBankOptions()
const envTypeOptions = computed(() => dictStore.getDictByType('env_type').map(item => ({ label: item.dictLabel, value: item.dictValue })))

// 银行名称联动：选了银行类型只展示该类型下的银行名称，未选则展示全部
const formBankNameOptions = computed(() => getBankNameOptions(formData.bankType))

// 银行类型变更时清空银行名称
const onBankTypeChange = () => {
  formData.bankName = ''
}

const formData = reactive({
  bankType: '',
  bankName: '',
  envType: '',
  nodeName: '',
  nodeIp: '',
  username: '',
  isMutualTrust: ''
})

const rules = {
  bankName: [{ required: true, message: '请选择银行名称', trigger: 'change' }],
  envType: [{ required: true, message: '请选择环境类型', trigger: 'change' }],
  nodeName: [{ required: true, message: '请输入节点简称', trigger: 'blur' }],
  nodeIp: [{ required: true, message: '请输入节点IP', trigger: 'blur' }]
}

const goBack = () => {
  router.push('/prod-nodes')
}

const loadDetail = async () => {
  if (!isEdit.value) return
  pageLoading.value = true
  try {
    const res = await getProdNodeDetail(route.params.id)
    const data = res.data || {}
    Object.assign(formData, {
      bankType: data.bankType || '',
      bankName: data.bankName || '',
      envType: data.envType || '',
      nodeName: data.nodeName || '',
      nodeIp: data.nodeIp || '',
      username: data.username || '',
      isMutualTrust: data.isMutualTrust || ''
    })
  } catch (e) {
    ElMessage.error('加载数据失败')
    goBack()
  } finally {
    pageLoading.value = false
  }
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updateProdNode(route.params.id, formData)
        ElMessage.success('更新成功')
      } else {
        await createProdNode(formData)
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
  loadDetail()
})
</script>

<style scoped>
.prod-node-edit {
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

.form-actions {
  margin-top: 32px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
  text-align: center;
}
</style>
