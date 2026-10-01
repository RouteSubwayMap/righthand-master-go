<template>
  <div class="batch-task-detail">
    <el-card v-loading="pageLoading">
      <template #header>
        <div class="card-header">
          <span>批量任务详情</span>
          <el-button @click="goBack" :icon="ArrowLeft" plain>返回列表</el-button>
        </div>
      </template>

      <!-- 基本信息 -->
      <el-divider content-position="left">基本信息</el-divider>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="银行类型">{{ detailData.bankType }}</el-descriptions-item>
        <el-descriptions-item label="银行名称">{{ detailData.bankName }}</el-descriptions-item>
        <el-descriptions-item label="所属系统">{{ detailData.systemName }}</el-descriptions-item>
        <el-descriptions-item label="任务类型">{{ detailData.taskType }}</el-descriptions-item>
        <el-descriptions-item label="执行时间">{{ detailData.execTime || '—' }}</el-descriptions-item>
        <el-descriptions-item label="任务名">{{ detailData.taskName }}</el-descriptions-item>
        <el-descriptions-item label="任务描述" :span="3">{{ detailData.taskDesc || '—' }}</el-descriptions-item>
      </el-descriptions>

      <!-- 依赖信息 -->
      <el-divider content-position="left">依赖信息</el-divider>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="上游依赖系统">{{ detailData.upstreamSystem || '—' }}</el-descriptions-item>
        <el-descriptions-item label="下游依赖系统">{{ detailData.downstreamSystem || '—' }}</el-descriptions-item>
        <el-descriptions-item label="上游依赖信息">{{ detailData.upstreamInfo || '—' }}</el-descriptions-item>
        <el-descriptions-item label="下游依赖信息">{{ detailData.downstreamInfo || '—' }}</el-descriptions-item>
      </el-descriptions>

      <!-- 其他 -->
      <el-divider content-position="left">其他</el-divider>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="备注">{{ detailData.remark || '—' }}</el-descriptions-item>
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
import { getBatchTaskDetail } from '@/api/batchTask'

const route = useRoute()
const router = useRouter()

const pageLoading = ref(false)
const detailData = ref({})

const goBack = () => {
  router.push('/batch-tasks')
}

const goEdit = () => {
  router.push(`/batch-tasks/${route.params.id}/edit`)
}

const loadDetail = async () => {
  pageLoading.value = true
  try {
    const res = await getBatchTaskDetail(route.params.id)
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
.batch-task-detail {
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
