<template>
  <div class="applications-view">
    <div class="page-header">
      <h2 class="page-title">应用管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <PlusOutlined />
        新建应用
      </a-button>
    </div>

    <a-table
      :columns="columns"
      :data-source="applications"
      :loading="loading"
      :pagination="pagination"
      :row-key="(record: any) => `${record.owner}/${record.name}`"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'clientId'">
          <a-space>
            <span style="font-family: monospace;">{{ record.clientId }}</span>
            <a-button 
              type="link" 
              size="small" 
              @click="copyToClipboard(record.clientId, '客户端 ID')"
            >
              <CopyOutlined />
            </a-button>
          </a-space>
        </template>
        <template v-else-if="column.key === 'clientSecret'">
          <a-space>
            <span v-if="visibleSecrets[record.clientId]" style="font-family: monospace;">
              {{ record.clientSecret }}
            </span>
            <span v-else style="color: #999;">
              ••••••••••••••••
            </span>
            <a-button 
              type="link" 
              size="small" 
              @click="toggleSecretVisibility(record.clientId)"
            >
              <EyeOutlined v-if="!visibleSecrets[record.clientId]" />
              <EyeInvisibleOutlined v-else />
            </a-button>
            <a-button 
              type="link" 
              size="small" 
              @click="copyToClipboard(record.clientSecret, '客户端密钥')"
            >
              <CopyOutlined />
            </a-button>
          </a-space>
        </template>
        <template v-else-if="column.key === 'grantTypes'">
          <a-tag 
            v-for="type in (Array.isArray(record.grantTypes) ? record.grantTypes : [])" 
            :key="type" 
            color="blue"
          >
            {{ type }}
          </a-tag>
          <span v-if="!record.grantTypes || record.grantTypes.length === 0" style="color: #999">-</span>
        </template>
        <template v-else-if="column.key === 'createdTime'">
          {{ formatDate(record.createdTime) }}
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="showEditModal(record)">
              编辑
            </a-button>
            <a-popconfirm
              title="确定要删除此应用吗？"
              ok-text="确定"
              cancel-text="取消"
              @confirm="handleDelete(record)"
            >
              <a-button type="link" size="small" danger>
                删除
              </a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
      :confirm-loading="modalLoading"
    >
      <a-form
        :model="formState"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 16 }"
      >
        <a-form-item
          label="应用名称"
          :rules="[{ required: true, message: '请输入应用名称' }]"
        >
          <a-input 
            v-model:value="formState.name" 
            placeholder="请输入应用名称（英文标识）"
            :disabled="!!editingApp"
          />
        </a-form-item>
        <a-form-item label="显示名称">
          <a-input v-model:value="formState.displayName" placeholder="请输入显示名称" />
        </a-form-item>
        <a-form-item label="应用图标">
          <a-input v-model:value="formState.logo" placeholder="请输入应用图标链接（显示在授权页）" />
          <template #extra>
            <span style="color: #8c8c8c; font-size: 12px;">
              图标将显示在授权页面，建议使用正方形图片，支持 http/https 链接
            </span>
          </template>
        </a-form-item>
        <a-form-item
          label="重定向 URI"
          :rules="[{ required: true, message: '请至少添加一个重定向 URI' }]"
        >
          <a-space direction="vertical" style="width: 100%">
            <a-input
              v-model:value="formState.redirectUriInput"
              placeholder="请输入重定向 URI"
              @press-enter="addRedirectUri"
            >
              <template #suffix>
                <a-button type="link" size="small" @click="addRedirectUri">添加</a-button>
              </template>
            </a-input>
            <a-tag
              v-for="(uri, index) in formState.redirectUris"
              :key="index"
              closable
              @close="removeRedirectUri(index)"
            >
              {{ uri }}
            </a-tag>
          </a-space>
        </a-form-item>
        <a-form-item label="授权类型">
          <a-select
            v-model:value="formState.grantTypes"
            mode="multiple"
            placeholder="请选择授权类型"
            style="width: 100%"
          >
            <a-select-option value="authorization_code">授权码模式</a-select-option>
            <a-select-option value="implicit">隐含模式</a-select-option>
            <a-select-option value="client_credentials">客户端模式</a-select-option>
            <a-select-option value="password">密码模式</a-select-option>
            <a-select-option value="refresh_token">刷新令牌</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="作用域">
          <a-select
            v-model:value="formState.scopes"
            mode="multiple"
            placeholder="请选择作用域"
            style="width: 100%"
          >
            <a-select-option value="openid">openid</a-select-option>
            <a-select-option value="profile">profile</a-select-option>
            <a-select-option value="email">email</a-select-option>
            <a-select-option value="offline_access">offline_access</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { PlusOutlined, EyeOutlined, EyeInvisibleOutlined, CopyOutlined } from '@ant-design/icons-vue'
