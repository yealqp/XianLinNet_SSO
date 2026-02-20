<template>
  <div class="profile-view">
    <h2 class="page-title">个人资料</h2>
    
    <a-row :gutter="16">
      <a-col :xs="24" :md="8">
        <a-card title="头像预览">
          <div class="avatar-preview">
            <a-avatar :size="120" :src="avatarUrl">
              <template #icon><UserOutlined /></template>
            </a-avatar>
            <p class="avatar-hint">{{ avatarSource }}</p>
          </div>
        </a-card>
      </a-col>

      <a-col :xs="24" :md="16">
        <a-card title="基本信息">
          <a-form
            :model="formState"
            name="profile"
            @finish="handleUpdate"
            layout="vertical"
          >
            <a-form-item label="用户名">
              <a-input
                v-model:value="formState.username"
                placeholder="用户名"
                size="large"
              >
                <template #prefix>
                  <UserOutlined />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item label="邮箱">
              <a-input
                v-model:value="formState.email"
                placeholder="邮箱"
                size="large"
                disabled
              >
                <template #prefix>
                  <MailOutlined />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item label="QQ 号">
              <a-input
                v-model:value="formState.qq"
                placeholder="输入 QQ 号可自动使用 QQ 头像"
                size="large"
                @change="updateAvatarPreview"
              >
                <template #prefix>
                  <MessageOutlined />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item label="自定义头像 URL（优先使用）">
              <a-input
                v-model:value="formState.avatar"
                placeholder="头像 URL，留空则使用 QQ 头像"
                size="large"
                @change="updateAvatarPreview"
              >
                <template #prefix>
                  <LinkOutlined />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item>
              <a-space>
                <a-button
                  type="primary"
                  html-type="submit"
                  size="large"
                  :loading="loading"
                >
                  保存修改
                </a-button>
                <a-button size="large" @click="loadData">
                  重置
                </a-button>
              </a-space>
            </a-form-item>
          </a-form>
        </a-card>
      </a-col>
    </a-row>

    <a-card title="账户信息" class="mt-4">
      <a-descriptions bordered :column="{ xs: 1, sm: 2 }">
        <a-descriptions-item label="用户 ID">
          {{ userInfo?.id }}
        </a-descriptions-item>
        <a-descriptions-item label="账户状态">
          <a-badge status="success" text="正常" />
        </a-descriptions-item>
        <a-descriptions-item label="管理员权限">
          <a-tag :color="userInfo?.isAdmin ? 'red' : 'default'">
            {{ userInfo?.isAdmin ? '是' : '否' }}
          </a-tag>
        </a-descriptions-item>
      </a-descriptions>
    </a-card>
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
  LinkOutlined
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

const loadData = () => {
  const user = authStore.userInfo
  if (user) {
    formState.username = user.username || ''
    formState.email = user.email || ''
    formState.qq = (user as any).qq || ''
    formState.avatar = (user as any).avatar || ''
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
      if (authStore.userInfo && response.data?.user) {
        authStore.userInfo.username = response.data.user.username
        authStore.userInfo.qq = response.data.user.qq
        authStore.userInfo.avatar = response.data.user.avatar
        // 保存到 localStorage
        storage.setUserInfo(authStore.userInfo)
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

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 24px;
  color: #1f1f1f;
}

.avatar-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 0;
}

.avatar-hint {
  margin-top: 16px;
  color: #8c8c8c;
  font-size: 14px;
}

.mt-4 {
  margin-top: 24px;
}

:deep(.ant-form-item-label > label) {
  color: #595959;
  font-size: 14px;
}

:deep(.ant-input-affix-wrapper) {
  background: #ffffff;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
}

:deep(.ant-input-affix-wrapper:hover) {
  border-color: #ff4db3;
}

:deep(.ant-input-affix-wrapper:focus),
:deep(.ant-input-affix-wrapper-focused) {
  border-color: #f6339a;
  box-shadow: 0 0 0 2px rgba(246, 51, 154, 0.1);
}

:deep(.ant-btn-primary) {
  background: #f6339a;
  border: none;
}

:deep(.ant-btn-primary:hover) {
  background: #ff4db3;
}

:deep(.ant-btn-primary:active) {
  background: #e02987;
}
</style>
