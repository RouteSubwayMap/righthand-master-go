<template>
  <div class="life-event">
    <el-button v-permission="'life:add'" type="primary" size="small" @click="handleAdd" style="margin-bottom: 15px">
      <el-icon><Plus /></el-icon>
      新增生活事件
    </el-button>
    
    <el-table :data="tableData" stripe style="width: 100%">
      <el-table-column prop="title" label="标题" />
      <el-table-column prop="occurDate" label="发生时间" width="120" />
      <el-table-column prop="location" label="地点" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button v-permission="'life:edit'" type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button v-permission="'life:delete'" type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      v-model:current-page="pagination.pageNum"
      v-model:page-size="pagination.pageSize"
      :page-sizes="[10, 20, 50, 100]"
      :total="pagination.total"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      style="margin-top: 20px; justify-content: flex-end"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { getLifeEventPage } from '@/api/experience'

const tableData = ref([])

// 分页
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

const loadData = async () => {
  try {
    const res = await getLifeEventPage({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize
    })
    if (res.data) {
      tableData.value = res.data.list || res.data.records || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    // 静默失败
  }
}

const handleAdd = () => {}
const handleEdit = (row) => {}
const handleDelete = (row) => {}

// 分页大小变化
const handleSizeChange = (val) => {
  pagination.pageSize = val
  pagination.pageNum = 1
  loadData()
}

// 当前页变化
const handleCurrentChange = (val) => {
  pagination.pageNum = val
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.life-event {
  padding: 10px;
}
</style>
