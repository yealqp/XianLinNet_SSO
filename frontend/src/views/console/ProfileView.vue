<template>
  <div class="profile-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <UserOutlined class="title-icon" />
        个人资料
      </h1>
      <p class="page-subtitle">管理你的个人信息和偏好设置</p>
    </div>

    <a-row :gutter="[24, 24]">
      <!-- 左侧：头像和基本信息 -->
      <a-col :xs="24" :lg="8">
        <!-- 头像卡片 -->
        <a-card class="avatar-card" :bordered="false">
          <div class="avatar-section">
            <div class="avatar-wrapper">
              <a-avatar :size="140" :src="avatarUrl" class="main-avatar">
                <template #icon><UserOutlined /></template>
              </a-avatar>
              <div class="avatar-badge">
                <CheckCircleFilled v-if="userInfo?.isRealName" class="verified-icon" />
              </div>
            </div>
            <h3 class="user-name">{{ formState.username || '未设置' }}</h3>
            <p class="user-email">{{ formState.email }}</p>
            <div class="avatar-source">
              <PictureOutlined />
              {{ avatarSource }}
            </div>
          </div>
        </a-card>

        <!-- 账户状态卡片 -->
        <a-card class="status-card" :bordered="false">
          <template #title>
            <div class="card-title">
              <SafetyOutlined class="title-icon" />
              账户状态
            </div>
          </template>
          <div class="status-items">
            <div class="status-item">
              <div class="status-label">用户 ID</div>
              <div class="status-value">
                <a-tag color="blue">{{ userInfo?.id }}</a-tag>
              </div>
            </div>
            <div class="status-item">
              <div class="status-label">账户状态</div>
              <div class="status-value">
                <a-badge status="success" text="正常" />
              </div>
            </div>
            <div class="status-item">
              <div class="status-label">实名认证</div>
              <div class="status-value">
                <a-tag :color="userInfo?.isRealName ? 'success' : 'default'">
                  {{ userInfo?.isRealName ? '已认证' : '未认证' }}
                </a-tag>
              </div>
            </div>
            <div class="status-item">
              <div class="status-label">管理员权限</div>
              <div class="status-value">
                <a-tag :color="userInfo?.isAdmin ? 'red' : 'default'">
                  {{ userInfo?.isAdmin ? '是' : '否' }}
                </a-tag>
              </div>
            </div>
          </div>
        </a-card>
      </a-col>

      <!-- 右侧：编辑表单 -->
      <a-col :xs="24" :lg="16">
        <a-card class="form-card" :bordered="false">
          <template #title>
            <div class="card-title">
              <EditOutlined class="title-icon" />
              编辑资料
            </div>
          </template>

          <a-form
            :model="formState"
            name="profile"
            @finish="handleUpdate"
            layout="vertical"
            class="profile-form"
          >
            <!-- 用户名 -->
            <a-form-item label="用户名" class="form-item">
              <a-input
                v-model:value="formState.username"
                placeholder="请输入用户名"
                size="large"
                class="form-input"
              >
                <template #prefix>
                  <UserOutlined class="input-icon" />
                </template>
              </a-input>
              <template #extra>
                <span class="form-hint">用户名将显示在你的个人资料中</span>
              </template>
            </a-form-item>

            <!-- 邮箱 -->
            <a-form-item label="邮箱地址" class="form-item">
              <a-input
                v-model:value="formState.email"
                placeholder="邮箱地址"
                size="large"
                disabled
                class="form-input"
              >
                <template #prefix>
                  <MailOutlined class="input-icon" />
                </template>
              </a-input>
              <template #extra>
                <span class="form-hint">邮箱地址不可修改，用于登录和接收通知</span>
              </template>
            </a-form-item>

            <!-- QQ 号 -->
            <a-form-item label="QQ 号码" class="form-item">
              <a-input
                v-model:value="formState.qq"
                placeholder="输入 QQ 号可自动使用 QQ 头像"
                size="large"
                class="form-input"
                @change="updateAvatarPreview"
              >
                <template #prefix>
                  <MessageOutlined class="input-icon" />
                </template>
                <template #suffix>
                  <a-tooltip title="输入 QQ 号后将自动使用 QQ 头像">
                    <QuestionCircleOutlined class="help-icon" />
                  </a-tooltip>
                </template>
              </a-input>
              <template #extra>
                <span class="form-hint">绑定 QQ 号后可自动获取 QQ 头像</span>
              </template>
            </a-form-item>

            <!-- 自定义头像 -->
            <a-form-item label="自定义头像" class="form-item">
              <a-input
                v-model:value="formState.avatar"
                placeholder="输入头像 URL，留空则使用 QQ 头像"
                size="large"
                class="form-input"
                @change="updateAvatarPreview"
              >
                <template #prefix>
                  <LinkOutlined class="input-icon" />
                </template>
                <template #suffix>
                  <a-tooltip title="自定义头像优先级高于 QQ 头像">
                    <QuestionCircleOutlined class="help-icon" />
                  </a-tooltip>
                </template>
              </a-input>
              <template #extra>
                <span class="form-hint">自定义头像 URL，优先级高于 QQ 头像</span>
              </template>
            </a-form-item>

            <!-- 头像预览 -->
            <a-form-item label="头像预览" class="form-item">
              <div class="avatar-preview-box">
                <a-avatar :size="80" :src="avatarUrl" class="preview-avatar">
                  <template #icon><UserOutlined /></template>
                </a-avatar>
                <div class="preview-info">
                  <div class="preview-label">当前头像</div>
                  <div class="preview-source">{{ avatarSource }}</div>
                </div>
              </div>
            </a-form-item>

            <!-- 操作按钮 -->
            <a-form-item class="form-actions">
              <a-space :size="12">
                <a-button
                  type="primary"
                  html-type="submit"
                  size="large"
                  :loading="loading"
                  class="submit-btn"
                >
                  <SaveOutlined />
                  保存修改
                </a-button>
                <a-button size="large" @click="loadData" class="reset-btn">
                  <ReloadOutlined />
                  重置
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { storage } from '@/utils/storage'
import {
  UserOutlined,
  MailOutlined,
  MessageOutlined,
  LinkOutlined,
  EditOutlined,
  SafetyOutlined,
  SaveOutlined,
  ReloadOutlined,
  PictureOutlined,
  QuestionCircleOutlined,
  CheckCircleFilled
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const authStore = useAuthStore()

const formState = reactive({
  username: '',
  email: '',
  qq: '',
  avatar: ''
})

const loading = ref(false)
const userInfo = computed(() => authStore.userInfo)

// 获取 QQ 头像 URL
const getQQAvatar = (qq: string): string => {
  if (!qq || !/^\d+$/.test(qq)) return ''
  return `https://q1.qlogo.cn/g?b=qq&nk=${qq}&s=640`
}

// 计算头像 URL（优先级：自定义头像 > QQ 头像）
const avatarUrl = computed(() => {
  if (formState.avatar && formState.avatar.trim()) {
    return formState.avatar.trim()
  }
  if (formState.qq && /^\d+$/.test(formState.qq)) {
    return getQQAvatar(formState.qq)
  }
  return ''
})

// 头像来源提示
const avatarSource = computed(() => {
  if (formState.avatar && formState.avatar.trim()) {
    return '使用自定义头像'
  }
  if (formState.qq && /^\d+$/.test(formState.qq)) {
    return '使用 QQ 头像'
  }
  return '未设置头像'
})

const updateAvatarPreview = () => {
  // 触发响应式更新
}

const loadData = async () => {
  try {
    // Fetch fresh user data from API
    const response = await authApi.getUserInfo()
    if (response.status === 'ok' && response.data) {
      const user = response.data
      // 映射后端字段到前端字段
      formState.username = user.username || user.name || user.preferred_username || ''
      formState.email = user.email || ''
      formState.qq = user.qq || ''
      formState.avatar = user.avatar || user.picture || ''
      
      // 更新 store，使用标准化的字段
      const normalizedUser = {
        id: String(user.id || user.sub || ''),
        username: user.username || user.name || user.preferred_username || '',
        email: user.email || '',
        qq: user.qq || '',
        avatar: user.avatar || user.picture || '',
        isAdmin: user.is_admin || false,
        isRealName: user.is_real_name || false
      }
      authStore.userInfo = normalizedUser as any
      storage.setUserInfo(normalizedUser)
    }
  } catch (error) {
    console.error('Failed to load user info:', error)
    // Fallback to store data if API fails
    const user = authStore.userInfo
    if (user) {
      formState.username = user.username || ''
      formState.email = user.email || ''
      formState.qq = (user as any).qq || ''
      formState.avatar = (user as any).avatar || ''
    }
  }
}

const handleUpdate = async () => {
  if (!userInfo.value?.id) {
    message.error('用户信息获取失败')
    return
  }

  loading.value = true
  try {
    const response = await authApi.updateProfile({
      userId: userInfo.value.id,
      username: formState.username,
      qq: formState.qq,
      avatar: formState.avatar
    })

    if (response.status === 'ok') {
      message.success('个人资料更新成功')
      
      // 更新本地存储的用户信息
      if (authStore.userInfo && response.status === 'ok') {
        // 处理可能的嵌套 ApiResponse
        let userData: any
        if ('data' in response && response.data) {
          userData = response.data
        } else {
          userData = response
        }
        
        if (userData?.user) {
          authStore.userInfo.username = userData.user.username
          authStore.userInfo.qq = userData.user.qq
          authStore.userInfo.avatar = userData.user.avatar
          // 保存到 localStorage
          storage.setUserInfo(authStore.userInfo)
        }
      }
    } else {
      message.error(response.msg || '更新失败')
    }
  } catch (error: any) {
    console.error('Update profile failed:', error)
    message.error(error.response?.data?.msg || '更新失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.profile-view {
  padding: 0;
}

/* 页面标题 */
.page-header {
  margin-bottom: 32px;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: #1f1f1f;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
}

.page-title .title-icon {
  margin-right: 12px;
  color: #667eea;
}

.page-subtitle {
  font-size: 15px;
  color: #8c8c8c;
  margin: 0;
}

/* 卡片通用样式 */
.avatar-card,
.status-card,
.form-card {
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  overflow: hidden;
}

.card-title {
  display: flex;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
  color: #1f1f1f;
}

.card-title .title-icon {
  margin-right: 8px;
  color: #667eea;
}

/* 头像卡片 */
.avatar-section {
  text-align: center;
  padding: 24px 0;
}

.avatar-wrapper {
  position: relative;
  display: inline-block;
  margin-bottom: 20px;
}

.main-avatar {
  border: 4px solid #f0f0f0;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.avatar-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  width: 32px;
  height: 32px;
  background: #ffffff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.verified-icon {
  font-size: 20px;
  color: #52c41a;
}

.user-name {
  font-size: 22px;
  font-weight: 600;
  color: #1f1f1f;
  margin: 0 0 8px 0;
}

.user-email {
  font-size: 14px;
  color: #8c8c8c;
  margin: 0 0 16px 0;
}

.avatar-source {
  display: inline-flex;
  align-items: center;
  padding: 6px 16px;
  background: #f0f0f0;
  border-radius: 20px;
  font-size: 13px;
  color: #595959;
  gap: 6px;
}

/* 状态卡片 */
.status-items {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.status-item:last-child {
  border-bottom: none;
}

.status-label {
  font-size: 14px;
  color: #595959;
  font-weight: 500;
}

.status-value {
  font-size: 14px;
  color: #1f1f1f;
}

/* 表单卡片 */
.profile-form {
  max-width: 600px;
}

.form-item {
  margin-bottom: 28px;
}

.form-input {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.form-input:hover {
  border-color: #667eea;
}

.form-input:focus,
.form-input:focus-within {
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.1);
}

.input-icon {
  color: #8c8c8c;
}

.help-icon {
  color: #bfbfbf;
  cursor: help;
}

.form-hint {
  font-size: 13px;
  color: #8c8c8c;
}

/* 头像预览框 */
.avatar-preview-box {
  display: flex;
  align-items: center;
  padding: 20px;
  background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
  border-radius: 12px;
  gap: 20px;
}

.preview-avatar {
  border: 3px solid #ffffff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.preview-info {
  flex: 1;
}

.preview-label {
  font-size: 14px;
  font-weight: 600;
  color: #1f1f1f;
  margin-bottom: 4px;
}

.preview-source {
  font-size: 13px;
  color: #667eea;
}

/* 操作按钮 */
.form-actions {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.submit-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  font-weight: 500;
  height: 44px;
  padding: 0 32px;
}

.submit-btn:hover {
  background: linear-gradient(135deg, #5568d3 0%, #6a3f8f 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.reset-btn {
  height: 44px;
  padding: 0 24px;
  border-radius: 8px;
}

/* 深度样式覆盖 */
:deep(.ant-form-item-label > label) {
  font-size: 14px;
  font-weight: 600;
  color: #1f1f1f;
}

:deep(.ant-input-disabled) {
  background: #fafafa;
  color: #8c8c8c;
}

/* 响应式 */
@media (max-width: 992px) {
  .page-title {
    font-size: 24px;
  }

  .avatar-section {
    padding: 20px 0;
  }

  .main-avatar {
    width: 100px !important;
    height: 100px !important;
  }

  .user-name {
    font-size: 20px;
  }
}

@media (max-width: 576px) {
  .page-header {
    margin-bottom: 24px;
  }

  .page-title {
    font-size: 22px;
  }

  .avatar-preview-box {
    flex-direction: column;
    text-align: center;
  }

  .submit-btn,
  .reset-btn {
    width: 100%;
  }

  .form-actions :deep(.ant-space) {
    width: 100%;
    flex-direction: column;
  }

  .form-actions :deep(.ant-space-item) {
    width: 100%;
  }
}
</style>
