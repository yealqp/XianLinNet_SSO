<template>
  <div class="roles-view">
    <div class="page-header">
      <h2 class="page-title">角色管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <PlusOutlined />
        新建角色
      </a-button>
    </div>

    <a-table
      :columns="columns"
      :data-source="roles"
      :loading="loading"
      :pagination="pagination"
      row-key="id"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'permissions'">
          <a-tag v-for="perm in record.permissions" :key="perm.id" color="blue">
            {{ perm.name }}
          </a-tag>
        </template>
        <template v-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="showPermissionsModal(record)">
              查看权限
            </a-button>
            <a-popconfirm
              title="确定要删除此角色吗？"
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
          label="角色名称"
          :rules="[{ required: true, message: '请输入角色名称' }]"
        >
          <a-input v-model:value="formState.name" placeholder="请输入角色名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea v-model:value="formState.description" placeholder="请输入角色描述" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="permissionsModalVisible"
      title="角色权限"
      width="800px"
      @cancel="handlePermissionsModalCancel"
    >
      <a-row>
        <a-col :span="12">
          <h3>可用权限</h3>
          <a-transfer
            :data-source="availablePermissions"
            :target-keys="selectedPermissionKeys"
            :render="(item: any) => item.name"
            @change="handleTransferChange"
            :titles="['可用权限', '已选权限']"
          />
        </a-col>
        <a-col :span="12">
          <h3>已选权限</h3>
          <a-list
            size="small"
            :data-source="selectedPermissions"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <a-space>
                  <a-tag color="blue">{{ item.name }}</a-tag>
                  <span>{{ item.description }}</span>
                  <a-button
                    type="link"
                    danger
                    size="small"
                    @click="removePermission(item)"
                  >
                    移除
                  </a-button>
                </a-space>
              </a-list-item>
            </template>
          </a-list>
        </a-col>
      </a-row>
      <template #footer>
        <a-button @click="handlePermissionsModalCancel">关闭</a-button>
        <a-button type="primary" @click="handleSavePermissions">保存</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { adminApi } from '@/api/admin'
import type { Role, Permission, CreateRoleRequest } from '@/api/types'
import { message } from 'ant-design-vue'

interface Column {
  title: string
  dataIndex?: string
  key: string
  width?: number
}

const columns: Column[] = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '角色名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '权限', key: 'permissions' },
  { title: '操作', key: 'actions', width: 200 }
]

const roles = ref<Role[]>([])
const allPermissions = ref<Permission[]>([])
const loading = ref(false)
const modalVisible = ref(false)
const modalLoading = ref(false)
const permissionsModalVisible = ref(false)
const currentRole = ref<Role | null>(null)
const selectedPermissionKeys = ref<string[]>([])

const formState = reactive({
  name: '',
  description: ''
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const modalTitle = ref('新建角色')

const availablePermissions = computed(() => {
  return allPermissions.value.map(p => ({
    key: String(p.id),
    ...p
  }))
})

const selectedPermissions = computed(() => {
  return allPermissions.value.filter(p =>
    selectedPermissionKeys.value.includes(String(p.id))
  )
})

const loadData = async () => {
  loading.value = true
  try {
    const [rolesResponse, permsResponse] = await Promise.all([
      adminApi.getRoles(),
      adminApi.getPermissions()
    ])
    if (rolesResponse.status === 'ok') {
      roles.value = rolesResponse.data || []
      pagination.total = roles.value.length
    }
    if (permsResponse.status === 'ok') {
      allPermissions.value = permsResponse.data || []
    }
  } catch (error) {
    console.error('Failed to load roles:', error)
    message.error('加载角色列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateModal = () => {
  currentRole.value = null
  modalTitle.value = '新建角色'
  formState.name = ''
  formState.description = ''
  modalVisible.value = true
}

const showPermissionsModal = (role: Role) => {
  currentRole.value = role
  selectedPermissionKeys.value = role.permissions.map(p => String(p.id))
  permissionsModalVisible.value = true
}

const handleModalOk = async () => {
  modalLoading.value = true
  try {
    const createData: CreateRoleRequest = {
      name: formState.name,
      description: formState.description
    }
    await adminApi.createRole(createData)
    message.success('角色创建成功')
    modalVisible.value = false
    loadData()
  } catch (error) {
    console.error('Create role failed:', error)
    message.error('创建失败')
  } finally {
    modalLoading.value = false
  }
}

const handleModalCancel = () => {
  modalVisible.value = false
}

const handlePermissionsModalCancel = () => {
  permissionsModalVisible.value = false
  currentRole.value = null
}

const handleTransferChange = (keys: string[]) => {
  selectedPermissionKeys.value = keys
}

const removePermission = (permission: Permission) => {
  selectedPermissionKeys.value = selectedPermissionKeys.value.filter(
    key => key !== String(permission.id)
  )
}

const handleSavePermissions = async () => {
  if (!currentRole.value) return
  
  try {
    const currentKeys = currentRole.value.permissions.map(p => String(p.id))
    const newKeys = selectedPermissionKeys.value
    
    const toAdd = newKeys.filter(k => !currentKeys.includes(k))
    const toRemove = currentKeys.filter(k => !newKeys.includes(k))
    
    for (const key of toAdd) {
      await adminApi.assignRolePermission(currentRole.value.id, Number(key))
    }
    
    for (const key of toRemove) {
      await adminApi.removeRolePermission(currentRole.value.id, Number(key))
    }
    
    message.success('权限更新成功')
    permissionsModalVisible.value = false
    loadData()
  } catch (error) {
    console.error('Update permissions failed:', error)
    message.error('更新权限失败')
  }
}

const handleDelete = async (role: Role) => {
  try {
    await adminApi.deleteRole(role.id)
    message.success('角色删除成功')
    loadData()
  } catch (error) {
    console.error('Delete role failed:', error)
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
.roles-view {
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

h3 {
  margin-bottom: 16px;
}
</style>
