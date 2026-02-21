<template>
  <AuthLayout>
    <div class="authorize-view">
      <div v-if="loading" class="loading-container">
        <a-spin size="large" />
        <p>加载中...</p>
      </div>

      <div v-else-if="error" class="error-container">
        <a-result
          status="error"
          :title="error"
          :sub-title="errorDetail"
        >
          <template #extra>
            <a-button type="primary" @click="$router.push('/login')">
              返回登录
            </a-button>
          </template>
        </a-result>
      </div>

      <div v-else class="authorize-content">
        <div class="app-info">
          <a-avatar :size="64" :src="appInfo.logo">
            <template #icon><AppstoreOutlined /></template>
          </a-avatar>
          <h2 class="app-name">{{ appInfo.displayName || appInfo.name }}</h2>
          <p class="app-description">{{ appInfo.description || '第三方应用' }}</p>
        </div>

        <a-divider />

        <div class="consent-info">
          <h3>授权请求</h3>
          <p class="consent-text">
            <strong>{{ appInfo.displayName || appInfo.name }}</strong> 
            请求访问您的账户信息
          </p>

          <div class="scope-list">
            <h4>该应用将获得以下权限：</h4>
            <a-list :data-source="requestedScopes" size="small">
              <template #renderItem="{ item }">
                <a-list-item>
                  <a-list-item-meta>
                    <template #avatar>
                      <CheckCircleOutlined style="color: #52c41a; font-size: 18px;" />
                    </template>
                    <template #title>{{ item.name }}</template>
                    <template #description>{{ item.description }}</template>
                  </a-list-item-meta>
                </a-list-item>
              </template>
            </a-list>
          </div>

          <div class="user-info">
            <a-alert
              message="授权账户"
              :description="`您将以 ${userInfo?.username} (${userInfo?.email}) 的身份授权此应用`"
              type="info"
              show-icon
            />
          </div>
        </div>

        <a-divider />

        <div class="action-buttons">
          <a-space direction="vertical" style="width: 100%;" :size="12">
            <a-button
              type="primary"
              size="large"
              block
              :loading="authorizing"
              @click="handleAuthorize"
            >
              授权
            </a-button>
            <a-button
              size="large"
              block
              @click="handleCancel"
            >
              取消
            </a-button>
          </a-space>
        </div>

        <div class="security-notice">
          <a-alert
            message="安全提示"
            description="授权后，该应用将能够访问您的个人信息。请确保您信任此应用。"
            type="warning"
            show-icon
          />
        </div>
      </div>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { AppstoreOutlined, CheckCircleOutlined } from '@ant-design/icons-vue'
import AuthLayout from '@/components/layout/AuthLayout.vue'
import { message } from 'ant-design-vue'

const route = useRoute()
const authStore = useAuthStore()

// 状态
const loading = ref(true)
const authorizing = ref(false)
const error = ref('')
const errorDetail = ref('')

// OAuth 参数
const clientId = ref('')
const responseType = ref('')
const redirectUri = ref('')
const scope = ref('')
const state = ref('')
const nonce = ref('')
const codeChallenge = ref('')

// 应用信息
const appInfo = ref<any>({
  name: '',
  displayName: '',
  description: '',
  logo: ''
})

// 用户信息
const userInfo = computed(() => authStore.userInfo)

// Scope 定义
const scopeDefinitions: Record<string, { name: string; description: string }> = {
  openid: {
    name: 'OpenID',
    description: '获取您的唯一标识符'
  },
  profile: {
    name: '基本资料',
    description: '访问您的用户名、头像等基本信息'
  },
  email: {
    name: '邮箱地址',
    description: '访问您的邮箱地址'
  },
  offline_access: {
    name: '离线访问',
    description: '在您离线时访问您的信息'
  }
}

// 请求的权限列表
const requestedScopes = computed(() => {
  if (!scope.value) return []
  
  return scope.value.split(' ')
    .filter(s => s)
    .map(s => scopeDefinitions[s] || { name: s, description: '未知权限' })
})

