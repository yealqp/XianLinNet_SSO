<template>
  <div class="dashboard-view">
    <h2 class="page-title">仪表板</h2>
    
    <a-row :gutter="[16, 16]">
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card">
          <a-statistic
            title="用户总数"
            :value="statistics.userCount"
            :value-style="{ color: '#3f8600' }"
          >
            <template #prefix>
              <UserOutlined />
            </template>
            <template #suffix>
              <span>人</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card">
          <a-statistic
            title="应用总数"
            :value="statistics.applicationCount"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <AppstoreOutlined />
            </template>
            <template #suffix>
              <span>个</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card">
          <a-statistic
            title="令牌总数"
            :value="statistics.tokenCount"
            :value-style="{ color: '#722ed1' }"
          >
            <template #prefix>
              <KeyOutlined />
            </template>
            <template #suffix>
              <span>个</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="12" :lg="6">
        <a-card class="stat-card">
          <a-statistic
            title="活动令牌"
            :value="statistics.activeTokenCount"
            :value-style="{ color: '#fa8c16' }"
          >
            <template #prefix>
              <CheckCircleOutlined />
            </template>
            <template #suffix>
              <span>个</span>
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <a-card title="系统信息" class="mt-4">
      <a-descriptions bordered :column="{ xs: 1, sm: 2, lg: 4 }">
        <a-descriptions-item label="系统版本">
          <a-tag color="blue">{{ systemInfo.version }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="运行时间">
          {{ formatUptime(systemInfo.uptime) }}
        </a-descriptions-item>
        <a-descriptions-item label="Redis 连接">
          <a-badge :status="systemInfo.redisConnected ? 'success' : 'error'" :text="systemInfo.redisConnected ? '已连接' : '未连接'" />
        </a-descriptions-item>
      </a-descriptions>
    </a-card>
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
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 24px;
  color: #1f1f1f;
}

.stat-card {
  text-align: center;
}

.mt-4 {
  margin-top: 24px;
}
</style>
