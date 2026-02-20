<template>
  <AuthLayout>
    <div class="login-view">
      <h2 class="title">登录 XianlinNet ID</h2>
      <a-form
        :model="formState"
        name="login"
        @finish="handleLogin"
        layout="vertical"
      >
        <a-form-item
          name="email"
          :rules="[
            { required: true, message: '请输入邮箱' },
            { type: 'email', message: '请输入有效的邮箱地址' }
          ]"
        >
          <a-input
            v-model:value="formState.email"
            placeholder="邮箱"
            size="large"
          >
            <template #prefix>
              <MailOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          name="password"
          :rules="[{ required: true, message: '请输入密码' }]"
        >
          <a-input-password
            v-model:value="formState.password"
            placeholder="密码"
            size="large"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <!-- Captcha Widget -->
        <a-form-item>
          <Captcha 
            :site-key="captchaSiteKey"
            :api-endpoint="captchaApiEndpoint"
            @success="handleCaptchaSuccess"
            @error="handleCaptchaError"
            ref="captchaRef"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            :disabled="!captchaToken"
            block
          >
            登录
          </a-button>
        </a-form-item>

        <div class="extra-links">
          <router-link to="/register">注册账号</router-link>
          <router-link to="/forgot-password">忘记密码？</router-link>
        </div>
      </a-form>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { MailOutlined, LockOutlined } from '@ant-design/icons-vue'
import AuthLayout from '@/components/layout/AuthLayout.vue'
import Captcha from '@/components/Captcha.vue'
import { message } from 'ant-design-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// Captcha 配置
const captchaSiteKey = import.meta.env.VITE_CAPTCHA_SITE_KEY || '1cbf106b94'
const captchaApiEndpoint = import.meta.env.VITE_CAPTCHA_API_ENDPOINT || 'https://captcha.yealqp.cn/1cbf106b94'

const formState = reactive({
  email: '',
  password: ''
})

const loading = ref(false)
const captchaToken = ref('')
const captchaRef = ref<InstanceType<typeof Captcha> | null>(null)

const handleCaptchaSuccess = (token: string) => {
  captchaToken.value = token
}

const handleCaptchaError = (error: any) => {
  console.error('Captcha error:', error)
  message.error('人机验证失败，请刷新页面重试')
}

const handleLogin = async () => {
  if (!captchaToken.value) {
    message.error('请完成人机验证')
    return
  }

  loading.value = true
  try {
    await authStore.login({
      email: formState.email,
      password: formState.password,
      captchaToken: captchaToken.value
    })
    message.success('登录成功')
    
    const redirect = route.query.redirect as string
    router.push(redirect || '/console/dashboard')
  } catch (error: any) {
    console.error('Login failed:', error)
    const errorMessage = error.message || error.response?.data?.msg || '登录失败，请检查邮箱和密码'
    message.error(errorMessage)
    
    // 重置 captcha
    captchaToken.value = ''
    if (captchaRef.value) {
      captchaRef.value.reset()
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-view {
  width: 100%;
}

.title {
  font-size: 26px;
  font-weight: 600;
  text-align: left;
  margin-bottom: 24px;
  color: #ffffff;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.extra-links {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
}

.extra-links a {
  color: #ffd6ed;
  text-decoration: none;
  font-size: 13px;
  transition: all 0.3s;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.extra-links a:hover {
  color: #ffffff;
  text-shadow: 0 2px 6px rgba(246, 51, 154, 0.6);
}

:deep(.ant-form-item-label > label) {
  color: rgba(255, 255, 255, 0.9);
  font-size: 13px;
  font-weight: 500;
}

:deep(.ant-input-affix-wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  padding: 10px 14px;
  transition: all 0.3s;
}

:deep(.ant-input-affix-wrapper:hover) {
  background: rgba(255, 255, 255, 0.98);
  border-color: #ff4db3;
  box-shadow: 0 2px 8px rgba(246, 51, 154, 0.15);
}

:deep(.ant-input-affix-wrapper:focus),
:deep(.ant-input-affix-wrapper-focused) {
  background: #ffffff;
  border-color: #f6339a;
  box-shadow: 0 0 0 3px rgba(246, 51, 154, 0.15);
}

:deep(.ant-input) {
  background: transparent;
  color: #1f1f1f;
  font-size: 14px;
}

:deep(.ant-input::placeholder) {
  color: #999999;
}

:deep(.ant-input-prefix) {
  color: #666666;
}

:deep(.ant-btn-primary) {
  height: 44px;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  background: linear-gradient(135deg, #f6339a 0%, #ff4db3 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(246, 51, 154, 0.4);
  transition: all 0.3s;
  margin-top: 4px;
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

:deep(.ant-btn-primary:disabled) {
  background: rgba(255, 255, 255, 0.3);
  box-shadow: none;
}

:deep(.ant-form-item) {
  margin-bottom: 16px;
}

:deep(.ant-form-item-explain-error) {
  color: #ffccc7;
  background: rgba(255, 77, 79, 0.15);
  padding: 4px 8px;
  border-radius: 4px;
  margin-top: 4px;
  font-size: 12px;
}

@media (max-width: 768px) {
  .title {
    font-size: 22px;
    margin-bottom: 20px;
  }

  :deep(.ant-btn-primary) {
    height: 42px;
    font-size: 14px;
  }

  .extra-links a {
    font-size: 12px;
  }
}
</style>
