<template>
  <div class="authorizations-view">
    <div class="page-header">
      <h1 class="page-title">
        <SafetyOutlined class="title-icon" />
        授权管理
      </h1>
      <p class="page-subtitle">管理你授权过的应用和访问令牌</p>
    </div>

    <a-tabs v-model:activeKey="activeTab" class="auth-tabs">
      <!-- 授权应用 -->
      <a-tab-pane key="applications" tab="授权应用">
        <a-spin :spinning="appsLoading">
          <div v-if="applications.length === 0" class="empty-state">
            <AppstoreOutlined class="empty-icon" />
            <p class="empty-text">暂无授权应用</p>
            <p class="empty-hint">当你使用第三方应用登录时，会在这里显示</p>
          </div>

          <div v-else class="app-grid">
            <a-card
              v-for="app in applications"
              :key="`${app.owner}/${app.name}`"
              class="app-card"
              :bordered="false"
            >
              <div class="app-header">
                <a-avatar :size="64" :src="app.logo" class="app-logo">
                  <template #icon><AppstoreOutlined /></template>
                </a-avatar>
                <div class="app-info">
                  <h3 class="app-name">{{ app.displayName }}</h3>
                  <p class="app-desc">{{ app.description || '暂无描述' }}</p>
                </div>
              </div>

              <div class="app-meta">
                <div class="meta-item">
                  <span class="meta-label">首次授权</span>
                  <span class="meta-value">{{ formatDate(app.firstAuth) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">最后授权</span>
                  <span class="meta-value">{{ formatDate(app.lastAuth) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">令牌数量</span>
                  <span class="meta-value">{{ app.tokenCount }}</span>
                </div>
              </div>

              <div class="app-scopes">
                <span class="scopes-label">授权范围：</span>
                <a-tag v-for="scope in app.scopes" :key="scope" color="blue">
                  {{ scope }}
                </a-tag>
              </div>

              <div class="app-actions">
                <a-button
                  v-if="app.homepageUrl"
                  type="link"
                  :href="app.homepageUrl"
                  target="_blank"
                >
                  <LinkOutlined />
                  访问应用
                </a-button>
              </div>
            </a-card>
          </div>
        </a-spin>
      </a-tab-pane>

      <!-- 访问令牌 -->
      <a-tab-pane key="tokens" tab="访问令牌">
        <a-spin :spinning="tokensLoading">
          <div v-if="tokens.length === 0" class="empty-state">
            <KeyOutlined class="empty-icon" />
            <p class="empty-text">暂无访问令牌</p>
            <p class="empty-hint">当你授权应用访问你的账户时，会生成访问令牌</p>
          </div>

          <a-table
            v-else
            :columns="tokenColumns"
            :data-source="tokens"
            :pagination="{ pageSize: 10 }"
            :row-key="(record: any) => record.name"
            class="tokens-table"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'application'">
                <div class="app-cell">
                  <a-avatar :size="32" :src="record.applicationLogo">
                    <template #icon><AppstoreOutlined /></template>
                  </a-avatar>
                  <span class="app-cell-name">{{ record.applicationName }}</span>
                </div>
              </template>

              <template v-else-if="column.key === 'status'">
                <a-tag v-if="record.isRevoked" color="red">已撤销</a-tag>
                <a-tag v-else-if="record.isExpired" color="orange">已过期</a-tag>
                <a-tag v-else color="green">有效</a-tag>
              </template>

              <template v-else-if="column.key === 'expiresAt'">
                {{ formatDate(record.expiresAt) }}
              </template>

              <template v-else-if="column.key === 'createdTime'">
                {{ formatDate(record.createdTime) }}
              </template>

              <template v-else-if="column.key === 'actions'">
                <a-button
                  v-if="!record.isRevoked && !record.isExpired"
                  type="link"
                  danger
                  @click="handleRevokeToken(record)"
                >
                  撤销
                </a-button>
                <span v-else class="disabled-text">-</span>
              </template>
            </template>
          </a-table>
        </a-spin>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authApi } from '@/api/auth'
import {
  SafetyOutlined,
  AppstoreOutlined,
  KeyOutlined,
  LinkOutlined
} from '@ant-design/icons-vue'
import { message, Modal } from 'ant-design-vue'

const activeTab = ref('applications')
const appsLoading = ref(false)
const tokensLoading = ref(false)

const applications = ref<any[]>([])
const tokens = ref<any[]>([])

const tokenColumns = [
  {
    title: '应用',
    key: 'application',
    dataIndex: 'applicationName',
    width: 200
  },
  {
    title: '授权范围',
    key: 'scope',
    dataIndex: 'scope',
    width: 150
  },
  {
    title: '状态',
    key: 'status',
    width: 100
  },
  {
    title: '创建时间',
    key: 'createdTime',
    dataIndex: 'createdTime',
    width: 180
  },
  {
    title: '过期时间',
    key: 'expiresAt',
    dataIndex: 'expiresAt',
    width: 180
  },
  {
    title: '操作',
    key: 'actions',
    width: 100
  }
]

const loadApplications = async () => {
  appsLoading.value = true
  try {
    const response = await authApi.getUserApplications()
    if (response.status === 'ok' && response.data) {
      applications.value = response.data as unknown as any[]
    }
  } catch (error: any) {
    console.error('Failed to load applications:', error)
    message.error(error.message || '加载授权应用失败')
  } finally {
    appsLoading.value = false
  }
}

const loadTokens = async () => {
  tokensLoading.value = true
  try {
    const response = await authApi.getUserTokens()
    if (response.status === 'ok' && response.data) {
      tokens.value = response.data as unknown as any[]
    }
  } catch (error: any) {
    console.error('Failed to load tokens:', error)
    message.error(error.message || '加载访问令牌失败')
  } finally {
    tokensLoading.value = false
  }
}

const handleRevokeToken = (token: any) => {
  Modal.confirm({
    title: '确认撤销令牌',
    content: `确定要撤销应用 "${token.applicationName}" 的访问令牌吗？撤销后该应用将无法继续访问你的账户。`,
    okText: '确认撤销',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        const response = await authApi.revokeUserToken(token.name)
        if (response.status === 'ok') {
          message.success('令牌已撤销')
          await loadTokens()
        } else {
          message.error(response.msg || '撤销失败')
        }
      } catch (error: any) {
        console.error('Failed to revoke token:', error)
        message.error(error.message || '撤销失败')
      }
    }
  })
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    const date = new Date(dateStr)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateStr
  }
}

