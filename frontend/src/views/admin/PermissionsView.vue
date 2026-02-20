<template>
  <div class="permissions-view">
    <div class="page-header">
      <h2 class="page-title">权限管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <PlusOutlined />
        新建权限
      </a-button>
    </div>

    <a-table
      :columns="columns"
      :data-source="permissions"
      :loading="loading"
      :pagination="pagination"
      row-key="id"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'actions'">
          <a-popconfirm
            title="确定要删除此权限吗？"
            ok-text="确定"
            cancel-text="取消"
            @confirm="handleDelete(record)"
          >
            <a-button type="link" size="small" danger>
              删除
            </a-button>
          </a-popconfirm>
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
          label="权限名称"
          :rules="[{ required: true, message: '请输入权限名称' }]"
        >
          <a-input v-model:value="formState.name" placeholder="请输入权限名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea v-model:value="formState.description" placeholder="请输入权限描述" :rows="3" />
        </a-form-item>
        <a-form-item
          label="资源"
          :rules="[{ required: true, message: '请输入资源' }]"
        >
          <a-input v-model:value="formState.resource" placeholder="例如：users, applications" />
        </a-form-item>
        <a-form-item
          label="操作"
          :rules="[{ required: true, message: '请输入操作' }]"
        >
          <a-input v-model:value="formState.action" placeholder="例如：read, write, delete" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { adminApi } from '@/api/admin'
import type { Permission, CreatePermissionRequest } from '@/api/types'
import { message } from 'ant-design-vue'

interface Column {
  title: string
  dataIndex?: string
  key: string
  width?: number
}

const columns: Column[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '权限名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '资源', dataIndex: 'resource', key: 'resource' },
  { title: '操作', dataIndex: 'action', key: 'action' },
  { title: '操作', key: 'actions', width: 100 }
]

const permissions = ref<Permission[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const modalLoading = ref(false)

const formState = reactive({
  name: '',
  description: '',
  resource: '',
  action: ''
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const modalTitle = ref('新建权限')

const loadData = async () => {
  loading.value = true
  try {
    const response = await adminApi.getPermissions()
    if (response.status === 'ok') {
      permissions.value = response.data || []
      pagination.total = permissions.value.length
    }
  } catch (error) {
    console.error('Failed to load permissions:', error)
    message.error('加载权限列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateModal = () => {
  formState.name = ''
  formState.description = ''
  formState.resource = ''
  formState.action = ''
  modalTitle.value = '新建权限'
  modalVisible.value = true
}

const handleModalOk = async () => {
  modalLoading.value = true
  try {
    const createData: CreatePermissionRequest = {
      name: formState.name,
      description: formState.description,
      resource: formState.resource,
      action: formState.action
    }
    await adminApi.createPermission(createData)
    message.success('权限创建成功')
    modalVisible.value = false
    loadData()
  } catch (error) {
    console.error('Create permission failed:', error)
    message.error('创建失败')
  } finally {
    modalLoading.value = false
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
}

const handleDelete = async (permission: Permission) => {
  try {
    await adminApi.deletePermission(permission.id)
    message.success('权限删除成功')
    loadData()
  } catch (error) {
    console.error('Delete permission failed:', error)
    message.error('删除失败')
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
.permissions-view {
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
