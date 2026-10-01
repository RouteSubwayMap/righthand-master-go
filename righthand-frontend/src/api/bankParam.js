import request from '@/utils/request'

// 获取银行机构参数列表（分页）
export function getBankParamList(params) {
  return request({
    url: '/api/bank-params',
    method: 'get',
    params: {
      pageNum: params.pageNum,
      pageSize: params.pageSize,
      bankType: params.bankType || '',
      bankName: params.bankName || '',
      bankCode: params.bankCode || ''
    }
  })
}

// 获取银行机构参数详情
export function getBankParam(id) {
  return request({
    url: `/api/bank-params/${id}`,
    method: 'get'
  })
}

// 创建银行机构参数
export function createBankParam(data) {
  return request({
    url: '/api/bank-params',
    method: 'post',
    data
  })
}

// 更新银行机构参数
export function updateBankParam(id, data) {
  return request({
    url: `/api/bank-params/${id}`,
    method: 'put',
    data
  })
}

// 删除银行机构参数
export function deleteBankParam(id) {
  return request({
    url: `/api/bank-params/${id}`,
    method: 'delete'
  })
}

// 批量导入银行机构参数
export function importBankParams(data) {
  return request({
    url: '/api/bank-params/import',
    method: 'post',
    data: { params: data }
  })
}