onMounted(() => {
  loadApplications()
  loadTokens()
})
</script>

<style scoped>
.authorizations-view {
  padding: 0;
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

.page-header {
  margin-bottom: 32px;
  animation: slideDown 0.5s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  color: #1f1f1f;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
}

.page-title .title-icon {
  margin-right: 12px;
  color: #ec4899;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

.page-subtitle {
  font-size: 15px;
  color: #64748B;
  margin: 0;
}

.auth-tabs {
  background: #ffffff;
  border-radius: 20px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: box-shadow 0.3s ease;
}

.auth-tabs:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 80px 20px;
  animation: fadeIn 0.6s ease-out;
}

.empty-icon {
  font-size: 64px;
  color: #d1d5db;
  margin-bottom: 16px;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.empty-text {
  font-size: 18px;
  font-weight: 600;
  color: #6b7280;
  margin: 0 0 8px 0;
  animation: fadeIn 0.8s ease-out;
}

.empty-hint {
  font-size: 14px;
  color: #9ca3af;
  margin: 0;
  animation: fadeIn 1s ease-out;
}

/* 应用网格 */
.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 24px;
}

.app-card {
  border-radius: 16px;
  border: 1px solid #e5e7eb;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  animation: slideUp 0.5s ease-out backwards;
  animation-delay: calc(var(--index, 0) * 0.1s);
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.app-card:hover {
  box-shadow: 0 12px 32px rgba(236, 72, 153, 0.12);
  border-color: #f472b6;
  transform: translateY(-4px);
}

.app-card:active {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(236, 72, 153, 0.1);
}

.app-header {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.app-logo {
  flex-shrink: 0;
  border: 2px solid #fce7f3;
  transition: all 0.3s ease;
}

.app-card:hover .app-logo {
  transform: scale(1.05);
  border-color: #f9a8d4;
  box-shadow: 0 4px 12px rgba(236, 72, 153, 0.2);
}

.app-info {
  flex: 1;
  min-width: 0;
}

.app-name {
  font-size: 18px;
  font-weight: 600;
  color: #1f1f1f;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-desc {
  font-size: 14px;
  color: #64748B;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.app-meta {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: #f8fafc;
  border-radius: 12px;
  transition: all 0.3s ease;
}

.app-card:hover .app-meta {
  background: #fdf2f8;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.meta-label {
  font-size: 13px;
  color: #64748B;
}

.meta-value {
  font-size: 14px;
  font-weight: 500;
  color: #1f1f1f;
}

.app-scopes {
  margin-bottom: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.scopes-label {
  font-size: 13px;
  color: #64748B;
  margin-right: 8px;
}

.app-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
}

.app-actions :deep(.ant-btn-link) {
  transition: all 0.3s ease;
}

.app-actions :deep(.ant-btn-link:hover) {
  transform: translateX(4px);
}

/* 令牌表格 */
.tokens-table {
  background: transparent;
}

.app-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.app-cell-name {
  font-weight: 500;
  color: #1f1f1f;
}

.disabled-text {
  color: #9ca3af;
}

/* 响应式 */
@media (max-width: 768px) {
  .page-title {
    font-size: 24px;
  }

  .app-grid {
    grid-template-columns: 1fr;
  }

  .auth-tabs {
    padding: 16px;
  }
}
</style>
