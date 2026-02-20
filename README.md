# OAuth 2.1 Server

一个从Casdoor提取的独立OAuth 2.1认证服务器，支持完整的OAuth2/OIDC协议。

## ✅ 安全状态更新 (2024-02-19)

**重要安全修复已完成！**

本项目已完成OAuth 2.1合规性审查并修复了所有严重安全问题：

✅ **已修复的严重问题**:
- ✅ 密码验证已实现（支持bcrypt、SHA256+salt）
- ✅ 令牌撤销已完整实现（RFC 7009）
- ✅ PKCE强制要求（公共客户端必须使用）
- ✅ 重定向URI验证已加强（精确匹配优先）
- ✅ State参数强制要求（CSRF防护）
- ✅ 刷新令牌重用检测（令牌家族撤销）

**合规性评分**: 从 55/100 提升至 85/100 ⬆️ +30

详细信息请查看:
- [OAuth 2.1 合规性报告](OAUTH_COMPLIANCE_REPORT.md)
- [合规性修复详情](COMPLIANCE_FIXES.md)
- [安全修复计划](SECURITY_FIXES.md)

⚠️ **剩余工作**:
- 用户同意页面（需要UI实现）
- Client Secret哈希存储
- HTTP Basic Auth客户端认证

---

## 功能特性

### OAuth 2.1 核心功能
- ✅ Authorization Code Flow (授权码模式)
- ✅ PKCE Support (RFC 7636) - 公共客户端强制要求
- ✅ Client Credentials Flow (客户端凭证模式)
- ✅ Password Flow (密码模式) - 带密码验证
- ✅ Refresh Token (刷新令牌) - 带重用检测
- ✅ Token Revocation (令牌撤销) - RFC 7009
- ✅ Token Exchange (RFC 8693)
- ✅ Device Authorization Flow (设备授权)
- ✅ Resource Indicators (RFC 8707)

### OIDC 支持
- ✅ OpenID Connect Discovery
- ✅ UserInfo Endpoint
- ✅ JWKS Endpoint
- ✅ ID Token

### 高级功能
- ✅ Dynamic Client Registration (RFC 7591)
- ✅ Token Introspection
- ✅ Multiple IdP Integration (30+ providers)
- ✅ JWT Token Format
- ✅ Custom Scopes
- ✅ Multi-tenancy Support
- ✅ Redis Cache Support
- ✅ Admin Management API
- ✅ Role-Based Access Control (RBAC)
- ✅ Multi-User Support
- ✅ Fine-Grained Permissions

## 快速开始

### 环境要求
- Go 1.23+
- MySQL/PostgreSQL/SQLite
- Redis (可选，用于缓存)

### 平台支持
- ✅ Linux
- ✅ macOS
- ✅ Windows (查看 [Windows 使用指南](WINDOWS.md))

### 安装

#### Linux/macOS

```bash
# 克隆项目
git clone <repository>
cd oauth-server

# 安装依赖
make install

# 或使用 go 命令
go mod download

# 配置数据库
cp conf/app.conf.example conf/app.conf
# 编辑 conf/app.conf 配置数据库连接

# 初始化数据库
make init

# 启动服务
make dev
```

#### Windows

```powershell
# 克隆项目
git clone <repository>
cd oauth-server

# 安装依赖
.\build.ps1 install

# 配置数据库
copy conf\app.conf.example conf\app.conf
# 编辑 conf\app.conf 配置数据库连接

# 初始化数据库
.\build.ps1 init

# 启动服务
.\build.ps1 dev
```

详细的 Windows 使用说明请查看 [WINDOWS.md](WINDOWS.md)
```

### 配置文件

编辑 `conf/app.conf`:

```ini
appname = oauth-server
httpport = 8080
runmode = dev

# 数据库配置
driverName = mysql
dataSourceName = root:password@tcp(localhost:3306)/oauth_server?charset=utf8mb4

# JWT密钥
jwtSecret = your-secret-key-here

# Origin配置
origin = http://localhost:8080
```

## API 端点

### OAuth 2.0 端点

```
# 授权端点
GET  /oauth/authorize

# 令牌端点
POST /api/oauth/token

# 令牌撤销
POST /api/oauth/revoke

# 令牌内省
POST /api/oauth/introspect

# 设备授权
POST /api/oauth/device/authorize
POST /api/oauth/device/token
```

### OIDC 端点

```
# Discovery
GET /.well-known/openid-configuration

# JWKS
GET /.well-known/jwks

# UserInfo
GET /api/userinfo

# 动态客户端注册
POST /api/oauth/register
```

## 使用示例

### 1. 授权码模式

```bash
# 步骤1: 获取授权码
curl "http://localhost:8080/oauth/authorize?\
client_id=YOUR_CLIENT_ID&\
response_type=code&\
redirect_uri=http://localhost:3000/callback&\
scope=openid profile email&\
state=random_state"

# 步骤2: 用授权码换取令牌
curl -X POST http://localhost:8080/api/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "redirect_uri=http://localhost:3000/callback"
```

### 2. 客户端凭证模式

```bash
curl -X POST http://localhost:8080/api/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=read write"
```

### 3. 刷新令牌

```bash
curl -X POST http://localhost:8080/api/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

### 4. 动态客户端注册

```bash
curl -X POST http://localhost:8080/api/oauth/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "My Application",
    "redirect_uris": ["http://localhost:3000/callback"],
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code"],
    "scope": "openid profile email"
  }'
```

## 项目结构

```
oauth-server/
├── main.go                 # 入口文件
├── conf/
│   └── app.conf           # 配置文件
├── controllers/           # 控制器层
│   ├── auth.go           # 认证控制器
│   ├── token.go          # 令牌控制器
│   └── oidc.go           # OIDC控制器
├── models/               # 数据模型
│   ├── user.go
│   ├── application.go
│   ├── token.go
│   └── organization.go
├── services/             # 业务逻辑
│   ├── oauth.go         # OAuth服务
│   ├── jwt.go           # JWT服务
│   └── idp/             # IdP集成
├── middleware/           # 中间件
│   └── auth.go
├── util/                # 工具函数
└── routers/             # 路由配置
```

## 数据库表结构

主要表：
- `users` - 用户表
- `applications` - 应用（客户端）表
- `tokens` - 令牌表
- `organizations` - 组织表
- `providers` - IdP提供商表

## 安全特性

- PKCE支持防止授权码拦截
- 授权码5分钟过期
- 防重放攻击（授权码只能使用一次）
- 客户端密钥验证
- Scope验证
- Token哈希存储

## 第三方IdP集成

支持30+种OAuth提供商：
- GitHub, Google, Facebook
- WeChat, DingTalk, Weibo
- Azure AD, Okta, Auth0
- 自定义OAuth提供商

## 开发

```bash
# 运行测试
go test ./...

# 构建
go build -o oauth-server main.go

# 运行
./oauth-server
```

## License

Apache-2.0

## 致谢

本项目基于 [Casdoor](https://github.com/casdoor/casdoor) 的OAuth实现提取而来。
