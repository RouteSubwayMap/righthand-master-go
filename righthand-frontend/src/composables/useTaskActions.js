import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { deleteTask, getWorkflowByType, executeTransition } from '@/api/task'

/**
 * 任务操作公共逻辑
 * 提供任务相关的通用操作方法，供多个组件复用
 */
export function useTaskActions(refreshCallback, options = {}) {
  const router = useRouter()
  
  // 从 options 中获取组件特有的引用
  const { 
    currentTask, 
    workflowVisible, 
    workflowTransitions, 
    workflowStatuses,
    detailVisible,
    dialogTitle,
    dialogVisible,
    formData
  } = options

  /**
   * 查看任务详情
   * @param {Object} row - 任务数据对象
   */
  const handleView = (row) => {
    // 这个方法需要各个组件自己实现，因为详情弹窗的展示方式不同
    console.log('查看任务:', row)
  }

  /**
   * 编辑任务
   * @param {Object} row - 任务数据对象  
   */
  const handleEdit = (row) => {
    if (dialogTitle) {
      dialogTitle.value = '编辑任务'
    }
    // 拷贝所有字段，确保 dueDate 和 statusId 正确赋值
    Object.assign(formData, {
      id: row.id,
      title: row.title,
      taskType: row.taskType,
      priority: row.priority,
      assigneeId: row.assigneeId,
      statusId: row.statusId,  // 保持当前状态，避免流转异常
      dueDate: row.dueDate,
      description: row.description
    })
    if (dialogVisible) {
      dialogVisible.value = true
    }
  }

  /**
   * 删除任务
   * @param {Object} row - 任务数据对象
   */
  const handleDelete = async (row) => {
    try {
      await ElMessageBox.confirm(
        `确定要删除任务 "${row.title}" 吗？`,
        '提示',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
      
      await deleteTask(row.id)
      ElMessage.success('删除成功')
      
      // 调用刷新回调
      if (refreshCallback) {
        await refreshCallback()
      }
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除失败:', error)
        ElMessage.error('删除失败')
      }
    }
  }

  /**
   * 工作流流转
   * @param {Object} row - 任务数据对象
   */
  const handleWorkflow = async (row) => {
    // 如果提供了 currentTask，则使用复杂的工作流弹窗逻辑（TaskList.vue）
    if (currentTask && workflowVisible && workflowTransitions && workflowStatuses) {
      currentTask.value = row
      try {
        // 直接调用API获取工作流数据
        const res = await getWorkflowByType(row.taskType)
        
        if (res.code === 200 && res.data) {
          const workflow = res.data
          // 保存所有状态列表（用于流程图显示）
          workflowStatuses.value = workflow.statuses || []
          
          // 过滤出从当前状态可以执行的转换
          const transitions = workflow.transitions?.filter(
            t => t.fromStatusId === row.statusId
          ) || []
          workflowTransitions.value = transitions
          
          if (workflowStatuses.value.length === 0) {
            ElMessage.warning('未找到工作流状态配置')
            return
          }
          
          if (transitions.length === 0) {
            ElMessage.warning('当前状态没有可执行的操作')
          }
          
          workflowVisible.value = true
        } else {
          ElMessage.error('获取工作流配置失败')
        }
      } catch (error) {
        console.error('获取流转操作失败:', error)
        ElMessage.error('获取可执行操作失败：' + (error.message || '未知错误'))
      }
    } else {
      // 否则使用简单的路由跳转逻辑（TaskBoard.vue）
      router.push(`/task/workflow?taskId=${row.id}`)
    }
  }

  /**
   * 执行工作流转换（仅 TaskList.vue 需要）
   */
  const executeTransitionClick = async (transition, taskId, comment) => {
    try {
      await executeTransition(taskId, transition.id, comment)
      ElMessage.success('流转成功')
      
      // 关闭弹窗
      if (workflowVisible) {
        workflowVisible.value = false
      }
      
      // 重新加载数据
      if (refreshCallback) {
        await refreshCallback()
      }
      
      // 刷新详情弹窗中的数据
      if (detailVisible && currentTask) {
        await handleView(currentTask.value)
      }
    } catch (error) {
      console.error('执行流转失败:', error)
      ElMessage.error(error.message || '流转失败')
    }
  }

  return {
    handleView,
    handleEdit,
    handleDelete,
    handleWorkflow,
    executeTransitionClick
  }
}
