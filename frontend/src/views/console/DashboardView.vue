<template>
  <div class="dashboard-view">
    <div class="dashboard-header">
      <h2 class="page-title">仪表板</h2>
      <p class="page-subtitle">系统概览和统计数据</p>
    </div>
    
    <!-- 统计卡片网格 -->
    <div class="stats-grid">
      <div class="stat-card stat-card-users">
        <div class="stat-icon">
          <UserOutlined />
        </div>
        <div class="stat-content">
          <div class="stat-label">用户总数</div>
          <div class="stat-value">{{ statistics.userCount }}</div>
          <div class="stat-suffix">人</div>
        </div>
      </div>

      <div class="stat-card stat-card-apps">
        <div class="stat-icon">
          <AppstoreOutlined />
        </div>
        <div class="stat-content">
          <div class="stat-label">应用总数</div>
          <div class="stat-value">{{ statistics.applicationCount }}</div>
          <div class="stat-suffix">个</div>
        </div>
      </div>

      <div class="stat-card stat-card-tokens">
        <div class="stat-icon">
          <KeyOutlined />
        </div>
        <div class="stat-content">
          <div class="stat-label">令牌总数</div>
          <div class="stat-value">{{ statistics.tokenCount }}</div>
          <div class="stat-suffix">个</div>
        </div>
      </div>

      <div class="stat-card stat-card-active">
        <div class="stat-icon">
          <CheckCircleOutlined />
        </div>
        <div class="stat-content">
          <div class="stat-label">活动令牌</div>
          <div class="stat-value">{{ statistics.activeTokenCount }}</div>
          <div class="stat-suffix">个</div>
        </div>
      </div>
    </div>

    <!-- 系统信息卡片 -->
    <div class="system-card">
      <div class="card-header">
        <div class="card-icon">
          <SettingOutlined />
        </div>
        <h3 class="card-title">系统信息</h3>
      </div>
      <div class="system-info-grid">
        <div class="info-item">
          <div class="info-label">系统版本</div>
          <div class="info-value">
            <a-tag color="blue">{{ systemInfo.version }}</a-tag>
          </div>
        </div>
        <div class="info-item">
          <div class="info-label">运行时间</div>
          <div class="info-value">{{ formatUptime(systemInfo.uptime) }}</div>
        </div>
        <div class="info-item">
          <div class="info-label">Redis 连接</div>
          <div class="info-value">
            <a-badge 
              :status="systemInfo.redisConnected ? 'success' : 'error'" 
              :text="systemInfo.redisConnected ? '已连接' : '未连接'" 
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  UserOutlined,
  AppstoreOutlined,
  KeyOutlined,
  CheckCircleOutlined,
  SettingOutlined
} from '@ant-design/icons-vue'
import { adminApi } from '@/api/admin'
import type { Statistics, SystemInfo } from '@/api/types'
import { message } from 'ant-design-vue'

const statistics = ref<Statistics>({
  userCount: 0,
  applicationCount: 0,
  tokenCount: 0,
  activeTokenCount: 0
})

const systemInfo = ref<SystemInfo>({
  version: '-',
  uptime: 0,
  redisConnected: false
})

const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  
  if (days > 0) {
    return `${days}天${hours}小时`
  } else if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

const loadData = async () => {
  try {
    const [statsResponse, infoResponse] = await Promise.all([
      adminApi.getStatistics(),
      adminApi.getSystemInfo()
    ])
    
    console.log('Stats response:', statsResponse)
    console.log('Info response:', infoResponse)
    
    if (statsResponse.status === 'ok' && statsResponse.data) {
      statistics.value = statsResponse.data
    }
    
    if (infoResponse.status === 'ok' && infoResponse.data) {
      systemInfo.value = infoResponse.data
    }
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
    message.error('加载数据失败')
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard-view {
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

/* 统计卡片网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

/* 统计卡片 */
.stat-card {
  background: #FFFFFF;
  border-radius: 20px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(226, 232, 240, 0.8);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 100px;
  height: 100px;
  border-radius: 50%;
  opacity: 0.1;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.12), 0 2px 6px rgba(0, 0, 0, 0.08);
}

/* 统计卡片图标 */
.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #FFFFFF;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

/* 统计卡片内容 */
.stat-content {
  flex: 1;
  position: relative;
  z-index: 1;
}

.stat-label {
  font-size: 13px;
  color: #64748B;
  font-weight: 500;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-suffix {
  font-size: 13px;
  color: #94A3B8;
  font-weight: 500;
}

/* 用户卡片 */
.stat-card-users .stat-icon {
  background: linear-gradient(135deg, #10B981 0%, #34D399 100%);
}

.stat-card-users .stat-value {
  color: #10B981;
}

.stat-card-users::before {
  background: #10B981;
}

/* 应用卡片 */
.stat-card-apps .stat-icon {
  background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
}

.stat-card-apps .stat-value {
  color: #2563EB;
}

.stat-card-apps::before {
  background: #2563EB;
}

/* 令牌卡片 */
.stat-card-tokens .stat-icon {
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
}

.stat-card-tokens .stat-value {
  color: #7C3AED;
}

.stat-card-tokens::before {
  background: #7C3AED;
}

/* 活动令牌卡片 */
.stat-card-active .stat-icon {
  background: linear-gradient(135deg, #F97316 0%, #FB923C 100%);
}

.stat-card-active .stat-value {
  color: #F97316;
}

.stat-card-active::before {
  background: #F97316;
}

/* 系统信息卡片 */
.system-card {
  background: #FFFFFF;
  border-radius: 20px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(226, 232, 240, 0.8);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
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

/* 系统信息网格 */
.system-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.info-item {
  padding: 16px;
  background: linear-gradient(135deg, #F8FAFC 0%, #F1F5F9 100%);
  border-radius: 12px;
  border: 1px solid #E2E8F0;
}

.info-label {
  font-size: 13px;
  color: #64748B;
  font-weight: 500;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 15px;
  color: #1E293B;
  font-weight: 600;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .page-title {
    font-size: 28px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .stat-card {
    padding: 20px;
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    font-size: 24px;
  }

  .stat-value {
    font-size: 28px;
  }

  .system-info-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}

@media (max-width: 480px) {
  .dashboard-header {
    margin-bottom: 24px;
  }

  .page-title {
    font-size: 24px;
  }

  .stat-card {
    flex-direction: column;
    text-align: center;
    gap: 16px;
  }

  .stat-icon {
    width: 48px;
    height: 48px;
    font-size: 20px;
  }

  .stat-value {
    font-size: 24px;
  }
}
</style>
