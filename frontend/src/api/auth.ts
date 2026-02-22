import { apiClient } from './client'
import type { LoginResponse, UserInfoResponse, ApiResponse } from './types'

export const authApi = {
  async login(data: { email: string; password: string; captchaToken?: string }) {
    const response = await apiClient.post<LoginResponse>('/auth/login', data)
    return response.data.data || response.data
  },

  async getUserInfo() {
    // 注意：/api/userinfo 端点直接返回用户信息对象，不使用 ApiResponse 包装
    // 这符合 OIDC UserInfo 端点标准
    const response = await apiClient.get<UserInfoResponse>('/userinfo')
    return response.data
  },

  async register(data: { email: string; password: string; username?: string; verificationCode: string }) {
    const response = await apiClient.post<ApiResponse<void>>('/auth/register', data)
    return response.data
  },

  async sendVerificationCode(email: string, purpose: 'register' | 'reset_password', captchaToken: string) {
    const response = await apiClient.post<ApiResponse<{ message: string; code?: string }>>('/auth/send-code', {
      email,
      purpose,
      captchaToken
    })
    return response.data
  },

  async resetPassword(data: { email: string; verificationCode: string; newPassword: string }) {
    const response = await apiClient.post<ApiResponse<{ message: string }>>('/auth/reset-password', data)
    return response.data
  },

  async refreshToken(refreshToken: string) {
    const response = await apiClient.post<LoginResponse>('/oauth/token', {
      grant_type: 'refresh_token',
      refresh_token: refreshToken
    })
    return response.data
  },

  async verifyRealName(data: { name: string; idcard: string }) {
    const response = await apiClient.post<ApiResponse<{ message: string; order_no: string }>>('/realname/verify', data)
    return response.data
  },

  async submitRealName(data: { userId: string | number; name: string; idcard: string }) {
    const response = await apiClient.post<ApiResponse<{ message: string; order_no: string }>>('/realname/submit', data)
    return response.data
  },

  async getRealNameInfo() {
    const response = await apiClient.get<ApiResponse<{ isRealName: boolean; name?: string; idcard?: string }>>('/realname/verify')
    return response.data
  },

  async updateProfile(data: { userId: string | number; username: string; qq: string; avatar: string }) {
    const response = await apiClient.post<ApiResponse<{ message: string; user: any }>>('/auth/update-profile', data)
    return response.data
  },

  async getApplicationInfo(clientId: string) {
    const response = await apiClient.get<ApiResponse<{
      name: string
      displayName: string
      logo: string
      description: string
      homepageUrl: string
      organization: string
    }>>(`/auth/application-info?client_id=${clientId}`)
    return response.data
  }
}
