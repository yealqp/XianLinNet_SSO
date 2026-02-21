export interface User {
  id: number
  owner: string
  username: string
  email: string
  qq?: string
  avatar?: string
  isRealName?: boolean
  isAdmin?: boolean
  isForbidden?: boolean
  isDeleted?: boolean
  createdTime: string
  updatedTime: string
  type?: string
  countryCode?: string
}

export interface CreateUserRequest {
  username: string
  password: string
  email: string
  qq?: string
  avatar?: string
}

export interface UpdateUserRequest {
  username?: string
  email?: string
  qq?: string
  avatar?: string
}

export interface Application {
  owner: string
  name: string
  createdTime: string
  displayName?: string
  logo?: string
  homepageUrl?: string
  description?: string
  organization?: string
  cert?: string
  enablePassword?: boolean
  enableSignUp?: boolean
  enableCodeSignin?: boolean
  grantTypes: string[]
  tags?: string[]
  clientId: string
  clientSecret?: string
  redirectUris: string[]
  tokenFormat?: string
  expireInHours?: number
  refreshExpireInHours?: number
  scopes: string[]
}

export interface CreateApplicationRequest {
  name: string
  redirectUris?: string[]
  grantTypes?: string[]
  scopes?: string[]
  displayName?: string
  logo?: string
  organization?: string
}

export interface UpdateApplicationRequest {
  name?: string
  redirectUris?: string[]
  grantTypes?: string[]
  scopes?: string[]
  displayName?: string
  logo?: string
  organization?: string
}

export interface Token {
  owner: string
  name: string
  createdTime: string
  application: string
  organization: string
  user: string
  code: string
  accessToken: string
  refreshToken: string
  expiresIn: number
  scope: string
  tokenType: string
  codeIsUsed: boolean
  codeExpireIn: number
  refreshTokenUsed: boolean
  tokenFamily: string
}

export interface RevokeTokenRequest {
  owner: string
  name: string
}

export interface RevokeUserTokensRequest {
  owner: string
  username: string
}

export interface Permission {
  id?: string
  owner: string
  name: string
  createdTime: string
  displayName: string
  description: string
  resource: string
  action: string
  effect: string
  isEnabled: boolean
}

export interface CreatePermissionRequest {
  owner?: string
  name: string
  displayName?: string
  description?: string
  resource: string
  action: string
  effect?: string
  isEnabled?: boolean
}

export interface Role {
  id?: string
  owner: string
  name: string
  createdTime: string
  updatedTime?: string
  displayName: string
  description: string
  isEnabled: boolean
  type: string
  organization: string
  permissions?: Permission[]
}

export interface CreateRoleRequest {
  owner?: string
  name: string
  displayName?: string
  description?: string
  isEnabled?: boolean
  type?: string
  organization?: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token?: string
  expires_in: number
  token_type: string
  user: {
    id: number
    email: string
    username: string
    isAdmin: boolean
    isRealName?: boolean
    qq?: string
    avatar?: string
  }
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
}

export interface UserInfoResponse {
  sub?: string
  id?: number
  username?: string
  name?: string
  preferred_username?: string
  given_name?: string
  email?: string
  picture?: string
  avatar?: string
  qq?: string
  is_real_name?: boolean
  roles?: string[]
  permissions?: string[]
}

export interface Statistics {
  userCount: number
  applicationCount: number
  tokenCount: number
  activeTokenCount: number
}

export interface SystemInfo {
  version: string
  uptime: number
  redisConnected: boolean
}

export interface ApiResponse<T> {
  status: string
  msg?: string
  data?: T
  data2?: T
}

export interface ApiError {
  status: string
  msg: string
}
