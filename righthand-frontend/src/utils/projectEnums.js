/**
 * 项目管理 - 枚举映射工具（仅用于展示：中文标签与颜色）
 *
 * 说明：操作能力（哪些按钮可用）完全由后端 allowedActions 决定，
 *      本文件不复制任何状态机逻辑，只负责「值 → 文案/颜色」的展示映射。
 */

/** 项目状态 */
export const PROJECT_STATUS = {
  STARTING: { label: '启动', tagType: 'info', color: '#909399' },
  IN_PROGRESS: { label: '进行中', tagType: 'primary', color: '#409eff' },
  PAUSED: { label: '已暂停', tagType: 'warning', color: '#e6a23c' },
  CLOSED: { label: '已结项', tagType: 'success', color: '#67c23a' },
  TERMINATED: { label: '已终止', tagType: 'danger', color: '#f56c6c' }
}

/** 项目状态下拉选项（筛选器用） */
export const PROJECT_STATUS_OPTIONS = Object.entries(PROJECT_STATUS).map(([value, v]) => ({
  value,
  label: v.label
}))

/** 风险等级（与 requirement 模块同字典口径） */
export const RISK_LEVEL = {
  HIGH: { label: '高', tagType: 'danger', color: '#f56c6c' },
  MEDIUM: { label: '中', tagType: 'warning', color: '#e6a23c' },
  LOW: { label: '低', tagType: 'success', color: '#67c23a' }
}

export const RISK_LEVEL_OPTIONS = Object.entries(RISK_LEVEL).map(([value, v]) => ({
  value,
  label: v.label
}))

/** 交付状态 */
export const DELIVERY_STATUS = {
  NORMAL: { label: '正常', tagType: 'success', color: '#67c23a' },
  AT_RISK: { label: '有延期风险', tagType: 'warning', color: '#e6a23c' },
  CANNOT_FINISH: { label: '预计无法完成', tagType: 'danger', color: '#f56c6c' }
}

export const DELIVERY_STATUS_OPTIONS = Object.entries(DELIVERY_STATUS).map(([value, v]) => ({
  value,
  label: v.label
}))

/** 周计划状态 */
export const WEEK_STATUS = {
  NOT_STARTED: { label: '未开始', tagType: 'info', color: '#909399' },
  IN_PROGRESS: { label: '进行中', tagType: 'primary', color: '#409eff' },
  DONE: { label: '已完成', tagType: 'success', color: '#67c23a' },
  SKIPPED: { label: '已跳过', tagType: 'info', color: '#b1b3b8' }
}

/** 事项状态（项目任务 Tab 展示用，与事项模块口径一致） */
export const TASK_STATUS = {
  TODO: { label: '待办', tagType: 'info', color: '#909399' },
  DOING: { label: '进行中', tagType: 'primary', color: '#409eff' },
  DONE: { label: '已完成', tagType: 'success', color: '#67c23a' },
  CANCELLED: { label: '已取消', tagType: 'info', color: '#b1b3b8' }
}

/** 事项优先级 */
export const TASK_PRIORITY = {
  HIGH: { label: '高', tagType: 'danger', color: '#f56c6c' },
  MEDIUM: { label: '中', tagType: 'warning', color: '#e6a23c' },
  LOW: { label: '低', tagType: 'success', color: '#67c23a' }
}

/** 事项类型（与事项模块口径一致） */
export const TASK_ITEM_TYPE = {
  TASK: { label: '任务', tagType: 'primary' },
  MEMO: { label: '备忘', tagType: 'success' },
  REQUIREMENT: { label: '需求', tagType: 'warning' }
}

/** 项目级操作（allowedActions 值 → 按钮文案） */
export const PROJECT_ACTION_LABELS = {
  start: '开始',
  pause: '暂停',
  resume: '恢复',
  close: '结项',
  terminate: '终止',
  edit: '编辑',
  delete: '删除',
  editRisk: '修改风险',
  manageMember: '成员管理',
  addWeek: '新增周计划'
}

/** 周计划级操作（allowedActions 值 → 按钮文案） */
export const WEEK_ACTION_LABELS = {
  edit: '编辑',
  delete: '删除',
  changeStatus: '状态流转',
  submitSummary: '填写总结',
  unlockSummary: '解锁总结'
}

/** 周状态流转动作（action → 文案，用于下拉菜单） */
export const WEEK_STATUS_ACTION_LABELS = {
  start: '开始',
  done: '标记完成',
  skip: '跳过本周',
  reopen: '重新打开'
}

/** 变更历史 action 枚举 → 中文标签 */
export const CHANGE_LOG_ACTION_LABELS = {
  CREATE: '创建项目',
  UPDATE: '编辑项目',
  STATUS: '状态流转',
  RISK_CHANGE: '风险变更',
  MEMBER_ADD: '添加成员',
  MEMBER_REMOVE: '移除成员',
  WEEK_ADD: '新增周计划',
  WEEK_EDIT: '编辑周计划',
  WEEK_DELETE: '删除周计划',
  WEEK_STATUS: '周状态变更',
  SUMMARY_SUBMIT: '提交周总结',
  SUMMARY_UNLOCK: '解锁周总结',
  STAGE_ADD: '新增阶段',
  STAGE_EDIT: '编辑阶段',
  STAGE_DELETE: '删除阶段'
}

/** 汇总各项动作与状态的展示文案（按钮渲染用） */
export const ALL_ACTION_LABELS = {
  ...PROJECT_ACTION_LABELS,
  ...WEEK_ACTION_LABELS
}

export function projectStatusLabel(status) {
  return PROJECT_STATUS[status]?.label || status || '-'
}

export function riskLevelLabel(level) {
  return RISK_LEVEL[level]?.label || level || '-'
}

export function deliveryStatusLabel(status) {
  return DELIVERY_STATUS[status]?.label || status || '-'
}

export function weekStatusLabel(status) {
  return WEEK_STATUS[status]?.label || status || '-'
}

export function taskStatusLabel(status) {
  return TASK_STATUS[status]?.label || status || '-'
}

export function taskPriorityLabel(priority) {
  return TASK_PRIORITY[priority]?.label || priority || '-'
}

export function changeLogActionLabel(action) {
  return CHANGE_LOG_ACTION_LABELS[action] || action || '-'
}

export function actionLabel(action) {
  return ALL_ACTION_LABELS[action] || action
}

/** 周总结 JSON 归一化：兼容字符串（DB TEXT）与对象两种形态 */
export function parseSummary(summary) {
  if (!summary) return null
  if (typeof summary === 'object') return summary
  try {
    return JSON.parse(summary)
  } catch {
    return null
  }
}
