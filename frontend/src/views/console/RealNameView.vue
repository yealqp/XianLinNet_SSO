<template>
  <div class="realname-view">
    <div class="page-header">
      <h2 class="page-title">实名认证</h2>
      <p class="page-subtitle">完成实名认证，获得更多功能权限</p>
    </div>

    <div v-if="!userInfo?.isRealName" class="realname-card">
      <a-alert
        message="实名认证说明"
        description="完成实名认证后，您将获得更多功能权限。请确保填写的信息真实有效。"
        type="info"
        show-icon
        class="info-alert"
      />

      <a-form
        :model="formState"
        name="realname"
        @finish="handleSubmit"
        layout="vertical"
        class="realname-form"
      >
        <a-form-item
          label="真实姓名"
          name="name"
          :rules="[{ required: true, message: '请输入真实姓名' }]"
          class="form-item"
        >
          <a-input
            v-model:value="formState.name"
            placeholder="请输入真实姓名"
            size="large"
            class="form-input"
          >
            <template #prefix>
              <UserOutlined class="input-icon" />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          label="身份证号"
          name="idcard"
          :rules="[
            { required: true, message: '请输入身份证号' },
            { pattern: /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/, message: '请输入有效的身份证号' }
          ]"
          class="form-item"
        >
          <a-input
            v-model:value="formState.idcard"
            placeholder="请输入18位身份证号"
            size="large"
            maxlength="18"
            class="form-input"
          >
            <template #prefix>
              <IdcardOutlined class="input-icon" />
            </template>
          </a-input>
        </a-form-item>

        <div class="form-actions">
          <a-button
            type="primary"
            html-type="submit"
            :loading="loading"
            class="submit-btn"
          >
            <SafetyOutlined />
            提交认证
          </a-button>
          <a-button @click="resetForm" class="reset-btn">
            <ReloadOutlined />
            重置
          </a-button>
        </div>
      </a-form>

      <div class="tips-box">
        <h4 class="tips-title">
          <InfoCircleOutlined class="tips-icon" />
          温馨提示
        </h4>
        <ul class="tips-list">
          <li>请确保填写的姓名和身份证号真实有效</li>
          <li>身份证信息将用于实名认证，不会泄露给第三方</li>
          <li>实名认证成功后，您的账号将升级为实名用户</li>
          <li>如有疑问，请联系客服</li>
        </ul>
      </div>
    </div>

    <div v-else class="success-card">
      <div class="success-icon">
        <CheckCircleOutlined />
      </div>
      <h3 class="success-title">您已完成实名认证</h3>
      <p class="success-subtitle">您的账号已通过实名认证，可以使用所有功能</p>
      <a-button type="primary" @click="$router.push('/console/dashboard')" class="success-btn">
        <HomeOutlined />
        返回首页
      </a-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { storage } from '@/utils/storage'
