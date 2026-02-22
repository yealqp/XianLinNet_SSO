<template>
  <a-layout class="console-layout">
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      class="console-sider"
    >
      <div class="logo">
        <span v-if="!collapsed">XianlinNet ID</span>
        <span v-else>XN</span>
      </div>
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
      >
        <a-menu-item key="dashboard" @click="$router.push('/console/dashboard')">
          <DashboardOutlined />
          <span>仪表板</span>
        </a-menu-item>
        <a-menu-item key="profile" @click="$router.push('/console/profile')">
          <UserOutlined />
          <span>个人资料</span>
        </a-menu-item>
        <a-menu-item key="realname" @click="$router.push('/console/realname')">
          <IdcardOutlined />
          <span>实名认证</span>
        </a-menu-item>
        <a-menu-item key="authorizations" @click="$router.push('/console/authorizations')">
          <SafetyOutlined />
          <span>授权管理</span>
        </a-menu-item>
        <a-sub-menu v-if="isAdmin" key="admin">
          <template #title>
            <SettingOutlined />
            <span>管理</span>
          </template>
          <a-menu-item key="admin-dashboard" @click="$router.push('/admin/dashboard')">
            <DashboardOutlined />
            <span>管理仪表板</span>
          </a-menu-item>
          <a-menu-item key="users" @click="$router.push('/admin/users')">
            <TeamOutlined />
            <span>用户管理</span>
          </a-menu-item>
          <a-menu-item key="applications" @click="$router.push('/admin/applications')">
            <AppstoreOutlined />
            <span>应用管理</span>
          </a-menu-item>
          <a-menu-item key="tokens" @click="$router.push('/admin/tokens')">
            <KeyOutlined />
            <span>令牌管理</span>
          </a-menu-item>
        </a-sub-menu>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="console-header" :class="{ collapsed }">
        <div class="header-left">
          <menu-unfold-outlined
            v-if="collapsed"
            class="trigger"
            @click="() => (collapsed = !collapsed)"
          />
          <menu-fold-outlined
            v-else
            class="trigger"
            @click="() => (collapsed = !collapsed)"
          />
        </div>
        <div class="header-right">
          <a-dropdown>
            <span class="user-info">
              <a-avatar :size="32" :src="userAvatarUrl">
                <template #icon><UserOutlined /></template>
              </a-avatar>
              <span class="username">{{ userInfo?.username || 'User' }}</span>
            </span>
            <template #overlay>
              <a-menu>
                <a-menu-item key="profile" @click="$router.push('/console/profile')">
                  <UserOutlined />
                  个人资料
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout" @click="handleLogout">
                  <LogoutOutlined />
                  退出登录
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      <a-layout-content class="console-content" :class="{ collapsed }">
        <router-view></router-view>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  DashboardOutlined,
  UserOutlined,
  SettingOutlined,
  TeamOutlined,
  AppstoreOutlined,
  KeyOutlined,
  IdcardOutlined,
  SafetyOutlined,
  MenuUnfoldOutlined,
  MenuFoldOutlined,
  LogoutOutlined
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const collapsed = ref(false)
const selectedKeys = ref<string[]>([])

const userInfo = computed(() => authStore.userInfo)
const isAdmin = computed(() => authStore.isAdmin)

// 获取用户头像 URL
const userAvatarUrl = computed(() => {
  if (!userInfo.value) return ''
  
  const user = userInfo.value as any
  
  // 优先使用自定义头像
  if (user.avatar && user.avatar.trim()) {
    return user.avatar.trim()
  }
  
  // 其次使用 QQ 头像
  if (user.qq && /^\d+$/.test(user.qq)) {
    return `https://q1.qlogo.cn/g?b=qq&nk=${user.qq}&s=100`
  }
  
  return ''
})

const updateSelectedKeys = () => {
  const path = route.path
  if (path === '/admin/dashboard') {
    selectedKeys.value = ['admin-dashboard']
  } else if (path.includes('/admin/')) {
    const subPath = path.replace('/admin/', '')
    selectedKeys.value = [subPath]
  } else if (path.includes('/console/')) {
    const subPath = path.replace('/console/', '')
    selectedKeys.value = [subPath]
  }
}

