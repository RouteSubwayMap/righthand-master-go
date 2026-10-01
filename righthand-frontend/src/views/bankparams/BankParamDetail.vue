<template>
  <div class="bank-param-detail">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>银行机构参数详情</span>
          <el-button @click="goBack" :icon="ArrowLeft" plain>返回列表</el-button>
        </div>
      </template>

      <!-- 基本信息 -->
      <el-divider content-position="left">基本信息</el-divider>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="银行类型">{{ detailData.bankType || '—' }}</el-descriptions-item>
        <el-descriptions-item label="银行名称">{{ detailData.bankName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="银行机构号">{{ detailData.bankCode || '—' }}</el-descriptions-item>
      </el-descriptions>

      <!-- 备用字段 -->
      <el-divider content-position="left">备用字段</el-divider>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="备用字段1">{{ detailData.reserve1 || '—' }}</el-descriptions-item>
        <el-descriptions-item label="备用字段2">{{ detailData.reserve2 || '—' }}</el-descriptions-item>
        <el-descriptions-item label="备用字段3">{{ detailData.reserve3 || '—' }}</el-descriptions-item>
        <el-descriptions-item label="备用字段4">{{ detailData.reserve4 || '—' }}</el-descriptions-item>
        <el-descriptions-item label="备用字段5" :span="2">{{ detailData.reserve5 || '—' }}</el-descriptions-item>
      </el-descriptions>

      <!-- 其他信息 -->
      <el-divider content-position="left">其他信息</el-divider>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="备注">{{ detailData.remark || '—' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detailData.createTime) }}</el-descriptions-item>
        <el-descriptions-item label="修改时间">{{ formatTime(detailData.updateTime) }}</el-descriptions-item>
      </el-descriptions>

      <!-- 操作按钮 -->
      <div class="form-actions">
        <el-button @click="goBack">关闭</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getBankParam } from '@/api/bankParam'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()

const pageLoading = ref(false)
const detailData = ref({})

const goBack = () => {
  router.push('/bank-params')
}

const goEdit = () => {
  router.push(`/bank-params/${route.params.id}/edit`)
}

const formatTime = (time) => {
  if (!time) return '—'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const loadDetail = async () => {
  pageLoading.value = true
  try {
    const res = await getBankParam(route.params.id)
    detailData.value = res.data || {}
  } catch (e) {
    ElMessage.error('加载数据失败')
    goBack()
  } finally {
    pageLoading.value = false
  }
}

onMounted(() => {
  loadDetail()
})
</script>

<style scoped>
.bank-param-detail {
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

.form-actions {
  margin-top: 32px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
  text-align: center;
}
</style>