import { 
  UserOutlined, 
  IdcardOutlined,
  SafetyOutlined,
  ReloadOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  HomeOutlined
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const authStore = useAuthStore()
const userInfo = computed(() => authStore.userInfo)

const formState = reactive({
  name: '',
  idcard: ''
})

const loading = ref(false)

const handleSubmit = async () => {
  if (!userInfo.value?.id) {
    message.error('用户信息获取失败')
    return
  }

  loading.value = true
  try {
    const response = await authApi.submitRealName({
      userId: userInfo.value.id,
      name: formState.name,
      idcard: formState.idcard
    })

    if (response.status === 'ok') {
      message.success('实名认证成功！')
      
      // 更新本地用户信息
      if (authStore.userInfo) {
        authStore.userInfo.isRealName = true
        // 保存到 localStorage
        storage.setUserInfo(authStore.userInfo)
      }

      // 重置表单
      resetForm()
    } else {
      message.error(response.msg || '实名认证失败')
    }
  } catch (error: any) {
    console.error('Real name verification failed:', error)
    message.error(error.response?.data?.msg || '实名认证失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  formState.name = ''
  formState.idcard = ''
}
</script>

<style scoped>
.realname-view {
  padding: 0;
}

/* 页面头部 */
.page-header {
  margin-bottom: 32px;
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #1E293B;
  background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.page-subtitle {
  font-size: 15px;
  color: #64748B;
  margin: 0;
}

/* 卡片样式 */
.realname-card {
  background: #FFFFFF;
  border-radius: 20px;
  padding: 32px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(226, 232, 240, 0.8);
  max-width: 600px;
  margin: 0 auto;
}

/* 提示信息 */
.info-alert {
  margin-bottom: 32px;
  border-radius: 12px;
  border: 1px solid #DBEAFE;
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
}

/* 表单样式 */
.realname-form {
  margin-bottom: 32px;
}

.form-item {
  margin-bottom: 24px;
}

.form-input {
  height: 48px;
  border-radius: 12px;
  border: 1px solid #E2E8F0;
  transition: all 0.3s ease;
}

.form-input:hover {
  border-color: #3B82F6;
}

.form-input:focus,
.form-input:focus-within {
  border-color: #2563EB;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.input-icon {
  color: #64748B;
  font-size: 16px;
}

/* 按钮组 */
.form-actions {
  display: flex;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid #F1F5F9;
}

.submit-btn {
  flex: 1;
  height: 48px;
  border-radius: 12px;
  font-weight: 500;
  background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
}

.submit-btn:hover {
  background: linear-gradient(135deg, #1d4ed8 0%, #2563EB 100%);
  box-shadow: 0 6px 16px rgba(37, 99, 235, 0.3);
  transform: translateY(-1px);
}

.reset-btn {
  height: 48px;
  border-radius: 12px;
  border: 1px solid #E2E8F0;
  background: #FFFFFF;
}

.reset-btn:hover {
  border-color: #3B82F6;
  color: #2563EB;
}

/* 提示框 */
.tips-box {
  background: linear-gradient(135deg, #F8FAFC 0%, #F1F5F9 100%);
  padding: 24px;
  border-radius: 16px;
  border: 1px solid #E2E8F0;
}

.tips-title {
  font-size: 16px;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.tips-icon {
  color: #3B82F6;
}

.tips-list {
  margin: 0;
  padding-left: 20px;
  color: #64748B;
  font-size: 14px;
  line-height: 1.8;
}

.tips-list li {
  margin-bottom: 8px;
}

.tips-list li:last-child {
  margin-bottom: 0;
}

/* 成功状态 */
.success-card {
  background: #FFFFFF;
  border-radius: 20px;
  padding: 48px 32px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(226, 232, 240, 0.8);
  text-align: center;
}

.success-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, #10B981 0%, #34D399 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  font-size: 40px;
}

.success-title {
  font-size: 24px;
  font-weight: 700;
  color: #1E293B;
  margin: 0 0 12px 0;
}

.success-subtitle {
  font-size: 15px;
  color: #64748B;
  margin: 0 0 32px 0;
}

.success-btn {
  height: 48px;
  padding: 0 32px;
  border-radius: 12px;
  font-weight: 500;
  background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
  border: none;
}

.success-btn:hover {
  background: linear-gradient(135deg, #1d4ed8 0%, #2563EB 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

/* 深度样式覆盖 */
:deep(.ant-form-item-label > label) {
  font-size: 14px;
  font-weight: 600;
  color: #1E293B;
}

:deep(.ant-alert-info) {
  border: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-title {
    font-size: 28px;
  }

  .realname-card,
  .success-card {
    padding: 24px;
    border-radius: 16px;
  }

  .form-actions {
    flex-direction: column;
  }

  .submit-btn,
  .reset-btn {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .page-header {
    margin-bottom: 24px;
  }

  .page-title {
    font-size: 24px;
  }

  .realname-card {
    padding: 20px;
  }

  .tips-box {
    padding: 20px;
  }
}
</style>
