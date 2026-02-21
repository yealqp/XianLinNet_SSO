import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import { storage } from '../utils/storage'
import type { ApiResponse, ApiError } from './types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

export class ApiClient {
  private instance: AxiosInstance

  constructor() {
    this.instance = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json'
      }
    })

    this.setupInterceptors()
  }

  private setupInterceptors(): void {
    this.instance.interceptors.request.use(
      (config) => {
        const token = storage.getAccessToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    this.instance.interceptors.response.use(
      (response: AxiosResponse<ApiResponse<unknown>>) => {
        // 检查后端返回的 status 字段
        const data = response.data as ApiResponse<unknown>
        if (data && data.status === 'error') {
          // 后端返回错误，抛出异常
          const error = new Error(data.msg || 'Request failed') as any
          error.response = response
          error.isBusinessError = true
          return Promise.reject(error)
        }
        return response
      },
      (error: AxiosError<ApiResponse<ApiError>>) => {
        if (error.response) {
          const { status, data } = error.response
          
          // 如果后端返回了错误消息，使用它
          const errorMessage = data?.msg || 'Request failed'
          
          switch (status) {
            case 401:
              storage.clear()
              window.location.href = '/login'
              break
            case 403:
              console.error('Access forbidden:', errorMessage)
              break
            case 400:
              console.error('Bad request:', errorMessage)
              break
            case 500:
              console.error('Server error:', errorMessage)
              break
          }
          
          // 创建一个包含错误消息的错误对象
          const enhancedError = new Error(errorMessage) as any
          enhancedError.response = error.response
          enhancedError.originalError = error
          return Promise.reject(enhancedError)
        }
        return Promise.reject(error)
      }
    )
  }

  public async get<T>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return this.instance.get(url, config)
  }

  public async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> {
    return this.instance.post(url, data, config)
  }
}

export const apiClient = new ApiClient()
