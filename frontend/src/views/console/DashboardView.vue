<template>
  <div class="dashboard-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">首页</h1>
    </div>
    
    <!-- 统计卡片网格 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon-wrapper">
          <div class="stat-icon stat-icon-primary">
            <UserOutlined />
          </div>
        </div>
        <div class="stat-info">
          <div class="stat-label">用户总数</div>
          <div class="stat-value">{{ statistics.userCount }} 人</div>
        </div>
        <div class="stat-badge">
          <a-tag color="pink">查看</a-tag>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper">
          <div class="stat-icon stat-icon-success">
            <AppstoreOutlined />
          </div>
        </div>
        <div class="stat-info">
          <div class="stat-label">应用总数</div>
          <div class="stat-value">{{ statistics.applicationCount }} 个</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper">
          <div class="stat-icon stat-icon-warning">
            <KeyOutlined />
          </div>
        </div>
        <div class="stat-info">
          <div class="stat-label">令牌总数 / 可用令牌</div>
          <div class="stat-value">{{ statistics.tokenCount }} / {{ statistics.activeTokenCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper">
          <div class="stat-icon stat-icon-info">
            <CheckCircleOutlined />
          </div>
        </div>
        <div class="stat-info">
          <div class="stat-label">活动令牌</div>
          <div class="stat-value">{{ statistics.activeTokenCount }}</div>
        </div>
      </div>
    </div>

    <!-- 系统信息卡片 -->
    <div class="system-section">
      <h3 class="section-title">系统信息</h3>
      <div class="system-card">
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  UserOutlined,
  AppstoreOutlined,
  KeyOutlined,
  CheckCircleOutlined
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
  max-width: 1200px;
}

/* 页面头部 */
.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1f1f1f;
  margin: 0;
}

/* 统计卡片网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.2s ease;
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.stat-icon-wrapper {
  flex-shrink: 0;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #ffffff;
}

.stat-icon-primary {
  background: linear-gradient(135deg, #ec4899 0%, #f472b6 100%);
}

.stat-icon-success {
  background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
}

.stat-icon-warning {
  background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%);
}

.stat-icon-info {
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-label {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #1f1f1f;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.stat-badge {
  flex-shrink: 0;
}

/* 系统信息区域 */
.system-section {
  margin-top: 24px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f1f1f;
  margin: 0 0 16px 0;
}

.system-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
}

.system-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.info-value {
  font-size: 15px;
  color: #1f1f1f;
  font-weight: 600;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .stat-card {
    padding: 16px;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
    font-size: 18px;
  }

  .stat-value {
    font-size: 16px;
  }

  .system-info-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
