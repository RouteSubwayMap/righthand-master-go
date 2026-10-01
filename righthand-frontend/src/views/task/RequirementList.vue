<template>
  <div class="requirement-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>需求管理</span>
          <div class="header-actions">
            <el-button type="primary" @click="downloadTemplateFile">
              <el-icon><Download /></el-icon>
              下载模板
            </el-button>
            <el-button type="warning" @click="handleExport">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
            <el-button type="primary" @click="openImportDialog">
              <el-icon><Upload /></el-icon>
              批量导入
            </el-button>
            <el-button type="success" @click="handleCreate">
              <el-icon><Plus /></el-icon>
              新增需求
            </el-button>
          </div>
        </div>
      </template>

      <!-- 筛选条件 -->
      <div class="filter-bar">
        <el-form :model="queryParams">
          <el-row :gutter="20">
            <el-col :span="6">
              <el-form-item label="需求名称">
                <el-input v-model="queryParams.title" placeholder="请输入需求名称" clearable />
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="所属银行">
                <el-input v-model="queryParams.bankName" placeholder="请输入所属银行" clearable />
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="需求编号">
                <el-input v-model="queryParams.requirementNo" placeholder="请输入需求编号" clearable />
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="状态">
                <el-select v-model="queryParams.status" placeholder="全部" clearable>
                  <el-option label="待澄清" value="PENDING_CLARIFY" />
                  <el-option label="待排期" value="PENDING_SCHEDULING" />
                  <el-option label="开发中" value="DEVELOPING" />
                  <el-option label="测试中" value="TESTING" />
                  <el-option label="待投产" value="PENDING_RELEASE" />
                  <el-option label="已完成" value="COMPLETED" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="20">
            <el-col :span="6">
              <el-form-item label="优先级">
                <el-select v-model="queryParams.priority" placeholder="全部" clearable>
                  <el-option label="高" value="HIGH" />
                  <el-option label="中" value="MEDIUM" />
                  <el-option label="低" value="LOW" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="风险等级">
                <el-select v-model="queryParams.riskLevel" placeholder="全部" clearable>
                  <el-option label="高" value="HIGH" />
                  <el-option label="中" value="MEDIUM" />
                  <el-option label="低" value="LOW" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="需求类型">
                <el-select v-model="queryParams.requirementType" placeholder="全部" clearable>
                  <el-option label="行内需求" value="BANK" />
                  <el-option label="内部优化" value="INTERNAL" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="所属小组">
                <el-select v-model="queryParams.groupType" placeholder="全部" clearable>
                  <el-option label="贷中" value="LOAN_MIDDLE" />
                  <el-option label="贷前" value="LOAN_PRE" />
                  <el-option label="风险" value="RISK" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="20">
            <el-col :span="6">
              <el-form-item label="需求跟进人">
                <el-select v-model="queryParams.followerName" placeholder="全部" clearable filterable>
                  <el-option
                    v-for="user in userList"
                    :key="user.id"
                    :label="user.realName || user.username"
                    :value="user.realName || user.username"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="创建人">
                <el-select v-model="queryParams.createdBy" placeholder="全部" clearable filterable>
                  <el-option
                    v-for="user in userList"
                    :key="user.id"
                    :label="user.realName || user.username"
                    :value="user.id"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="完成时间">
                <el-date-picker
                  v-model="queryParams.completionDateRange"
                  type="daterange"
                  range-separator="至"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  value-format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item>
                <el-button type="primary" @click="loadData">查询</el-button>
                <el-button @click="resetQuery">重置</el-button>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </div>

      <!-- 需求列表 -->
      <el-table :data="tableData" stripe style="width: 100%" v-loading="loading">
        <el-table-column prop="requirementNo" label="需求编号" width="120" />
        <el-table-column prop="title" label="需求名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="bankName" label="所属银行" width="120" show-overflow-tooltip />
        
        <el-table-column prop="priority" label="优先级" width="100" sortable>
          <template #default="{ row }">
            <el-tag 
              :type="row.priority === 'HIGH' ? 'danger' : row.priority === 'MEDIUM' ? 'warning' : 'info'"
              size="small"
            >
              {{ row.priority === 'HIGH' ? '高' : row.priority === 'MEDIUM' ? '中' : '低' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="riskLevel" label="风险等级" width="80">
          <template #default="{ row }">
            <el-tag 
              :type="row.riskLevel === 'HIGH' ? 'danger' : row.riskLevel === 'MEDIUM' ? 'warning' : 'info'"
              size="small"
            >
              {{ row.riskLevel === 'HIGH' ? '高' : row.riskLevel === 'MEDIUM' ? '中' : '低' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="requirementType" label="需求类型" width="100">
          <template #default="{ row }">
            <el-tag 
              :type="row.requirementType === 'BANK' ? 'info' : 'success'"
              size="small"
            >
              {{ row.requirementType === 'BANK' ? '行内需求' : row.requirementType === 'INTERNAL' ? '内部优化' : row.requirementType || '-' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="groupType" label="所属小组" width="80">
          <template #default="{ row }">
            {{ getGroupLabel(row.groupType) }}
          </template>
        </el-table-column>

        <el-table-column prop="followerName" label="需求跟进人" width="100" />
        <el-table-column label="创建人" width="100">
          <template #default="{ row }">
            {{ getCreatorName(row) }}
          </template>
        </el-table-column>
        <el-table-column label="排期" width="150" sortable>
          <template #default="{ row }">
            {{ formatDate(row.scheduleDate) }}
          </template>
        </el-table-column>
        <el-table-column label="完成时间" width="160" sortable>
          <template #default="{ row }">
            {{ row.completionDate ? formatDateTime(row.completionDate) : '—' }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleView(row)">
              查看
            </el-button>
            <el-button
              v-if="!isRequirementCompleted(row)"
              type="success"
              size="small"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="!isRequirementCompleted(row)"
              type="warning"
              size="small"
              @click="handleWorkflow(row)"
            >
              流转
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="pagination.total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑需求弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="1200px"
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="140px">
        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="需求名称" prop="title">
              <el-input v-model="formData.title" placeholder="请输入需求名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属银行" prop="bankName">
              <el-input v-model="formData.bankName" placeholder="请输入所属银行" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="需求编号" prop="requirementNo">
              <el-input v-model="formData.requirementNo" placeholder="请输入需求编号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="formData.priority" placeholder="请选择优先级" style="width: 100%">
                <el-option label="高" value="HIGH" />
                <el-option label="中" value="MEDIUM" />
                <el-option label="低" value="LOW" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="风险等级" prop="riskLevel">
              <el-select v-model="formData.riskLevel" placeholder="请选择风险等级" style="width: 100%">
                <el-option label="高" value="HIGH" />
                <el-option label="中" value="MEDIUM" />
                <el-option label="低" value="LOW" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="需求类型" prop="requirementType">
              <el-select v-model="formData.requirementType" placeholder="请选择需求类型" style="width: 100%">
                <el-option label="行内需求" value="BANK" />
                <el-option label="内部优化" value="INTERNAL" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-select v-model="formData.status" placeholder="请选择状态" style="width: 100%">
                <el-option label="待澄清" value="PENDING_CLARIFY" />
                <el-option label="待排期" value="PENDING_SCHEDULING" />
                <el-option label="开发中" value="DEVELOPING" />
                <el-option label="测试中" value="TESTING" />
                <el-option label="待投产" value="PENDING_RELEASE" />
                <el-option label="已完成" value="COMPLETED" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="主 R 系统" prop="mainSystem">
              <el-input v-model="formData.mainSystem" placeholder="请输入主 R 系统" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="开发主 R|参与者" prop="developer">
              <el-input v-model="formData.developer" placeholder="请输入开发主 R|参与者" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="涉及系统" prop="involvedSystems">
              <el-input v-model="formData.involvedSystems" placeholder="请输入涉及系统" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="所属小组" prop="groupType">
              <el-select v-model="formData.groupType" placeholder="请选择所属小组" style="width: 100%">
                <el-option label="贷中" value="LOAN_MIDDLE" />
                <el-option label="贷前" value="LOAN_PRE" />
                <el-option label="风险" value="RISK" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="需求跟进人" prop="followerId">
              <el-select v-model="formData.followerId" placeholder="请选择需求跟进人" filterable style="width: 100%">
                <el-option
                  v-for="user in userList"
                  :key="user.id"
                  :label="user.realName || user.username"
                  :value="user.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="提交人" prop="submitterName">
              <el-select v-model="formData.submitterName" placeholder="请选择提交人" clearable filterable>
                <el-option
                  v-for="user in userList"
                  :key="user.id"
                  :label="user.realName || user.username"
                  :value="user.realName || user.username"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="需求提出时间" prop="submitTime">
              <el-date-picker
                v-model="formData.submitTime"
                type="datetime"
                placeholder="选择需求提出时间"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="需求完成时间" prop="completionDate">
              <el-date-picker
                v-model="formData.completionDate"
                type="datetime"
                placeholder="选择需求完成时间"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="30">
          <el-col :span="12">
            <el-form-item label="排期" prop="scheduleDate">
              <el-date-picker
                v-model="formData.scheduleDate"
                type="date"
                placeholder="请选择排期日期"
                clearable
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预计工作量（人天）" prop="estimatedWorkload">
              <el-input-number v-model="formData.estimatedWorkload" :min="0" :precision="2" :step="0.5" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="需求描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="6"
            :autosize="{ minRows: 4, maxRows: 15 }"
            placeholder="请输入需求详细描述"
          />
        </el-form-item>

        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="4"
            :autosize="{ minRows: 2, maxRows: 10 }"
            placeholder="备注信息（可选）"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 需求详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="需求详情"
      width="900px"
    >
      <div v-if="currentRequirement" class="requirement-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="需求编号">{{ currentRequirement.requirementNo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="需求名称">{{ currentRequirement.title }}</el-descriptions-item>
          <el-descriptions-item label="所属银行">{{ currentRequirement.bankName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="优先级">
            <el-tag :type="currentRequirement.priority === 'HIGH' ? 'danger' : currentRequirement.priority === 'MEDIUM' ? 'warning' : 'info'" size="small">
              {{ currentRequirement.priority === 'HIGH' ? '高' : currentRequirement.priority === 'MEDIUM' ? '中' : '低' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <el-tag :type="currentRequirement.riskLevel === 'HIGH' ? 'danger' : currentRequirement.riskLevel === 'MEDIUM' ? 'warning' : 'info'" size="small">
              {{ currentRequirement.riskLevel === 'HIGH' ? '高' : currentRequirement.riskLevel === 'MEDIUM' ? '中' : '低' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="需求类型">
            <el-tag :type="currentRequirement.requirementType === 'BANK' ? 'primary' : 'success'" size="small">
              {{ currentRequirement.requirementType === 'BANK' ? '行内需求' : '内部优化' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusTag(currentRequirement.status)">
              {{ getStatusLabel(currentRequirement.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="主 R 系统">{{ currentRequirement.mainSystem || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开发主 R|参与者">{{ currentRequirement.developer || '-' }}</el-descriptions-item>
          <el-descriptions-item label="涉及系统">{{ currentRequirement.involvedSystems || '-' }}</el-descriptions-item>
          <el-descriptions-item label="所属小组">{{ getGroupLabel(currentRequirement.groupType) }}</el-descriptions-item>
          <el-descriptions-item label="需求跟进人">{{ currentRequirement.followerName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ getCreatorName(currentRequirement) }}</el-descriptions-item>
          <el-descriptions-item label="提交人">{{ currentRequirement.submitterName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="需求提出时间">{{ formatDateTime(currentRequirement.submitTime) }}</el-descriptions-item>
          <el-descriptions-item label="需求完成时间">{{ formatDateTime(currentRequirement.completionDate) }}</el-descriptions-item>
          <el-descriptions-item label="预计工作量（人天）">{{ getEstimatedWorkload(currentRequirement.estimatedWorkload) }}</el-descriptions-item>
        </el-descriptions>
        
        <!-- 需求描述和备注单独展示 -->
        <el-divider content-position="left">需求描述</el-divider>
        <div class="detail-text-content">
          {{ currentRequirement.description || '无' }}
        </div>
        
        <el-divider content-position="left">备注</el-divider>
        <div class="detail-text-content">
          {{ currentRequirement.remark || '-' }}
        </div>

        <!-- 关联事项列表 -->
        <el-divider content-position="left">关联事项</el-divider>
        <el-table :data="requirementTasks" v-loading="tasksLoading" border stripe size="small" style="width: 100%">
          <el-table-column type="index" label="序号" width="60" align="center" />
          <el-table-column prop="title" label="事项标题" min-width="200" show-overflow-tooltip />
          <el-table-column label="类型" width="80" align="center">
            <template #default="{ row }">
              <el-tag size="small" :type="row.itemType === 'REQUIREMENT' ? 'warning' : row.itemType === 'MEMO' ? 'success' : undefined">
                {{ row.itemType === 'REQUIREMENT' ? '需求' : row.itemType === 'MEMO' ? '备忘' : '任务' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="优先级" width="80" align="center">
            <template #default="{ row }">
              <el-tag size="small" :type="row.priority === 'HIGH' ? 'danger' : row.priority === 'MEDIUM' ? 'warning' : 'info'">
                {{ row.priority === 'HIGH' ? '高' : row.priority === 'MEDIUM' ? '中' : '低' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag size="small" :type="row.statusCode === 'TODO' ? 'info' : row.statusCode === 'DOING' ? 'warning' : row.statusCode === 'DONE' ? 'success' : undefined">
                {{ row.statusCode === 'TODO' ? '待办' : row.statusCode === 'DOING' ? '进行中' : row.statusCode === 'DONE' ? '已完成' : '已取消' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="截止时间" width="160" align="center">
            <template #default="{ row }">
              {{ row.deadline ? formatDateTime(row.deadline) : '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" align="center" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" link @click="viewTaskDetail(row)">查看</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="!tasksLoading && requirementTasks.length === 0" description="暂无关联事项" :image-size="60" />
      </div>
    </el-dialog>

    <!-- 事项详情弹窗 -->
    <el-dialog
      v-model="taskDetailVisible"
      title="事项详情"
      width="800px"
      :destroy-on-close="true"
    >
      <div v-loading="taskDetailLoading">
        <template v-if="taskDetailData">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="标题" :span="2">{{ taskDetailData.title || '—' }}</el-descriptions-item>
            <el-descriptions-item label="类型">
              <el-tag size="small" :type="taskDetailData.itemType === 'REQUIREMENT' ? 'warning' : taskDetailData.itemType === 'MEMO' ? 'success' : undefined">
                {{ taskDetailData.itemType === 'REQUIREMENT' ? '需求' : taskDetailData.itemType === 'MEMO' ? '备忘' : '任务' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="优先级">
              <el-tag size="small" :type="taskDetailData.priority === 'HIGH' ? 'danger' : taskDetailData.priority === 'MEDIUM' ? 'warning' : 'info'">
                {{ taskDetailData.priority === 'HIGH' ? '高' : taskDetailData.priority === 'MEDIUM' ? '中' : '低' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="分类">{{ dictStore.getDictLabel('memo_type', taskDetailData.category) || '—' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag size="small" :type="taskDetailData.statusCode === 'TODO' ? 'info' : taskDetailData.statusCode === 'DOING' ? 'warning' : taskDetailData.statusCode === 'DONE' ? 'success' : undefined">
                {{ taskDetailData.statusCode === 'TODO' ? '待办' : taskDetailData.statusCode === 'DOING' ? '进行中' : taskDetailData.statusCode === 'DONE' ? '已完成' : '已取消' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="截止时间" :span="2">{{ taskDetailData.deadline ? formatDateTime(taskDetailData.deadline) : '—' }}</el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">详细描述</el-divider>
          <div class="detail-text-content">{{ taskDetailData.description || '暂无描述' }}</div>

          <el-divider content-position="left">周期性设置</el-divider>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="重复类型">{{ taskDetailData.recurrenceType === 'DAILY' ? '每天' : taskDetailData.recurrenceType === 'WEEKLY' ? '每周' : taskDetailData.recurrenceType === 'MONTHLY' ? '每月' : '不重复' }}</el-descriptions-item>
            <el-descriptions-item label="结束条件">{{ taskDetailData.recurrenceEndType === 'NEVER' ? '永不结束' : taskDetailData.recurrenceEndType === 'DATE' ? '到指定日期' : taskDetailData.recurrenceEndType === 'COUNT' ? '重复次数' : '永不结束' }}</el-descriptions-item>
            <el-descriptions-item v-if="taskDetailData.recurrenceEndType === 'DATE'" label="结束日期" :span="2">{{ taskDetailData.recurrenceEndDate ? formatDateTime(taskDetailData.recurrenceEndDate) : '—' }}</el-descriptions-item>
          </el-descriptions>

          <template v-if="taskDetailData.itemType === 'REQUIREMENT' && taskDetailData.requirementId">
            <el-divider content-position="left">需求关联</el-divider>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="关联需求">
                [{{ taskDetailData.requirementNo || '—' }}] {{ taskDetailData.requirementTitle || '—' }}
              </el-descriptions-item>
            </el-descriptions>
          </template>
        </template>
      </div>
      <template #footer>
        <el-button @click="taskDetailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="批量导入需求"
      width="600px"
    >
      <el-form label-width="120px">
        <el-form-item label="上传文件">
          <el-upload
            ref="uploadRef"
            drag
            :auto-upload="false"
            :on-change="handleFileChange"
            :limit="1"
            accept=".xlsx,.xls"
          >
            <el-icon class="el-icon--upload"><Upload-Filled /></el-icon>
            <div class="el-upload__text">
              将文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                只能上传 Excel 文件，请先点击"下载模板"获取标准格式
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleImport" :loading="importLoading">
          开始导入
        </el-button>
      </template>
    </el-dialog>

    <!-- 工作流流转弹窗 -->
    <el-dialog
      v-model="workflowVisible"
      title="工作流流转"
      width="800px"
    >
      <div v-if="currentRequirement" class="workflow-transition">
        <!-- 当前状态提示 -->
        <el-alert
          :title="`当前状态：${getStatusLabel(currentRequirement.status)}`"
          type="warning"
          :closable="false"
          show-icon
          style="margin-bottom: 20px;"
        />

        <!-- 流程图简化展示 -->
        <div class="workflow-diagram">
          <div 
            v-for="(status, index) in workflowStatuses" 
            :key="status.id"
            class="status-node"
            :class="{
              'current': status.statusCode === currentRequirement.status,
              'reachable': isReachableStatus(status)
            }"
          >
            <div class="status-indicator">
              <span class="status-number">{{ index + 1 }}</span>
            </div>
            <div class="status-info">
              <div class="status-name">{{ status.statusName }}</div>
              <div class="status-arrow" v-if="index < workflowStatuses.length - 1">→</div>
            </div>
          </div>
        </div>

        <el-divider />

        <!-- 可执行的流转操作 -->
        <div class="transition-section">
          <h4>可执行的操作</h4>
          
          <!-- 上一步操作 -->
          <div v-if="previousTransitions.length > 0" class="transition-group">
            <div class="transition-label">
              <el-icon><ArrowLeft /></el-icon>
              <span>返回上一步</span>
            </div>
            <div class="transition-buttons">
              <el-button
                v-for="transition in previousTransitions"
                :key="transition.id"
                type="warning"
                size="small"
                @click="executeTransitionClick(transition)"
              >
                {{ transition.name }}
              </el-button>
            </div>
          </div>

          <!-- 下一步操作 -->
          <div v-if="nextTransitions.length > 0" class="transition-group">
            <div class="transition-label">
              <el-icon><ArrowRight /></el-icon>
              <span>进入下一步</span>
            </div>
            <div class="transition-buttons">
              <el-button
                v-for="transition in nextTransitions"
                :key="transition.id"
                type="primary"
                size="small"
                @click="executeTransitionClick(transition)"
              >
                {{ transition.name }}
              </el-button>
            </div>
          </div>

          <!-- 直接完成 -->
          <div v-if="completeTransition" class="transition-group">
            <div class="transition-label">
              <el-icon><CircleCheck /></el-icon>
              <span>完成需求</span>
            </div>
            <div class="transition-buttons">
              <el-button
                type="success"
                size="small"
                @click="executeTransitionClick(completeTransition)"
              >
                {{ completeTransition.name }}
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Upload, Download, UploadFilled, ArrowLeft, ArrowRight, CircleCheck } from '@element-plus/icons-vue'
import {
  getRequirementPage,
  getRequirementDetail,
  getRequirementTasks,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  exportRequirements,
  batchImport
} from '@/api/requirement'
import { getTaskDetail } from '@/api/task'
import { getAllUsers } from '@/api/user'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'
import * as XLSX from 'xlsx'
import { useDictStore } from '@/stores/dict'

const dictStore = useDictStore()
const userStore = useUserStore()
// 注意：需求有自己的流转接口，不使用任务的 executeTransition
import { executeRequirementTransition } from '@/api/requirement'

const loading = ref(false)
const submitLoading = ref(false)
const importLoading = ref(false)
const dialogVisible = ref(false)
const detailVisible = ref(false)
const importDialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const uploadRef = ref(null)
const selectedFile = ref(null)

// 关联事项
const requirementTasks = ref([])
const tasksLoading = ref(false)

// 事项详情弹窗
const taskDetailVisible = ref(false)
const taskDetailLoading = ref(false)
const taskDetailData = ref(null)

const router = useRouter()

// 查询参数
const queryParams = reactive({
  id: '',
  title: '',
  bankName: '',
  requirementNo: '',
  status: '',
  priority: '',
  riskLevel: '',
  requirementType: '',
  groupType: '',
  followerName: '',
  createdBy: '',
  completionDateRange: []
})

// 分页
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 表格数据
const tableData = ref([])

// 表单数据
const formData = reactive({
  id: null,
  title: '',
  bankName: '',
  requirementNo: '',
  priority: '',
  riskLevel: '',
  requirementType: 'INTERNAL',
  status: '',
  scheduleDate: '',
  mainSystem: '',
  developer: '',
  involvedSystems: '',
  followerId: null,
  submitterName: '',
  submitTime: '',
  completionDate: '',
  estimatedWorkload: null,
  groupType: '',
  description: '',
  remark: ''
})

// 当前查看的需求
const currentRequirement = ref(null)

// 工作流流转相关
const workflowVisible = ref(false)
const workflowTransitions = ref([])
const workflowStatuses = ref([])

// 用户列表
const userList = ref([])

// 表单验证规则
const formRules = {
  title: [{ required: true, message: '请输入需求标题', trigger: 'blur' }],
  requirementType: [{ required: true, message: '请选择需求类型', trigger: 'change' }],
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }]
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const { completionDateRange, ...rest } = queryParams
    const params = {
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      ...rest,
      completionDateStart: completionDateRange && completionDateRange[0] ? dayjs(completionDateRange[0]).format('YYYY-MM-DD') : '',
      completionDateEnd: completionDateRange && completionDateRange[1] ? dayjs(completionDateRange[1]).format('YYYY-MM-DD') : ''
    }
    const res = await getRequirementPage(params)
    
    // 安全检查：确保 res 和 res.data 存在
    if (!res || !res.data) {
      tableData.value = []
      pagination.total = 0
      return
    }
    
    if (res.code === 200) {
      tableData.value = res.data.records || res.data.list || []
      pagination.total = res.data.total || 0
      
      // 如果当前页没有数据且不是第一页，自动回到第一页
      if (tableData.value.length === 0 && pagination.pageNum > 1) {
        pagination.pageNum = 1
        await loadData()
      }
    } else if (res.message) {
      ElMessage.warning(res.message)
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 重置查询
const resetQuery = () => {
  queryParams.id = ''
  queryParams.title = ''
  queryParams.bankName = ''
  queryParams.requirementNo = ''
  queryParams.status = ''
  queryParams.priority = ''
  queryParams.riskLevel = ''
  queryParams.requirementType = ''
  queryParams.groupType = ''
  queryParams.followerName = ''
  queryParams.createdBy = ''
  queryParams.completionDateRange = []
  pagination.pageNum = 1
  loadData()
}

// 处理每页条数变化
const handleSizeChange = () => {
  pagination.pageNum = 1 // 改变每页条数时，重置到第一页
  loadData()
}

// 处理页码变化
const handleCurrentChange = () => {
  loadData()
}

// 新建需求
const handleCreate = () => {
  dialogTitle.value = '新增需求'
  resetForm()
  // 设置默认状态为"新建"
  formData.status = 'PENDING_CLARIFY'
  // 设置默认优先级为"中"
  formData.priority = 'MEDIUM'
  dialogVisible.value = true
}

// 工作流流转
// 工作流流转
const handleWorkflow = async (row) => {
  currentRequirement.value = row
  try {
    // Go后端简化了状态流转,直接使用固定的5个状态
    // 定义需求管理的状态转换规则
    const statusTransitions = {
      'PENDING_CLARIFY': [
        { id: 'PENDING_SCHEDULING', name: '转为待排期', targetStatus: 'PENDING_SCHEDULING' },
        { id: 'COMPLETED', name: '直接完成', targetStatus: 'COMPLETED' }
      ],
      'PENDING_SCHEDULING': [
        { id: 'DEVELOPING', name: '开始开发', targetStatus: 'DEVELOPING' },
        { id: 'PENDING_CLARIFY', name: '返回待澄清', targetStatus: 'PENDING_CLARIFY' },
        { id: 'COMPLETED', name: '直接完成', targetStatus: 'COMPLETED' }
      ],
      'DEVELOPING': [
        { id: 'TESTING', name: '进入测试', targetStatus: 'TESTING' },
        { id: 'PENDING_SCHEDULING', name: '返回待排期', targetStatus: 'PENDING_SCHEDULING' },
        { id: 'COMPLETED', name: '直接完成', targetStatus: 'COMPLETED' }
      ],
      'TESTING': [
        { id: 'COMPLETED', name: '直接完成', targetStatus: 'COMPLETED' },
        { id: 'DEVELOPING', name: '返回开发', targetStatus: 'DEVELOPING' }
      ],
      'COMPLETED': [] // 已完成状态不能再流转
    }
    
    const availableTransitions = statusTransitions[row.status] || []
    workflowTransitions.value = availableTransitions
    
    // 设置所有状态列表(用于流程图显示)
    workflowStatuses.value = [
      { id: 1, statusCode: 'PENDING_CLARIFY', statusName: '待澄清' },
      { id: 2, statusCode: 'PENDING_SCHEDULING', statusName: '待排期' },
      { id: 3, statusCode: 'DEVELOPING', statusName: '开发中' },
      { id: 4, statusCode: 'TESTING', statusName: '测试中' },
      { id: 5, statusCode: 'COMPLETED', statusName: '已完成' }
    ]
    
    if (availableTransitions.length === 0 && row.status !== 'COMPLETED') {
      ElMessage.warning('当前状态无可用的流转操作')
      return
    }
    
    workflowVisible.value = true
  } catch (error) {
    console.error('加载工作流失败:', error)
    ElMessage.error('加载工作流失败')
  }
}

// 执行工作流转换
const executeTransitionClick = async (transition) => {
  try {
    await executeRequirementTransition(currentRequirement.value.id, transition.id, '')
    ElMessage.success('流转成功')
    
    // 关闭弹窗
    workflowVisible.value = false
    
    // 重新加载数据
    await loadData()
  } catch (error) {
    console.error('执行流转失败:', error)
    ElMessage.error(error.message || '流转失败')
  }
}

// 判断是否是可达到的状态
const isReachableStatus = (status) => {
  return workflowTransitions.value.some(t => t.toStatusId === status.id)
}

// 根据状态 ID 获取状态名称
const getStatusNameById = (statusId) => {
  const status = workflowStatuses.value.find(s => s.id === statusId)
  return status ? status.statusName : '未知状态'
}

// 状态顺序映射
const statusOrder = {
  'PENDING_CLARIFY': 1,
  'PENDING_SCHEDULING': 2,
  'DEVELOPING': 3,
  'TESTING': 4,
  'COMPLETED': 5
}

// 计算属性：上一步操作（返回到前一状态）
const previousTransitions = computed(() => {
  if (!currentRequirement.value) return []
  const currentStatus = currentRequirement.value.status
  const currentOrder = statusOrder[currentStatus]
  
  return workflowTransitions.value.filter(t => {
    const targetOrder = statusOrder[t.targetStatus]
    return targetOrder < currentOrder
  })
})

// 计算属性：下一步操作（进入下一状态）
const nextTransitions = computed(() => {
  if (!currentRequirement.value) return []
  const currentStatus = currentRequirement.value.status
  const currentOrder = statusOrder[currentStatus]
  
  return workflowTransitions.value.filter(t => {
    const targetOrder = statusOrder[t.targetStatus]
    return targetOrder > currentOrder && t.targetStatus !== 'COMPLETED'
  })
})

// 计算属性：直接完成操作
const completeTransition = computed(() => {
  if (!currentRequirement.value) return null
  return workflowTransitions.value.find(t => t.targetStatus === 'COMPLETED') || null
})

// 查看详情
const handleView = async (row) => {
  try {
    const res = await getRequirementDetail(row.id)
    if (res.code === 200) {
      currentRequirement.value = res.data
      detailVisible.value = true
      // 加载关联事项
      loadRequirementTasks(row.id)
    }
  } catch (error) {
    console.error('获取详情失败:', error)
    ElMessage.error('获取详情失败')
  }
}

// 加载需求关联的事项列表
const loadRequirementTasks = async (requirementId) => {
  tasksLoading.value = true
  requirementTasks.value = []
  try {
    const res = await getRequirementTasks(requirementId)
    if (res.code === 200) {
      requirementTasks.value = res.data || []
    }
  } catch (error) {
    console.error('加载关联事项失败:', error)
  } finally {
    tasksLoading.value = false
  }
}

// 查看事项详情（弹窗）
const viewTaskDetail = async (row) => {
  taskDetailVisible.value = true
  taskDetailLoading.value = true
  taskDetailData.value = null
  try {
    const res = await getTaskDetail(row.id)
    if (res.code === 200) {
      taskDetailData.value = res.data
    } else {
      ElMessage.error('加载事项详情失败')
    }
  } catch (error) {
    console.error('加载事项详情失败:', error)
    ElMessage.error('加载事项详情失败')
  } finally {
    taskDetailLoading.value = false
  }
}

// 检查 URL 参数，判断是否需要过滤特定需求
const checkUrlParamForFilter = () => {
  const urlParams = new URLSearchParams(window.location.search)
  const requirementId = urlParams.get('requirementId')
  const showDetail = urlParams.get('showDetail')

  if (requirementId) {
    // 清除 URL 参数，避免刷新时重复
    const newUrl = window.location.pathname
    window.history.replaceState({}, '', newUrl)

    if (showDetail === 'true') {
      // 直接弹出需求详情
      handleView({ id: parseInt(requirementId) })
    } else {
      // 设置查询参数为这个需求 ID
      queryParams.id = requirementId
      ElMessage.info(`已定位到指定需求`)
    }
  }
}

// 编辑需求
const handleEdit = (row) => {
  dialogTitle.value = '编辑需求'
  Object.assign(formData, row)
  // 处理 estimatedWorkload 字段，后端返回的是 sql.NullFloat64 对象
  if (row.estimatedWorkload && typeof row.estimatedWorkload === 'object') {
    formData.estimatedWorkload = row.estimatedWorkload.Float64 || null
  } else if (row.estimatedWorkload === null || row.estimatedWorkload === undefined) {
    formData.estimatedWorkload = null
  }
  dialogVisible.value = true
}

// 删除需求
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除这条需求吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteRequirement(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 提交表单
const submitForm = () => {
  formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      // 根据 followerId 获取 followerName
      let followerName = ''
      if (formData.followerId) {
        const selectedUser = userList.value.find(user => user.id === formData.followerId)
        if (selectedUser) {
          followerName = selectedUser.realName || selectedUser.username
        }
      }
      
      // 处理 estimatedWorkload，确保是数字类型或 null
      let estimatedWorkload = null
      if (formData.estimatedWorkload !== null && formData.estimatedWorkload !== undefined && formData.estimatedWorkload !== '') {
        estimatedWorkload = Number(formData.estimatedWorkload)
      }
      
      // 构建提交数据，只包含业务字段
      const submitData = {
        title: formData.title,
        bankName: formData.bankName,
        requirementNo: formData.requirementNo,
        priority: formData.priority,
        riskLevel: formData.riskLevel,
        requirementType: formData.requirementType,
        status: formData.status,
        scheduleDate: formData.scheduleDate,
        mainSystem: formData.mainSystem,
        developer: formData.developer,
        involvedSystems: formData.involvedSystems,
        followerId: formData.followerId,
        followerName: followerName,
        submitterName: formData.submitterName,
        submitTime: formData.submitTime,
        completionDate: formData.completionDate,
        estimatedWorkload: estimatedWorkload,
        groupType: formData.groupType,
        description: formData.description,
        remark: formData.remark
      }
      
      if (formData.id) {
        // 编辑时添加 ID
        submitData.id = formData.id
        await updateRequirement(formData.id, submitData)
        ElMessage.success('更新成功')
      } else {
        await createRequirement(submitData)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadData()
    } catch (error) {
      console.error('提交失败:', error)
      // 显示详细错误信息（如需求编号重复）
      const errorMsg = error.response?.data?.message || error.message || '操作失败'
      ElMessageBox.alert(
        errorMsg,
        '操作失败',
        {
          confirmButtonText: '确定',
          type: 'error'
        }
      )
    } finally {
      submitLoading.value = false
    }
  })
}

// 重置表单
const resetForm = () => {
  Object.assign(formData, {
    id: null,
    title: '',
    bankName: '',
    requirementNo: '',
    priority: '',
    riskLevel: '',
    status: '',
    scheduleDate: '',
    mainSystem: '',
    developer: '',
    involvedSystems: '',
    followerId: null,
    submitterName: '',
    submitTime: '',
    completionDate: '',
    estimatedWorkload: null,
    groupType: '',
    description: '',
    remark: ''
  })
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 加载用户列表
const loadUsers = async () => {
  try {
    const res = await getAllUsers()
    if (res.code === 200 && res.data) {
      // 后端返回的是分页结构，需要提取 list 字段（小写）
      const data = res.data
      userList.value = Array.isArray(data.list) ? data.list : (Array.isArray(data) ? data : [])
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
    userList.value = []
  }
}

// Excel 列定义（导入/导出/模版共用）
const excelColumns = [
  { key: 'title', header: '需求名称' },
  { key: 'bankName', header: '所属银行' },
  { key: 'requirementNo', header: '需求编号' },
  { key: 'priority', header: '优先级' },
  { key: 'riskLevel', header: '风险等级' },
  { key: 'requirementType', header: '需求类型' },
  { key: 'mainSystem', header: '主R系统' },
  { key: 'developer', header: '开发主R|参与者' },
  { key: 'involvedSystems', header: '涉及系统' },
  { key: 'followerName', header: '需求跟进人' },
  { key: 'submitterName', header: '提交人' },
  { key: 'estimatedWorkload', header: '预计工作量(人天)' },
  { key: 'groupType', header: '所属小组' },
  { key: 'description', header: '需求描述' },
  { key: 'remark', header: '备注' }
]

// 导出需求列表
const handleExport = async () => {
  try {
    const response = await exportRequirements(queryParams)

    if (response.code === 200 && response.data) {
      const requirements = response.data.requirements || []

      if (requirements.length === 0) {
        ElMessage.warning('没有可导出的数据')
        return
      }

      // 使用 XLSX 生成 Excel
      const aoa = [excelColumns.map(c => c.header)]
      requirements.forEach(row => {
        aoa.push(excelColumns.map(c => row[c.key] ?? ''))
      })
      const ws = XLSX.utils.aoa_to_sheet(aoa)
      ws['!cols'] = excelColumns.map(() => ({ wch: 18 }))
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, '需求列表')
      XLSX.writeFile(wb, `需求列表_${new Date().toLocaleDateString().replace(/\//g, '-')}.xlsx`)

      ElMessage.success(`成功导出 ${requirements.length} 条数据`)
    } else {
      ElMessage.error('导出数据失败')
    }
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败: ' + (error.message || '未知错误'))
  }
}

// 下载导入模板（带数据字典下拉选项）
const downloadTemplateFile = async () => {
  const ExcelJS = await import('exceljs')
  const workbook = new ExcelJS.Workbook()
  const ws = workbook.addWorksheet('需求导入模板')

  // 表头行
  ws.addRow(excelColumns.map(c => c.header))
  // 示例行
  ws.addRow(['示例需求', '上海银行', 'REQ-2024-001', 'HIGH', 'MEDIUM', 'INTERNAL', 'poseidon', '张三|R&D团队', '系统A;系统B', '李四', '王五', 5.0, 'LOAN_MIDDLE', '这是一个示例需求的描述', '这是备注信息'])
  // 列宽
  ws.columns = excelColumns.map(() => ({ width: 18 }))

  // 字典列定义：colIndex 为 0-based 列索引，dictType 为字典类型
  const dictCols = [
    { colIndex: 1, dictType: 'bank_name' },          // B - 所属银行
    { colIndex: 3, dictType: 'priority' },            // D - 优先级
    { colIndex: 4, dictType: 'risk_level' },          // E - 风险等级
    { colIndex: 5, dictType: 'requirement_type' },     // F - 需求类型
    { colIndex: 12, dictType: 'requirement_group' }   // M - 所属小组
  ]

  // 创建隐藏 sheet 存放字典值
  const dictSheet = workbook.addWorksheet('字典数据', { state: 'hidden' })
  const dictRefs = {}
  dictCols.forEach((dc, i) => {
    const items = dictStore.getDictByType(dc.dictType)
    if (items && items.length > 0) {
      const colLetter = String.fromCharCode(65 + i) // A, B, C, ...
      items.forEach((item, idx) => {
        dictSheet.getCell(`${colLetter}${idx + 1}`).value = item.dictValue
      })
      dictRefs[dc.colIndex] = `字典数据!$${colLetter}$1:$${colLetter}$${items.length}`
    }
  })

  // 为字典列添加下拉验证（第2-1000行）
  dictCols.forEach(dc => {
    const ref = dictRefs[dc.colIndex]
    if (!ref) return
    const colLetter = String.fromCharCode(65 + dc.colIndex)
    for (let row = 2; row <= 1000; row++) {
      ws.getCell(`${colLetter}${row}`).dataValidation = {
        type: 'list',
        allowBlank: true,
        formulae: [ref]
      }
    }
  })

  const buffer = await workbook.xlsx.writeBuffer()
  const blob = new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = '需求导入模板.xlsx'
  link.click()
  URL.revokeObjectURL(url)
}

// 处理文件选择
const handleFileChange = (file) => {
  selectedFile.value = file.raw
}

// 打开导入弹窗（清空上次选择的文件）
const openImportDialog = () => {
  selectedFile.value = null
  if (uploadRef.value) {
    uploadRef.value.clearFiles()
  }
  importDialogVisible.value = true
}

// 执行导入
const handleImport = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请选择要导入的文件')
    return
  }

  importLoading.value = true
  try {
    // 解析 Excel 文件
    const data = await selectedFile.value.arrayBuffer()
    const wb = XLSX.read(data)
    const ws = wb.Sheets[wb.SheetNames[0]]
    const rows = XLSX.utils.sheet_to_json(ws, { header: 1, defval: '' })

    if (rows.length < 2) {
      ElMessage.warning('Excel文件没有数据行')
      return
    }

    // 跳过表头，从第二行开始
    const requirements = []
    for (let i = 1; i < rows.length; i++) {
      const row = rows[i]
      if (!row[0]) continue
      requirements.push({
        title: String(row[0] || '').trim(),
        bankName: String(row[1] || '').trim(),
        requirementNo: String(row[2] || '').trim(),
        priority: String(row[3] || '').trim(),
        riskLevel: String(row[4] || '').trim(),
        requirementType: String(row[5] || '').trim(),
        mainSystem: String(row[6] || '').trim(),
        developer: String(row[7] || '').trim(),
        involvedSystems: String(row[8] || '').trim(),
        followerName: String(row[9] || '').trim(),
        submitterName: String(row[10] || '').trim(),
        estimatedWorkload: parseFloat(row[11]) || 0,
        groupType: String(row[12] || '').trim(),
        description: String(row[13] || '').trim(),
        remark: String(row[14] || '').trim()
      })
    }

    if (requirements.length === 0) {
      ElMessage.warning('没有有效的数据行')
      return
    }

    const res = await batchImport(requirements)

    if (res.code === 200 && res.data) {
      const result = res.data
      let message = `导入完成！成功: ${result.successCount}条, 失败: ${result.failCount}条`
      if (result.failCount > 0 && result.errors && result.errors.length > 0) {
        message += '<br><br><strong>失败详情：</strong><br>' + result.errors.join('<br>')
      }

      ElMessageBox.alert(
        message,
        '导入结果',
        {
          confirmButtonText: '确定',
          type: result.failCount > 0 ? 'warning' : 'success',
          dangerouslyUseHTMLString: true
        }
      ).then(() => {
        importDialogVisible.value = false
        selectedFile.value = null
        uploadRef.value?.clearFiles()
        loadData()
      })
    }
  } catch (error) {
    ElMessage.error('导入失败: ' + (error.message || '未知错误'))
  } finally {
    importLoading.value = false
  }
}

// 状态标签映射
const getStatusLabel = (status) => {
  const map = {
    'PENDING_CLARIFY': '待澄清',
    'PENDING_SCHEDULING': '待排期',
    'DEVELOPING': '开发中',
    'TESTING': '测试中',
    'PENDING_RELEASE': '待投产',
    'COMPLETED': '已完成'
  }
  return map[status] || status
}

const getStatusTag = (status) => {
  const map = {
    'PENDING_CLARIFY': 'info',
    'PENDING_SCHEDULING': 'warning',
    'DEVELOPING': 'primary',
    'TESTING': 'warning',
    'PENDING_RELEASE': 'danger',
    'COMPLETED': 'success'
  }
  return map[status] || 'info'
}

const getGroupLabel = (groupType) => {
  const map = {
    'LOAN_MIDDLE': '贷中',
    'LOAN_PRE': '贷前',
    'RISK': '风险'
  }
  return map[groupType] || groupType
}

// 根据用户ID获取用户名称（用于创建人列展示）
const getCreatorName = (row) => {
  const userId = row.createUser || row.userId
  if (!userId) return '-'
  const user = userList.value.find(u => u.id === userId)
  return user ? (user.realName || user.username) : '-'
}

// 格式化日期，只显示年月日
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  
  // 处理多种日期格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD 或 YYYY-MM-DDTHH:mm:ss
  const date = new Date(dateStr)
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    // 如果不是有效日期，尝试截取字符串
    const match = dateStr.match(/^(\d{4}-\d{2}-\d{2})/)
    return match ? match[1] : dateStr
  }
  
  // 格式化输出 YYYY-MM-DD
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  
  return `${year}-${month}-${day}`
}

// 获取预计工作量显示值
const getEstimatedWorkload = (workload) => {
  if (!workload) return '-'
  
  // 处理后端返回的 sql.NullFloat64 对象
  if (typeof workload === 'object' && workload.Float64 !== undefined) {
    if (workload.Valid) {
      return workload.Float64
    }
    return '-'
  }
  
  // 处理数字类型
  if (typeof workload === 'number') {
    return workload
  }
  
  // 处理字符串类型
  const num = parseFloat(workload)
  if (!isNaN(num)) {
    return num
  }
  
  return '-'
}

// 格式化日期时间，显示年月日时分秒
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  
  // 处理多种日期格式：YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD 或 YYYY-MM-DDTHH:mm:ss
  const date = new Date(dateStr)
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    // 如果不是有效日期，尝试截取字符串
    const match = dateStr.match(/^(\d{4}-\d{2}-\d{2})\s*[T\s]*(\d{2}:\d{2}:\d{2})?/)
    if (match) {
      if (match[2]) {
        return `${match[1]} ${match[2]}`
      }
      return match[1]
    }
    return dateStr
  }
  
  // 格式化输出 YYYY-MM-DD HH:mm:ss
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 判断需求是否已完成（根据当前需求的当前节点是否是流程中最后一个节点来判断）
const isRequirementCompleted = (requirement) => {
  if (!requirement || !requirement.status) {
    return false
  }
  
  // 需要获取工作流状态列表来判断
  // 由于需求管理使用 statusCode，需要从 workflowStatuses 中查找对应的状态
  if (workflowStatuses.value.length === 0) {
    // 如果工作流状态还未加载，返回 false
    return false
  }
  
  const status = workflowStatuses.value.find(s => s.statusCode === requirement.status)
  if (!status) {
    return false
  }
  
  // 使用工作流状态中的 is_final 字段判断是否为最终状态
  // 注意：后端返回的可能是 isFinal（驼峰）或 is_final（下划线）
  const isFinal = status.is_final === 1 || status.is_final === true || status.isFinal === 1 || status.isFinal === true
  return isFinal
}

// 预加载工作流状态（用于判断是否已完成）
const loadWorkflowStatuses = async () => {
  // Go后端简化了状态,直接使用固定的5个状态
  workflowStatuses.value = [
    { id: 1, statusCode: 'PENDING_CLARIFY', statusName: '待澄清' },
    { id: 2, statusCode: 'PENDING_SCHEDULING', statusName: '待排期' },
    { id: 3, statusCode: 'DEVELOPING', statusName: '开发中' },
    { id: 4, statusCode: 'TESTING', statusName: '测试中' },
    { id: 5, statusCode: 'COMPLETED', statusName: '已完成' }
  ]
}

onMounted(() => {
  // 先检查 URL 参数，设置过滤条件
  checkUrlParamForFilter()
  // 预加载工作流状态
  loadWorkflowStatuses()
  // 创建人默认填充当前登录人
  if (userStore.userInfo && userStore.userInfo.id) {
    queryParams.createdBy = userStore.userInfo.id
  }
  // 加载数据
  loadData()
  loadUsers()
  // 加载字典数据
  dictStore.loadAllDicts()
})
</script>

<style scoped>
.requirement-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.filter-bar {
  margin-bottom: 20px;
}

.pagination-bar {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.requirement-detail {
  padding: 20px;
}

.form-item-tip {
  margin-top: 5px;
  font-size: 12px;
  color: #909399;
}

.el-icon--upload {
  font-size: 67px;
  color: #8c939d;
  margin: 40px 0 16px;
}

/* 工作流流转按钮 */
.transition-group {
  margin-bottom: 20px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.transition-group:last-child {
  margin-bottom: 0;
}

.transition-label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  color: #606266;
  font-size: 14px;
  font-weight: 600;
}

.transition-label .el-icon {
  font-size: 18px;
}

.transition-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.transition-buttons .el-button {
  min-width: 180px;
  height: 40px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 6px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  position: relative;
  overflow: hidden;
}

.transition-buttons .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4);
}

.transition-buttons .el-button:active {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
}

.transition-section h4 {
  margin: 0 0 15px 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

/* 工作流流程图样式 */
.workflow-transition {
  padding: 10px;
}

.workflow-diagram {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 20px;
  overflow-x: auto;
}

.status-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  transition: all 0.3s;
}

.status-node.current .status-indicator {
  background-color: #409EFF;
  box-shadow: 0 0 12px rgba(64, 158, 255, 0.5);
}

.status-node.reachable .status-indicator {
  background-color: #67C23A;
  box-shadow: 0 0 12px rgba(103, 194, 58, 0.5);
}

.status-indicator {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #909399;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: bold;
  margin-bottom: 8px;
  transition: all 0.3s;
}

.status-number {
  font-size: 16px;
}

.status-info {
  text-align: center;
}

.status-name {
  font-size: 13px;
  color: #606266;
  white-space: nowrap;
}

.status-arrow {
  font-size: 18px;
  color: #909399;
  margin-top: 4px;
}

.transition-section h4 {
  margin: 0 0 15px 0;
  color: #303133;
  font-size: 14px;
}

/* 需求详情文本内容样式 */
.detail-text-content {
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 4px;
  line-height: 1.8;
  color: #606266;
  font-size: 14px;
  white-space: pre-wrap;
  word-break: break-word;
  min-height: 60px;
  max-height: 400px;
  overflow-y: auto;
}

/* 导入错误提示框样式 */
:deep(.import-error-dialog) {
  .el-message-box__content {
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.8;
    max-height: 400px;
    overflow-y: auto;
  }
  
  .el-message-box__message {
    white-space: pre-wrap;
    word-break: break-word;
  }
}
</style>
