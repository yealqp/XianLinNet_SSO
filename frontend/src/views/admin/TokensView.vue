<template>
  <div class="tokens-view">
    <div class="page-header">
      <h2 class="page-title">令牌管理</h2>
    </div>

    <a-table
      :columns="columns"
      :data-source="tokens"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: any) => `${record.owner}/${record.name}`"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'actions'">
          <a-space>
            <a-popconfirm
              title="确定要撤销此令牌吗？"
              ok-text="确定"
              cancel-text="取消"
              @confirm="handleRevoke(record)"
            >
              <a-button type="link" size="small" danger>
                撤销
              </a-button>
            </a-popconfirm>
            <a-popconfirm
              title="确定要撤销此用户的所有令牌吗？"
              ok-text="确定"
              cancel-text="取消"
              @confirm="handleRevokeAll(record)"
            >
              <a-button type="link" size="small" danger>
                撤销所有
              </a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import type { Token } from '@/api/types'
import { message } from 'ant-design-vue'

interface Column {
  title: string
  dataIndex?: string
  key: string
  width?: number
}

const columns: Column[] = [
  { title: '名称', dataIndex: 'name', key: 'name', width: 150 },
  { title: '用户', dataIndex: 'user', key: 'user', width: 120 },
  { title: '应用', dataIndex: 'application', key: 'application', width: 150 },
  { title: '作用域', dataIndex: 'scope', key: 'scope', width: 200 },
  { title: '令牌类型', dataIndex: 'tokenType', key: 'tokenType', width: 100 },
  { title: '过期时间(秒)', dataIndex: 'expiresIn', key: 'expiresIn', width: 120 },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime', width: 180 },
  { title: '操作', key: 'actions', width: 200 }
]

const tokens = ref<Token[]>([])
const loading = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const loadData = async () => {
  loading.value = true
  try {
    const response = await adminApi.getTokens()
    if (response.status === 'ok') {
      tokens.value = response.data || []
      pagination.total = tokens.value.length
    }
  } catch (error) {
    console.error('Failed to load tokens:', error)
    message.error('加载令牌列表失败')
  } finally {
    loading.value = false
  }
}

const handleRevoke = async (token: Token) => {
  try {
    await adminApi.revokeToken({ owner: token.owner, name: token.name })
    message.success('令牌撤销成功')
    loadData()
  } catch (error) {
    console.error('Revoke token failed:', error)
    message.error('撤销失败')
  }
}

const handleRevokeAll = async (token: Token) => {
  try {
    await adminApi.revokeUserTokens({ owner: token.owner, username: token.user })
    message.success('用户所有令牌已撤销')
    loadData()
  } catch (error) {
    console.error('Revoke all tokens failed:', error)
    message.error('撤销失败')
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.tokens-view {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #1f1f1f;
  margin: 0;
}
</style>
