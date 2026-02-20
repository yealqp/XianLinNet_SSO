<template>
  <div class="users-view">
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <PlusOutlined />
        新建用户
      </a-button>
    </div>

    <a-table
      :columns="columns"
      :data-source="users"
      :loading="loading"
      :pagination="pagination"
      row-key="id"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'isAdmin'">
          <a-tag :color="record.isAdmin ? 'green' : 'default'">
            {{ record.isAdmin ? '是' : '否' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'isRealName'">
          <a-tag :color="record.isRealName ? 'blue' : 'default'">
            {{ record.isRealName ? '已认证' : '未认证' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'createdTime'">
          {{ formatDate(record.createdTime) }}
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="showEditModal(record)">
              编辑
            </a-button>
            <a-button 
              v-if="record.isRealName" 
              type="link" 
              size="small" 
              @click="showRealNameInfo(record)"
            >
              查看实名
            </a-button>
            <a-popconfirm
              title="确定要删除此用户吗？"
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
          label="用户名"
          :rules="[{ required: true, message: '请输入用户名' }]"
        >
          <a-input v-model:value="formState.username" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item
          v-if="!editingUser"
          label="密码"
          :rules="[{ required: true, message: '请输入密码' }]"
        >
          <a-input-password v-model:value="formState.password" placeholder="请输入密码" />
        </a-form-item>
        <a-form-item
          label="邮箱"
          :rules="[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '请输入有效的邮箱' }]"
        >
          <a-input v-model:value="formState.email" placeholder="请输入邮箱" />
        </a-form-item>
        <a-form-item label="QQ号">
          <a-input v-model:value="formState.qq" placeholder="请输入QQ号" />
        </a-form-item>
        <a-form-item label="头像 URL">
          <a-input v-model:value="formState.avatar" placeholder="请输入头像 URL" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="realNameModalVisible"
      title="实名信息"
      :footer="null"
      @cancel="realNameModalVisible = false"
    >
      <a-spin :spinning="realNameLoading">
        <a-descriptions bordered :column="1">
          <a-descriptions-item label="用户名">
            {{ currentRealNameUser?.username }}
          </a-descriptions-item>
          <a-descriptions-item label="真实姓名">
            {{ realNameInfo.name || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="身份证号">
            {{ realNameInfo.idcard || '-' }}
          </a-descriptions-item>
        </a-descriptions>
      </a-spin>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { adminApi } from '@/api/admin'
import type { User, CreateUserRequest, UpdateUserRequest } from '@/api/types'
import { message } from 'ant-design-vue'

interface Column {
  title: string
  dataIndex?: string
  key: string
  width?: number
}

const columns: Column[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '用户名', dataIndex: 'username', key: 'username', width: 120 },
  { title: '邮箱', dataIndex: 'email', key: 'email', width: 200 },
  { title: 'QQ号', dataIndex: 'qq', key: 'qq', width: 120 },
  { title: '管理员', dataIndex: 'isAdmin', key: 'isAdmin', width: 80 },
  { title: '实名认证', dataIndex: 'isRealName', key: 'isRealName', width: 100 },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime', width: 180 },
  { title: '操作', key: 'actions', width: 150 }
]

const users = ref<User[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const modalLoading = ref(false)
const editingUser = ref<User | null>(null)

const realNameModalVisible = ref(false)
const realNameLoading = ref(false)
const currentRealNameUser = ref<User | null>(null)
const realNameInfo = ref({
  name: '',
  idcard: ''
})

const formState = reactive({
  username: '',
  password: '',
  email: '',
  qq: '',
  avatar: ''
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const modalTitle = ref('新建用户')

const loadData = async () => {
  loading.value = true
  try {
    const response = await adminApi.getUsers()
    console.log('Users response:', response)
    if (response.status === 'ok' && response.data) {
      users.value = response.data
      pagination.total = response.data.length
    }
  } catch (error) {
    console.error('Failed to load users:', error)
    message.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateModal = () => {
  editingUser.value = null
  modalTitle.value = '新建用户'
  formState.username = ''
  formState.password = ''
  formState.email = ''
  formState.qq = ''
  formState.avatar = ''
  modalVisible.value = true
}

const showEditModal = (user: User) => {
  editingUser.value = user
  modalTitle.value = '编辑用户'
  formState.username = user.username
  formState.password = ''
  formState.email = user.email || ''
  formState.qq = user.qq || ''
  formState.avatar = user.avatar || ''
  modalVisible.value = true
}

const handleModalOk = async () => {
  modalLoading.value = true
  try {
    if (editingUser.value) {
      const updateData: UpdateUserRequest = {
        username: formState.username,
        email: formState.email,
        qq: formState.qq,
        avatar: formState.avatar
      }
      await adminApi.updateUser(editingUser.value.id, updateData)
      message.success('用户更新成功')
    } else {
      const createData: CreateUserRequest = {
        username: formState.username,
        password: formState.password,
        email: formState.email,
        qq: formState.qq,
        avatar: formState.avatar
      }
      await adminApi.createUser(createData)
      message.success('用户创建成功')
    }
    modalVisible.value = false
    loadData()
  } catch (error) {
    console.error('User operation failed:', error)
    message.error(editingUser.value ? '更新失败' : '创建失败')
  } finally {
    modalLoading.value = false
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
}

const handleDelete = async (user: User) => {
  try {
    await adminApi.deleteUser(user.id)
    message.success('用户删除成功')
    loadData()
  } catch (error) {
    console.error('Delete user failed:', error)
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

const showRealNameInfo = async (user: User) => {
  if (!user.isRealName) {
    message.warning('该用户未进行实名认证')
    return
  }

  currentRealNameUser.value = user
  realNameModalVisible.value = true
  realNameLoading.value = true

  try {
    const response = await adminApi.getRealNameInfo(user.id)
    console.log('RealName response:', response)
    if (response.status === 'ok' && response.data) {
      realNameInfo.value = response.data
    }
  } catch (error) {
    console.error('Failed to load real name info:', error)
    message.error('获取实名信息失败')
  } finally {
    realNameLoading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.users-view {
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