import { adminApi } from '@/api/admin'
import type { Application, CreateApplicationRequest, UpdateApplicationRequest } from '@/api/types'
import { message } from 'ant-design-vue'

interface Column {
  title: string
  dataIndex?: string
  key: string
  width?: number
}

const columns: Column[] = [
  { title: '应用名称', dataIndex: 'name', key: 'name', width: 150 },
  { title: '显示名称', dataIndex: 'displayName', key: 'displayName', width: 150 },
  { title: '客户端 ID', key: 'clientId', width: 250 },
  { title: '客户端密钥', key: 'clientSecret', width: 250 },
  { title: '组织', dataIndex: 'organization', key: 'organization', width: 120 },
  { title: '授权类型', key: 'grantTypes', width: 200 },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime', width: 180 },
  { title: '操作', key: 'actions', width: 150 }
]

const applications = ref<Application[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const modalLoading = ref(false)
const editingApp = ref<Application | null>(null)
const visibleSecrets = ref<Record<string, boolean>>({})

const formState = reactive({
  name: '',
  displayName: '',
  logo: '',
  redirectUris: [] as string[],
  redirectUriInput: '',
  grantTypes: [] as string[],
  scopes: [] as string[]
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const modalTitle = ref('新建应用')

const loadData = async () => {
  loading.value = true
  try {
    const response = await adminApi.getApplications()
    console.log('Applications response:', response)
    if (response.status === 'ok' && response.data) {
      applications.value = response.data
      pagination.total = response.data.length
    }
  } catch (error) {
    console.error('Failed to load applications:', error)
    message.error('加载应用列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateModal = () => {
  editingApp.value = null
  modalTitle.value = '新建应用'
  formState.name = ''
  formState.displayName = ''
  formState.logo = ''
  formState.redirectUris = []
  formState.redirectUriInput = ''
  formState.grantTypes = []
  formState.scopes = []
  modalVisible.value = true
}

const showEditModal = (app: Application) => {
  editingApp.value = app
  modalTitle.value = '编辑应用'
  formState.name = app.name
  formState.displayName = app.displayName || ''
  formState.logo = app.logo || ''
  formState.redirectUris = Array.isArray(app.redirectUris) ? [...app.redirectUris] : []
  formState.redirectUriInput = ''
  formState.grantTypes = Array.isArray(app.grantTypes) ? [...app.grantTypes] : []
  formState.scopes = Array.isArray(app.scopes) ? [...app.scopes] : []
  modalVisible.value = true
}

const addRedirectUri = () => {
  const uri = formState.redirectUriInput.trim()
  if (uri && !formState.redirectUris.includes(uri)) {
    formState.redirectUris.push(uri)
    formState.redirectUriInput = ''
  }
}

const removeRedirectUri = (index: number) => {
  formState.redirectUris.splice(index, 1)
}

const handleModalOk = async () => {
  if (formState.redirectUris.length === 0) {
    message.error('请至少添加一个重定向 URI')
    return
  }

  modalLoading.value = true
  try {
    if (editingApp.value) {
      // 编辑模式：不发送 name 字段（主键不可修改）
      const updateData: UpdateApplicationRequest = {
        displayName: formState.displayName,
        logo: formState.logo,
        redirectUris: formState.redirectUris,
        grantTypes: formState.grantTypes,
        scopes: formState.scopes
      }
      await adminApi.updateApplication(editingApp.value.owner, editingApp.value.name, updateData)
      message.success('应用更新成功')
    } else {
      // 创建模式：需要 name 字段
      const createData: CreateApplicationRequest = {
        name: formState.name,
        displayName: formState.displayName,
        logo: formState.logo,
        redirectUris: formState.redirectUris,
        grantTypes: formState.grantTypes,
        scopes: formState.scopes
      }
      await adminApi.createApplication(createData)
      message.success('应用创建成功')
    }
    modalVisible.value = false
    loadData()
  } catch (error) {
    console.error('Application operation failed:', error)
    message.error(editingApp.value ? '更新失败' : '创建失败')
  } finally {
    modalLoading.value = false
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
}

const handleDelete = async (app: Application) => {
  try {
    await adminApi.deleteApplication(app.owner, app.name)
    message.success('应用删除成功')
    loadData()
  } catch (error) {
    console.error('Delete application failed:', error)
    message.error('删除失败')
  }
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
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

const toggleSecretVisibility = (clientId: string) => {
  visibleSecrets.value[clientId] = !visibleSecrets.value[clientId]
}

const copyToClipboard = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text)
    message.success(`${label}已复制到剪贴板`)
  } catch (err) {
    message.error('复制失败')
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.applications-view {
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
