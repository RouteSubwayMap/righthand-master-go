import request from '@/utils/request'

export const getProdNodeList = (params) => {
  return request({
    url: '/api/prod-nodes',
    method: 'get',
    params: {
      pageNum: params.pageNum,
      pageSize: params.pageSize,
      bankType: params.bankType || '',
      bankName: params.bankName || '',
      envType: params.envType || '',
      nodeName: params.nodeName || '',
      nodeIP: params.nodeIP || '',
      username: params.username || ''
    }
  })
}

export const getProdNodeDetail = (id) => {
  return request({
    url: `/api/prod-nodes/${id}`,
    method: 'get'
  })
}

export const createProdNode = (data) => {
  return request({
    url: '/api/prod-nodes',
    method: 'post',
    data
  })
}

export const updateProdNode = (id, data) => {
  return request({
    url: `/api/prod-nodes/${id}`,
    method: 'put',
    data
  })
}

export const deleteProdNode = (id) => {
  return request({
    url: `/api/prod-nodes/${id}`,
    method: 'delete'
  })
}

export const importProdNodes = (nodes) => {
  return request({
    url: '/api/prod-nodes/import',
    method: 'post',
    data: { nodes }
  })
}