// 加载应用信息
const loadAppInfo = async () => {
  try {
    const apiResponse = await authApi.getApplicationInfo(clientId.value)
    // 处理可能的嵌套 ApiResponse
    let data: any
    if ('data' in apiResponse && apiResponse.data) {
      data = apiResponse.data
    } else {
      data = apiResponse
    }
    
    if (data) {
      appInfo.value = {
        name: data.name || clientId.value,
        displayName: data.displayName || data.name || '第三方应用',
        description: data.description || '第三方应用',
        logo: data.logo || '',
        homepageUrl: data.homepageUrl || '',
        organization: data.organization || ''
      }
    } else {
      // Fallback to default values
      appInfo.value = {
        name: clientId.value,
        displayName: '第三方应用',
        description: '一个使用 OAuth2 的第三方应用',
        logo: ''
      }
    }
  } catch (err: any) {
    console.error('Failed to load app info:', err)
    // Use default values on error
    appInfo.value = {
      name: clientId.value,
      displayName: '第三方应用',
      description: '一个使用 OAuth2 的第三方应用',
      logo: ''
    }
  }
}

// 验证参数
const validateParams = () => {
  if (!clientId.value) {
    error.value = '缺少必需参数'
    errorDetail.value = '缺少 client_id 参数'
    return false
  }

  if (!responseType.value) {
    error.value = '缺少必需参数'
    errorDetail.value = '缺少 response_type 参数'
    return false
  }

  if (!redirectUri.value) {
    error.value = '缺少必需参数'
    errorDetail.value = '缺少 redirect_uri 参数'
    return false
  }

  if (!authStore.isAuthenticated) {
    error.value = '未登录'
    errorDetail.value = '请先登录后再进行授权'
    return false
  }

  return true
}

// 初始化
onMounted(async () => {
  // 从 URL 获取参数
  clientId.value = route.query.client_id as string || ''
  responseType.value = route.query.response_type as string || ''
  redirectUri.value = route.query.redirect_uri as string || ''
  scope.value = route.query.scope as string || 'openid profile email'
  state.value = route.query.state as string || ''
  nonce.value = route.query.nonce as string || ''
  codeChallenge.value = route.query.code_challenge as string || ''

  // 验证参数
  if (!validateParams()) {
    loading.value = false
    return
  }

  // 加载应用信息
  await loadAppInfo()
  
  loading.value = false
})

