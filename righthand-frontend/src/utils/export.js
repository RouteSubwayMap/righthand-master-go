/**
 * 数据导出工具
 */

/**
 * 将JSON数据导出为CSV文件
 * @param {Array} data - 数据数组
 * @param {String} filename - 文件名(不含扩展名)
 * @param {Array} headers - 表头配置 [{key: '字段名', title: '显示名称'}]
 */
export function exportToCSV(data, filename = 'export', headers = []) {
  if (!data || data.length === 0) {
    throw new Error('没有可导出的数据')
  }

  // 如果没有提供headers,自动从第一条数据生成
  if (headers.length === 0) {
    const firstRow = data[0]
    headers = Object.keys(firstRow).map(key => ({
      key,
      title: key
    }))
  }

  // 构建CSV内容
  const csvRows = []
  
  // 添加表头
  csvRows.push(headers.map(h => h.title).join(','))
  
  // 添加数据行
  data.forEach(row => {
    const values = headers.map(header => {
      let value = row[header.key]
      
      // 处理null/undefined
      if (value === null || value === undefined) {
        value = ''
      }
      
      // 处理日期对象
      if (value instanceof Date) {
        value = value.toISOString().slice(0, 19).replace('T', ' ')
      }
      
      // 转换为字符串并处理特殊字符
      value = String(value)
      
      // 如果包含逗号、引号或换行符,需要用引号包裹
      if (value.includes(',') || value.includes('"') || value.includes('\n')) {
        value = '"' + value.replace(/"/g, '""') + '"'
      }
      
      return value
    })
    
    csvRows.push(values.join(','))
  })
  
  // 添加BOM以支持中文
  const BOM = '\uFEFF'
  const csvContent = BOM + csvRows.join('\n')
  
  // 创建Blob并下载
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  
  link.setAttribute('href', url)
  link.setAttribute('download', `${filename}.csv`)
  link.style.visibility = 'hidden'
  
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  URL.revokeObjectURL(url)
}

/**
 * 将需求数据导出为CSV
 * @param {Array} requirements - 需求数据数组
 * @param {String} filename - 文件名
 */
export function exportRequirementsToCSV(requirements, filename = '需求列表') {
  const headers = [
    { key: 'requirementNo', title: '需求编号' },
    { key: 'title', title: '需求名称' },
    { key: 'bankName', title: '所属银行' },
    { key: 'requirementType', title: '需求类型' },
    { key: 'status', title: '状态' },
    { key: 'priority', title: '优先级' },
    { key: 'riskLevel', title: '风险等级' },
    { key: 'groupType', title: '所属小组' },
    { key: 'mainSystem', title: '主R系统' },
    { key: 'developer', title: '开发人员' },
    { key: 'submitterName', title: '提交人' },
    { key: 'followerName', title: '跟进人' },
    { key: 'submitTime', title: '提交时间' },
    { key: 'completionDate', title: '完成时间' },
    { key: 'estimatedWorkload', title: '预计工作量' },
    { key: 'description', title: '需求描述' },
    { key: 'remark', title: '备注' }
  ]
  
  // 转换数据格式
  const exportData = requirements.map(req => ({
    ...req,
    submitTime: req.submitTime ? new Date(req.submitTime) : '',
    completionDate: req.completionDate ? new Date(req.completionDate) : '',
    estimatedWorkload: req.estimatedWorkload || ''
  }))
  
  exportToCSV(exportData, filename, headers)
}
