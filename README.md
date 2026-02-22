# OAuth Server

[English](#english) | [中文](#中文)

---

## English

A full-featured OAuth 2.0 / OpenID Connect server built with Go (Fiber v2) and Vue 3, supporting user authentication, authorization, role-based access control (RBAC), and real-name verification.

### Features

- 🔐 OAuth 2.0 & OpenID Connect support
- 👥 User management with role-based access control (RBAC)
- 🎫 Token management and validation
- 📧 Email verification and password reset
- 🔑 JWT-based authentication
- 🎨 Modern Vue 3 + Ant Design Vue frontend
- 💾 PostgreSQL database support
- 🚀 Redis caching support (optional)
- 🔒 RSA encryption for sensitive data
- 🌐 CORS support for cross-origin requests
- 📱 Real-name verification API integration (optional)
- ⚡ High-performance Fiber v2 web framework
- 🛡️ Request timeout and connection pool management

### Tech Stack

**Backend:**
- Go 1.23+
- Fiber v2 web framework (Express-inspired)
- XORM for database operations
- JWT for token management
- Redis for caching (optional)
- PostgreSQL connection pooling

**Frontend:**
- Vue 3 with TypeScript
- Ant Design Vue 4
- Vite build tool
- Pinia for state management
- Vue Router for routing

### Prerequisites

- Go 1.23 or higher
- Node.js 18+ and pnpm
- PostgreSQL 12 or higher
- Redis (optional, for caching)

### Quick Start

#### 1. Clone the repository

```bash
git clone <repository-url>
cd oauth-server
```

#### 2. Backend Setup

**For Development:**

```bash
# Install Go dependencies
go mod download

# Setup PostgreSQL database
# See docs/postgresql-setup.md for detailed instructions
createdb oauth_server

# Copy development configuration
copy .env.development .env

# Edit .env and configure:
# - Set APP_ENV=development (enables CORS middleware)
# - PostgreSQL connection string (REQUIRED)
# - JWT secret (minimum 32 characters)
# - Admin credentials (REQUIRED for first run)
# - SMTP settings (optional for development)
# - Redis settings (optional)

# Initialize database and create admin user
go run main.go init

# Start the backend server
go run main.go
```

**For Production:**

```bash
# Copy production configuration
copy .env.example .env

# Edit .env and configure:
# - Set APP_ENV=production (CORS handled by nginx)
# - All required settings (see Configuration section)

# Initialize database
go run main.go init

# Build and run
go build -o oauth-server
./oauth-server
```

The backend server will start on `http://localhost:8080` (or your configured port).

**Important:** 
- Development mode (`APP_ENV=development`): CORS middleware is enabled in the backend
- Production mode (`APP_ENV=production`): CORS should be handled by nginx to avoid duplicate headers
go run main.go init

# Start the backend server
go run main.go
```

The backend server will start on `http://localhost:8080` (or your configured port).

#### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
pnpm install

# Start development server
pnpm dev
```

The frontend will start on `http://localhost:5173`.

For production build:

```bash
pnpm build
```

### Configuration

Edit `.env` to configure:

**Required Settings:**
- Database configuration (PostgreSQL):
  - `DB_HOST` - PostgreSQL server address (default: localhost)
  - `DB_PORT` - PostgreSQL port (default: 5432)
  - `DB_USER` - Database username
  - `DB_PASSWORD` - Database password
  - `DB_NAME` - Database name
  - `DB_SSLMODE` - SSL mode (disable/require/verify-ca/verify-full)
- Database connection pool (recommended):
  - `DB_MAX_OPEN_CONNS` - Maximum open connections (default: 50)
  - `DB_MAX_IDLE_CONNS` - Maximum idle connections (default: 10)
  - `DB_CONN_MAX_LIFETIME` - Connection max lifetime (default: 1h)
  - `DB_CONN_MAX_IDLE_TIME` - Connection max idle time (default: 10m)
  - `DB_QUERY_TIMEOUT` - Query timeout (default: 5s)
- `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_USERNAME` - Initial admin user credentials
- `JWT_SECRET` - JWT signing key (minimum 32 characters)

**Database Configuration Example:**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=oauth_server
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=10m
DB_QUERY_TIMEOUT=5s
```

See `docs/postgresql-setup.md` for detailed setup instructions

**Optional Settings:**
- Redis cache configuration (with timeout settings)
- SMTP email settings
- CORS and origin settings
- Real-name verification API
- Server timeouts (READ_TIMEOUT, WRITE_TIMEOUT)

### Project Structure

```
.
├── handlers/         # HTTP request handlers (Fiber)
├── models/           # Database models and operations
├── services/         # Business logic layer
├── routers/          # Route definitions (Fiber)
├── middlewares/      # Fiber middlewares (CORS, Auth, Logger, etc.)
├── types/            # Type definitions and response structures
├── config/           # Configuration management
├── keys/             # RSA key pairs for encryption
├── docs/             # Documentation
├── frontend/         # Vue 3 frontend application
│   ├── src/
│   │   ├── api/      # API client
│   │   ├── components/  # Vue components
│   │   ├── views/    # Page views
│   │   ├── router/   # Route configuration
│   │   └── stores/   # Pinia stores
│   └── dist/         # Production build output
└── main.go           # Application entry point
```

### API Endpoints

**Authentication:**
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `POST /api/auth/send-code` - Send verification code
- `POST /api/auth/reset-password` - Reset password
- `GET /api/auth/application-info` - Get application info

**OAuth 2.0 / OIDC:**
- `GET /oauth/authorize` - OAuth authorization endpoint
- `POST /oauth/authorize` - OAuth authorization (POST)
- `POST /api/oauth/token` - OAuth token endpoint
- `POST /api/oauth/introspect` - Token introspection
- `POST /api/oauth/revoke` - Token revocation
- `GET /api/userinfo` - OIDC UserInfo endpoint
- `GET /.well-known/openid-configuration` - OIDC Discovery
- `GET /.well-known/jwks` - JSON Web Key Set

**User Management (Authenticated):**
- `POST /api/auth/update-profile` - Update user profile
- `POST /api/realname/submit` - Submit real-name verification
- `GET /api/realname/verify` - Get real-name info

**Admin Endpoints (Admin Only):**
- `GET /api/admin/users` - List users
- `POST /api/admin/users` - Create user
- `GET /api/admin/applications` - List applications
- `GET /api/admin/tokens` - List tokens
- `GET /api/admin/stats` - System statistics
- `GET /api/admin/system` - System information
- `POST /api/admin/cache/clear` - Clear cache

**Health Check:**
- `GET /health` - Server health status

### Default Admin Account

After running `go run main.go init`, an admin account will be created using the credentials from `.env`:

- Email: As configured in `ADMIN_EMAIL`
- Password: As configured in `ADMIN_PASSWORD`
- Username: As configured in `ADMIN_USERNAME`

**Important:** Change these default credentials immediately after first login!

### Performance & Reliability

**Connection Pool Management:**
- Database connection pooling with configurable limits
- Redis connection pooling with timeout controls
- Automatic connection lifecycle management

**Timeout Protection:**
- Request read/write timeouts (configurable)
- Database query timeout (default: 5s)
- Redis operation timeout (default: 3s)
- Graceful shutdown support

**Monitoring:**
- Slow query logging (>100ms)
- Request/response logging
- Error tracking and recovery
- Health check endpoint

See `docs/backend-hang-fix.md` for detailed performance optimization guide.

### Development

```bash
# Run backend with hot reload (requires air or similar tool)
go run main.go

# Run frontend development server
cd frontend && pnpm dev

# Build frontend for production
cd frontend && pnpm build
```

### License

Apache License 2.0

---

## 中文

一个功能完整的 OAuth 2.0 / OpenID Connect 服务器，使用 Go (Fiber v2) 和 Vue 3 构建，支持用户认证、授权、基于角色的访问控制 (RBAC) 和实名验证。

### 功能特性

- 🔐 支持 OAuth 2.0 和 OpenID Connect
- 👥 用户管理与基于角色的访问控制 (RBAC)
- 🎫 令牌管理和验证
- 📧 邮箱验证和密码重置
- 🔑 基于 JWT 的身份认证
- 🎨 现代化的 Vue 3 + Ant Design Vue 前端
- 💾 PostgreSQL 数据库支持
- 🚀 Redis 缓存支持（可选）
- 🔒 敏感数据 RSA 加密
- 🌐 支持跨域请求 (CORS)
- 📱 实名验证 API 集成（可选）
- ⚡ 高性能 Fiber v2 Web 框架
- 🛡️ 请求超时和连接池管理

### 技术栈

**后端：**
- Go 1.23+
- Fiber v2 Web 框架（Express 风格）
- XORM 数据库操作
- JWT 令牌管理
- Redis 缓存（可选）
- PostgreSQL 连接池管理

**前端：**
- Vue 3 + TypeScript
- Ant Design Vue 4
- Vite 构建工具
- Pinia 状态管理
- Vue Router 路由管理

### 环境要求

- Go 1.23 或更高版本
- Node.js 18+ 和 pnpm
- PostgreSQL 12 或更高版本
- Redis（可选，用于缓存）

### 快速开始

#### 1. 克隆仓库

```bash
git clone <repository-url>
cd oauth-server
```

#### 2. 后端设置

**开发环境：**

```bash
# 安装 Go 依赖
go mod download

# 设置 PostgreSQL 数据库
# 详细说明请参见 docs/postgresql-setup.md
createdb oauth_server

# 复制开发环境配置
copy .env.development .env

# 编辑 .env 并配置：
# - 设置 APP_ENV=development（启用 CORS 中间件）
# - PostgreSQL 连接字符串（必需）
# - JWT 密钥（至少 32 个字符）
# - 管理员凭据（首次运行必需）
# - SMTP 设置（开发环境可选）
# - Redis 设置（可选）

# 初始化数据库并创建管理员用户
go run main.go init

# 启动后端服务器
go run main.go
```

**生产环境：**

```bash
# 复制生产环境配置
copy .env.example .env

# 编辑 .env 并配置：
# - 设置 APP_ENV=production（CORS 由 nginx 处理）
# - 所有必需设置（参见配置说明部分）

# 初始化数据库
go run main.go init

# 编译并运行
go build -o oauth-server
./oauth-server
```

后端服务器将在 `http://localhost:8080`（或您配置的端口）启动。

**重要提示：**
- 开发模式（`APP_ENV=development`）：后端启用 CORS 中间件
- 生产模式（`APP_ENV=production`）：CORS 由 nginx 处理，避免重复添加 CORS 头

#### 3. 前端设置

```bash
cd frontend

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
```

前端将在 `http://localhost:5173` 启动。

生产环境构建：

```bash
pnpm build
```

### 配置说明

编辑 `.env` 进行配置：

**必需设置：**
- 数据库配置（PostgreSQL）：
  - `DB_HOST` - PostgreSQL 服务器地址（默认：localhost）
  - `DB_PORT` - PostgreSQL 端口（默认：5432）
  - `DB_USER` - 数据库用户名
  - `DB_PASSWORD` - 数据库密码
  - `DB_NAME` - 数据库名称
  - `DB_SSLMODE` - SSL 模式（disable/require/verify-ca/verify-full）
- 数据库连接池（推荐配置）：
  - `DB_MAX_OPEN_CONNS` - 最大连接数（默认：50）
  - `DB_MAX_IDLE_CONNS` - 最大空闲连接数（默认：10）
  - `DB_CONN_MAX_LIFETIME` - 连接最大生命周期（默认：1h）
  - `DB_CONN_MAX_IDLE_TIME` - 连接最大空闲时间（默认：10m）
  - `DB_QUERY_TIMEOUT` - 查询超时时间（默认：5s）
- `ADMIN_EMAIL`、`ADMIN_PASSWORD`、`ADMIN_USERNAME` - 初始管理员用户凭据
- `JWT_SECRET` - JWT 签名密钥（至少 32 个字符）

**数据库配置示例：**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=oauth_server
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=10m
DB_QUERY_TIMEOUT=5s
```

详细设置说明请参见 `docs/postgresql-setup.md`

**可选设置：**
- Redis 缓存配置（包含超时设置）
- SMTP 邮件设置
- CORS 和源站设置
- 实名验证 API
- 服务器超时设置（READ_TIMEOUT、WRITE_TIMEOUT）

### 项目结构

```
.
├── handlers/         # HTTP 请求处理器（Fiber）
├── models/           # 数据库模型和操作
├── services/         # 业务逻辑层
├── routers/          # 路由定义（Fiber）
├── middlewares/      # Fiber 中间件（CORS、认证、日志等）
├── types/            # 类型定义和响应结构
├── config/           # 配置管理
├── keys/             # RSA 密钥对（用于加密）
├── docs/             # 文档
├── frontend/         # Vue 3 前端应用
│   ├── src/
│   │   ├── api/      # API 客户端
│   │   ├── components/  # Vue 组件
│   │   ├── views/    # 页面视图
│   │   ├── router/   # 路由配置
│   │   └── stores/   # Pinia 状态存储
│   └── dist/         # 生产构建输出
└── main.go           # 应用程序入口
```

### API 端点

**认证相关：**
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/send-code` - 发送验证码
- `POST /api/auth/reset-password` - 重置密码
- `GET /api/auth/application-info` - 获取应用信息

**OAuth 2.0 / OIDC：**
- `GET /oauth/authorize` - OAuth 授权端点
- `POST /oauth/authorize` - OAuth 授权（POST）
- `POST /api/oauth/token` - OAuth 令牌端点
- `POST /api/oauth/introspect` - 令牌内省
- `POST /api/oauth/revoke` - 令牌撤销
- `GET /api/userinfo` - OIDC 用户信息端点
- `GET /.well-known/openid-configuration` - OIDC 发现
- `GET /.well-known/jwks` - JSON Web 密钥集

**用户管理（需认证）：**
- `POST /api/auth/update-profile` - 更新用户资料
- `POST /api/realname/submit` - 提交实名认证
- `GET /api/realname/verify` - 获取实名信息

**管理员端点（仅管理员）：**
- `GET /api/admin/users` - 用户列表
- `POST /api/admin/users` - 创建用户
- `GET /api/admin/applications` - 应用列表
- `GET /api/admin/tokens` - 令牌列表
- `GET /api/admin/stats` - 系统统计
- `GET /api/admin/system` - 系统信息
- `POST /api/admin/cache/clear` - 清除缓存

**健康检查：**
- `GET /health` - 服务器健康状态

### 默认管理员账户

运行 `go run main.go init` 后，将使用 `.env` 中的凭据创建管理员账户：

- 邮箱：在 `ADMIN_EMAIL` 中配置
- 密码：在 `ADMIN_PASSWORD` 中配置
- 用户名：在 `ADMIN_USERNAME` 中配置

**重要提示：** 首次登录后请立即更改这些默认凭据！

### 性能与可靠性

**连接池管理：**
- 数据库连接池，可配置连接限制
- Redis 连接池，带超时控制
- 自动连接生命周期管理

**超时保护：**
- 请求读写超时（可配置）
- 数据库查询超时（默认：5s）
- Redis 操作超时（默认：3s）
- 优雅关闭支持

**监控：**
- 慢查询日志（>100ms）
- 请求/响应日志
- 错误跟踪和恢复
- 健康检查端点

详细的性能优化指南请参见 `docs/backend-hang-fix.md`。

### 开发

```bash
# 运行后端（需要 air 或类似工具实现热重载）
go run main.go

# 运行前端开发服务器
cd frontend && pnpm dev

# 构建生产环境前端
cd frontend && pnpm build
```

### 许可证

Apache License 2.0
