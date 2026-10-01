import request from '@/utils/request'

// 获取联系人列表（分页）
export function getContactList(params) {
  return request({
    url: '/api/contacts',
    method: 'get',
    params: {
      pageNum: params.pageNum,
      pageSize: params.pageSize,
      keyword: params.name || '',
      group: params.contactGroup || '',
      position: params.position || ''
    }
  })
}

// 获取联系人详情
export function getContactDetail(id) {
  return request({
    url: `/api/contacts/${id}`,
    method: 'get'
  })
}

// 创建联系人
export function createContact(data) {
  return request({
    url: '/api/contacts',
    method: 'post',
    data
  })
}

// 更新联系人
export function updateContact(id, data) {
  return request({
    url: `/api/contacts/${id}`,
    method: 'put',
    data
  })
}

// 删除联系人
export function deleteContact(id) {
  return request({
    url: `/api/contacts/${id}`,
    method: 'delete'
  })
}

// 获取联系人分组列表
export function getContactGroups() {
  return request({
    url: '/api/contacts/groups',
    method: 'get'
  })
}