watch(() => route.path, updateSelectedKeys, { immediate: true })

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.console-layout {
  min-height: 100vh;
  width: 100%;
  background: #ffffff;
}

.console-sider {
  overflow: auto;
  height: 100vh;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 100;
  background: #ffffff !important;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 覆盖 Ant Design 侧栏暗色主题 */
.console-sider :deep(.ant-layout-sider-children) {
  background: #ffffff;
}

.console-sider :deep(.ant-menu-dark) {
  background: #ffffff;
}

.console-sider :deep(.ant-menu-dark .ant-menu-item) {
  color: #666666;
  background: transparent;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.console-sider :deep(.ant-menu-dark .ant-menu-item::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #ec4899;
  transform: scaleY(0);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.console-sider :deep(.ant-menu-dark .ant-menu-item:hover) {
  color: #ec4899;
  background: #fdf2f8;
  transform: translateX(4px);
}

.console-sider :deep(.ant-menu-dark .ant-menu-item-selected) {
  color: #ec4899;
  background: linear-gradient(90deg, #fdf2f8 0%, #fce7f3 100%);
  border-right: 3px solid #ec4899;
  font-weight: 600;
  transform: translateX(0);
}

.console-sider :deep(.ant-menu-dark .ant-menu-item-selected::before) {
  transform: scaleY(1);
}

.console-sider :deep(.ant-menu-dark .ant-menu-submenu-title) {
  color: #666666;
}

.console-sider :deep(.ant-menu-dark .ant-menu-submenu-title:hover) {
  color: #ec4899;
  background: #fdf2f8;
}

.console-sider :deep(.ant-menu-dark .ant-menu-submenu-selected > .ant-menu-submenu-title) {
  color: #ec4899;
}

.console-sider :deep(.ant-menu-dark .ant-menu-item-selected .anticon) {
  color: #ec4899;
}

.console-sider :deep(.ant-menu-dark .ant-menu-submenu-open) {
  color: #ec4899;
}

.console-sider :deep(.ant-menu-dark .ant-menu-sub) {
  background: #fafafa;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ec4899;
  font-size: 20px;
  font-weight: bold;
  background: linear-gradient(135deg, #fdf2f8 0%, #fce7f3 100%);
  white-space: nowrap;
  border-bottom: 1px solid #fbcfe8;
  transition: all 0.3s ease;
  cursor: pointer;
}

.logo:hover {
  background: linear-gradient(135deg, #fce7f3 0%, #fbcfe8 100%);
  letter-spacing: 1px;
}

.console-header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  z-index: 1;
  position: sticky;
  top: 0;
  margin-left: 200px;
  border-bottom: 1px solid #f3f4f6;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.console-header.collapsed {
  margin-left: 80px;
}

.header-left {
  display: flex;
  align-items: center;
}

.trigger {
  font-size: 18px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  color: #666666;
  padding: 8px;
  border-radius: 8px;
}

.trigger:hover {
  color: #ec4899;
  background: #fdf2f8;
  transform: scale(1.1);
}

.trigger:active {
  transform: scale(0.95);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.user-info:hover {
  background: #fdf2f8;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(236, 72, 153, 0.1);
}

.user-info:active {
  transform: translateY(0);
}

.username {
  margin-left: 8px;
  white-space: nowrap;
  color: #1f1f1f;
  font-weight: 500;
}

.console-content {
  margin: 16px;
  margin-left: 216px;
  padding: 24px;
  background: #ffffff;
  min-height: calc(100vh - 112px);
  border-radius: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.console-content.collapsed {
  margin-left: 96px;
}

@media (max-width: 1200px) {
  .console-content {
    margin: 12px;
    padding: 20px;
  }
}

@media (max-width: 992px) {
  .header-left {
    flex: 1;
  }

  .console-content {
    margin: 8px;
    padding: 16px;
  }
}

@media (max-width: 768px) {
  .console-sider {
    z-index: 1000;
  }

  .username {
    display: none;
  }

  .console-header {
    padding: 0 16px;
  }

  .console-content {
    margin: 0;
    padding: 12px;
    border-radius: 0;
  }
}

@media (max-width: 480px) {
  .console-header {
    padding: 0 12px;
  }

  .console-content {
    padding: 8px;
  }
}
</style>
