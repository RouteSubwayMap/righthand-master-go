import request from '@/utils/request'

/**
 * 获取未读消息数量
 * 返回 { code, data: { count } }
 */
export function getUnreadCount() {
  return request({
    url: '/api/messages/unread-count',
    method: 'get'
  })
}

/**
 * 分页查询当前用户的消息列表
 * @param {Object} params - { pageNum, pageSize, isRead }
 */
export function getMessages(params) {
  return request({
    url: '/api/messages',
    method: 'get',
    params
  })
}

/**
 * 标记单条消息已读
 */
export function markMessageRead(id) {
  return request({
    url: `/api/messages/${id}/read`,
    method: 'put'
  })
}

/**
 * 标记全部消息已读
 */
export function markAllMessagesRead() {
  return request({
    url: '/api/messages/read-all',
    method: 'put'
  })
}
