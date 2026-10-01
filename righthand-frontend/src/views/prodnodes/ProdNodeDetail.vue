<template>
  <div class="prod-node-detail">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>节点详情</span>
          <el-button @click="goBack" :icon="ArrowLeft" plain>返回列表</el-button>
        </div>
      </template>

      <!-- 基本信息 -->
      <el-divider content-position="left">基本信息</el-divider>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="银行类型">{{ detailData.bankType }}</el-descriptions-item>
        <el-descriptions-item label="银行名称">{{ detailData.bankName }}</el-descriptions-item>
        <el-descriptions-item label="环境类型">{{ detailData.envType }}</el-descriptions-item>
        <el-descriptions-item label="节点简称">{{ detailData.nodeName }}</el-descriptions-item>
        <el-descriptions-item label="节点IP">{{ detailData.nodeIp }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ detailData.username || '—' }}</el-descriptions-item>
        <el-descriptions-item label="是否互信">
          <el-tag v-if="detailData.isMutualTrust === '是' || detailData.isMutualTrust === 'Y'" type="success">是</el-tag>
          <el-tag v-else type="info">否</el-tag>
        </el-descriptions-item>
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
import { getProdNodeDetail } from '@/api/prodNode'

const route = useRoute()
const router = useRouter()

const pageLoading = ref(false)
const detailData = ref({})

const goBack = () => {
  router.push('/prod-nodes')
}

const goEdit = () => {
  router.push(`/prod-nodes/${route.params.id}/edit`)
}

const loadDetail = async () => {
  pageLoading.value = true
  try {
    const res = await getProdNodeDetail(route.params.id)
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
.prod-node-detail {
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
