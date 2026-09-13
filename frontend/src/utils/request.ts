import axios from 'axios'
import { ElMessage } from 'element-plus'
import { ErrorCode } from '@/constants/errorCodes'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) {
        return body.data
      }
      ElMessage.error(body.message || '请求失败')
      if (body.code === ErrorCode.UNAUTHORIZED) {
        localStorage.removeItem('token')
        if (!location.pathname.startsWith('/login')) {
          location.href = '/login'
        }
      }
      return Promise.reject(new Error(body.message || '请求失败'))
    }
    return body
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '网络错误'
    ElMessage.error(message)
    if (status === 401) {
      localStorage.removeItem('token')
      if (!location.pathname.startsWith('/login')) {
        location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export default request
