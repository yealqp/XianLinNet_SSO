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
      const apiResponse = await authApi.login(data)
      // 处理可能的嵌套 ApiResponse
      let loginData: any
      if ('data' in apiResponse && apiResponse.data) {
        loginData = apiResponse.data
      } else {
        loginData = apiResponse
      }
      
      // 检查是否有 access_token
      if (!loginData || !loginData.access_token) {
        throw new Error('登录失败：未收到访问令牌')
      }
      
      accessToken.value = loginData.access_token
      storage.setAccessToken(loginData.access_token)
      
      if (loginData.refresh_token) {
        refreshToken.value = loginData.refresh_token
        storage.setRefreshToken(loginData.refresh_token)
      }
      
      if (loginData.user) {
        userInfo.value = {
          id: String(loginData.user.id),
          email: loginData.user.email,
          username: loginData.user.username,
          isAdmin: loginData.user.isAdmin,
          isRealName: loginData.user.isRealName,
          qq: loginData.user.qq,
          avatar: loginData.user.avatar
        }
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
      const apiResponse = await authApi.getUserInfo()
      // 处理可能的嵌套 ApiResponse
      let data: any
      if ('data' in apiResponse && apiResponse.data) {
        data = apiResponse.data
      } else {
        data = apiResponse
      }
      
      userInfo.value = {
        id: data?.sub || String(data?.id) || '',
        email: data?.email || '',
        username: data?.username || data?.name || data?.preferred_username || '',
        isAdmin: false, // Will be determined by roles/permissions
        isRealName: data?.is_real_name || false,
        qq: data?.qq || '',
        avatar: data?.picture || data?.avatar || ''
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
