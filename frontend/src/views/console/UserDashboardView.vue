<template>
  <div class="user-dashboard">
    <h2 class="page-title">欢迎回来，{{ userInfo?.username }}</h2>
    
    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :md="12">
        <a-card title="个人信息" :bordered="false">
          <a-descriptions :column="1">
            <a-descriptions-item label="邮箱">
              {{ userInfo?.email }}
            </a-descriptions-item>
            <a-descriptions-item label="显示名称">
              {{ userInfo?.username }}
            </a-descriptions-item>
            <a-descriptions-item label="账户状态">
              <a-badge status="success" text="正常" />
            </a-descriptions-item>
          </a-descriptions>
          <a-button type="primary" @click="$router.push('/console/profile')" style="margin-top: 16px;">
            编辑个人资料
          </a-button>
        </a-card>
      </a-col>

      <a-col :xs="24" :md="12">
        <a-card title="快速操作" :bordered="false">
          <a-space direction="vertical" style="width: 100%;" :size="12">
            <a-button block @click="$router.push('/console/profile')">
              <UserOutlined />
              个人资料
            </a-button>
            <a-button block @click="handleChangePassword">
              <LockOutlined />
              修改密码
            </a-button>
          </a-space>
        </a-card>
      </a-col>
    </a-row>

    <a-card title="最近活动" style="margin-top: 16px;" :bordered="false">
      <a-empty description="暂无活动记录" />
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const authStore = useAuthStore()
const userInfo = computed(() => authStore.userInfo)

const handleChangePassword = () => {
  message.info('密码修改功能开发中')
}
</script>

<style scoped>
.user-dashboard {
  padding: 0;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 24px;
  color: #1f1f1f;
}
</style>