// 处理授权
const handleAuthorize = async () => {
  authorizing.value = true

  try {
    // 构建授权请求参数
    const params = new URLSearchParams({
      client_id: clientId.value,
      response_type: responseType.value,
      redirect_uri: redirectUri.value,
      scope: scope.value,
      state: state.value
    })

    if (nonce.value) {
      params.append('nonce', nonce.value)
    }

    if (codeChallenge.value) {
      params.append('code_challenge', codeChallenge.value)
    }

    // 调用后端授权接口（需要带上 Authorization header）
    // 使用完整的 API URL
    const token = authStore.accessToken
    const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api'
    // 构建完整的授权 URL
    let authorizeUrl: string
    if (apiBaseUrl.startsWith('http')) {
      // 生产环境：https://api.account.idcxl.cn/api -> https://api.account.idcxl.cn/oauth/authorize
      const baseUrl = apiBaseUrl.replace(/\/api$/, '')
      authorizeUrl = `${baseUrl}/oauth/authorize`
    } else {
      // 开发环境：/api -> /oauth/authorize
      authorizeUrl = '/oauth/authorize'
    }
    
    const response = await fetch(`${authorizeUrl}?${params.toString()}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    const result = await response.json()

    // 后端应该返回重定向 URL 或授权码
    if (result.status === 'ok') {
      const data = result.data
      
      // 如果有重定向 URL，直接跳转
      if (data.redirect_uri) {
        window.location.href = data.redirect_uri
      } else if (data.code) {
        // 如果只返回授权码，手动构建重定向 URL
        const redirectUrl = new URL(redirectUri.value)
        redirectUrl.searchParams.set('code', data.code)
        if (state.value) {
          redirectUrl.searchParams.set('state', state.value)
        }
        window.location.href = redirectUrl.toString()
      }
    } else {
      message.error(result.msg || '授权失败')
    }
  } catch (err: any) {
    console.error('Authorization failed:', err)
    const errorMessage = err.message || '授权失败'
    message.error(errorMessage)
  } finally {
    authorizing.value = false
  }
}

// 处理取消
const handleCancel = () => {
  // 重定向回应用，带上 error 参数
  const redirectUrl = new URL(redirectUri.value)
  redirectUrl.searchParams.set('error', 'access_denied')
  redirectUrl.searchParams.set('error_description', 'User denied authorization')
  if (state.value) {
    redirectUrl.searchParams.set('state', state.value)
  }
  
  window.location.href = redirectUrl.toString()
}
</script>

<style scoped>
.authorize-view {
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
}

.loading-container,
.error-container {
  text-align: center;
  padding: 40px 20px;
}

.loading-container p {
  margin-top: 16px;
  color: rgba(255, 255, 255, 0.8);
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.authorize-content {
  width: 100%;
}

.app-info {
  text-align: center;
  padding: 16px 0;
}

.app-name {
  font-size: 22px;
  font-weight: 600;
  margin: 14px 0 6px;
  color: #ffffff;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.app-description {
  color: rgba(255, 255, 255, 0.75);
  font-size: 13px;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.consent-info {
  padding: 16px 0;
}

.consent-info h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
  color: #ffffff;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

.consent-text {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
  margin-bottom: 20px;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.scope-list {
  margin: 20px 0;
}

.scope-list h4 {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 10px;
  color: rgba(255, 255, 255, 0.9);
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.user-info {
  margin: 20px 0;
}

.action-buttons {
  margin: 20px 0;
}

.security-notice {
  margin-top: 20px;
}

:deep(.ant-btn-primary) {
  background: linear-gradient(135deg, #f6339a 0%, #ff4db3 100%);
  border: none;
  height: 44px;
  font-size: 15px;
  font-weight: 600;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(246, 51, 154, 0.4);
  transition: all 0.3s;
}

:deep(.ant-btn-primary:hover:not(:disabled)) {
  background: linear-gradient(135deg, #ff4db3 0%, #ff66c4 100%);
  box-shadow: 0 6px 16px rgba(246, 51, 154, 0.5);
  transform: translateY(-2px);
}

:deep(.ant-btn-primary:active) {
  background: linear-gradient(135deg, #e02987 0%, #f6339a 100%);
  transform: translateY(0);
}

:deep(.ant-btn-default) {
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: #ffffff;
  height: 44px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 8px;
  backdrop-filter: blur(10px);
  transition: all 0.3s;
}

:deep(.ant-btn-default:hover) {
  background: rgba(255, 255, 255, 0.25);
  border-color: rgba(255, 255, 255, 0.5);
  color: #ffffff;
}

:deep(.ant-list-item-meta-title) {
  font-size: 13px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.95);
}

:deep(.ant-list-item-meta-description) {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
}

:deep(.ant-list-item) {
  border-bottom-color: rgba(255, 255, 255, 0.1);
  padding: 10px 0;
}

:deep(.ant-alert) {
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
}

:deep(.ant-alert-message) {
  color: rgba(255, 255, 255, 0.95);
  font-weight: 500;
}

:deep(.ant-alert-description) {
  color: rgba(255, 255, 255, 0.8);
}

:deep(.ant-alert-icon) {
  color: rgba(255, 255, 255, 0.9);
}

:deep(.ant-divider) {
  border-color: rgba(255, 255, 255, 0.15);
  margin: 16px 0;
}

@media (max-width: 768px) {
  .app-name {
    font-size: 20px;
  }

  .consent-info h3 {
    font-size: 15px;
  }

  .consent-text {
    font-size: 13px;
  }

  :deep(.ant-btn-primary),
  :deep(.ant-btn-default) {
    height: 42px;
    font-size: 14px;
  }
}
</style>
