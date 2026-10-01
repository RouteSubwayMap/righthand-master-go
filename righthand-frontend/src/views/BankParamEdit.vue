<template>
  <div class="bank-param-edit">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑银行机构参数' : '新建银行机构参数' }}</span>
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
          <el-col :span="12">
            <el-form-item label="银行类型" prop="bankType">
              <el-select
                v-model="formData.bankType"
                placeholder="请选择银行类型"
                filterable
                style="width: 100%"
                @change="onBankTypeChange"
              >
                <el-option
                  v-for="item in bankTypeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="银行名称" prop="bankName">
              <el-select
                v-model="formData.bankName"
                placeholder="请选择银行名称"
                filterable
                allow-create
                style="width: 100%"
              >
                <el-option
                  v-for="item in formBankNameOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="银行机构号" prop="bankCode">
              <el-input v-model="formData.bankCode" placeholder="请输入银行机构号" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 备用字段 -->
        <el-divider content-position="left">备用字段</el-divider>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="备用字段1">
              <el-input v-model="formData.reserve1" placeholder="备用字段1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="备用字段2">
              <el-input v-model="formData.reserve2" placeholder="备用字段2" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="备用字段3">
              <el-input v-model="formData.reserve3" placeholder="备用字段3" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="备用字段4">
              <el-input v-model="formData.reserve4" placeholder="备用字段4" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="备用字段5">
              <el-input v-model="formData.reserve5" placeholder="备用字段5" />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 其他信息 -->
        <el-divider content-position="left">其他信息</el-divider>

        <el-form-item label="备注">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注信息"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>

        <!-- 操作按钮 -->
        <div class="form-actions">
          <el-button type="primary" @click="submitForm" :loading="submitting">
            {{ isEdit ? '保存修改' : '创建参数' }}
          </el-button>
          <el-button @click="goBack">取消</el-button>
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
import { createBankParam, updateBankParam, getBankParam } from '@/api/bankParam'
import { useBankOptions } from '@/composables/useBankOptions'

const route = useRoute()
const router = useRouter()

const formRef = ref(null)
const pageLoading = ref(false)
const submitting = ref(false)

// 是否为编辑模式（路由中带有 id 参数）
const paramId = computed(() => route.params.id)
const isEdit = computed(() => !!paramId.value)

// 银行类型 / 银行名称 下拉选项来自银行机构参数管理（不再使用数据字典关联）
const { bankTypeOptions, getBankNameOptions } = useBankOptions()

// 银行名称联动：选了银行类型只展示该类型下的银行名称，未选则展示全部
const formBankNameOptions = computed(() => getBankNameOptions(formData.bankType))

// 表单校验规则
const rules = {
  bankType: [{ required: true, message: '请选择银行类型', trigger: 'change' }],
  bankName: [{ required: true, message: '请选择银行名称', trigger: 'change' }],
  bankCode: [{ required: true, message: '请输入银行机构号', trigger: 'blur' }]
}

// 表单数据
const formData = reactive({
  id: null,
  bankType: '',
  bankName: '',
  bankCode: '',
  reserve1: '',
  reserve2: '',
  reserve3: '',
  reserve4: '',
  reserve5: '',
  remark: ''
})

// 银行类型变更时清空银行名称
const onBankTypeChange = () => {
  formData.bankName = ''
}

// 重置表单
const resetForm = () => {
  Object.assign(formData, {
    id: null,
    bankType: '',
    bankName: '',
    bankCode: '',
    reserve1: '',
    reserve2: '',
    reserve3: '',
    reserve4: '',
    reserve5: '',
    remark: ''
  })
  formRef.value?.clearValidate()
}

// 加载参数详情（编辑模式）
const loadParamDetail = async () => {
  pageLoading.value = true
  try {
    const res = await getBankParam(paramId.value)
    if (res.code === 200) {
      const row = res.data
      Object.assign(formData, {
        id: row.id,
        bankType: row.bankType || '',
        bankName: row.bankName || '',
        bankCode: row.bankCode || '',
        reserve1: row.reserve1 || '',
        reserve2: row.reserve2 || '',
        reserve3: row.reserve3 || '',
        reserve4: row.reserve4 || '',
        reserve5: row.reserve5 || '',
        remark: row.remark || ''
      })
    } else {
      ElMessage.error('加载参数详情失败')
      goBack()
    }
  } catch (error) {
    ElMessage.error('加载参数详情失败')
    goBack()
  } finally {
    pageLoading.value = false
  }
}

// 返回列表页
const goBack = () => {
  router.push('/bank-params')
}

// 提交表单
const submitForm = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const payload = {
        bankType: formData.bankType,
        bankName: formData.bankName,
        bankCode: formData.bankCode,
        reserve1: formData.reserve1,
        reserve2: formData.reserve2,
        reserve3: formData.reserve3,
        reserve4: formData.reserve4,
        reserve5: formData.reserve5,
        remark: formData.remark
      }

      if (isEdit.value) {
        await updateBankParam(formData.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createBankParam(payload)
        ElMessage.success('创建成功')
      }
      router.push('/bank-params')
    } catch (error) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    } finally {
      submitting.value = false
    }
  })
}

onMounted(() => {
  if (isEdit.value) {
    loadParamDetail()
  } else {
    resetForm()
  }
})
</script>

<style scoped lang="scss">
.bank-param-edit {
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

    .el-button {
      min-width: 120px;
      margin: 0 8px;
    }
  }
}
</style>
