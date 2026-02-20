import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { storage } from '@/utils/storage'

interface LoginRequest {
  email: string
  password: string
  captchaToken?: string
}

interface UserInfo {
  id: string
  email: string
  username: string
  isAdmin: boolean
  isRealName?: boolean
  qq?: string
  avatar?: string
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(storage.getAccessToken())
  const refreshToken = ref<string | null>(storage.getRefreshToken())
  const userInfo = ref<UserInfo | null>(storage.getUserInfo() as UserInfo | null)

  const isAuthenticated = computed(() => !!accessToken.value)
  const isAdmin = computed(() => {
    if (!userInfo.value) return false
    return userInfo.value.isAdmin || false
  })

  async function login(data: LoginRequest) {
    try {
      const loginData = await authApi.login(data)
      
      // 检查是否有 access_token
      if (!loginData.access_token) {
        throw new Error('登录失败：未收到访问令牌')
      }
      
      accessToken.value = loginData.access_token
      storage.setAccessToken(loginData.access_token)
      
      if (loginData.refresh_token) {
        refreshToken.value = loginData.refresh_token
        storage.setRefreshToken(loginData.refresh_token)
      }
      
      if (loginData.user) {
        userInfo.value = loginData.user
        storage.setUserInfo(userInfo.value)
      }
      
      return loginData
    } catch (error: any) {
      // 清理可能的残留数据
      accessToken.value = null
      refreshToken.value = null
      userInfo.value = null
      storage.clear()
      
      // 重新抛出错误，让调用者处理
      throw error
    }
  }

  async function logout() {
    accessToken.value = null
    refreshToken.value = null
    userInfo.value = null
    storage.clear()
  }

  async function fetchUserInfo() {
    try {
      const response = await authApi.getUserInfo()
      userInfo.value = {
        id: response.sub || '',
        email: response.email || '',
        username: response.username || response.name || response.preferred_username || '',
        isAdmin: false, // Will be determined by roles/permissions
        isRealName: response.is_real_name || false,
        qq: response.qq || '',
        avatar: response.picture || ''
      }
      storage.setUserInfo(userInfo.value)
      return userInfo.value
    } catch (error) {
      console.error('Failed to fetch user info:', error)
      throw error
    }
  }

  function setTokens(accessTokenValue: string, refreshTokenValue?: string) {
    accessToken.value = accessTokenValue
    storage.setAccessToken(accessTokenValue)
    
    if (refreshTokenValue) {
      refreshToken.value = refreshTokenValue
      storage.setRefreshToken(refreshTokenValue)
    }
  }

  return {
    accessToken,
    refreshToken,
    userInfo,
    isAuthenticated,
    isAdmin,
    login,
    logout,
    fetchUserInfo,
    setTokens
  }
})
