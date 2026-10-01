import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 存储待处理的请求控制器，用于组件卸载时取消请求
const pendingRequests = new Map()

/**
 * 生成请求唯一标识
 */
function generateRequestKey(config) {
  return `${config.method}:${config.url}:${JSON.stringify(config.params || {})}:${JSON.stringify(config.data || {})}`
}

/**
 * 将 Date 对象转换为 yyyy-MM-dd HH:mm:ss 格式
 */
function formatDateToDateTimeString(date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

/**
 * 标准化日期时间字符串为 yyyy-MM-dd HH:mm:ss 格式
 */
function normalizeDateTimeString(str) {
  try {
    // 如果是 ISO 格式（包含 T），转换为空格分隔
    if (str.includes('T')) {
      return str.replace('T', ' ').substring(0, 19)
    }
    // 如果已经包含时分秒，截取前 19 位
    return str.substring(0, 19)
  } catch (e) {
    return str
  }
}

const request = axios.create({
  baseURL: '',  // 空基础路径，由各 API 文件自己控制完整路径
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('righthand_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    // 创建 AbortController 用于请求取消
    const controller = new AbortController()
    config.signal = controller.signal
    
    // 存储控制器以便后续取消
    const requestKey = generateRequestKey(config)
    // 如果存在相同的未完成的请求，先取消它
    if (pendingRequests.has(requestKey)) {
      pendingRequests.get(requestKey).abort()
      pendingRequests.delete(requestKey)
    }
    pendingRequests.set(requestKey, controller)
    
    // 处理请求数据中的日期时间格式
    if (config.data && typeof config.data === 'object') {
      // 将 Date 对象或日期字符串转换为 yyyy-MM-dd HH:mm:ss 格式
      Object.keys(config.data).forEach(key => {
        const value = config.data[key]
        if (value instanceof Date) {
          // Date 对象转字符串
          config.data[key] = formatDateToDateTimeString(value)
        } else if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}/.test(value)) {
          // 已经是日期时间字符串，确保格式正确
          config.data[key] = normalizeDateTimeString(value)
        }
      })
    }
    
    return config
  },
  error => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    // 清理已完成的请求
    const requestKey = generateRequestKey(response.config)
    pendingRequests.delete(requestKey)
    
    const res = response.data
    
    // 如果是文件下载，直接返回
    if (response.config.responseType === 'blob') {
      return response
    }
    
    if (res.code !== 200 && res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      
      // 401: 未授权或 token 过期
      if (res.code === 401) {
        localStorage.removeItem('righthand_token')
        localStorage.removeItem('righthand_user_info')
        router.push('/login')
      }
      
      return Promise.reject(new Error(res.message || 'Error'))
    }
    
    return res
  },
  error => {
    // 清理失败的请求
    if (error.config) {
      const requestKey = generateRequestKey(error.config)
      pendingRequests.delete(requestKey)
    }
    
    // 如果是请求被取消，静默处理，不显示错误提示，也不抛出异常
    if (axios.isCancel(error) || error?.name === 'CanceledError' || error?.message === 'canceled') {
      // 静默忽略取消错误，返回一个已解决的 Promise
      return Promise.resolve({ code: 200, data: null, message: 'Request canceled' })
    }
    
    console.error('Response error:', error)
    
    if (error.response) {
      switch (error.response.status) {
        case 401:
          ElMessage.error('未授权，请重新登录')
          localStorage.removeItem('righthand_token')
          localStorage.removeItem('righthand_user_info')
          router.push('/login')
          break
        case 403:
          ElMessage.error('拒绝访问')
          break
        case 404:
          ElMessage.error('请求地址不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(error.message || '请求失败')
      }
    } else {
      // 网络错误（排除取消错误）
      ElMessage.error('网络异常，请检查网络连接')
    }
    
    return Promise.reject(error)
  }
)

/**
 * 取消指定请求
 * @param {Object} config - axios 配置对象
 */
export function cancelRequest(config) {
  const requestKey = generateRequestKey(config)
  if (pendingRequests.has(requestKey)) {
    pendingRequests.get(requestKey).abort()
    pendingRequests.delete(requestKey)
  }
}

/**
 * 取消所有待处理的请求
 */
export function cancelAllRequests() {
  pendingRequests.forEach((controller) => {
    controller.abort()
  })
  pendingRequests.clear()
}

/**
 * 判断是否为请求取消错误
 * @param {Error} error - 错误对象
 * @returns {boolean}
 */
export function isCancelError(error) {
  return axios.isCancel(error) || 
         error?.name === 'CanceledError' || 
         error?.message === 'canceled'
}

export default request
