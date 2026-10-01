import request from '@/utils/request'

/* ================================ 项目 ================================ */

/**
 * 分页查询项目列表
 * @param {Object} params - { pageNum, pageSize, status(逗号分隔多选), riskLevel, managerId,
 *                            participantId, startDate, endDate, keyword, sortBy, order }
 */
export function getProjectList(params) {
  return request({
    url: '/api/projects',
    method: 'get',
    params
  })
}

/**
 * 新建项目
 * @param {Object} data - { projectName, projectKey, startDate, endDate, managerId, planDesc, description }
 */
export function createProject(data) {
  return request({
    url: '/api/projects',
    method: 'post',
    data
  })
}

/**
 * 获取项目详情（含周计划、成员、变更日志、allowedActions）
 */
export function getProjectDetail(id) {
  return request({
    url: `/api/projects/${id}`,
    method: 'get'
  })
}

/**
 * 更新项目基本信息
 */
export function updateProject(id, data) {
  return request({
    url: `/api/projects/${id}`,
    method: 'put',
    data
  })
}

/**
 * 删除项目（仅 STARTING 状态允许）
 */
export function deleteProject(id) {
  return request({
    url: `/api/projects/${id}`,
    method: 'delete'
  })
}

/**
 * 项目状态流转
 * @param {number} id
 * @param {Object} data - { action: start|pause|resume|close|terminate, reason? }
 */
export function changeProjectStatus(id, data) {
  return request({
    url: `/api/projects/${id}/status`,
    method: 'put',
    data
  })
}

/**
 * 修改项目风险等级/交付状态（必填原因）
 * @param {Object} data - { riskLevel?, deliveryStatus?, riskNote?, reason }
 */
export function changeProjectRisk(id, data) {
  return request({
    url: `/api/projects/${id}/risk`,
    method: 'put',
    data
  })
}

/* ================================ 成员 ================================ */

/**
 * 获取项目成员列表
 */
export function getProjectMembers(id) {
  return request({
    url: `/api/projects/${id}/members`,
    method: 'get'
  })
}

/**
 * 添加项目成员
 * @param {Object} data - { userId, memberRole: PM|MEMBER }
 */
export function addProjectMember(id, data) {
  return request({
    url: `/api/projects/${id}/members`,
    method: 'post',
    data
  })
}

/**
 * 移除项目成员
 */
export function removeProjectMember(id, memberId) {
  return request({
    url: `/api/projects/${id}/members/${memberId}`,
    method: 'delete'
  })
}

/* ================================ 项目阶段 ================================ */

/**
 * 获取项目阶段列表（按开始时间正序）
 */
export function getProjectStages(id) {
  return request({
    url: `/api/projects/${id}/stages`,
    method: 'get'
  })
}

/**
 * 新增项目阶段
 * @param {Object} data - { stageName, startDate, endDate, description }
 */
export function createProjectStage(id, data) {
  return request({
    url: `/api/projects/${id}/stages`,
    method: 'post',
    data
  })
}

/**
 * 编辑项目阶段（字段均可选）
 */
export function updateProjectStage(id, stageId, data) {
  return request({
    url: `/api/projects/${id}/stages/${stageId}`,
    method: 'put',
    data
  })
}

/**
 * 删除项目阶段（软删除）
 */
export function deleteProjectStage(id, stageId) {
  return request({
    url: `/api/projects/${id}/stages/${stageId}`,
    method: 'delete'
  })
}

/* ================================ 周计划 ================================ */

/**
 * 新增周计划（week_no 由服务端自增）
 * @param {Object} data - { title, planDesc, startDate, endDate, ownerId, remark }
 */
export function createWeekPlan(id, data) {
  return request({
    url: `/api/projects/${id}/weeks`,
    method: 'post',
    data
  })
}

/**
 * 编辑周计划
 */
export function updateWeekPlan(id, weekId, data) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}`,
    method: 'put',
    data
  })
}

/**
 * 删除周计划
 */
export function deleteWeekPlan(id, weekId) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}`,
    method: 'delete'
  })
}

/**
 * 周状态流转
 * @param {Object} data - { action: start|done|skip|reopen }
 */
export function changeWeekStatus(id, weekId, data) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}/status`,
    method: 'put',
    data
  })
}

/**
 * 提交周总结（提交即锁定）
 * @param {Object} data - { summary: { finished, nextPlan, risks, support } }
 */
export function submitWeekSummary(id, weekId, data) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}/summary`,
    method: 'post',
    data
  })
}

/**
 * 解锁周总结（仅 PM/ADMIN）
 */
export function unlockWeekSummary(id, weekId) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}/summary/unlock`,
    method: 'put'
  })
}

/* ============================ 任务-周关联 ============================ */

/**
 * 查询项目任务
 * @param {number} id - 项目ID
 * @param {Object} params - { weekPlanId?, pageNum, pageSize }
 *   不传 weekPlanId 为「全部」视图：按任务去重并返回所属周列表（weekNos）
 */
export function getProjectTasks(id, params) {
  return request({
    url: `/api/projects/${id}/tasks`,
    method: 'get',
    params
  })
}

/**
 * 在指定周内新建任务（创建事项 + 建立周关联）
 * @param {Object} data - { title, description, priority, deadline, userId }
 */
export function createTaskInWeek(id, weekId, data) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}/tasks`,
    method: 'post',
    data
  })
}

/**
 * 将已有事项加入指定周（幂等）
 */
export function linkTaskToWeek(id, weekId, taskId) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}/tasks/${taskId}`,
    method: 'post'
  })
}

/**
 * 将事项从指定周移出
 */
export function unlinkTaskFromWeek(id, weekId, taskId) {
  return request({
    url: `/api/projects/${id}/weeks/${weekId}/tasks/${taskId}`,
    method: 'delete'
  })
}

/**
 * 将任务移动到项目内的其他周
 * @param {Object} data - { weekPlanId }
 */
export function moveTaskToWeek(id, taskId, data) {
  return request({
    url: `/api/projects/${id}/tasks/${taskId}/week`,
    method: 'put',
    data
  })
}
