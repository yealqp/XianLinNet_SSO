<template>
  <div class="user-dashboard">
    <div class="dashboard-header">
      <h2 class="page-title">欢迎回来，{{ userInfo?.username }}</h2>
      <p class="page-subtitle">管理你的账户和偏好设置</p>
    </div>
    
    <div class="bento-grid">
      <!-- 个人信息卡片 -->
      <div class="bento-card profile-card">
        <div class="card-header">
          <div class="card-icon">
            <UserOutlined />
          </div>
          <h3 class="card-title">个人信息</h3>
        </div>
        <div class="card-content">
          <div class="info-item">
            <span class="info-label">邮箱</span>
            <span class="info-value">{{ userInfo?.email }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">显示名称</span>
            <span class="info-value">{{ userInfo?.username }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">账户状态</span>
            <span class="info-value">
              <a-badge status="success" text="正常" />
            </span>
          </div>
        </div>
        <div class="card-footer">
          <a-button type="primary" @click="$router.push('/console/profile')" class="action-btn">
            <EditOutlined />
            编辑个人资料
          </a-button>
        </div>
      </div>

      <!-- 快速操作卡片 -->
      <div class="bento-card actions-card">
        <div class="card-header">
          <div class="card-icon">
            <ThunderboltOutlined />
          </div>
          <h3 class="card-title">快速操作</h3>
        </div>
        <div class="card-content">
          <div class="action-grid">
            <button class="quick-action-btn" @click="$router.push('/console/profile')">
              <div class="action-icon">
                <UserOutlined />
              </div>
              <span class="action-text">个人资料</span>
            </button>
            <button class="quick-action-btn" @click="handleChangePassword">
              <div class="action-icon">
                <LockOutlined />
              </div>
              <span class="action-text">修改密码</span>
            </button>
            <button class="quick-action-btn" @click="$router.push('/console/realname')">
              <div class="action-icon">
                <SafetyOutlined />
              </div>
              <span class="action-text">实名认证</span>
            </button>
            <button class="quick-action-btn" @click="handleLogout">
              <div class="action-icon">
                <LogoutOutlined />
              </div>
              <span class="action-text">退出登录</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 最近活动卡片 -->
      <div class="bento-card activity-card">
        <div class="card-header">
          <div class="card-icon">
            <ClockCircleOutlined />
          </div>
          <h3 class="card-title">最近活动</h3>
        </div>
        <div class="card-content">
          <a-empty description="暂无活动记录" :image="Empty.PRESENTED_IMAGE_SIMPLE" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { 
  UserOutlined, 
  LockOutlined, 
  EditOutlined,
  ThunderboltOutlined,
  SafetyOutlined,
  LogoutOutlined,
  ClockCircleOutlined
} from '@ant-design/icons-vue'
import { message, Empty } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()
const userInfo = computed(() => authStore.userInfo)

const handleChangePassword = () => {
  message.info('密码修改功能开发中')
}

const handleLogout = () => {
  authStore.logout()
  router.push('/auth/login')
  message.success('已退出登录')
}
</script>

<style scoped>
.user-dashboard {
  padding: 0;
}

/* 页面头部 */
.dashboard-header {
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

/* Bento Grid 布局 */
.bento-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 24px;
  width: 100%;
}

/* 卡片基础样式 */
.bento-card {
  background: #FFFFFF;
  border-radius: 20px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(226, 232, 240, 0.8);
}

.bento-card:hover {
  box-shadow: 0 8px 24px rgba(37, 99, 235, 0.12), 0 2px 6px rgba(37, 99, 235, 0.08);
  transform: translateY(-2px);
}

/* 卡片头部 */
.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.card-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  font-size: 18px;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

/* 卡片内容 */
.card-content {
  margin-bottom: 20px;
}

/* 个人信息卡片 */
.profile-card {
  grid-column: span 1;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid #F1F5F9;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 14px;
  color: #64748B;
  font-weight: 500;
}

.info-value {
  font-size: 14px;
  color: #1E293B;
  font-weight: 500;
}

.card-footer {
  padding-top: 16px;
  border-top: 1px solid #F1F5F9;
}

.action-btn {
  width: 100%;
  height: 44px;
  border-radius: 12px;
  font-weight: 500;
  background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
}

.action-btn:hover {
  background: linear-gradient(135deg, #1d4ed8 0%, #2563EB 100%);
  box-shadow: 0 6px 16px rgba(37, 99, 235, 0.3);
  transform: translateY(-1px);
}

/* 快速操作卡片 */
.actions-card {
  grid-column: span 1;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px 16px;
  background: linear-gradient(135deg, #F8FAFC 0%, #F1F5F9 100%);
  border: 1px solid #E2E8F0;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.quick-action-btn:hover {
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
  border-color: #3B82F6;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.action-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  font-size: 20px;
}

.action-text {
  font-size: 13px;
  font-weight: 500;
  color: #1E293B;
}

/* 最近活动卡片 */
.activity-card {
  grid-column: span 2;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .activity-card {
    grid-column: span 1;
  }
}

@media (max-width: 768px) {
  .page-title {
    font-size: 28px;
  }

  .bento-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .bento-card {
    padding: 20px;
    border-radius: 16px;
  }

  .action-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .quick-action-btn {
    padding: 16px 12px;
  }

  .action-icon {
    width: 40px;
    height: 40px;
    font-size: 18px;
  }

  .action-text {
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .dashboard-header {
    margin-bottom: 24px;
  }

  .page-title {
    font-size: 24px;
  }

  .action-grid {
    grid-template-columns: 1fr;
  }
}
</style>
