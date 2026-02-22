import { apiClient } from './client'
import type {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  Application,
  CreateApplicationRequest,
  UpdateApplicationRequest,
  Token,
  RevokeTokenRequest,
  RevokeUserTokensRequest,
  Statistics,
  SystemInfo,
  Role,
  CreateRoleRequest,
  Permission,
  CreatePermissionRequest
} from './types'

export const adminApi = {
  async getUsers() {
    const response = await apiClient.get<User[]>('/admin/users')
    return response.data
  },

  async getUser(id: number) {
    const response = await apiClient.get<User>(`/admin/users/${id}`)
    return response.data
  },

  async createUser(data: CreateUserRequest) {
    const response = await apiClient.post<User>('/admin/users', data)
    return response.data
  },

  async updateUser(id: number, data: UpdateUserRequest) {
    const response = await apiClient.post<User>(`/admin/users/${id}/update`, data)
    return response.data
  },

  async deleteUser(id: number) {
    const response = await apiClient.post<void>(`/admin/users/${id}/delete`)
    return response.data
  },

  async banUser(id: number) {
    const response = await apiClient.post<void>(`/admin/users/${id}/ban`)
    return response.data
  },

  async unbanUser(id: number) {
    const response = await apiClient.post<void>(`/admin/users/${id}/unban`)
    return response.data
  },

  async getApplications() {
    const response = await apiClient.get<Application[]>('/admin/applications')
    return response.data
  },

  async getApplication(owner: string, name: string) {
    const response = await apiClient.get<Application>(`/admin/applications/${owner}/${name}`)
    return response.data
  },

  async createApplication(data: CreateApplicationRequest) {
    const response = await apiClient.post<Application>('/admin/applications', data)
    return response.data
  },

  async updateApplication(owner: string, name: string, data: UpdateApplicationRequest) {
    const response = await apiClient.post<Application>(`/admin/applications/${owner}/${name}/update`, data)
    return response.data
  },

  async deleteApplication(owner: string, name: string) {
    const response = await apiClient.post<void>(`/admin/applications/${owner}/${name}/delete`)
    return response.data
  },

  async getTokens() {
    const response = await apiClient.get<Token[]>('/admin/tokens')
    return response.data
  },

  async revokeToken(data: RevokeTokenRequest) {
    const response = await apiClient.post<void>(`/admin/tokens/${data.owner}/${data.name}/revoke`)
    return response.data
  },

  async revokeUserTokens(data: RevokeUserTokensRequest) {
    const response = await apiClient.post<void>(`/admin/tokens/user/${data.owner}/${data.username}/revoke`)
    return response.data
  },

  async getStatistics() {
    const response = await apiClient.get<Statistics>('/admin/stats')
    return response.data
  },

  async getSystemInfo() {
    const response = await apiClient.get<SystemInfo>('/admin/system')
    return response.data
  },

  async getRoles() {
    const response = await apiClient.get<Role[]>('/roles')
    return response.data
  },

  async createRole(data: CreateRoleRequest) {
    const response = await apiClient.post<Role>('/roles', data)
    return response.data
  },

  async deleteRole(owner: string, name: string) {
    const response = await apiClient.post<void>(`/roles/${owner}/${name}/delete`)
    return response.data
  },

  async getRolePermissions(owner: string, name: string) {
    const response = await apiClient.get<Permission[]>(`/roles/${owner}/${name}/permissions`)
    return response.data
  },

  async assignRolePermission(roleOwner: string, roleName: string, permOwner: string, permName: string) {
    const response = await apiClient.post<void>(`/roles/${roleOwner}/${roleName}/permissions`, {
      permOwner,
      permName
    })
    return response.data
  },

  async removeRolePermission(roleOwner: string, roleName: string, permOwner: string, permName: string) {
    const response = await apiClient.post<void>(`/roles/${roleOwner}/${roleName}/permissions/remove`, {
      permOwner,
      permName
    })
    return response.data
  },

  async getPermissions() {
    const response = await apiClient.get<Permission[]>('/permissions')
    return response.data
  },

  async createPermission(data: CreatePermissionRequest) {
    const response = await apiClient.post<Permission>('/permissions', data)
    return response.data
  },

  async deletePermission(owner: string, name: string) {
    const response = await apiClient.post<void>(`/permissions/${owner}/${name}/delete`)
    return response.data
  },

  async getRealNameInfo(userId: number) {
    const response = await apiClient.get<{ name: string; idcard: string }>(`/admin/realname/${userId}`)
    return response.data
  }
}
